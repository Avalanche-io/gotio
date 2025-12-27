# API Reference

Complete API reference for gotio.

## Package: opentime

```go
import "github.com/mrjoshuak/gotio/opentime"
```

### Types

#### RationalTime

A point in time represented as a rational number.

```go
type RationalTime struct {
    // contains filtered or unexported fields
}
```

**Constructors:**

| Function | Description |
|----------|-------------|
| `NewRationalTime(value, rate float64) RationalTime` | Create from value and rate |
| `NewRationalTimeFromSeconds(seconds, rate float64) RationalTime` | Create from seconds |
| `FromTimecode(tc string, rate float64) (RationalTime, error)` | Parse timecode string |
| `FromTimeString(s string, rate float64) (RationalTime, error)` | Parse time string |

**Methods:**

| Method | Description |
|--------|-------------|
| `Value() float64` | Get the value component |
| `Rate() float64` | Get the rate component |
| `ToSeconds() float64` | Convert to seconds |
| `ToTimecode(rate float64, df IsDropFrameRate) (string, error)` | Convert to timecode |
| `ToTimeString() string` | Convert to string representation |
| `RescaledTo(newRate float64) RationalTime` | Convert to new rate |
| `Add(other RationalTime) RationalTime` | Add two times |
| `Sub(other RationalTime) RationalTime` | Subtract times |
| `Neg() RationalTime` | Negate time |
| `Abs() RationalTime` | Absolute value |
| `Equal(other RationalTime) bool` | Check equality |
| `Compare(other RationalTime) int` | Compare (-1, 0, 1) |
| `AlmostEqual(other RationalTime, delta float64) bool` | Approximate equality |

---

#### TimeRange

A range of time with start and duration.

```go
type TimeRange struct {
    // contains filtered or unexported fields
}
```

**Constructors:**

| Function | Description |
|----------|-------------|
| `NewTimeRange(start, duration RationalTime) TimeRange` | Create from start and duration |
| `NewTimeRangeFromStartEndTime(start, end RationalTime) TimeRange` | Create from start/end (inclusive) |
| `NewTimeRangeFromStartEndTimeExclusive(start, end RationalTime) TimeRange` | Create from start/end (exclusive) |

**Methods:**

| Method | Description |
|--------|-------------|
| `StartTime() RationalTime` | Get start time |
| `Duration() RationalTime` | Get duration |
| `EndTimeExclusive() RationalTime` | Get exclusive end time |
| `EndTimeInclusive() RationalTime` | Get inclusive end time |
| `DurationExtendedBy(other RationalTime) TimeRange` | Extend duration |
| `Extended(other TimeRange) TimeRange` | Get union of ranges |
| `Clamped(other TimeRange) TimeRange` | Get intersection |
| `Contains(time RationalTime) bool` | Check if time is in range |
| `ContainsRange(other TimeRange) bool` | Check if range contains another |
| `Overlaps(other TimeRange) bool` | Check if ranges overlap |
| `OverlapsRange(other TimeRange) bool` | Alias for Overlaps |
| `Intersects(other TimeRange, epsilon float64) bool` | Check intersection with epsilon |
| `Before(time RationalTime, epsilon float64) bool` | Check if range is before time |
| `After(time RationalTime, epsilon float64) bool` | Check if range is after time |
| `ClampedTime(time RationalTime) RationalTime` | Clamp time to range |
| `Equal(other TimeRange) bool` | Check equality |

---

#### TimeTransform

A linear transformation for times.

```go
type TimeTransform struct {
    // contains filtered or unexported fields
}
```

**Constructors:**

| Function | Description |
|----------|-------------|
| `NewTimeTransform(offset RationalTime, scale, rate float64) TimeTransform` | Create transform |
| `NewTimeTransformWithRate(rate float64) TimeTransform` | Create identity transform |

**Methods:**

| Method | Description |
|--------|-------------|
| `Offset() RationalTime` | Get offset |
| `Scale() float64` | Get scale factor |
| `Rate() float64` | Get rate |
| `AppliedToTime(time RationalTime) RationalTime` | Transform a time |
| `AppliedToTimeRange(range TimeRange) TimeRange` | Transform a range |
| `AppliedTo(other TimeTransform) TimeTransform` | Compose transforms |
| `Equal(other TimeTransform) bool` | Check equality |

---

#### Constants

```go
const DefaultEpsilon = 0.00001  // Default epsilon for comparisons

type IsDropFrameRate int
const (
    ForceNo       IsDropFrameRate = 0  // Non-drop-frame
    ForceYes      IsDropFrameRate = 1  // Drop-frame
    InferFromRate IsDropFrameRate = 2  // Auto-detect from rate
)
```

---

## Package: opentimelineio

```go
import "github.com/mrjoshuak/gotio/opentimelineio"
```

### Core Types

#### SerializableObject (interface)

Base interface for all OTIO types.

```go
type SerializableObject interface {
    SchemaName() string
    SchemaVersion() int
    Clone() SerializableObject
    IsEquivalentTo(other SerializableObject) bool
}
```

---

#### Timeline

The root container for editorial content.

```go
type Timeline struct {
    // contains filtered or unexported fields
}
```

**Constructor:**

```go
func NewTimeline(name string, globalStartTime *opentime.RationalTime, metadata AnyDictionary) *Timeline
```

**Methods:**

| Method | Description |
|--------|-------------|
| `Name() string` | Get timeline name |
| `SetName(name string)` | Set timeline name |
| `GlobalStartTime() *opentime.RationalTime` | Get global start time |
| `SetGlobalStartTime(time *opentime.RationalTime)` | Set global start time |
| `Tracks() *Stack` | Get tracks stack |
| `SetTracks(stack *Stack)` | Set tracks stack |
| `Metadata() AnyDictionary` | Get metadata |
| `VideoTracks() []*Track` | Get video tracks |
| `AudioTracks() []*Track` | Get audio tracks |
| `FindClips(search *opentime.TimeRange, shallow bool) []*Clip` | Find clips |
| `FindChildren(search *opentime.TimeRange, descend bool) []Composable` | Find children |
| `Duration() (opentime.RationalTime, error)` | Get duration |
| `RangeOfChild(child Composable) (opentime.TimeRange, error)` | Get child's range |
| `Clone() SerializableObject` | Deep copy |

---

#### Track

Sequential arrangement of items.

```go
type Track struct {
    // contains filtered or unexported fields
}
```

**Constructor:**

```go
func NewTrack(name string, sourceRange *opentime.TimeRange, kind string, metadata AnyDictionary, color *Color) *Track
```

**Constants:**

```go
const (
    TrackKindVideo = "Video"
    TrackKindAudio = "Audio"
)
```

**Methods:**

| Method | Description |
|--------|-------------|
| `Name() string` | Get name |
| `SetName(name string)` | Set name |
| `Kind() string` | Get kind (Video/Audio) |
| `SetKind(kind string)` | Set kind |
| `SourceRange() *opentime.TimeRange` | Get source range |
| `SetSourceRange(r *opentime.TimeRange)` | Set source range |
| `Children() []Composable` | Get children |
| `AppendChild(child Composable) error` | Add child at end |
| `InsertChild(index int, child Composable) error` | Insert child |
| `RemoveChild(index int) error` | Remove child |
| `SetChild(index int, child Composable) error` | Replace child |
| `IndexOfChild(child Composable) (int, error)` | Find child index |
| `RangeOfChildAtIndex(index int) (opentime.TimeRange, error)` | Get child range |
| `TrimmedRangeOfChildAtIndex(index int) (opentime.TimeRange, error)` | Get trimmed range |
| `AvailableRange() (opentime.TimeRange, error)` | Get total range |
| `Duration() (opentime.RationalTime, error)` | Get duration |
| `ChildAtTime(time opentime.RationalTime, shallow bool) (Composable, error)` | Find child at time |
| `ChildrenInRange(range opentime.TimeRange) ([]Composable, error)` | Find children in range |
| `NeighborsOf(item Composable, policy NeighborGapPolicy) (Composable, Composable, error)` | Get neighbors |
| `RangeOfAllChildren() (map[Composable]opentime.TimeRange, error)` | Map of all ranges |

---

#### Stack

Layered arrangement of items.

```go
type Stack struct {
    // contains filtered or unexported fields
}
```

**Constructor:**

```go
func NewStack(name string, sourceRange *opentime.TimeRange, markers []*Marker, effects []Effect, metadata AnyDictionary, color *Color) *Stack
```

**Methods:**

Same as Track, plus:

| Method | Description |
|--------|-------------|
| `AvailableRange() (opentime.TimeRange, error)` | Max duration of children |

---

#### Clip

A segment of media.

```go
type Clip struct {
    // contains filtered or unexported fields
}
```

**Constructor:**

```go
func NewClip(
    name string,
    mediaReference MediaReference,
    sourceRange *opentime.TimeRange,
    metadata AnyDictionary,
    effects []Effect,
    markers []*Marker,
    activeMediaReferenceKey string,
    color *Color,
) *Clip
```

**Methods:**

| Method | Description |
|--------|-------------|
| `Name() string` | Get name |
| `SetName(name string)` | Set name |
| `MediaReference() MediaReference` | Get media reference |
| `SetMediaReference(ref MediaReference)` | Set media reference |
| `SourceRange() *opentime.TimeRange` | Get source range |
| `SetSourceRange(r *opentime.TimeRange)` | Set source range |
| `ActiveMediaReferenceKey() string` | Get active reference key |
| `SetActiveMediaReferenceKey(key string)` | Set active reference key |
| `AvailableRange() (opentime.TimeRange, error)` | Get available range |
| `TrimmedRange() (opentime.TimeRange, error)` | Get effective range |
| `VisibleRange() (opentime.TimeRange, error)` | Get visible range |
| `Duration() (opentime.RationalTime, error)` | Get duration |
| `RangeInParent() (opentime.TimeRange, error)` | Get range in parent |
| `TrimmedRangeInParent() (*opentime.TimeRange, error)` | Get trimmed range in parent |
| `TransformedTime(t RationalTime, toItem Item) (RationalTime, error)` | Transform time to another item's coordinate space |
| `TransformedTimeRange(tr TimeRange, toItem Item) (TimeRange, error)` | Transform time range to another item's coordinate space |
| `Effects() []Effect` | Get effects |
| `SetEffects(effects []Effect)` | Set effects |
| `Markers() []*Marker` | Get markers |
| `SetMarkers(markers []*Marker)` | Set markers |
| `AvailableImageBounds() (*Box2d, error)` | Get image bounds |

---

#### Gap

Empty space in a track.

```go
type Gap struct {
    // contains filtered or unexported fields
}
```

**Constructors:**

```go
func NewGap(name string, sourceRange *opentime.TimeRange, metadata AnyDictionary, effects []Effect, markers []*Marker, color *Color) *Gap
func NewGapWithDuration(duration opentime.RationalTime) *Gap
```

**Methods:**

| Method | Description |
|--------|-------------|
| `Duration() (opentime.RationalTime, error)` | Get duration |
| `Visible() bool` | Always returns true |
| `SetSourceRange(r *opentime.TimeRange)` | Set source range |

---

#### Transition

Blend between adjacent items.

```go
type Transition struct {
    // contains filtered or unexported fields
}
```

**Constructor:**

```go
func NewTransition(name, transitionType string, inOffset, outOffset opentime.RationalTime, metadata AnyDictionary) *Transition
```

**Constants:**

```go
const (
    TransitionTypeSMPTEDissolve = "SMPTE_Dissolve"
    TransitionTypeCustom        = "Custom_Transition"
)
```

**Methods:**

| Method | Description |
|--------|-------------|
| `TransitionType() string` | Get type |
| `SetTransitionType(t string)` | Set type |
| `InOffset() opentime.RationalTime` | Get in offset |
| `SetInOffset(offset opentime.RationalTime)` | Set in offset |
| `OutOffset() opentime.RationalTime` | Get out offset |
| `SetOutOffset(offset opentime.RationalTime)` | Set out offset |
| `Duration() (opentime.RationalTime, error)` | In + out offsets |
| `Visible() bool` | Always returns false |
| `Overlapping() bool` | Always returns true |

---

### Media References

#### ExternalReference

Reference to external media file.

```go
type ExternalReference struct {
    // contains filtered or unexported fields
}
```

**Constructor:**

```go
func NewExternalReference(name, targetURL string, availableRange *opentime.TimeRange, metadata AnyDictionary) *ExternalReference
```

**Methods:**

| Method | Description |
|--------|-------------|
| `TargetURL() string` | Get URL |
| `SetTargetURL(url string)` | Set URL |
| `AvailableRange() *opentime.TimeRange` | Get range |
| `SetAvailableRange(r *opentime.TimeRange)` | Set range |
| `IsMissingReference() bool` | Always false |
| `AvailableImageBounds() *Box2d` | Get image bounds |
| `SetAvailableImageBounds(b *Box2d)` | Set image bounds |

---

#### MissingReference

Placeholder for missing media.

```go
type MissingReference struct {
    // contains filtered or unexported fields
}
```

**Constructor:**

```go
func NewMissingReference(name string, availableRange *opentime.TimeRange, metadata AnyDictionary) *MissingReference
```

**Methods:**

| Method | Description |
|--------|-------------|
| `IsMissingReference() bool` | Always true |
| `AvailableRange() *opentime.TimeRange` | Get range |

---

#### GeneratorReference

Procedurally generated media.

```go
type GeneratorReference struct {
    // contains filtered or unexported fields
}
```

**Constructor:**

```go
func NewGeneratorReference(name, generatorKind string, parameters AnyDictionary, availableRange *opentime.TimeRange, metadata AnyDictionary) *GeneratorReference
```

**Methods:**

| Method | Description |
|--------|-------------|
| `GeneratorKind() string` | Get generator type |
| `SetGeneratorKind(kind string)` | Set generator type |
| `Parameters() AnyDictionary` | Get parameters |
| `SetParameters(p AnyDictionary)` | Set parameters |

---

#### ImageSequenceReference

Numbered image sequence.

```go
type ImageSequenceReference struct {
    // contains filtered or unexported fields
}
```

**Constructor:**

```go
func NewImageSequenceReference(
    name, targetURLBase, namePrefix, nameSuffix string,
    startFrame, frameStep int,
    rate float64,
    frameZeroPadding int,
    availableRange *opentime.TimeRange,
    metadata AnyDictionary,
    missingFramePolicy MissingFramePolicy,
) *ImageSequenceReference
```

**Methods:**

| Method | Description |
|--------|-------------|
| `TargetURLBase() string` | Get base URL |
| `NamePrefix() string` | Get filename prefix |
| `NameSuffix() string` | Get filename suffix |
| `StartFrame() int` | Get start frame number |
| `FrameStep() int` | Get frame step |
| `Rate() float64` | Get frame rate |
| `FrameZeroPadding() int` | Get zero padding |
| `TargetURLForImageNumber(n int) string` | Get URL for frame |
| `MissingFramePolicy() MissingFramePolicy` | Get policy |

---

### Effects

#### Effect (interface)

```go
type Effect interface {
    SerializableObjectWithMetadata
    EffectName() string
    SetEffectName(name string)
}
```

---

#### BasicEffect

Generic effect.

```go
func NewEffect(name, effectName string, metadata AnyDictionary) *BasicEffect
```

---

#### LinearTimeWarp

Speed change effect.

```go
func NewLinearTimeWarp(name, effectName string, timeScalar float64, metadata AnyDictionary) *LinearTimeWarp
```

**Methods:**

| Method | Description |
|--------|-------------|
| `TimeScalar() float64` | Get speed multiplier (2.0 = 2x speed) |
| `SetTimeScalar(s float64)` | Set speed multiplier |

---

#### FreezeFrame

Freeze frame effect (timeScalar = 0).

```go
func NewFreezeFrame(name string, metadata AnyDictionary) *FreezeFrame
```

---

### Markers

#### Marker

Annotation attached to items.

```go
type Marker struct {
    // contains filtered or unexported fields
}
```

**Constructor:**

```go
func NewMarker(name string, markedRange opentime.TimeRange, color MarkerColor, comment string, metadata AnyDictionary) *Marker
```

**Methods:**

| Method | Description |
|--------|-------------|
| `MarkedRange() opentime.TimeRange` | Get position/duration |
| `SetMarkedRange(r opentime.TimeRange)` | Set position |
| `Color() MarkerColor` | Get color |
| `SetColor(c MarkerColor)` | Set color |
| `Comment() string` | Get comment |
| `SetComment(c string)` | Set comment |

**Colors:**

```go
const (
    MarkerColorPink    MarkerColor = "PINK"
    MarkerColorRed     MarkerColor = "RED"
    MarkerColorOrange  MarkerColor = "ORANGE"
    MarkerColorYellow  MarkerColor = "YELLOW"
    MarkerColorGreen   MarkerColor = "GREEN"
    MarkerColorCyan    MarkerColor = "CYAN"
    MarkerColorBlue    MarkerColor = "BLUE"
    MarkerColorPurple  MarkerColor = "PURPLE"
    MarkerColorMagenta MarkerColor = "MAGENTA"
    MarkerColorBlack   MarkerColor = "BLACK"
    MarkerColorWhite   MarkerColor = "WHITE"
)
```

---

### Serialization

#### File I/O

```go
// Read from file
func FromJSONFile(filename string) (SerializableObject, error)

// Read from bytes
func FromJSONBytes(data []byte) (SerializableObject, error)

// Read from string
func FromJSONString(jsonStr string) (SerializableObject, error)

// Write to file
func ToJSONFile(obj SerializableObject, filename, indent string) error

// Write to bytes
func ToJSONBytes(obj SerializableObject) ([]byte, error)

// Write to bytes with indent
func ToJSONBytesIndent(obj SerializableObject, indent string) ([]byte, error)

// Write to string
func ToJSONString(obj SerializableObject) (string, error)
```

---

#### Schema Registry

```go
// Register a schema type
func RegisterSchema(schema Schema, factory SchemaFactory)

// Register an alias
func RegisterSchemaAlias(alias, canonicalName string)

// Check if registered
func IsSchemaRegistered(schemaName string) bool

// Create instance
func CreateSchema(schemaName string) (SerializableObject, error)

// Parse schema string
func ParseSchema(schemaStr string) (name string, version int, err error)
```

---

### Other Types

#### SerializableCollection

Container for arbitrary serializable objects.

```go
func NewSerializableCollection(name string, children []SerializableObject, metadata AnyDictionary) *SerializableCollection
```

---

#### UnknownSchema

Preserves unknown schema types during round-trip.

```go
func NewUnknownSchema(schemaStr string, data AnyDictionary) *UnknownSchema
```

---

#### AnyDictionary

Arbitrary metadata storage.

```go
type AnyDictionary map[string]any
```

---

## Package: algorithms

```go
import "github.com/mrjoshuak/gotio/algorithms"
```

### Track Algorithms

```go
// Trim track to range
func TrackTrimmedToRange(track *opentimelineio.Track, trimRange opentime.TimeRange) (*opentimelineio.Track, error)

// Expand transitions
func TrackWithExpandedTransitions(track *opentimelineio.Track) (*opentimelineio.Track, error)
```

### Stack Algorithms

```go
// Flatten stack to single track
func FlattenStack(stack *opentimelineio.Stack) (*opentimelineio.Track, error)

// Flatten multiple tracks
func FlattenTracks(tracks []*opentimelineio.Track) (*opentimelineio.Track, error)

// Get topmost clip at time
func TopClipAtTime(stack *opentimelineio.Stack, t opentime.RationalTime) *opentimelineio.Clip
```

### Timeline Algorithms

```go
// Trim timeline to range
func TimelineTrimmedToRange(timeline *opentimelineio.Timeline, trimRange opentime.TimeRange) (*opentimelineio.Timeline, error)

// Get video tracks
func TimelineVideoTracks(timeline *opentimelineio.Timeline) []*opentimelineio.Track

// Get audio tracks
func TimelineAudioTracks(timeline *opentimelineio.Timeline) []*opentimelineio.Track

// Flatten video tracks
func FlattenTimelineVideoTracks(timeline *opentimelineio.Timeline) (*opentimelineio.Timeline, error)
```

### Filtering

```go
// Filter function type
type FilterFunc func(obj opentimelineio.SerializableObject) bool

// Context-aware filter
type ContextFilterFunc func(obj opentimelineio.SerializableObject, ctx FilterContext) bool

// Filter composition
func FilteredComposition(root opentimelineio.SerializableObject, filter FilterFunc, typesToPrune []reflect.Type) opentimelineio.SerializableObject

// Filter with context
func FilteredWithSequenceContext(root opentimelineio.SerializableObject, filter ContextFilterFunc, typesToPrune []reflect.Type) opentimelineio.SerializableObject

// Built-in filters
func KeepFilter() FilterFunc                          // Keep everything
func PruneFilter() FilterFunc                         // Prune everything
func TypeFilter(types ...reflect.Type) FilterFunc     // Filter by type
func NameFilter(pattern string) FilterFunc            // Filter by name pattern
```

---

## Error Types

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

// Missing reference
type MissingReferenceError struct{}

// Type mismatch
type TypeMismatchError struct {
    Expected string
    Got      string
}
```

All error types implement the `error` interface.
