# Timeline Structure

An OpenTimelineIO `Timeline` object can contain many tracks, nested stacks, clips, gaps, and transitions. This document clarifies how these objects nest within each other and how they work together to represent an audio/video timeline.

## Rendering Model

### Video Tracks

Rendering of video tracks in a timeline uses **painter's order**. Layers in a stack are composited from bottom (first entry) to top (last entry). Higher layers overlay lower layers using alpha compositing over a transparent background.

Within a track, clips may overlap via transitions. During a transition, the contribution is a blend of the elements being transitioned between.

### Audio Tracks

Rendering of audio tracks is **additive**. All audio from all tracks is summed together. Applications should process the summed audio through a limiter to prevent clipping.

## Simple Cut List

Let's start with a simple cut list of a few clips. This is stored as a single `Timeline` with a single `Track` containing several `Clip` children, spliced end-to-end.

```
Timeline: "My Edit"
└── tracks: Stack
    └── Track: "V1"
        ├── Clip: "Shot_001" [frames 0-100]
        │   └── ExternalReference: "shot_001.mov"
        ├── Clip: "Shot_002" [frames 0-150]
        │   └── ExternalReference: "shot_002.mov"
        └── Clip: "Shot_003" [frames 0-75]
            └── ExternalReference: "shot_003.mov"
```

**Key Concepts:**

1. A `Timeline` always has a top-level `Stack` to hold its tracks
2. Each `Clip` has a `source_range` specifying which portion of the media to use
3. Each media reference has an `available_range` indicating the full extent of available media
4. Clips in a track are arranged sequentially - each starts where the previous one ends

### Code Example

```go
// Create a simple cut list
timeline := opentimelineio.NewTimeline("My Edit", nil, nil)
track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
timeline.Tracks().AppendChild(track)

// Add clips
shots := []struct {
    name     string
    file     string
    duration float64
}{
    {"Shot_001", "shot_001.mov", 100},
    {"Shot_002", "shot_002.mov", 150},
    {"Shot_003", "shot_003.mov", 75},
}

for _, shot := range shots {
    sourceRange := opentime.NewTimeRange(
        opentime.NewRationalTime(0, 24),
        opentime.NewRationalTime(shot.duration, 24),
    )
    ref := opentimelineio.NewExternalReference("", shot.file, &sourceRange, nil)
    clip := opentimelineio.NewClip(shot.name, ref, &sourceRange, nil, nil, nil, "", nil)
    track.AppendChild(clip)
}

// Query track duration (sum of all clips)
dur, _ := track.Duration()
fmt.Printf("Track duration: %.2f seconds\n", dur.ToSeconds())  // 13.54 seconds
```

## Clip Timing

Each `Clip` has several time-related concepts:

### available_range

The range of media that exists in the referenced file. For example, if a video file contains frames 0-500, the available_range would span 500 frames.

```go
ref := clip.MediaReference()
if extRef, ok := ref.(*opentimelineio.ExternalReference); ok {
    if ar := extRef.AvailableRange(); ar != nil {
        fmt.Printf("Available: frames %d to %d\n",
            int(ar.StartTime().Value()),
            int(ar.EndTimeExclusive().Value()))
    }
}
```

### source_range

The portion of the available media that is actually used in the edit. This trims the clip to only show a subset of the available media.

```go
if sr := clip.SourceRange(); sr != nil {
    fmt.Printf("Using: frames %d to %d\n",
        int(sr.StartTime().Value()),
        int(sr.EndTimeExclusive().Value()))
}
```

### trimmed_range

The effective range after applying source_range. If source_range is nil, this equals available_range.

```go
tr, _ := clip.TrimmedRange()
fmt.Printf("Trimmed range: %v\n", tr)
```

### range_in_parent

The position of the clip within its parent track.

```go
rip, _ := clip.RangeInParent()
fmt.Printf("In track: starts at %v, duration %v\n",
    rip.StartTime(), rip.Duration())
```

## Transitions

A `Transition` blends two adjacent items on the same track. The most common case is a cross-dissolve between two clips.

```
Track: "V1"
├── Clip: "Shot_001" [frames 0-100]
├── Transition: "Dissolve" [in: 12 frames, out: 12 frames]
└── Clip: "Shot_002" [frames 0-150]
```

**Key Points:**

1. Transitions don't add or remove time from the track
2. `in_offset` specifies how many frames of the incoming clip are used
3. `out_offset` specifies how many frames of the outgoing clip are used
4. The transition "borrows" frames from adjacent clips

### Code Example

```go
// Add a transition between clips
inOffset := opentime.NewRationalTime(12, 24)   // 12 frames (0.5 sec)
outOffset := opentime.NewRationalTime(12, 24)  // 12 frames (0.5 sec)

transition := opentimelineio.NewTransition(
    "Dissolve",
    opentimelineio.TransitionTypeSMPTEDissolve,
    inOffset,
    outOffset,
    nil,
)

// Insert between index 0 and 1
track.InsertChild(1, transition)
// Track now: [Clip] [Transition] [Clip]
```

### Transition Properties

```go
// Transitions are not "visible" (don't occupy time slots)
fmt.Printf("Visible: %v\n", transition.Visible())  // false

// But they are "overlapping"
fmt.Printf("Overlapping: %v\n", transition.Overlapping())  // true

// Duration is sum of offsets
dur, _ := transition.Duration()
fmt.Printf("Duration: %v\n", dur)  // 24 frames (1 second)
```

## Gaps

A `Gap` represents empty space in a track. Gaps are transparent, so lower tracks show through.

```
Stack
├── Track: "V1"
│   ├── Clip: "Overlay" [frames 0-50]
│   └── Gap [50 frames]    ← transparent, V2 shows through
└── Track: "V2"
    └── Clip: "Background" [frames 0-100]
```

### Code Example

```go
// Create a gap
gapDuration := opentime.NewRationalTime(50, 24)
gap := opentimelineio.NewGapWithDuration(gapDuration)
track.AppendChild(gap)

// Gaps can have metadata
gap.Metadata()["reason"] = "scene break"
```

## Multiple Tracks (Stacks)

A `Stack` layers its children vertically. In a typical timeline, the `tracks` field is a `Stack` containing multiple `Track` objects.

```
Timeline: "Multi-track Edit"
└── tracks: Stack
    ├── Track: "V1" (bottom, renders first)
    │   └── Clip: "Background"
    ├── Track: "V2"
    │   ├── Gap [offset]
    │   └── Clip: "Foreground"
    └── Track: "V3" (top, renders last)
        └── Clip: "Title"
```

**Compositing Order:**
1. V1 (Background) renders first
2. V2 (Foreground) composites over V1
3. V3 (Title) composites over V2

### Code Example

```go
timeline := opentimelineio.NewTimeline("Multi-track Edit", nil, nil)

// Create video tracks
v1 := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
v2 := opentimelineio.NewTrack("V2", nil, opentimelineio.TrackKindVideo, nil, nil)
v3 := opentimelineio.NewTrack("V3", nil, opentimelineio.TrackKindVideo, nil, nil)

// Add tracks (order matters: first = bottom, last = top)
timeline.Tracks().AppendChild(v1)
timeline.Tracks().AppendChild(v2)
timeline.Tracks().AppendChild(v3)

// V2 starts later - use a gap for offset
gap := opentimelineio.NewGapWithDuration(opentime.NewRationalTime(24, 24))
v2.AppendChild(gap)
v2.AppendChild(foregroundClip)
```

### Stack vs Track Duration

- **Track Duration**: Sum of all children's durations
- **Stack Duration**: Maximum of all children's durations

```go
// Stack duration is the longest track
stackDur, _ := timeline.Tracks().Duration()

// Individual track durations
for _, child := range timeline.Tracks().Children() {
    if track, ok := child.(*opentimelineio.Track); ok {
        dur, _ := track.Duration()
        fmt.Printf("%s: %.2f seconds\n", track.Name(), dur.ToSeconds())
    }
}
```

## Nested Compositions

Compositions can be nested - a `Track` can contain a `Stack`, and a `Stack` can contain `Track`s.

```
Timeline
└── tracks: Stack
    └── Track: "V1"
        ├── Clip: "Intro"
        ├── Stack: "VFX Shot"    ← nested composition
        │   ├── Track: "Plate"
        │   │   └── Clip: "background.exr"
        │   └── Track: "CG"
        │       └── Clip: "character.exr"
        └── Clip: "Outro"
```

**Key Points:**

1. Nested compositions behave like clips in the parent track
2. Setting `source_range` on a nested composition trims it
3. The nested composition's internal timing is independent

### Code Example

```go
// Create a nested stack
nestedStack := opentimelineio.NewStack("VFX Shot", nil, nil, nil, nil, nil)

plateTrack := opentimelineio.NewTrack("Plate", nil, opentimelineio.TrackKindVideo, nil, nil)
cgTrack := opentimelineio.NewTrack("CG", nil, opentimelineio.TrackKindVideo, nil, nil)

nestedStack.AppendChild(plateTrack)
nestedStack.AppendChild(cgTrack)

// Add clips to nested tracks
plateTrack.AppendChild(plateClip)
cgTrack.AppendChild(cgClip)

// Add nested stack to main track (acts like a single item)
mainTrack.AppendChild(nestedStack)

// Optionally trim the nested composition
trimRange := opentime.NewTimeRange(
    opentime.NewRationalTime(10, 24),
    opentime.NewRationalTime(50, 24),
)
nestedStack.SetSourceRange(&trimRange)
```

## Coordinate Space Transformations

Each item in an OTIO timeline has its own local coordinate space. The `TransformedTime` and `TransformedTimeRange` methods convert times between different coordinate spaces in the hierarchy.

### Understanding Coordinate Spaces

Consider a clip with a `source_range` of frames 100-200 placed in a track:

```
Track: "V1" (track coordinate space: 0-100)
└── Clip: "Shot_001" (clip coordinate space: 100-200)
    └── source_range: [100, 100] (start at frame 100, duration 100 frames)
```

- **Clip's coordinate space**: Frame 100 is the start of the clip's media
- **Track's coordinate space**: Frame 0 is where the clip begins in the track

### TransformedTime

Converts a time from one item's coordinate space to another's:

```go
// Create a track with a clip that has an offset source range
track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)

sourceRange := opentime.NewTimeRange(
    opentime.NewRationalTime(100, 24),  // Start at frame 100
    opentime.NewRationalTime(100, 24),  // Duration: 100 frames
)
clip := opentimelineio.NewClip("Shot_001", mediaRef, &sourceRange, nil, nil, nil, "", nil)
track.AppendChild(clip)

// Transform time from clip to track coordinate space
clipTime := opentime.NewRationalTime(150, 24)  // Frame 150 in clip's space
trackTime, err := clip.TransformedTime(clipTime, track)
// trackTime = frame 50 (150 - 100 source offset = 50 frames into track)

// Transform time from track to clip coordinate space
trackTime = opentime.NewRationalTime(25, 24)  // Frame 25 in track
clipTime, err = track.TransformedTime(trackTime, clip)
// clipTime = frame 125 (25 + 100 source offset = 125 in clip's media)
```

### Transformations Through Nested Hierarchies

TransformedTime works through any depth of nesting:

```
Timeline
└── tracks: Stack
    └── Track: "V1"
        └── Stack: "Nested"
            └── Track: "Inner"
                └── Clip: "Deep"
```

```go
// Transform from deeply nested clip to outer track
deepClipTime := opentime.NewRationalTime(50, 24)
outerTrackTime, err := deepClip.TransformedTime(deepClipTime, outerTrack)
```

The algorithm:
1. Walks UP from the source item to the common ancestor
2. At each level: subtracts the item's trimmed_range start, adds the parent's range_of_child start
3. Walks DOWN from the common ancestor to the target item
4. At each level: subtracts the parent's range_of_child start, adds the item's trimmed_range start

### TransformedTimeRange

Transforms a time range, preserving duration:

```go
// Transform a range from clip to track space
clipRange := opentime.NewTimeRange(
    opentime.NewRationalTime(120, 24),  // Start at frame 120 in clip
    opentime.NewRationalTime(30, 24),   // Duration: 30 frames
)
trackRange, err := clip.TransformedTimeRange(clipRange, track)
// trackRange starts at frame 20, duration 30 (preserved)
```

### Sibling Item Transformations

Transform times between clips in the same track:

```go
track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)

// Clip 1: source range 0-100, at track position 0-100
sr1 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
clip1 := opentimelineio.NewClip("Clip1", ref1, &sr1, nil, nil, nil, "", nil)

// Clip 2: source range 50-150, at track position 100-200
sr2 := opentime.NewTimeRange(opentime.NewRationalTime(50, 24), opentime.NewRationalTime(100, 24))
clip2 := opentimelineio.NewClip("Clip2", ref2, &sr2, nil, nil, nil, "", nil)

track.AppendChild(clip1)
track.AppendChild(clip2)

// Transform frame 80 in clip1 to clip2's coordinate space
clip1Time := opentime.NewRationalTime(80, 24)
clip2Time, err := clip1.TransformedTime(clip1Time, clip2)
// Calculation: clip1(80) -> track(80) -> clip2(80 - 100 + 50 = 30)
```

### Use Cases

**1. Syncing markers across items:**
```go
// Copy a marker from one clip to another, adjusting position
markerTime := marker.MarkedRange().StartTime()
newTime, _ := sourceClip.TransformedTime(markerTime, targetClip)
newRange := opentime.NewTimeRange(newTime, marker.MarkedRange().Duration())
newMarker := opentimelineio.NewMarker(marker.Name(), newRange, marker.Color(), marker.Comment(), nil)
```

**2. Finding corresponding frames in different clips:**
```go
// What frame in the B-roll matches frame 100 in the A-roll?
aRollTime := opentime.NewRationalTime(100, 24)
bRollTime, _ := aRollClip.TransformedTime(aRollTime, bRollClip)
```

**3. Converting playhead position to clip coordinates:**
```go
// User clicks at timeline position 500
playheadTime := opentime.NewRationalTime(500, 24)
clipUnderPlayhead := track.ChildAtTime(playheadTime, true)
if clip, ok := clipUnderPlayhead.(*opentimelineio.Clip); ok {
    clipTime, _ := track.TransformedTime(playheadTime, clip)
    fmt.Printf("Frame %v in clip %s\n", clipTime.Value(), clip.Name())
}
```

## Markers

`Marker` objects can be attached to any `Item` (Clip, Track, Stack, Gap, etc.).

```go
// Create a marker on a clip
markedRange := opentime.NewTimeRange(
    opentime.NewRationalTime(50, 24),  // position within clip
    opentime.NewRationalTime(1, 24),   // duration (0 for point marker)
)

marker := opentimelineio.NewMarker(
    "Fix Color",
    markedRange,
    opentimelineio.MarkerColorYellow,
    "Color is too saturated here",
    nil,
)

clip.SetMarkers(append(clip.Markers(), marker))

// Query markers
for _, m := range clip.Markers() {
    fmt.Printf("Marker: %s at %v (%s)\n",
        m.Name(), m.MarkedRange().StartTime(), m.Comment())
}
```

### Marker Colors

```go
opentimelineio.MarkerColorPink
opentimelineio.MarkerColorRed
opentimelineio.MarkerColorOrange
opentimelineio.MarkerColorYellow
opentimelineio.MarkerColorGreen
opentimelineio.MarkerColorCyan
opentimelineio.MarkerColorBlue
opentimelineio.MarkerColorPurple
opentimelineio.MarkerColorMagenta
opentimelineio.MarkerColorBlack
opentimelineio.MarkerColorWhite
```

## Effects

`Effect` objects modify the appearance or timing of items.

### Basic Effects

```go
effect := opentimelineio.NewEffect(
    "blur",
    "GaussianBlur",
    opentimelineio.AnyDictionary{
        "radius": 5.0,
    },
)
clip.SetEffects(append(clip.Effects(), effect))
```

### Time Effects

```go
// Speed up 2x
speedUp := opentimelineio.NewLinearTimeWarp("fast", "LinearTimeWarp", 2.0, nil)
clip.SetEffects(append(clip.Effects(), speedUp))

// Slow down 0.5x
slowMo := opentimelineio.NewLinearTimeWarp("slow", "LinearTimeWarp", 0.5, nil)
clip.SetEffects(append(clip.Effects(), slowMo))

// Freeze frame (time_scalar = 0)
freeze := opentimelineio.NewFreezeFrame("freeze", nil)
clip.SetEffects(append(clip.Effects(), freeze))
```

## Complete Example

Here's a complete example creating a multi-track timeline:

```go
package main

import (
    "fmt"
    "log"

    "github.com/mrjoshuak/gotio/opentime"
    "github.com/mrjoshuak/gotio/opentimelineio"
)

func main() {
    // Create timeline
    timeline := opentimelineio.NewTimeline("Complete Example", nil, nil)

    // Create video track
    videoTrack := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
    timeline.Tracks().AppendChild(videoTrack)

    // Create audio track
    audioTrack := opentimelineio.NewTrack("A1", nil, opentimelineio.TrackKindAudio, nil, nil)
    timeline.Tracks().AppendChild(audioTrack)

    // Add clips to video track
    clips := []struct {
        name string
        file string
        dur  float64
    }{
        {"Opening", "opening.mov", 48},
        {"Scene 1", "scene1.mov", 240},
        {"Scene 2", "scene2.mov", 180},
        {"Closing", "closing.mov", 72},
    }

    for i, c := range clips {
        sr := opentime.NewTimeRange(
            opentime.NewRationalTime(0, 24),
            opentime.NewRationalTime(c.dur, 24),
        )
        ref := opentimelineio.NewExternalReference("", c.file, &sr, nil)
        clip := opentimelineio.NewClip(c.name, ref, &sr, nil, nil, nil, "", nil)

        // Add marker to first clip
        if i == 0 {
            markerRange := opentime.NewTimeRange(
                opentime.NewRationalTime(24, 24),
                opentime.NewRationalTime(0, 24),
            )
            marker := opentimelineio.NewMarker(
                "Title Card",
                markerRange,
                opentimelineio.MarkerColorGreen,
                "Add title overlay",
                nil,
            )
            clip.SetMarkers(append(clip.Markers(), marker))
        }

        videoTrack.AppendChild(clip)

        // Add transitions between clips (except after last)
        if i < len(clips)-1 {
            trans := opentimelineio.NewTransition(
                "Dissolve",
                opentimelineio.TransitionTypeSMPTEDissolve,
                opentime.NewRationalTime(12, 24),
                opentime.NewRationalTime(12, 24),
                nil,
            )
            videoTrack.AppendChild(trans)
        }
    }

    // Print timeline info
    fmt.Printf("Timeline: %s\n", timeline.Name())
    fmt.Printf("Video tracks: %d\n", len(timeline.VideoTracks()))
    fmt.Printf("Audio tracks: %d\n", len(timeline.AudioTracks()))

    dur, _ := timeline.Duration()
    fmt.Printf("Duration: %.2f seconds\n", dur.ToSeconds())

    allClips := timeline.FindClips(nil, false)
    fmt.Printf("Total clips: %d\n", len(allClips))

    // Write to file
    err := opentimelineio.ToJSONFile(timeline, "output.otio", "  ")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Saved to output.otio")
}
```
