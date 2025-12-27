# Quick Start Guide

This guide will help you get started with gotio, the Go implementation of OpenTimelineIO.

## Installation

Add gotio to your Go project:

```bash
go get github.com/mrjoshuak/gotio
```

Import the packages you need:

```go
import (
    "github.com/mrjoshuak/gotio/opentime"
    "github.com/mrjoshuak/gotio/opentimelineio"
    "github.com/mrjoshuak/gotio/algorithms"
)
```

## Reading OTIO Files

### Load a Timeline from a File

```go
obj, err := opentimelineio.FromJSONFile("my_project.otio")
if err != nil {
    log.Fatal(err)
}

// Type assert to Timeline
timeline, ok := obj.(*opentimelineio.Timeline)
if !ok {
    log.Fatal("Expected a Timeline")
}

fmt.Printf("Loaded timeline: %s\n", timeline.Name())
```

### Load from a JSON String

```go
jsonStr := `{
    "OTIO_SCHEMA": "Timeline.1",
    "name": "My Timeline",
    "tracks": {
        "OTIO_SCHEMA": "Stack.1",
        "children": []
    }
}`

obj, err := opentimelineio.FromJSONString(jsonStr)
if err != nil {
    log.Fatal(err)
}

timeline := obj.(*opentimelineio.Timeline)
```

## Exploring a Timeline

### Get Basic Information

```go
// Timeline name
fmt.Printf("Name: %s\n", timeline.Name())

// Global start time (if set)
if gst := timeline.GlobalStartTime(); gst != nil {
    fmt.Printf("Global Start: %v\n", gst)
}

// Get the tracks stack
tracks := timeline.Tracks()
fmt.Printf("Number of tracks: %d\n", len(tracks.Children()))
```

### Iterate Over Tracks

```go
for i, child := range timeline.Tracks().Children() {
    track, ok := child.(*opentimelineio.Track)
    if !ok {
        continue
    }

    dur, _ := track.Duration()
    fmt.Printf("Track %d: %s (%s) - %.2f seconds\n",
        i, track.Name(), track.Kind(), dur.ToSeconds())
}
```

### Find All Clips

```go
// Find all clips in the timeline (including nested compositions)
clips := timeline.FindClips(nil, false)

for _, clip := range clips {
    dur, _ := clip.Duration()
    fmt.Printf("Clip: %s (%.2f seconds)\n", clip.Name(), dur.ToSeconds())

    // Check media reference
    ref := clip.MediaReference()
    if ref != nil {
        if extRef, ok := ref.(*opentimelineio.ExternalReference); ok {
            fmt.Printf("  Media: %s\n", extRef.TargetURL())
        }
    }
}
```

### Get Video and Audio Tracks

```go
// Get only video tracks
videoTracks := timeline.VideoTracks()
fmt.Printf("Video tracks: %d\n", len(videoTracks))

// Get only audio tracks
audioTracks := timeline.AudioTracks()
fmt.Printf("Audio tracks: %d\n", len(audioTracks))
```

## Creating Timelines

### Basic Timeline Creation

```go
// Create a new timeline
timeline := opentimelineio.NewTimeline("My Project", nil, nil)

// Create a video track
videoTrack := opentimelineio.NewTrack(
    "V1",                              // name
    nil,                               // source_range (nil = use full range)
    opentimelineio.TrackKindVideo,     // kind
    nil,                               // metadata
    nil,                               // color
)

// Add the track to the timeline
timeline.Tracks().AppendChild(videoTrack)
```

### Creating Clips

```go
// Define the source range (which portion of the media to use)
sourceRange := opentime.NewTimeRange(
    opentime.NewRationalTime(0, 24),    // start at frame 0
    opentime.NewRationalTime(100, 24),  // duration of 100 frames
)

// Create a clip with a media reference
clip := opentimelineio.NewClip(
    "My Clip",                          // name
    nil,                                // media reference (set separately)
    &sourceRange,                       // source range
    nil,                                // metadata
    nil,                                // effects
    nil,                                // markers
    "",                                 // active media reference key
    nil,                                // color
)

// Set the media reference
availableRange := opentime.NewTimeRange(
    opentime.NewRationalTime(0, 24),
    opentime.NewRationalTime(500, 24),
)
ref := opentimelineio.NewExternalReference(
    "",                                 // name
    "file:///path/to/video.mov",        // target URL
    &availableRange,                    // available range
    nil,                                // metadata
)
clip.SetMediaReference(ref)

// Add clip to track
videoTrack.AppendChild(clip)
```

### Creating Gaps

```go
// Create a gap (empty space)
gapDuration := opentime.NewRationalTime(24, 24)  // 1 second at 24fps
gap := opentimelineio.NewGapWithDuration(gapDuration)

// Gaps can be named and have metadata too
gap.SetName("Scene Break")

videoTrack.AppendChild(gap)
```

### Creating Transitions

```go
// Create a cross-dissolve transition
inOffset := opentime.NewRationalTime(12, 24)   // 12 frames
outOffset := opentime.NewRationalTime(12, 24)  // 12 frames

transition := opentimelineio.NewTransition(
    "Dissolve",
    opentimelineio.TransitionTypeSMPTEDissolve,
    inOffset,
    outOffset,
    nil,
)

// Insert between two clips (transitions go between items)
// Track: [Clip1] [Transition] [Clip2]
videoTrack.InsertChild(1, transition)
```

## Working with Time

### RationalTime

```go
// Create times at different frame rates
frame24 := opentime.NewRationalTime(100, 24)   // Frame 100 at 24fps
frame30 := opentime.NewRationalTime(100, 30)   // Frame 100 at 30fps

// Convert to seconds
seconds := frame24.ToSeconds()  // 4.166666...

// Convert between frame rates
frame24at30 := frame24.RescaledTo(30)  // Frame 125 at 30fps

// Arithmetic
sum := frame24.Add(frame30.RescaledTo(24))
diff := frame24.Sub(opentime.NewRationalTime(50, 24))

// Comparisons
if frame24.Equal(frame30.RescaledTo(24)) {
    fmt.Println("Times are equal")
}
```

### TimeRange

```go
// Create a time range
start := opentime.NewRationalTime(100, 24)
duration := opentime.NewRationalTime(50, 24)
timeRange := opentime.NewTimeRange(start, duration)

// Query the range
fmt.Printf("Start: %v\n", timeRange.StartTime())
fmt.Printf("Duration: %v\n", timeRange.Duration())
fmt.Printf("End (exclusive): %v\n", timeRange.EndTimeExclusive())
fmt.Printf("End (inclusive): %v\n", timeRange.EndTimeInclusive())

// Check containment
testTime := opentime.NewRationalTime(120, 24)
if timeRange.Contains(testTime) {
    fmt.Println("Time is within range")
}

// Check overlap
otherRange := opentime.NewTimeRange(
    opentime.NewRationalTime(140, 24),
    opentime.NewRationalTime(20, 24),
)
if timeRange.Overlaps(otherRange) {
    fmt.Println("Ranges overlap")
}
```

### TimeTransform

```go
// Create a transform (offset + scale)
transform := opentime.NewTimeTransform(
    opentime.NewRationalTime(10, 24),  // offset
    2.0,                                // rate (2x speed)
    24,
)

// Apply to a time
original := opentime.NewRationalTime(100, 24)
transformed := transform.AppliedToTime(original)

// Apply to a range
originalRange := opentime.NewTimeRange(
    opentime.NewRationalTime(0, 24),
    opentime.NewRationalTime(100, 24),
)
transformedRange := transform.AppliedToTimeRange(originalRange)
```

## Writing OTIO Files

### Write to File

```go
// Write with indentation (human-readable)
err := opentimelineio.ToJSONFile(timeline, "output.otio", "  ")
if err != nil {
    log.Fatal(err)
}

// Write without indentation (compact)
err = opentimelineio.ToJSONFile(timeline, "output.otio", "")
```

### Convert to JSON String

```go
// Get JSON string
jsonStr, err := opentimelineio.ToJSONString(timeline)
if err != nil {
    log.Fatal(err)
}

// Get formatted JSON bytes
jsonBytes, err := opentimelineio.ToJSONBytesIndent(timeline, "  ")
```

## Adding Metadata

All OTIO objects support arbitrary metadata:

```go
// Set metadata on creation
metadata := opentimelineio.AnyDictionary{
    "author":     "Jane Doe",
    "project_id": 12345,
    "tags":       []string{"vfx", "final"},
}
timeline := opentimelineio.NewTimeline("Project", nil, metadata)

// Access metadata
author := timeline.Metadata()["author"]

// Modify metadata
timeline.Metadata()["version"] = "2.0"
```

## Adding Markers

```go
// Create a marker
markedRange := opentime.NewTimeRange(
    opentime.NewRationalTime(100, 24),
    opentime.NewRationalTime(1, 24),  // 0 duration for point marker
)

marker := opentimelineio.NewMarker(
    "VFX Note",
    markedRange,
    opentimelineio.MarkerColorRed,
    "Add explosion here",
    nil,
)

// Add marker to a clip
clip.SetMarkers(append(clip.Markers(), marker))
```

## Adding Effects

```go
// Create a basic effect
effect := opentimelineio.NewEffect(
    "color_correction",
    "Color Correction",
    opentimelineio.AnyDictionary{
        "brightness": 1.2,
        "contrast":   1.1,
    },
)
clip.SetEffects(append(clip.Effects(), effect))

// Create a speed change effect (2x speed)
speedEffect := opentimelineio.NewLinearTimeWarp(
    "speed_up",
    "LinearTimeWarp",
    2.0,  // time scalar
    nil,
)
clip.SetEffects(append(clip.Effects(), speedEffect))

// Create a freeze frame effect
freeze := opentimelineio.NewFreezeFrame("freeze", nil)
clip.SetEffects(append(clip.Effects(), freeze))
```

## Next Steps

- [Architecture Overview](architecture.md) - Understand the library structure
- [Timeline Structure](timeline-structure.md) - Deep dive into the data model
- [Time and Ranges](time-ranges.md) - Detailed time manipulation guide
- [API Reference](api-reference.md) - Complete API documentation
- [Algorithms](algorithms.md) - Timeline manipulation algorithms
- [Examples](../examples/) - Working code examples
