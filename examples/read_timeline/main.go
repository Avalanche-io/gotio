// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

// read_timeline demonstrates reading and exploring an OTIO file.
//
// Usage:
//
//	go run main.go input.otio
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Avalanche-io/gotio/opentime"
	"github.com/Avalanche-io/gotio"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <input.otio>")
		os.Exit(1)
	}
	inputPath := os.Args[1]

	// Load the timeline
	obj, err := gotio.FromJSONFile(inputPath)
	if err != nil {
		log.Fatalf("Failed to load %s: %v", inputPath, err)
	}

	// Type switch to handle different root types
	switch root := obj.(type) {
	case *gotio.Timeline:
		printTimeline(root)
	case *gotio.SerializableCollection:
		printCollection(root)
	default:
		fmt.Printf("Root object is %T\n", obj)
	}
}

func printTimeline(timeline *gotio.Timeline) {
	fmt.Println("=== Timeline ===")
	fmt.Printf("Name: %s\n", timeline.Name())

	// Global start time
	if gst := timeline.GlobalStartTime(); gst != nil {
		fmt.Printf("Global Start Time: %v\n", gst)
	}

	// Duration
	if dur, err := timeline.Duration(); err == nil {
		fmt.Printf("Duration: %.2f seconds\n", dur.ToSeconds())
	}

	// Metadata
	if len(timeline.Metadata()) > 0 {
		fmt.Println("\nMetadata:")
		for k, v := range timeline.Metadata() {
			fmt.Printf("  %s: %v\n", k, v)
		}
	}

	// Tracks
	fmt.Println("\n--- Tracks ---")
	videoTracks := timeline.VideoTracks()
	audioTracks := timeline.AudioTracks()
	fmt.Printf("Video tracks: %d\n", len(videoTracks))
	fmt.Printf("Audio tracks: %d\n", len(audioTracks))

	// Print each track
	for i, child := range timeline.Tracks().Children() {
		track, ok := child.(*gotio.Track)
		if !ok {
			continue
		}
		printTrack(track, i)
	}

	// All clips summary
	fmt.Println("\n--- All Clips ---")
	clips := timeline.FindClips(nil, false)
	for i, clip := range clips {
		printClip(clip, i)
	}
}

func printTrack(track *gotio.Track, index int) {
	fmt.Printf("\nTrack %d: %s (%s)\n", index, track.Name(), track.Kind())

	dur, _ := track.Duration()
	fmt.Printf("  Duration: %.2f seconds\n", dur.ToSeconds())
	fmt.Printf("  Children: %d\n", len(track.Children()))

	// Print children summary
	for i, child := range track.Children() {
		childRange, _ := track.RangeOfChildAtIndex(i)

		switch c := child.(type) {
		case *gotio.Clip:
			fmt.Printf("    [%d] Clip: %s @ %.2fs (%.2fs)\n",
				i, c.Name(),
				childRange.StartTime().ToSeconds(),
				childRange.Duration().ToSeconds())
		case *gotio.Gap:
			fmt.Printf("    [%d] Gap @ %.2fs (%.2fs)\n",
				i,
				childRange.StartTime().ToSeconds(),
				childRange.Duration().ToSeconds())
		case *gotio.Transition:
			fmt.Printf("    [%d] Transition: %s (in: %.2fs, out: %.2fs)\n",
				i, c.TransitionType(),
				c.InOffset().ToSeconds(),
				c.OutOffset().ToSeconds())
		case *gotio.Stack:
			fmt.Printf("    [%d] Nested Stack: %s\n", i, c.Name())
		case *gotio.Track:
			fmt.Printf("    [%d] Nested Track: %s\n", i, c.Name())
		default:
			fmt.Printf("    [%d] %T\n", i, child)
		}
	}
}

func printClip(clip *gotio.Clip, index int) {
	fmt.Printf("\nClip %d: %s\n", index, clip.Name())

	// Duration
	if dur, err := clip.Duration(); err == nil {
		fmt.Printf("  Duration: %.2f seconds\n", dur.ToSeconds())
	}

	// Source range
	if sr := clip.SourceRange(); sr != nil {
		fmt.Printf("  Source Range: %v - %v\n",
			sr.StartTime(), sr.EndTimeExclusive())
	}

	// Media reference
	ref := clip.MediaReference()
	if ref != nil {
		switch r := ref.(type) {
		case *gotio.ExternalReference:
			fmt.Printf("  Media: %s\n", r.TargetURL())
			if ar := r.AvailableRange(); ar != nil {
				fmt.Printf("  Available: %v - %v\n",
					ar.StartTime(), ar.EndTimeExclusive())
			}
		case *gotio.MissingReference:
			fmt.Println("  Media: MISSING")
		case *gotio.GeneratorReference:
			fmt.Printf("  Generator: %s\n", r.GeneratorKind())
		case *gotio.ImageSequenceReference:
			fmt.Printf("  Image Sequence: %s\n", r.TargetURLBase())
		}
	}

	// Effects
	if effects := clip.Effects(); len(effects) > 0 {
		fmt.Printf("  Effects: %d\n", len(effects))
		for _, e := range effects {
			fmt.Printf("    - %s (%s)\n", e.Name(), e.EffectName())
			if ltw, ok := e.(*gotio.LinearTimeWarp); ok {
				fmt.Printf("      Time Scalar: %.2f\n", ltw.TimeScalar())
			}
		}
	}

	// Markers
	if markers := clip.Markers(); len(markers) > 0 {
		fmt.Printf("  Markers: %d\n", len(markers))
		for _, m := range markers {
			fmt.Printf("    - %s @ %v (%s): %s\n",
				m.Name(),
				m.MarkedRange().StartTime(),
				m.Color(),
				m.Comment())
		}
	}
}

func printCollection(coll *gotio.SerializableCollection) {
	fmt.Println("=== Serializable Collection ===")
	fmt.Printf("Name: %s\n", coll.Name())
	fmt.Printf("Children: %d\n", len(coll.Children()))

	for i, child := range coll.Children() {
		fmt.Printf("  [%d] %s (%s)\n", i, child.SchemaName(),
			getChildName(child))
	}
}

func getChildName(obj gotio.SerializableObject) string {
	switch o := obj.(type) {
	case *gotio.Timeline:
		return o.Name()
	case *gotio.Clip:
		return o.Name()
	case *gotio.Track:
		return o.Name()
	case *gotio.Stack:
		return o.Name()
	default:
		return ""
	}
}

// formatTimecode converts a RationalTime to timecode string
func formatTimecode(t opentime.RationalTime) string {
	tc, err := t.ToTimecode(t.Rate(), opentime.ForceNo)
	if err != nil {
		return fmt.Sprintf("%.2fs", t.ToSeconds())
	}
	return tc
}
