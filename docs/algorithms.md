# Algorithms

The `algorithms` package provides functions for manipulating timelines, tracks, and stacks. These algorithms are essential for common editorial operations like trimming, flattening, and filtering.

```go
import "github.com/mrjoshuak/gotio/algorithms"
```

## Track Algorithms

### TrackTrimmedToRange

Trims a track to a specific time range, keeping only the portions of children that fall within the range.

```go
func TrackTrimmedToRange(track *opentimelineio.Track, trimRange opentime.TimeRange) (*opentimelineio.Track, error)
```

**Use Cases:**
- Extracting a segment of a timeline
- Creating a subclip
- Removing unwanted content at the beginning or end

**Example:**

```go
// Original track: 10 seconds
originalTrack := timeline.VideoTracks()[0]

// Trim to frames 48-144 (2-6 seconds at 24fps)
trimRange := opentime.NewTimeRange(
    opentime.NewRationalTime(48, 24),
    opentime.NewRationalTime(96, 24),
)

trimmedTrack, err := algorithms.TrackTrimmedToRange(originalTrack, trimRange)
if err != nil {
    log.Fatal(err)
}

// trimmedTrack now contains only 4 seconds of content
dur, _ := trimmedTrack.Duration()
fmt.Printf("Trimmed duration: %.2fs\n", dur.ToSeconds())  // 4.00s
```

**Behavior:**

1. Clips entirely within the range are kept unchanged
2. Clips partially within the range are trimmed
3. Clips entirely outside the range are removed
4. Gaps are adjusted accordingly
5. Transitions at the edges may be removed

```
Original: [Clip A: 0-72] [Clip B: 72-144] [Clip C: 144-216]
Trim 48-168:
Result:   [Clip A: 24-72] [Clip B: 72-144] [Clip C: 144-168]
          (trimmed)       (unchanged)      (trimmed)
```

---

### TrackWithExpandedTransitions

Expands transitions to include the media they "borrow" from adjacent clips.

```go
func TrackWithExpandedTransitions(track *opentimelineio.Track) (*opentimelineio.Track, error)
```

**Use Cases:**
- Analyzing actual media usage including transition handles
- Preparing for format conversion
- Quality checking available handles

**Example:**

```go
// Track with: [Clip A] [Dissolve 12f] [Clip B]
expandedTrack, err := algorithms.TrackWithExpandedTransitions(originalTrack)

// Now clip source ranges include the transition frames
```

---

## Stack Algorithms

### FlattenStack

Flattens a multi-layer stack into a single track by resolving what's visible at each point in time.

```go
func FlattenStack(stack *opentimelineio.Stack) (*opentimelineio.Track, error)
```

**Use Cases:**
- Converting multi-track timelines for formats that only support single tracks
- Analyzing what's actually visible in the final output
- Simplifying complex compositions

**Example:**

```go
// Stack with multiple tracks
stack := timeline.Tracks()

// Flatten to single track
flatTrack, err := algorithms.FlattenStack(stack)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Flattened to %d items\n", len(flatTrack.Children()))
```

**Compositing Rules:**

1. Higher tracks (later in the stack) obscure lower tracks
2. Gaps are transparent (lower tracks show through)
3. The result contains the topmost visible content at each time

```
Track V2: [    ] [Clip B] [    ]    <- top
Track V1: [Clip A      ] [Clip C]   <- bottom

Flattened: [A  ] [B    ] [C   ]
           (A shows where V2 has gap)
```

---

### FlattenTracks

Flattens multiple tracks into a single track.

```go
func FlattenTracks(tracks []*opentimelineio.Track) (*opentimelineio.Track, error)
```

**Example:**

```go
videoTracks := timeline.VideoTracks()
flatTrack, err := algorithms.FlattenTracks(videoTracks)
```

---

### TopClipAtTime

Returns the topmost visible clip at a specific time.

```go
func TopClipAtTime(stack *opentimelineio.Stack, t opentime.RationalTime) *opentimelineio.Clip
```

**Use Cases:**
- Finding what's playing at a specific frame
- Building frame-accurate reports
- Implementing playback simulation

**Example:**

```go
// What clip is visible at frame 100?
t := opentime.NewRationalTime(100, 24)
clip := algorithms.TopClipAtTime(timeline.Tracks(), t)

if clip != nil {
    fmt.Printf("At frame 100: %s\n", clip.Name())
} else {
    fmt.Println("At frame 100: (gap)")
}
```

**Behavior:**

- Searches from top track to bottom
- Returns the first clip that covers the given time
- Returns nil if the time falls in a gap on all tracks

---

## Timeline Algorithms

### TimelineTrimmedToRange

Trims an entire timeline to a specific range.

```go
func TimelineTrimmedToRange(timeline *opentimelineio.Timeline, trimRange opentime.TimeRange) (*opentimelineio.Timeline, error)
```

**Example:**

```go
// Extract the first minute
trimRange := opentime.NewTimeRange(
    opentime.NewRationalTime(0, 24),
    opentime.NewRationalTime(1440, 24),  // 60 seconds at 24fps
)

trimmed, err := algorithms.TimelineTrimmedToRange(timeline, trimRange)
if err != nil {
    log.Fatal(err)
}

dur, _ := trimmed.Duration()
fmt.Printf("Trimmed timeline: %.2fs\n", dur.ToSeconds())  // 60.00s
```

---

### TimelineVideoTracks / TimelineAudioTracks

Get only video or audio tracks from a timeline.

```go
func TimelineVideoTracks(timeline *opentimelineio.Timeline) []*opentimelineio.Track
func TimelineAudioTracks(timeline *opentimelineio.Timeline) []*opentimelineio.Track
```

**Example:**

```go
videoTracks := algorithms.TimelineVideoTracks(timeline)
audioTracks := algorithms.TimelineAudioTracks(timeline)

fmt.Printf("Video tracks: %d\n", len(videoTracks))
fmt.Printf("Audio tracks: %d\n", len(audioTracks))
```

---

### FlattenTimelineVideoTracks

Creates a new timeline with video tracks flattened to a single track.

```go
func FlattenTimelineVideoTracks(timeline *opentimelineio.Timeline) (*opentimelineio.Timeline, error)
```

**Example:**

```go
// Original timeline has 5 video tracks
original := loadTimeline("complex_edit.otio")

// Flatten to 1 video track
flattened, err := algorithms.FlattenTimelineVideoTracks(original)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Flattened video tracks: %d\n", len(flattened.VideoTracks()))  // 1
```

---

## Filtering

The filtering functions allow you to traverse and filter compositions based on custom criteria.

### FilterFunc

A function type for filtering:

```go
type FilterFunc func(obj opentimelineio.SerializableObject) bool
```

Return `true` to keep the object, `false` to remove it.

---

### FilteredComposition

Filters a composition tree, keeping only objects that match the filter.

```go
func FilteredComposition(
    root opentimelineio.SerializableObject,
    filter FilterFunc,
    typesToPrune []reflect.Type,
) opentimelineio.SerializableObject
```

**Parameters:**
- `root` - The root object to filter
- `filter` - Function that returns true for objects to keep
- `typesToPrune` - Types to completely remove (including children)

**Example: Keep Only Clips**

```go
import "reflect"

// Filter to keep only clips
clipFilter := func(obj opentimelineio.SerializableObject) bool {
    _, isClip := obj.(*opentimelineio.Clip)
    return isClip
}

filtered := algorithms.FilteredComposition(timeline, clipFilter, nil)
```

**Example: Remove Gaps**

```go
// Remove all gaps
noGapsFilter := func(obj opentimelineio.SerializableObject) bool {
    _, isGap := obj.(*opentimelineio.Gap)
    return !isGap  // Keep everything except gaps
}

filtered := algorithms.FilteredComposition(timeline, noGapsFilter, nil)
```

---

### Built-in Filters

#### KeepFilter

Keeps everything (useful as a default):

```go
filter := algorithms.KeepFilter()
```

#### PruneFilter

Removes everything:

```go
filter := algorithms.PruneFilter()
```

#### TypeFilter

Keeps objects of specific types:

```go
import "reflect"

// Keep only Clips and Gaps
filter := algorithms.TypeFilter(
    reflect.TypeOf(&opentimelineio.Clip{}),
    reflect.TypeOf(&opentimelineio.Gap{}),
)

filtered := algorithms.FilteredComposition(timeline, filter, nil)
```

#### NameFilter

Keeps objects matching a name pattern:

```go
// Keep clips with names starting with "VFX_"
filter := algorithms.NameFilter("VFX_*")
filtered := algorithms.FilteredComposition(timeline, filter, nil)
```

---

### FilteredWithSequenceContext

Like FilteredComposition, but the filter receives context about the object's position:

```go
type FilterContext struct {
    Index       int                                // Position in parent
    Parent      opentimelineio.SerializableObject  // Parent composition
    Neighbors   []opentimelineio.SerializableObject // Adjacent items
}

type ContextFilterFunc func(obj opentimelineio.SerializableObject, ctx FilterContext) bool
```

**Example: Keep Every Other Clip**

```go
everyOther := func(obj opentimelineio.SerializableObject, ctx algorithms.FilterContext) bool {
    if _, isClip := obj.(*opentimelineio.Clip); isClip {
        return ctx.Index%2 == 0  // Keep clips at even indices
    }
    return true  // Keep non-clips
}

filtered := algorithms.FilteredWithSequenceContext(timeline, everyOther, nil)
```

---

## Common Patterns

### Extract Segment

```go
func extractSegment(timeline *opentimelineio.Timeline, startSec, endSec float64) (*opentimelineio.Timeline, error) {
    rate := 24.0  // Assume 24fps
    trimRange := opentime.NewTimeRange(
        opentime.NewRationalTime(startSec*rate, rate),
        opentime.NewRationalTime((endSec-startSec)*rate, rate),
    )
    return algorithms.TimelineTrimmedToRange(timeline, trimRange)
}

// Extract seconds 10-30
segment, err := extractSegment(timeline, 10, 30)
```

### Find Clips with Markers

```go
func clipsWithMarkers(timeline *opentimelineio.Timeline) []*opentimelineio.Clip {
    var result []*opentimelineio.Clip
    for _, clip := range timeline.FindClips(nil, false) {
        if len(clip.Markers()) > 0 {
            result = append(result, clip)
        }
    }
    return result
}
```

### Build Edit List

```go
type EditEntry struct {
    ClipName  string
    MediaFile string
    InTC      string
    OutTC     string
    RecordIn  string
    RecordOut string
}

func buildEditList(timeline *opentimelineio.Timeline) []EditEntry {
    var entries []EditEntry

    for _, track := range timeline.VideoTracks() {
        recordTime := opentime.NewRationalTime(0, 24)

        for _, child := range track.Children() {
            clip, ok := child.(*opentimelineio.Clip)
            if !ok {
                continue
            }

            sr := clip.SourceRange()
            dur, _ := clip.Duration()

            var mediaURL string
            if ref, ok := clip.MediaReference().(*opentimelineio.ExternalReference); ok {
                mediaURL = ref.TargetURL()
            }

            entry := EditEntry{
                ClipName:  clip.Name(),
                MediaFile: mediaURL,
                InTC:      formatTC(sr.StartTime()),
                OutTC:     formatTC(sr.EndTimeExclusive()),
                RecordIn:  formatTC(recordTime),
                RecordOut: formatTC(recordTime.Add(dur)),
            }
            entries = append(entries, entry)

            recordTime = recordTime.Add(dur)
        }
    }

    return entries
}
```

### Conform Timeline

```go
// Replace all media references with new paths
func conformTimeline(timeline *opentimelineio.Timeline, pathMap map[string]string) {
    for _, clip := range timeline.FindClips(nil, false) {
        ref, ok := clip.MediaReference().(*opentimelineio.ExternalReference)
        if !ok {
            continue
        }

        oldPath := ref.TargetURL()
        if newPath, found := pathMap[oldPath]; found {
            ref.SetTargetURL(newPath)
        }
    }
}

// Usage
pathMap := map[string]string{
    "file:///old/path/video.mov": "file:///new/path/video.mov",
}
conformTimeline(timeline, pathMap)
```

---

## Performance Considerations

1. **Flattening is O(n*m)** where n = tracks and m = items per track
2. **TopClipAtTime is O(n)** where n = total items in stack
3. **Filtering creates copies** - original objects are not modified
4. **Large timelines** - Consider processing tracks in parallel

```go
// Parallel track processing
var wg sync.WaitGroup
results := make([]*opentimelineio.Track, len(tracks))

for i, track := range tracks {
    wg.Add(1)
    go func(idx int, t *opentimelineio.Track) {
        defer wg.Done()
        processed, _ := processTrack(t)
        results[idx] = processed
    }(i, track)
}

wg.Wait()
```

---

## Edit Algorithms

The edit algorithm suite provides operations for interactive timeline editing. These operations modify compositions in place and handle the complexities of splitting clips, managing transitions, and maintaining timeline integrity.

### Types and Errors

```go
// ReferencePoint determines how fill operations place clips
type ReferencePoint int

const (
    ReferencePointSource   ReferencePoint = iota // Use clip's natural duration
    ReferencePointSequence                        // Trim clip to fit gap exactly
    ReferencePointFit                             // Add time warp to stretch/compress
)

// EditError represents an error from edit operations
type EditError struct {
    Operation string
    Message   string
    Time      *opentime.RationalTime
    Item      opentimelineio.Composable
}
```

---

### Overwrite

Replaces content in a time range with a new item.

```go
func Overwrite(
    item opentimelineio.Item,
    composition opentimelineio.Composition,
    timeRange opentime.TimeRange,
    opts ...OverwriteOption,
) error

// Options
func WithRemoveTransitions(remove bool) OverwriteOption
func WithFillTemplate(template opentimelineio.Item) OverwriteOption
```

**Behavior:**
- If range starts after composition end: creates gap, appends item
- If range ends before composition start: inserts item at beginning
- Otherwise: splits items at boundaries, removes items in range, inserts new item

**Example:**

```go
// Create a clip to insert
clip := opentimelineio.NewClip("new_clip", mediaRef, nil, nil, nil, nil, "", nil)

// Overwrite frames 100-200
overwriteRange := opentime.NewTimeRange(
    opentime.NewRationalTime(100, 24),
    opentime.NewRationalTime(100, 24),
)

err := algorithms.Overwrite(clip, track, overwriteRange)
```

---

### Insert

Inserts an item at a specific time, growing the composition.

```go
func Insert(
    item opentimelineio.Item,
    composition opentimelineio.Composition,
    time opentime.RationalTime,
    opts ...InsertOption,
) error

// Options
func WithInsertRemoveTransitions(remove bool) InsertOption
func WithInsertFillTemplate(template opentimelineio.Item) InsertOption
```

**Behavior:**
- If time >= composition end: appends (with gap fill if needed)
- If time <= 0: prepends
- Otherwise: splits item at time, inserts between halves

**Example:**

```go
// Insert a clip at frame 50
insertTime := opentime.NewRationalTime(50, 24)
err := algorithms.Insert(clip, track, insertTime)
```

---

### Slice

Cuts an item at a specific time, creating two items.

```go
func Slice(
    composition opentimelineio.Composition,
    time opentime.RationalTime,
    opts ...SliceOption,
) error

// Options
func WithSliceRemoveTransitions(remove bool) SliceOption
```

**Behavior:**
- If time is at item boundary: no-op
- If time is within an item: splits into two items with adjusted source ranges
- Does not change composition duration

**Example:**

```go
// Cut the clip at frame 100
sliceTime := opentime.NewRationalTime(100, 24)
err := algorithms.Slice(track, sliceTime)
// Track now has two clips where there was one
```

---

### Trim

Adjusts an item's in/out points without affecting composition duration. Adjacent items are adjusted to compensate.

```go
func Trim(
    item opentimelineio.Item,
    composition opentimelineio.Composition,
    deltaIn opentime.RationalTime,
    deltaOut opentime.RationalTime,
    opts ...TrimOption,
) error

// Options
func WithTrimFillTemplate(template opentimelineio.Item) TrimOption
```

**Behavior:**
- deltaIn > 0: moves source start forward, previous item expands
- deltaIn < 0: moves source start backward, previous item contracts
- deltaOut > 0: extends duration, next item contracts
- deltaOut < 0: reduces duration, next item expands

**Example:**

```go
// Trim 10 frames off the head, extend 5 frames at the tail
err := algorithms.Trim(
    clip, track,
    opentime.NewRationalTime(10, 24),  // trim head
    opentime.NewRationalTime(5, 24),   // extend tail
)
```

---

### Slip

Moves an item's playhead through source media without changing position or duration.

```go
func Slip(item opentimelineio.Item, delta opentime.RationalTime) error
```

**Behavior:**
- Adds delta to source_range.start_time
- Clamps to available_range if present
- Duration and position in composition unchanged

**Example:**

```go
// Slip the clip 24 frames forward in source media
err := algorithms.Slip(clip, opentime.NewRationalTime(24, 24))
// Clip shows different content but stays in same position
```

---

### Slide

Moves an item's position by adjusting the previous item's duration.

```go
func Slide(
    item opentimelineio.Item,
    composition opentimelineio.Composition,
    delta opentime.RationalTime,
) error
```

**Behavior:**
- Positive delta: expands previous item, moves this item right
- Negative delta: contracts previous item, moves this item left
- Clamps to prevent negative durations
- First item cannot be slid

**Example:**

```go
// Slide clip 12 frames to the right
err := algorithms.Slide(clip, track, opentime.NewRationalTime(12, 24))
```

---

### Ripple

Adjusts an item's source range with clamping to available media. Unlike Trim, Ripple does not affect adjacent items.

```go
func Ripple(
    item opentimelineio.Item,
    deltaIn opentime.RationalTime,
    deltaOut opentime.RationalTime,
) error
```

**Behavior:**
- deltaIn adjusts source start with clamping
- deltaOut adjusts source end with clamping
- Available range bounds are checked
- No effect on adjacent items (composition duration changes)

**Example:**

```go
// Extend the clip's in-point by 10 frames
err := algorithms.Ripple(
    clip,
    opentime.NewRationalTime(-10, 24),  // extend head
    opentime.NewRationalTime(0, 24),    // no tail change
)
```

---

### Roll

Moves an edit point, adjusting both adjacent items. Total duration is preserved.

```go
func Roll(
    item opentimelineio.Item,
    composition opentimelineio.Composition,
    deltaIn opentime.RationalTime,
    deltaOut opentime.RationalTime,
) error
```

**Behavior:**
- deltaIn adjusts current item's head and previous item's tail
- deltaOut adjusts current item's tail and next item's head
- Edit point moves, but total duration is preserved

**Example:**

```go
// Roll the in-point 10 frames right (current clip shrinks, previous expands)
err := algorithms.Roll(
    clip, track,
    opentime.NewRationalTime(10, 24),  // roll in-point right
    opentime.NewRationalTime(0, 24),   // no out-point change
)
```

---

### Fill

Places an item into a gap using 3/4-point edit logic.

```go
func Fill(
    item opentimelineio.Item,
    composition opentimelineio.Composition,
    trackTime opentime.RationalTime,
    referencePoint ReferencePoint,
) error
```

**Behavior by ReferencePoint:**
- **Source**: Use clip's natural duration, perform overwrite from trackTime
- **Sequence**: Trim clip to fit gap exactly
- **Fit**: Add LinearTimeWarp effect to stretch/compress clip to gap

**Example:**

```go
// Fill a gap at frame 100, trimming clip to fit
err := algorithms.Fill(
    clip, track,
    opentime.NewRationalTime(100, 24),
    algorithms.ReferencePointSequence,
)
```

---

### Remove

Removes an item at a specific time and optionally fills the space.

```go
func Remove(
    composition opentimelineio.Composition,
    time opentime.RationalTime,
    opts ...RemoveOption,
) error

// Options
func WithFill(fill bool) RemoveOption
func WithRemoveFillTemplate(template opentimelineio.Item) RemoveOption
```

**Behavior:**
- If fill=true (default): inserts a gap in place of removed item
- If fill=false: adjacent items become adjacent (composition shrinks)

**Example:**

```go
// Remove item at frame 100, leaving a gap
err := algorithms.Remove(track, opentime.NewRationalTime(100, 24))

// Remove item at frame 100, collapsing the timeline
err := algorithms.Remove(
    track,
    opentime.NewRationalTime(100, 24),
    algorithms.WithFill(false),
)
```

---

### RemoveRange

Removes all items within a time range.

```go
func RemoveRange(
    composition opentimelineio.Composition,
    timeRange opentime.TimeRange,
    opts ...RemoveOption,
) error
```

**Behavior:**
- Items completely within range are removed
- Items partially within range are trimmed
- If fill=true: a single gap fills the removed space
- If fill=false: composition shrinks

**Example:**

```go
// Remove everything between frames 100-200
removeRange := opentime.NewTimeRange(
    opentime.NewRationalTime(100, 24),
    opentime.NewRationalTime(100, 24),
)
err := algorithms.RemoveRange(track, removeRange, algorithms.WithFill(false))
```

---

### Edit Algorithm Summary

| Operation | Affects Duration | Affects Adjacent | Use Case |
|-----------|-----------------|------------------|----------|
| Overwrite | No* | Yes | Replace content in timeline |
| Insert | Yes | Yes | Add new content, push existing right |
| Slice | No | No | Cut clip into two pieces |
| Trim | No | Yes | Adjust in/out, adjacent compensates |
| Slip | No | No | Change source content, keep position |
| Slide | No | Yes | Move clip by adjusting previous |
| Ripple | Yes | No | Adjust source range, timeline shrinks/grows |
| Roll | No | Yes | Move edit point between two clips |
| Fill | No* | Yes | Place clip into a gap |
| Remove | Varies | Yes | Delete content from timeline |

*Unless overwriting/filling beyond current duration
