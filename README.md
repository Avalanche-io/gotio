# gotio

[![Go Reference](https://pkg.go.dev/badge/github.com/mrjoshuak/gotio.svg)](https://pkg.go.dev/github.com/mrjoshuak/gotio)
[![Go Report Card](https://goreportcard.com/badge/github.com/mrjoshuak/gotio)](https://goreportcard.com/report/github.com/mrjoshuak/gotio)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

A pure Go implementation of [OpenTimelineIO](http://opentimeline.io/) - the interchange format and API for editorial cut information.

## Overview

gotio is a native Go implementation of OpenTimelineIO (OTIO), providing full compatibility with the OTIO file format while offering idiomatic Go APIs. This library enables Go applications to read, write, and manipulate editorial timeline data without any C/C++ dependencies (no cgo required).

**Key Features:**
- **Pure Go** - No cgo, no external dependencies, cross-platform compatible
- **Full OTIO Compatibility** - Reads and writes standard `.otio` JSON files
- **Complete Type System** - Timeline, Track, Stack, Clip, Gap, Transition, and more
- **Time Manipulation** - RationalTime, TimeRange, and TimeTransform support
- **Algorithm Support** - Track trimming, stack flattening, composition filtering
- **90%+ Test Coverage** - Comprehensive test suite with functional parity validation

## Installation

```bash
go get github.com/mrjoshuak/gotio
```

## Quick Start

### Reading a Timeline

```go
package main

import (
    "fmt"
    "log"

    "github.com/mrjoshuak/gotio/opentimelineio"
)

func main() {
    // Load a timeline from an OTIO file
    obj, err := opentimelineio.FromJSONFile("project.otio")
    if err != nil {
        log.Fatal(err)
    }

    timeline := obj.(*opentimelineio.Timeline)
    fmt.Printf("Timeline: %s\n", timeline.Name())

    // Find all clips in the timeline
    for _, clip := range timeline.FindClips(nil, false) {
        dur, _ := clip.Duration()
        fmt.Printf("  Clip: %s [%.2f seconds]\n", clip.Name(), dur.ToSeconds())
    }
}
```

### Creating a Timeline

```go
package main

import (
    "github.com/mrjoshuak/gotio/opentime"
    "github.com/mrjoshuak/gotio/opentimelineio"
)

func main() {
    // Create a new timeline
    timeline := opentimelineio.NewTimeline("My Project", nil, nil)

    // Create a video track
    track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
    timeline.Tracks().AppendChild(track)

    // Create clips with media references
    for i, file := range []string{"shot_001.mov", "shot_002.mov", "shot_003.mov"} {
        // Define the available media range
        availableRange := opentime.NewTimeRange(
            opentime.NewRationalTime(0, 24),
            opentime.NewRationalTime(100, 24), // 100 frames at 24fps
        )

        // Create an external reference to the media file
        ref := opentimelineio.NewExternalReference(
            "",
            "file:///path/to/media/"+file,
            &availableRange,
            nil,
        )

        // Create a clip using frames 10-90 of the media
        sourceRange := opentime.NewTimeRange(
            opentime.NewRationalTime(10, 24),
            opentime.NewRationalTime(80, 24),
        )

        clip := opentimelineio.NewClip(
            fmt.Sprintf("Shot %d", i+1),
            ref,
            &sourceRange,
            nil, nil, nil, "", nil,
        )

        track.AppendChild(clip)
    }

    // Write to file
    opentimelineio.ToJSONFile(timeline, "output.otio", "  ")
}
```

### Working with Time

```go
package main

import (
    "fmt"

    "github.com/mrjoshuak/gotio/opentime"
)

func main() {
    // Create rational times
    frame100 := opentime.NewRationalTime(100, 24)  // Frame 100 at 24fps
    frame200 := opentime.NewRationalTime(200, 24)  // Frame 200 at 24fps

    // Convert to seconds
    fmt.Printf("Frame 100 = %.4f seconds\n", frame100.ToSeconds())

    // Create a time range
    duration := opentime.NewRationalTime(50, 24)
    timeRange := opentime.NewTimeRange(frame100, duration)

    fmt.Printf("Range: %v to %v (duration: %v)\n",
        timeRange.StartTime(),
        timeRange.EndTimeExclusive(),
        timeRange.Duration())

    // Check if a time is within the range
    testTime := opentime.NewRationalTime(120, 24)
    fmt.Printf("Frame 120 in range: %v\n", timeRange.Contains(testTime))

    // Convert between frame rates
    at30fps := frame100.RescaledTo(30)
    fmt.Printf("Frame 100 @24fps = Frame %.0f @30fps\n", at30fps.Value())
}
```

## Package Structure

```
github.com/mrjoshuak/gotio/
├── opentime/           # Time representation (RationalTime, TimeRange, TimeTransform)
├── opentimelineio/     # Core OTIO types (Timeline, Track, Stack, Clip, etc.)
└── algorithms/         # Timeline manipulation algorithms
```

### opentime

The `opentime` package provides types for representing time:

- **RationalTime** - A point in time represented as a rational number (value/rate)
- **TimeRange** - A range of time with start time and duration
- **TimeTransform** - A transformation that can be applied to times (offset, scale)

### opentimelineio

The `opentimelineio` package provides the core OTIO data model:

**Compositions:**
- **Timeline** - The root container with tracks and global start time
- **Stack** - Layers children on top of each other (like Photoshop layers)
- **Track** - Arranges children sequentially in time

**Items:**
- **Clip** - A segment of media with a reference and source range
- **Gap** - Empty space (transparent) in a composition
- **Transition** - A blend between adjacent items (dissolve, wipe, etc.)

**Media References:**
- **ExternalReference** - Points to external media (file path or URL)
- **MissingReference** - Placeholder for missing media
- **GeneratorReference** - Procedurally generated media (solid colors, etc.)
- **ImageSequenceReference** - Numbered image sequence

**Metadata:**
- **Marker** - Annotation attached to an item
- **Effect** - Visual/audio effect applied to an item
- **LinearTimeWarp** - Speed change effect
- **FreezeFrame** - Freeze frame effect

### algorithms

The `algorithms` package provides functions for manipulating timelines:

- **TrackTrimmedToRange** - Trim a track to a specific time range
- **FlattenStack** - Flatten a multi-track stack to a single track
- **FlattenTracks** - Flatten multiple tracks to a single track
- **TopClipAtTime** - Get the topmost visible clip at a given time
- **TimelineVideoTracks/AudioTracks** - Get video or audio tracks from a timeline
- **FilteredComposition** - Filter a composition by type or name

## Documentation

- [Quick Start Guide](docs/quickstart.md)
- [Architecture Overview](docs/architecture.md)
- [Timeline Structure](docs/timeline-structure.md)
- [Time and Ranges](docs/time-ranges.md)
- [API Reference](docs/api-reference.md)
- [Algorithms](docs/algorithms.md)
- [Examples](examples/)

## Compatibility

gotio is compatible with OpenTimelineIO files created by:
- OpenTimelineIO (Python/C++)
- DaVinci Resolve
- Adobe Premiere Pro
- Avid Media Composer
- Final Cut Pro (via adapters)
- Any application that exports standard OTIO JSON

### Legacy Support

gotio supports legacy OTIO files including:
- **Sequence schema** - Legacy name for Track (pre-1.0 OTIO files)
- **Non-standard JSON values** - Files with `Inf`, `-Infinity`, `NaN` values (Python-generated)

## Contributing

Contributions are welcome! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## License

gotio is licensed under the Apache License 2.0. See [LICENSE](LICENSE) for details.

## Acknowledgments

This project is a Go implementation inspired by [OpenTimelineIO](https://github.com/AcademySoftwareFoundation/OpenTimelineIO), originally developed by Pixar Animation Studios and now maintained by the Academy Software Foundation.
