# gotio Implementation Plan

This document outlines the work required to achieve feature parity with the original OpenTimelineIO implementation.

## Current State

| Package | Test Coverage | Status |
|---------|--------------|--------|
| opentime | 97.9% | Complete |
| opentimelineio | 90.1% | Complete |
| algorithms | 92.0% | Complete |
| bundle | - | Complete |
| medialinker | - | Complete |
| adapters | - | Complete (6 adapters) |

## Completed Work

### Phase 1: Edit Algorithm Suite ✓
All 11 edit operations implemented:
- Overwrite, Insert, Trim, Slice
- Slip, Slide, Ripple, Roll
- Fill, Remove
- Core infrastructure and utilities

### Phase 2: File Bundle Support ✓
- OTIOZ format (ZIP) read/write
- OTIOD format (directory) read/write
- All media reference policies implemented

### Phase 3: Media Linker Interface ✓
- MediaLinker interface defined
- Registry with default linker support
- Built-in linkers (PathTemplate, Directory)

### Additional Completions:
- **TransformedTime**: Coordinate space transformations between items in hierarchy
- **Format Adapters**: CMX3600 EDL, FCP7 XML, FCPXML (FCP X), HLS, ALE, XGES

## Original Scope

This plan covered:
- Edit Algorithm Suite (11 operations) ✓
- File Bundle Support (.otioz, .otiod) ✓
- Media Linker Interface (standalone, without plugin system) ✓

Also completed:
- Python Bridge (plugin system via subprocess)
- AAF format support (via Python bridge)

---

## Phase 1: Edit Algorithm Suite ✓ COMPLETE

The edit algorithm suite provides essential operations for interactive timeline editing.

### 1.1 Core Infrastructure

Before implementing edit operations, add supporting infrastructure to the existing codebase.

**File:** `algorithms/edit_utils.go`

```go
// ReferencePoint determines how fill operations place clips
type ReferencePoint int

const (
    // ReferencePointSource uses clip's natural duration
    ReferencePointSource ReferencePoint = iota
    // ReferencePointSequence trims clip to fit gap exactly
    ReferencePointSequence
    // ReferencePointFit adds time warp to stretch clip to gap
    ReferencePointFit
)

// EditError represents errors from edit operations
type EditError struct {
    Operation string
    Message   string
}
```

**Required helper functions:**
- `findItemAtTime(comp, time) (Item, int, error)` - Find item and index at time
- `findItemsInRange(comp, range) []Item` - Find all items intersecting range
- `splitItemAtTime(item, time) (Item, Item, error)` - Split item into two parts
- `removeTransitionsInRange(track, range)` - Remove transitions in range
- `clampToAvailableRange(item, sourceRange) TimeRange` - Clamp to media bounds

**Estimated size:** ~300 lines

---

### 1.2 Overwrite Operation

Replaces content in a time range with a new item.

**File:** `algorithms/edit_overwrite.go`

**Signature:**
```go
func Overwrite(
    item Item,
    composition Composition,
    timeRange opentime.TimeRange,
    opts ...OverwriteOption,
) error

type OverwriteOption func(*overwriteConfig)

func WithRemoveTransitions(remove bool) OverwriteOption
func WithFillTemplate(template Item) OverwriteOption
```

**Behavior:**
1. If range starts after composition end: create gap, append item
2. If range ends before composition start: insert item at beginning
3. Otherwise: find items in range, split at boundaries, remove middle items, insert new item

**Test cases:**
- Overwrite at end (append)
- Overwrite at beginning (prepend)
- Overwrite middle (split + replace)
- Overwrite spanning multiple clips
- Overwrite with transitions present
- Overwrite with fill template

**Estimated size:** ~250 lines + ~200 lines tests

---

### 1.3 Insert Operation

Inserts an item at a specific time, growing the composition.

**File:** `algorithms/edit_insert.go`

**Signature:**
```go
func Insert(
    item Item,
    composition Composition,
    time opentime.RationalTime,
    opts ...InsertOption,
) error

type InsertOption func(*insertConfig)

func WithInsertRemoveTransitions(remove bool) InsertOption
func WithInsertFillTemplate(template Item) InsertOption
```

**Behavior:**
1. If time >= end: append (with gap fill if needed)
2. If time < start: prepend
3. Otherwise: split item at time, insert between halves

**Test cases:**
- Insert at end
- Insert at beginning
- Insert in middle of clip
- Insert at clip boundary
- Insert with transitions

**Estimated size:** ~150 lines + ~150 lines tests

---

### 1.4 Trim Operation

Adjusts an item's in/out points without affecting composition duration.

**File:** `algorithms/edit_trim.go`

**Signature:**
```go
func Trim(
    item Item,
    deltaIn opentime.RationalTime,
    deltaOut opentime.RationalTime,
    opts ...TrimOption,
) error

type TrimOption func(*trimConfig)

func WithTrimFillTemplate(template Item) TrimOption
```

**Behavior:**
1. Adjust source_range start by deltaIn
2. Adjust source_range duration by deltaOut
3. Compensate by adjusting previous/next items (usually gaps)
4. Create gaps if expanding into empty space

**Test cases:**
- Trim head (positive deltaIn)
- Extend head (negative deltaIn)
- Trim tail (negative deltaOut)
- Extend tail (positive deltaOut)
- Trim both ends
- Trim adjacent to gap vs clip

**Estimated size:** ~200 lines + ~200 lines tests

---

### 1.5 Slice Operation

Cuts an item at a specific time, creating two items.

**File:** `algorithms/edit_slice.go`

**Signature:**
```go
func Slice(
    composition Composition,
    time opentime.RationalTime,
    opts ...SliceOption,
) error

type SliceOption func(*sliceConfig)

func WithSliceRemoveTransitions(remove bool) SliceOption
```

**Behavior:**
1. Find item at time
2. If slice at boundary: no-op
3. Calculate split point in source media
4. Create two items with adjusted source ranges
5. Replace original with both items

**Test cases:**
- Slice in middle of clip
- Slice at clip boundary (no-op)
- Slice at composition start/end
- Slice through transition (error or remove)
- Slice gap

**Estimated size:** ~120 lines + ~150 lines tests

---

### 1.6 Slip Operation

Moves the playhead through source media without changing position or duration.

**File:** `algorithms/edit_slip.go`

**Signature:**
```go
func Slip(item Item, delta opentime.RationalTime) error
```

**Behavior:**
1. Add delta to source_range.start_time
2. Clamp to available_range if present
3. Duration and position unchanged

**Test cases:**
- Slip forward
- Slip backward
- Slip with clamping to available range
- Slip without available range

**Estimated size:** ~60 lines + ~80 lines tests

---

### 1.7 Slide Operation

Moves an item by adjusting the previous item's duration.

**File:** `algorithms/edit_slide.go`

**Signature:**
```go
func Slide(item Item, delta opentime.RationalTime) error
```

**Behavior:**
1. Get previous item
2. Adjust previous item's duration by delta
3. Item moves left (negative delta) or right (positive delta)
4. Clamp to prevent negative durations

**Test cases:**
- Slide right (expand previous)
- Slide left (contract previous)
- Slide first item (error or no-op)
- Slide with available range limits

**Estimated size:** ~80 lines + ~100 lines tests

---

### 1.8 Ripple Operation

Adjusts source range with clamping to available media.

**File:** `algorithms/edit_ripple.go`

**Signature:**
```go
func Ripple(
    item Item,
    deltaIn opentime.RationalTime,
    deltaOut opentime.RationalTime,
) error
```

**Behavior:**
1. Adjust source start with clamping
2. Adjust source end with clamping
3. Available range bounds checked
4. No effect on adjacent items

**Test cases:**
- Ripple in
- Ripple out
- Ripple both
- Ripple with clamping
- Ripple without available range

**Estimated size:** ~100 lines + ~120 lines tests

---

### 1.9 Roll Operation

Moves an edit point, adjusting both adjacent items.

**File:** `algorithms/edit_roll.go`

**Signature:**
```go
func Roll(
    item Item,
    deltaIn opentime.RationalTime,
    deltaOut opentime.RationalTime,
) error
```

**Behavior:**
1. For deltaIn: adjust current start AND previous end
2. For deltaOut: adjust current end AND next start
3. Edit point moves, total duration preserved
4. Both items' source ranges adjusted

**Test cases:**
- Roll in (move head edit point)
- Roll out (move tail edit point)
- Roll with adjacent gaps
- Roll with adjacent clips
- Roll at composition boundaries

**Estimated size:** ~150 lines + ~180 lines tests

---

### 1.10 Fill Operation (3/4-Point Edit)

Places a clip into a gap with configurable fitting.

**File:** `algorithms/edit_fill.go`

**Signature:**
```go
func Fill(
    item Item,
    track *opentimelineio.Track,
    trackTime opentime.RationalTime,
    referencePoint ReferencePoint,
) error
```

**Behavior by ReferencePoint:**

1. **Source**: Use clip's natural duration, overwrite from trackTime
2. **Sequence**: Trim clip to fit gap exactly
3. **Fit**: Add LinearTimeWarp effect to stretch/compress to gap

**Test cases:**
- Fill with Source mode
- Fill with Sequence mode (clip longer than gap)
- Fill with Sequence mode (clip shorter than gap)
- Fill with Fit mode
- Fill non-gap (error)

**Estimated size:** ~180 lines + ~200 lines tests

---

### 1.11 Remove Operation

Removes an item and optionally fills the space.

**File:** `algorithms/edit_remove.go`

**Signature:**
```go
func Remove(
    composition Composition,
    time opentime.RationalTime,
    opts ...RemoveOption,
) error

type RemoveOption func(*removeConfig)

func WithFill(fill bool) RemoveOption
func WithRemoveFillTemplate(template Item) RemoveOption
```

**Behavior:**
1. Find item at time
2. Remove from composition
3. If fill=true: insert gap or template
4. If fill=false: adjacent items become adjacent

**Test cases:**
- Remove with fill (gap)
- Remove without fill (collapse)
- Remove with template
- Remove at invalid time (error)

**Estimated size:** ~100 lines + ~120 lines tests

---

### Phase 1 Summary

| Component | Code Lines | Test Lines | Total |
|-----------|-----------|------------|-------|
| edit_utils.go | 300 | 100 | 400 |
| edit_overwrite.go | 250 | 200 | 450 |
| edit_insert.go | 150 | 150 | 300 |
| edit_trim.go | 200 | 200 | 400 |
| edit_slice.go | 120 | 150 | 270 |
| edit_slip.go | 60 | 80 | 140 |
| edit_slide.go | 80 | 100 | 180 |
| edit_ripple.go | 100 | 120 | 220 |
| edit_roll.go | 150 | 180 | 330 |
| edit_fill.go | 180 | 200 | 380 |
| edit_remove.go | 100 | 120 | 220 |
| **Total** | **1,690** | **1,600** | **3,290** |

---

## Phase 2: File Bundle Support ✓ COMPLETE

File bundles allow packaging timelines with their media for distribution.

### 2.1 Bundle Types

**File:** `opentimelineio/bundle/types.go`

```go
// MediaReferencePolicy determines how media references are handled
type MediaReferencePolicy int

const (
    // ErrorIfNotFile raises error for non-file references
    ErrorIfNotFile MediaReferencePolicy = iota
    // MissingIfNotFile replaces with MissingReference
    MissingIfNotFile
    // AllMissing replaces all references with MissingReference
    AllMissing
)

// BundleManifest maps absolute paths to references pointing to them
type BundleManifest map[string][]*ExternalReference
```

**Estimated size:** ~50 lines

---

### 2.2 Bundle Utilities

**File:** `opentimelineio/bundle/utils.go`

**Functions:**
```go
// PrepareForBundle processes media references according to policy
func PrepareForBundle(
    timeline *Timeline,
    policy MediaReferencePolicy,
) (*Timeline, BundleManifest, error)

// VerifyUniqueBasenames checks for basename collisions
func VerifyUniqueBasenames(manifest BundleManifest) error

// RelinkToBundle updates references to bundle paths
func RelinkToBundle(timeline *Timeline, manifest BundleManifest) error

// ConvertToAbsolutePaths converts relative bundle paths to absolute
func ConvertToAbsolutePaths(timeline *Timeline, bundleRoot string) error
```

**Behavior:**
1. Deep copy timeline to avoid mutation
2. Iterate all clips via FindClips()
3. Process each media reference per policy
4. Build manifest of files to include
5. Verify unique basenames
6. Relink to relative bundle paths

**Estimated size:** ~300 lines + ~200 lines tests

---

### 2.3 OTIOZ Format (ZIP)

**File:** `opentimelineio/bundle/otioz.go`

**Structure:**
```
file.otioz (ZIP archive)
├── content.otio (ZIP_DEFLATED)
├── version.txt  (ZIP_DEFLATED, contains "1.0.0")
└── media/       (directory)
    └── *.mov    (ZIP_STORED - no compression)
```

**Functions:**
```go
// ReadOTIOZ reads a .otioz bundle
func ReadOTIOZ(path string) (*Timeline, error)

// ReadOTIOZWithExtraction reads and extracts media to directory
func ReadOTIOZWithExtraction(path, extractDir string) (*Timeline, error)

// WriteOTIOZ writes a .otioz bundle
func WriteOTIOZ(
    timeline *Timeline,
    path string,
    policy MediaReferencePolicy,
) error

// WriteOTIOZDryRun returns total size without writing
func WriteOTIOZDryRun(
    timeline *Timeline,
    policy MediaReferencePolicy,
) (int64, error)
```

**Implementation notes:**
- Use Go's `archive/zip` package
- Media files stored uncompressed (ZIP_STORED) since they're already compressed
- content.otio and version.txt use ZIP_DEFLATED
- POSIX-style paths internally for cross-platform compatibility

**Estimated size:** ~250 lines + ~200 lines tests

---

### 2.4 OTIOD Format (Directory)

**File:** `opentimelineio/bundle/otiod.go`

**Structure:**
```
directory.otiod/
├── content.otio
└── media/
    └── *.mov
```

**Functions:**
```go
// ReadOTIOD reads a .otiod bundle directory
func ReadOTIOD(path string, absolutePaths bool) (*Timeline, error)

// WriteOTIOD writes a .otiod bundle directory
func WriteOTIOD(
    timeline *Timeline,
    path string,
    policy MediaReferencePolicy,
) error

// WriteOTIODDryRun returns total size without writing
func WriteOTIODDryRun(
    timeline *Timeline,
    policy MediaReferencePolicy,
) (int64, error)
```

**Implementation notes:**
- Simple file/directory operations
- No compression
- Media files copied directly

**Estimated size:** ~150 lines + ~150 lines tests

---

### Phase 2 Summary

| Component | Code Lines | Test Lines | Total |
|-----------|-----------|------------|-------|
| types.go | 50 | 0 | 50 |
| utils.go | 300 | 200 | 500 |
| otioz.go | 250 | 200 | 450 |
| otiod.go | 150 | 150 | 300 |
| **Total** | **750** | **550** | **1,300** |

---

## Phase 3: Media Linker Interface ✓ COMPLETE

The media linker resolves MissingReferences to actual media files.

### 3.1 Media Linker Interface

**File:** `opentimelineio/medialinker/linker.go`

```go
// MediaLinker resolves media references for clips
type MediaLinker interface {
    // LinkMediaReference resolves a clip's media reference
    LinkMediaReference(
        clip *Clip,
        args map[string]any,
    ) (MediaReference, error)
}

// LinkingPolicy determines when to apply media linking
type LinkingPolicy int

const (
    // DoNotLink skips media linking entirely
    DoNotLink LinkingPolicy = iota
    // UseDefaultLinker uses the registered default linker
    UseDefaultLinker
)
```

**Estimated size:** ~30 lines

---

### 3.2 Linker Registry

**File:** `opentimelineio/medialinker/registry.go`

```go
// Registry holds registered media linkers
type Registry struct {
    linkers       map[string]MediaLinker
    defaultLinker string
}

// Global registry instance
var defaultRegistry = &Registry{
    linkers: make(map[string]MediaLinker),
}

// Register adds a linker to the registry
func Register(name string, linker MediaLinker)

// Get retrieves a linker by name
func Get(name string) (MediaLinker, error)

// SetDefault sets the default linker name
func SetDefault(name string)

// Default returns the default linker
func Default() (MediaLinker, error)

// Available returns all registered linker names
func Available() []string
```

**Estimated size:** ~80 lines + ~50 lines tests

---

### 3.3 Linking Execution

**File:** `opentimelineio/medialinker/link.go`

```go
// LinkMedia applies a media linker to all clips in a timeline
func LinkMedia(
    timeline *Timeline,
    linkerName string,
    args map[string]any,
) error

// LinkMediaWithLinker applies a specific linker instance
func LinkMediaWithLinker(
    timeline *Timeline,
    linker MediaLinker,
    args map[string]any,
) error
```

**Behavior:**
1. Iterate all clips via timeline.FindClips()
2. For each clip, call linker.LinkMediaReference()
3. Replace clip's media reference with result
4. Continue on error or fail fast (configurable)

**Estimated size:** ~100 lines + ~100 lines tests

---

### 3.4 Built-in Linkers

**File:** `opentimelineio/medialinker/builtin.go`

```go
// PathTemplateLinker resolves paths using a template
type PathTemplateLinker struct {
    Template string // e.g., "/media/{name}.mov"
}

func (l *PathTemplateLinker) LinkMediaReference(
    clip *Clip,
    args map[string]any,
) (MediaReference, error)

// DirectoryLinker searches directories for matching files
type DirectoryLinker struct {
    SearchPaths []string
    Extensions  []string // e.g., [".mov", ".mp4"]
}

func (l *DirectoryLinker) LinkMediaReference(
    clip *Clip,
    args map[string]any,
) (MediaReference, error)
```

**Estimated size:** ~150 lines + ~150 lines tests

---

### Phase 3 Summary

| Component | Code Lines | Test Lines | Total |
|-----------|-----------|------------|-------|
| linker.go | 30 | 0 | 30 |
| registry.go | 80 | 50 | 130 |
| link.go | 100 | 100 | 200 |
| builtin.go | 150 | 150 | 300 |
| **Total** | **360** | **300** | **660** |

---

## Implementation Order

### Recommended Sequence

1. **Phase 1.1-1.2**: Core infrastructure + Overwrite
   - Foundation for all other operations
   - Most complex operation, validates infrastructure

2. **Phase 1.3-1.5**: Insert, Trim, Slice
   - Common editing operations
   - Build on overwrite infrastructure

3. **Phase 1.6-1.8**: Slip, Slide, Ripple
   - Simpler operations
   - Source range manipulation

4. **Phase 1.9-1.11**: Roll, Fill, Remove
   - Advanced operations
   - Complete the edit algorithm suite

5. **Phase 2**: File Bundle Support
   - Independent of edit algorithms
   - Useful for distribution/archival

6. **Phase 3**: Media Linker
   - Independent feature
   - Complements bundle support

---

## Testing Strategy

### Unit Tests

Each operation requires:
- Basic functionality test
- Edge case tests (boundaries, empty, single item)
- Error condition tests
- Round-trip tests (operation + inverse)

### Integration Tests

- Multi-operation sequences
- Complex timeline scenarios
- Compatibility with Python/C++ OTIO files

### Benchmark Tests

- Performance on large timelines (1000+ clips)
- Memory allocation patterns
- Comparison with original implementation

---

## API Design Principles

### Functional Options Pattern

All operations use the functional options pattern for configuration:

```go
func Operation(required args, opts ...OptionFunc) error

// Example usage
err := Overwrite(clip, track, timeRange,
    WithRemoveTransitions(true),
    WithFillTemplate(gap),
)
```

Benefits:
- Sensible defaults
- Backward compatible extension
- Clear, readable call sites

### Error Handling

Operations return descriptive errors:

```go
type EditError struct {
    Operation string
    Item      string
    Time      opentime.RationalTime
    Message   string
}

func (e *EditError) Error() string
```

### Immutability Options

Some operations can work in two modes:
- **Mutating**: Modify composition in place (default, matches original)
- **Copying**: Return modified copy, leave original unchanged

```go
func OverwriteCopy(...) (*Track, error)  // Returns new track
func Overwrite(...)    error             // Modifies in place
```

---

## Total Estimates

| Phase | Code Lines | Test Lines | Total |
|-------|-----------|------------|-------|
| Phase 1: Edit Algorithms | 1,690 | 1,600 | 3,290 |
| Phase 2: File Bundles | 750 | 550 | 1,300 |
| Phase 3: Media Linker | 360 | 300 | 660 |
| **Grand Total** | **2,800** | **2,450** | **5,250** |

This would increase the codebase from ~18,500 lines to ~23,750 lines, with test coverage maintained at 90%+.

---

## Success Criteria

### Phase 1 Complete When:
- All 11 edit operations implemented
- Test coverage > 90%
- All original test cases ported and passing
- Documentation with examples for each operation

### Phase 2 Complete When:
- .otioz read/write working
- .otiod read/write working
- All three media policies implemented
- Cross-platform path handling verified
- Round-trip tests passing

### Phase 3 Complete When:
- MediaLinker interface defined
- Registry working
- Built-in linkers functional
- Integration with bundle reading

---

## Python Bridge (Plugin System) ✓ COMPLETE

The Python bridge provides full interoperability with all Python OTIO adapters:

**Implementation:** `adapters/bridge.go` + `adapters/bridge.py`

```go
bridge, err := adapters.NewBridge(adapters.Config{})
defer bridge.Close()

// Discover all Python adapters
formats := bridge.AvailableFormats()

// Read/write using ANY Python adapter (including AAF)
timeline, err := bridge.Read("project.aaf")
err = bridge.Write(timeline, "output.fcpxml")
```

**Features:**
- Long-lived Python subprocess (avoids startup overhead)
- JSON-RPC communication over stdin/stdout
- Auto-discovers all Python OTIO adapters via plugin manifest
- Supports reading and writing any format Python supports
- **AAF support** via Python's AAF adapter

This design sidesteps Go's static compilation limitation by delegating to Python for formats that require complex parsing (AAF, etc.) while native Go adapters provide maximum performance for common formats.

---

## Future Work

The following remain for future development:

1. **Hook System**
   - Pre/post serialization hooks
   - Custom validation

## Completed Format Adapters

The following format adapters have been implemented with Python feature parity:

| Adapter | Description | Status |
|---------|-------------|--------|
| CMX3600 | CMX 3600 EDL format | ✓ Complete |
| FCP7XML | Final Cut Pro 7 XML | ✓ Complete |
| FCPXML | Final Cut Pro X XML | ✓ Complete |
| HLS | HTTP Live Streaming playlists | ✓ Complete |
| ALE | Avid Log Exchange | ✓ Complete |
| XGES | GStreamer Editing Services | ✓ Complete |
| AAF | Advanced Authoring Format | ✓ Via Python bridge |
| SVG | Scalable Vector Graphics | ✓ Complete (output-only) |
