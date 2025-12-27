# Architecture

## Overview

gotio is a pure Go implementation of OpenTimelineIO, organized into three main packages:

```
github.com/mrjoshuak/gotio/
├── opentime/           # Time representation
├── opentimelineio/     # Core OTIO data model
└── algorithms/         # Timeline manipulation algorithms
```

## Package Dependencies

```
algorithms
    ↓
opentimelineio
    ↓
opentime
```

The dependency flow is strictly downward:
- `opentime` has no dependencies on other gotio packages
- `opentimelineio` depends on `opentime`
- `algorithms` depends on both `opentimelineio` and `opentime`

## opentime Package

The `opentime` package provides fundamental types for representing time in media applications.

### Core Types

| Type | Description |
|------|-------------|
| `RationalTime` | A point in time as a rational number (value/rate) |
| `TimeRange` | A range with start time and duration |
| `TimeTransform` | A transformation (offset + scale) |
| `IsDropFrameRate` | Enum for drop frame handling |

### Design Principles

1. **Immutability** - All time types are value types with no mutable methods
2. **Rate Preservation** - Operations preserve rates where possible
3. **Epsilon Comparisons** - Floating-point comparisons use configurable epsilon

```go
// Example: Time types are immutable
t1 := opentime.NewRationalTime(100, 24)
t2 := t1.Add(opentime.NewRationalTime(10, 24))  // Returns new value
// t1 is unchanged
```

## opentimelineio Package

The `opentimelineio` package implements the OTIO data model with Go interfaces and types.

### Type Hierarchy

```
SerializableObject (interface)
├── SerializableObjectWithMetadata (interface)
│   ├── Composable (interface)
│   │   ├── Item (interface)
│   │   │   ├── Composition (interface)
│   │   │   │   ├── *Track
│   │   │   │   └── *Stack
│   │   │   ├── *Clip
│   │   │   └── *Gap
│   │   └── *Transition
│   ├── MediaReference (interface)
│   │   ├── *ExternalReference
│   │   ├── *MissingReference
│   │   ├── *GeneratorReference
│   │   └── *ImageSequenceReference
│   ├── *Marker
│   ├── Effect (interface)
│   │   ├── *BasicEffect
│   │   ├── *LinearTimeWarp
│   │   └── *FreezeFrame
│   ├── *Timeline
│   ├── *SerializableCollection
│   └── *UnknownSchema
```

### Key Interfaces

#### SerializableObject

The base interface for all OTIO types:

```go
type SerializableObject interface {
    SchemaName() string
    SchemaVersion() int
    Clone() SerializableObject
    IsEquivalentTo(other SerializableObject) bool
}
```

#### Composable

Items that can be placed in compositions:

```go
type Composable interface {
    SerializableObjectWithMetadata
    Parent() Composition
    SetParent(parent Composition)
    Duration() (opentime.RationalTime, error)
    Visible() bool
    Overlapping() bool
}
```

#### Composition

Containers that hold composable children:

```go
type Composition interface {
    Item
    Children() []Composable
    AppendChild(child Composable) error
    InsertChild(index int, child Composable) error
    RemoveChild(index int) error
    RangeOfChildAtIndex(index int) (opentime.TimeRange, error)
    ChildAtTime(time opentime.RationalTime, shallow bool) (Composable, error)
}
```

### Canonical Structure

The canonical OTIO structure follows this pattern:

```
Timeline
└── tracks: Stack
    ├── Track (video)
    │   ├── Clip
    │   ├── Transition
    │   ├── Clip
    │   └── Gap
    └── Track (audio)
        ├── Clip
        └── Clip
```

**Key Points:**
- A `Timeline` always has a `tracks` field which is a `Stack`
- The `Stack` contains `Track` objects
- `Track` objects contain `Clip`, `Gap`, `Transition`, or nested compositions
- `Stack` layers children vertically (like Photoshop layers)
- `Track` arranges children sequentially in time

### Schema Registry

Types are registered with a schema registry for JSON deserialization:

```go
// Registration happens in init()
func init() {
    RegisterSchema(ClipSchema, func() SerializableObject {
        return NewClip("", nil, nil, nil, nil, nil, "", nil)
    })
}

// Schema aliases support legacy formats
func init() {
    RegisterSchemaAlias("Sequence", "Track")  // Legacy name
}
```

### JSON Serialization

All types implement `json.Marshaler` and `json.Unmarshaler`:

```go
// JSON structure includes OTIO_SCHEMA field
type clipJSON struct {
    Schema          string              `json:"OTIO_SCHEMA"`
    Name            string              `json:"name"`
    MediaReference  json.RawMessage     `json:"media_reference"`
    SourceRange     *opentime.TimeRange `json:"source_range"`
    // ...
}
```

The schema field enables polymorphic deserialization:

```go
// Reading uses schema to determine type
obj, err := FromJSONBytes(data)
// obj could be *Timeline, *Clip, *Track, etc.
```

## algorithms Package

The `algorithms` package provides functions for manipulating timelines.

### Categories

#### Track Algorithms

```go
// Trim a track to a specific time range
func TrackTrimmedToRange(track *Track, trimRange TimeRange) (*Track, error)

// Expand transitions to include adjacent media
func TrackWithExpandedTransitions(track *Track) (*Track, error)
```

#### Stack Algorithms

```go
// Flatten a stack to a single track
func FlattenStack(stack *Stack) (*Track, error)

// Flatten multiple tracks
func FlattenTracks(tracks []*Track) (*Track, error)

// Find the topmost clip at a given time
func TopClipAtTime(stack *Stack, t RationalTime) *Clip
```

#### Timeline Algorithms

```go
// Trim a timeline to a specific range
func TimelineTrimmedToRange(timeline *Timeline, trimRange TimeRange) (*Timeline, error)

// Get only video/audio tracks
func TimelineVideoTracks(timeline *Timeline) []*Track
func TimelineAudioTracks(timeline *Timeline) []*Track

// Flatten all video tracks
func FlattenTimelineVideoTracks(timeline *Timeline) (*Timeline, error)
```

#### Filtering

```go
// Filter compositions by type or criteria
func FilteredComposition(root SerializableObject, filter FilterFunc, prune []reflect.Type) SerializableObject

// Built-in filters
func TypeFilter(types ...reflect.Type) FilterFunc
func NameFilter(pattern string) FilterFunc
```

## Error Handling

gotio uses typed errors for specific error conditions:

```go
// Schema not found
type SchemaError struct {
    Schema  string
    Message string
}

// JSON parsing error
type JSONError struct {
    Message string
}

// Index out of bounds
type IndexError struct {
    Index int
    Size  int
}

// Missing required reference
type MissingReferenceError struct{}

// Type mismatch
type TypeMismatchError struct {
    Expected string
    Got      string
}
```

## Thread Safety

The schema registry is thread-safe and uses `sync.RWMutex`. However, individual OTIO objects are **not** thread-safe. If you need concurrent access to a timeline, you should:

1. Use external synchronization (mutex)
2. Work with cloned copies in each goroutine
3. Use channels to serialize access

```go
// Safe: Clone before modifying
clone := timeline.Clone().(*Timeline)
// Modify clone in goroutine

// Safe: Use mutex
var mu sync.Mutex
mu.Lock()
timeline.SetName("New Name")
mu.Unlock()
```

## Memory Management

Go's garbage collector handles memory automatically. However, be aware of:

1. **Parent References** - Compositions hold references to children, and children hold references to parents (circular)
2. **Clone Deep Copies** - `Clone()` creates deep copies, which may use significant memory for large timelines

```go
// This is fine - GC handles circular refs
track.AppendChild(clip)  // clip.Parent() == track

// Be mindful of deep clones
for i := 0; i < 1000; i++ {
    copy := hugeTimeline.Clone()  // 1000 copies in memory
    // ...
}
```

## Extending gotio

### Custom Schema Types

To add custom schema types:

```go
// Define your type
type MyCustomType struct {
    opentimelineio.SerializableObjectWithMetadataBase
    customField string
}

// Implement required interfaces
func (m *MyCustomType) SchemaName() string { return "MyCustomType" }
func (m *MyCustomType) SchemaVersion() int { return 1 }
func (m *MyCustomType) Clone() SerializableObject { /* ... */ }
func (m *MyCustomType) IsEquivalentTo(other SerializableObject) bool { /* ... */ }

// Implement JSON marshaling
func (m *MyCustomType) MarshalJSON() ([]byte, error) { /* ... */ }
func (m *MyCustomType) UnmarshalJSON(data []byte) error { /* ... */ }

// Register the schema
func init() {
    opentimelineio.RegisterSchema(
        opentimelineio.Schema{Name: "MyCustomType", Version: 1},
        func() opentimelineio.SerializableObject {
            return &MyCustomType{}
        },
    )
}
```

### Custom Algorithms

Add algorithms by creating functions that operate on OTIO types:

```go
package myalgorithms

import (
    "github.com/mrjoshuak/gotio/opentimelineio"
    "github.com/mrjoshuak/gotio/opentime"
)

// FindClipsWithMarkers returns clips that have markers
func FindClipsWithMarkers(timeline *opentimelineio.Timeline) []*opentimelineio.Clip {
    var result []*opentimelineio.Clip
    for _, clip := range timeline.FindClips(nil, false) {
        if len(clip.Markers()) > 0 {
            result = append(result, clip)
        }
    }
    return result
}
```

## Performance Considerations

1. **JSON Parsing** - Large files benefit from streaming parsers; gotio uses standard `encoding/json`
2. **FindClips/FindChildren** - These traverse the entire tree; cache results if used repeatedly
3. **Cloning** - Deep clones are expensive; modify in place when possible
4. **Time Conversions** - `RescaledTo` creates new values; batch conversions when possible

```go
// Inefficient: Multiple traversals
for i := 0; i < 10; i++ {
    clips := timeline.FindClips(nil, false)
    // process clips[i]
}

// Efficient: Single traversal
clips := timeline.FindClips(nil, false)
for i := 0; i < 10; i++ {
    // process clips[i]
}
```

## Comparison with Python/C++ OTIO

| Aspect | Python/C++ OTIO | gotio |
|--------|-----------------|-------|
| Dependencies | C++ core, pybind11 | Pure Go, no cgo |
| Plugin System | Python plugins, adapters | Not supported (native .otio only) |
| Memory Management | Reference counting | Garbage collection |
| Thread Safety | GIL in Python | Manual synchronization needed |
| API Style | Snake_case | CamelCase (Go idiom) |
| Nil Handling | None/nullptr | nil, error returns |
