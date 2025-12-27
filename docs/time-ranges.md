# Time and Ranges

The `opentime` package provides types for representing time in media applications. This document explains how to work with `RationalTime`, `TimeRange`, and `TimeTransform`.

## Overview

Media applications need precise time representation because:

1. **Frame Accuracy** - Video is frame-based, requiring exact frame counts
2. **Rate Conversion** - Different formats use different frame rates (24, 25, 29.97, 30, etc.)
3. **No Floating-Point Errors** - Rational numbers avoid accumulating errors

The `opentime` package addresses these needs with three core types:

| Type | Purpose |
|------|---------|
| `RationalTime` | A point in time as value/rate |
| `TimeRange` | A range with start time and duration |
| `TimeTransform` | A transformation (offset + scale) |

## RationalTime

`RationalTime` represents a point in time as a rational number.

### Creating RationalTime

```go
import "github.com/mrjoshuak/gotio/opentime"

// Frame 100 at 24 fps
frame := opentime.NewRationalTime(100, 24)

// 2.5 seconds at 24 fps
seconds := opentime.NewRationalTimeFromSeconds(2.5, 24)

// From timecode
tc, err := opentime.FromTimecode("01:00:10:15", 24)  // 1 hour, 10 seconds, 15 frames
```

### Accessing Values

```go
t := opentime.NewRationalTime(100, 24)

// Get value and rate
fmt.Printf("Value: %.0f\n", t.Value())  // 100
fmt.Printf("Rate: %.0f\n", t.Rate())    // 24

// Convert to seconds
fmt.Printf("Seconds: %.4f\n", t.ToSeconds())  // 4.1667

// Convert to timecode
tc, _ := t.ToTimecode(24, opentime.ForceNo)
fmt.Printf("Timecode: %s\n", tc)  // "00:00:04:04"
```

### Arithmetic Operations

All operations return new values (immutable):

```go
a := opentime.NewRationalTime(100, 24)
b := opentime.NewRationalTime(50, 24)

// Addition
sum := a.Add(b)  // 150 frames

// Subtraction
diff := a.Sub(b)  // 50 frames

// Negation
neg := a.Neg()  // -100 frames

// Absolute value
abs := neg.Abs()  // 100 frames
```

### Rate Conversion

```go
t24 := opentime.NewRationalTime(100, 24)  // 100 frames at 24 fps

// Convert to 30 fps
t30 := t24.RescaledTo(30)  // 125 frames at 30 fps

// Both represent the same point in time
fmt.Printf("24fps: %.4f seconds\n", t24.ToSeconds())  // 4.1667
fmt.Printf("30fps: %.4f seconds\n", t30.ToSeconds())  // 4.1667
```

### Comparisons

```go
a := opentime.NewRationalTime(100, 24)
b := opentime.NewRationalTime(125, 30)  // Same time, different rate
c := opentime.NewRationalTime(100, 30)

// Equality (compares actual time, not just values)
fmt.Println(a.Equal(b))  // true (same time)
fmt.Println(a.Equal(c))  // false (different times)

// Comparisons
fmt.Println(a.Compare(c))  // 1 (a > c)
fmt.Println(c.Compare(a))  // -1 (c < a)
```

### Timecode Conversion

```go
// From timecode string
t, err := opentime.FromTimecode("01:23:45:12", 24)
if err != nil {
    log.Fatal(err)
}

// To timecode string (non-drop-frame)
tc, err := t.ToTimecode(24, opentime.ForceNo)
fmt.Println(tc)  // "01:23:45:12"

// To timecode string (drop-frame for 29.97 fps)
t2997 := opentime.NewRationalTime(107892, 29.97)
tcDF, err := t2997.ToTimecode(29.97, opentime.ForceYes)
fmt.Println(tcDF)  // "01:00:00;00" (note semicolon for drop-frame)
```

### Common Patterns

```go
// Zero time
zero := opentime.RationalTime{}
// or
zero = opentime.NewRationalTime(0, 24)

// Check if zero
if t.Value() == 0 {
    // is zero
}

// Duration in frames
frames := int(duration.Value())

// Duration in whole seconds
wholeSeconds := int(duration.ToSeconds())
```

## TimeRange

`TimeRange` represents a span of time with a start time and duration.

### Creating TimeRange

```go
start := opentime.NewRationalTime(100, 24)
duration := opentime.NewRationalTime(50, 24)

// Create range
tr := opentime.NewTimeRange(start, duration)

// Create from start and end times
startEnd := opentime.NewTimeRangeFromStartEndTime(
    opentime.NewRationalTime(100, 24),
    opentime.NewRationalTime(150, 24),
)

// Create from start and end (exclusive)
startEndExcl := opentime.NewTimeRangeFromStartEndTimeExclusive(
    opentime.NewRationalTime(100, 24),
    opentime.NewRationalTime(150, 24),  // exclusive end
)
```

### Accessing Components

```go
tr := opentime.NewTimeRange(
    opentime.NewRationalTime(100, 24),
    opentime.NewRationalTime(50, 24),
)

// Start time
start := tr.StartTime()  // frame 100

// Duration
dur := tr.Duration()  // 50 frames

// End times
endExcl := tr.EndTimeExclusive()  // frame 150 (first frame AFTER range)
endIncl := tr.EndTimeInclusive()  // frame 149 (last frame IN range)
```

### Containment Tests

```go
tr := opentime.NewTimeRange(
    opentime.NewRationalTime(100, 24),
    opentime.NewRationalTime(50, 24),
)

// Check if a time is within the range
t1 := opentime.NewRationalTime(120, 24)
t2 := opentime.NewRationalTime(200, 24)

fmt.Println(tr.Contains(t1))  // true
fmt.Println(tr.Contains(t2))  // false

// Check boundary behavior
boundary := opentime.NewRationalTime(150, 24)
fmt.Println(tr.Contains(boundary))  // false (150 is exclusive end)
```

### Range Operations

```go
range1 := opentime.NewTimeRange(
    opentime.NewRationalTime(100, 24),
    opentime.NewRationalTime(50, 24),
)  // 100-150

range2 := opentime.NewTimeRange(
    opentime.NewRationalTime(125, 24),
    opentime.NewRationalTime(50, 24),
)  // 125-175

// Check overlap
overlaps := range1.Overlaps(range2)  // true

// Get intersection (clamped range)
clamped := range1.Clamped(range2)  // 125-150

// Extended range (union)
extended := range1.Extended(range2)  // 100-175
```

### Time Within Range

```go
tr := opentime.NewTimeRange(
    opentime.NewRationalTime(100, 24),
    opentime.NewRationalTime(50, 24),
)

// Clamp a time to within the range
t := opentime.NewRationalTime(200, 24)
clamped := tr.ClampedTime(t)  // frame 149 (last valid frame)

// Get time before/after range
before := tr.Before(opentime.NewRationalTime(50, 24), opentime.DefaultEpsilon)  // true
after := tr.After(opentime.NewRationalTime(200, 24), opentime.DefaultEpsilon)   // true
```

### Range Comparisons

```go
range1 := opentime.NewTimeRange(
    opentime.NewRationalTime(100, 24),
    opentime.NewRationalTime(50, 24),
)

range2 := opentime.NewTimeRange(
    opentime.NewRationalTime(100, 24),
    opentime.NewRationalTime(50, 24),
)

// Equality
fmt.Println(range1.Equal(range2))  // true

// Contains another range
outer := opentime.NewTimeRange(
    opentime.NewRationalTime(0, 24),
    opentime.NewRationalTime(200, 24),
)
fmt.Println(outer.ContainsRange(range1))  // true
```

## TimeTransform

`TimeTransform` represents a linear transformation that can be applied to times.

### Creating TimeTransform

```go
// Transform with offset and rate
transform := opentime.NewTimeTransform(
    opentime.NewRationalTime(10, 24),  // offset: shift by 10 frames
    2.0,                                // rate: 2x speed
    24,                                 // result rate
)

// Identity transform
identity := opentime.NewTimeTransformWithRate(24)
```

### Applying Transforms

```go
transform := opentime.NewTimeTransform(
    opentime.NewRationalTime(10, 24),  // offset
    2.0,                                // rate (2x speed)
    24,
)

// Apply to a single time
original := opentime.NewRationalTime(100, 24)
transformed := transform.AppliedToTime(original)
// Result: (100 + 10) * 2 = 220 frames

// Apply to a range
originalRange := opentime.NewTimeRange(
    opentime.NewRationalTime(0, 24),
    opentime.NewRationalTime(100, 24),
)
transformedRange := transform.AppliedToTimeRange(originalRange)
// Start: 10 * 2 = 20, Duration: 100 * 2 = 200
```

### Composing Transforms

```go
// First transform: shift by 10
t1 := opentime.NewTimeTransform(
    opentime.NewRationalTime(10, 24),
    1.0,
    24,
)

// Second transform: 2x speed
t2 := opentime.NewTimeTransform(
    opentime.NewRationalTime(0, 24),
    2.0,
    24,
)

// Compose: apply t1 then t2
composed := t1.AppliedTo(t2)

// Apply composed transform
original := opentime.NewRationalTime(100, 24)
result := composed.AppliedToTime(original)  // (100 + 10) * 2 = 220
```

## Working with Clips

Understanding how time ranges apply to clips:

### Clip Time Relationships

```go
clip := timeline.FindClips(nil, false)[0]

// available_range: full extent of the media file
if ref := clip.MediaReference(); ref != nil {
    if extRef, ok := ref.(*opentimelineio.ExternalReference); ok {
        if ar := extRef.AvailableRange(); ar != nil {
            fmt.Printf("Available: %v - %v\n",
                ar.StartTime(), ar.EndTimeExclusive())
        }
    }
}

// source_range: portion used in the edit
if sr := clip.SourceRange(); sr != nil {
    fmt.Printf("Source: %v - %v\n",
        sr.StartTime(), sr.EndTimeExclusive())
}

// trimmed_range: effective range (source_range if set, else available_range)
tr, _ := clip.TrimmedRange()
fmt.Printf("Trimmed: %v - %v\n", tr.StartTime(), tr.EndTimeExclusive())

// range_in_parent: position within parent track
rip, _ := clip.RangeInParent()
fmt.Printf("In track: %v - %v\n", rip.StartTime(), rip.EndTimeExclusive())
```

### Duration Calculations

```go
// Clip duration
clipDur, _ := clip.Duration()

// Track duration (sum of visible children)
trackDur, _ := track.Duration()

// Stack duration (max of children)
stackDur, _ := stack.Duration()

// Timeline duration
timelineDur, _ := timeline.Duration()
```

## Track Time Methods

```go
track := timeline.VideoTracks()[0]

// Total duration of track
available, _ := track.AvailableRange()
fmt.Printf("Track range: %v\n", available)

// Range of specific child
childRange, _ := track.RangeOfChildAtIndex(0)
fmt.Printf("Child 0 range: %v\n", childRange)

// Range of all children (map)
allRanges, _ := track.RangeOfAllChildren()
for child, r := range allRanges {
    if clip, ok := child.(*opentimelineio.Clip); ok {
        fmt.Printf("%s: %v\n", clip.Name(), r)
    }
}
```

## Common Patterns

### Frame Rate Conversion

```go
// Convert entire timeline to different frame rate
func convertToFrameRate(t opentime.RationalTime, newRate float64) opentime.RationalTime {
    return t.RescaledTo(newRate)
}

// Convert range
func convertRangeToFrameRate(r opentime.TimeRange, newRate float64) opentime.TimeRange {
    return opentime.NewTimeRange(
        r.StartTime().RescaledTo(newRate),
        r.Duration().RescaledTo(newRate),
    )
}
```

### Working with Timecode

```go
// Parse timecode with different formats
func parseTimecode(tc string, rate float64) (opentime.RationalTime, error) {
    return opentime.FromTimecode(tc, rate)
}

// Format as timecode
func formatAsTimecode(t opentime.RationalTime, dropFrame bool) (string, error) {
    dfRate := opentime.ForceNo
    if dropFrame {
        dfRate = opentime.ForceYes
    }
    return t.ToTimecode(t.Rate(), dfRate)
}
```

### Finding Clips at Time

```go
// Find what's playing at a specific time
func clipAtTime(timeline *opentimelineio.Timeline, time opentime.RationalTime) *opentimelineio.Clip {
    for _, track := range timeline.VideoTracks() {
        child, _ := track.ChildAtTime(time, false)
        if clip, ok := child.(*opentimelineio.Clip); ok {
            return clip
        }
    }
    return nil
}

// Example usage
t := opentime.NewRationalTime(100, 24)
clip := clipAtTime(timeline, t)
if clip != nil {
    fmt.Printf("At frame 100: %s\n", clip.Name())
}
```

### Range Manipulation

```go
// Split a range at a specific time
func splitRange(r opentime.TimeRange, splitPoint opentime.RationalTime) (opentime.TimeRange, opentime.TimeRange) {
    if !r.Contains(splitPoint) {
        return r, opentime.TimeRange{}
    }

    first := opentime.NewTimeRange(
        r.StartTime(),
        splitPoint.Sub(r.StartTime()),
    )
    second := opentime.NewTimeRange(
        splitPoint,
        r.EndTimeExclusive().Sub(splitPoint),
    )
    return first, second
}

// Offset a range by a duration
func offsetRange(r opentime.TimeRange, offset opentime.RationalTime) opentime.TimeRange {
    return opentime.NewTimeRange(
        r.StartTime().Add(offset),
        r.Duration(),
    )
}
```

## Epsilon Comparisons

Floating-point comparisons use epsilon for tolerance:

```go
// Default epsilon
epsilon := opentime.DefaultEpsilon

// Check if times are approximately equal
t1 := opentime.NewRationalTime(100, 24)
t2 := opentime.NewRationalTimeFromSeconds(4.1666666, 24)

if t1.AlmostEqual(t2, epsilon) {
    fmt.Println("Times are approximately equal")
}

// Range overlaps with epsilon tolerance
overlaps := range1.Overlaps(range2, epsilon)
```
