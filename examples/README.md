# gotio Examples

This directory contains working examples demonstrating gotio usage.

## Examples

### build_simple_timeline

Creates a simple timeline from scratch with clips and markers.

```bash
cd build_simple_timeline
go run main.go output.otio
```

**Demonstrates:**
- Creating a Timeline
- Adding Tracks
- Creating Clips with ExternalReferences
- Setting source ranges
- Adding Markers
- Writing to file

---

### read_timeline

Reads and explores an OTIO file, printing detailed information.

```bash
cd read_timeline
go run main.go input.otio
```

**Demonstrates:**
- Loading OTIO files
- Type assertions for different schema types
- Iterating over tracks and clips
- Accessing media references
- Reading effects and markers

---

### flatten_tracks

Flattens multiple video tracks into a single track.

```bash
cd flatten_tracks
go run main.go input.otio output.otio
```

**Demonstrates:**
- Using the algorithms package
- Flattening multi-track timelines
- Understanding compositing order

---

### summarize_timing

Prints a detailed timing analysis of a timeline.

```bash
cd summarize_timing
go run main.go input.otio
```

**Demonstrates:**
- Calculating durations
- Analyzing media usage
- Working with timecode
- Iterating over all clips

---

### multitrack_edit

Creates a complex multi-track timeline with video, audio, transitions, effects, and markers.

```bash
cd multitrack_edit
go run main.go output.otio
```

**Demonstrates:**
- Multiple video and audio tracks
- Adding transitions between clips
- Applying effects (LinearTimeWarp)
- Adding markers
- Using gaps for timing
- Setting global start time
- Rich metadata

---

## Running Examples

From the examples directory:

```bash
# Run any example
go run ./build_simple_timeline output.otio
go run ./read_timeline ../test.otio
go run ./flatten_tracks input.otio output.otio
go run ./summarize_timing input.otio
go run ./multitrack_edit output.otio
```

Or from the individual example directory:

```bash
cd build_simple_timeline
go run main.go output.otio
```

## Sample Workflows

### Create and Inspect

```bash
# Create a timeline
go run ./build_simple_timeline my_timeline.otio

# Inspect it
go run ./read_timeline my_timeline.otio

# Get timing summary
go run ./summarize_timing my_timeline.otio
```

### Complex Edit

```bash
# Create a multi-track edit
go run ./multitrack_edit complex_edit.otio

# Flatten for delivery
go run ./flatten_tracks complex_edit.otio flattened.otio

# Compare
go run ./summarize_timing complex_edit.otio
go run ./summarize_timing flattened.otio
```
