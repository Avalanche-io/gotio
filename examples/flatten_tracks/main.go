// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

// flatten_tracks demonstrates flattening multiple video tracks into one.
//
// This is useful for:
// - Simplifying timelines for delivery
// - Converting multi-track edits to single-track EDL format
// - Analyzing what's visible at each point in time
//
// Usage:
//
//	go run main.go input.otio output.otio
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mrjoshuak/gotio/algorithms"
	"github.com/mrjoshuak/gotio/opentimelineio"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <input.otio> <output.otio>")
		os.Exit(1)
	}
	inputPath := os.Args[1]
	outputPath := os.Args[2]

	// Load the timeline
	obj, err := opentimelineio.FromJSONFile(inputPath)
	if err != nil {
		log.Fatalf("Failed to load %s: %v", inputPath, err)
	}

	timeline, ok := obj.(*opentimelineio.Timeline)
	if !ok {
		log.Fatalf("Expected Timeline, got %T", obj)
	}

	fmt.Printf("Loaded: %s\n", timeline.Name())

	// Count original tracks
	videoTracks := timeline.VideoTracks()
	audioTracks := timeline.AudioTracks()
	fmt.Printf("Original video tracks: %d\n", len(videoTracks))
	fmt.Printf("Original audio tracks: %d\n", len(audioTracks))

	// Print original structure
	fmt.Println("\n--- Original Structure ---")
	for i, track := range videoTracks {
		dur, _ := track.Duration()
		fmt.Printf("V%d: %s (%.2fs, %d items)\n",
			i+1, track.Name(), dur.ToSeconds(), len(track.Children()))
	}

	// Flatten video tracks
	fmt.Println("\n--- Flattening ---")
	flattened, err := algorithms.FlattenTimelineVideoTracks(timeline)
	if err != nil {
		log.Fatalf("Failed to flatten: %v", err)
	}

	// Print flattened structure
	flatVideoTracks := flattened.VideoTracks()
	fmt.Printf("Flattened video tracks: %d\n", len(flatVideoTracks))

	if len(flatVideoTracks) > 0 {
		flatTrack := flatVideoTracks[0]
		dur, _ := flatTrack.Duration()
		fmt.Printf("Flattened track: %s (%.2fs)\n",
			flatTrack.Name(), dur.ToSeconds())

		fmt.Println("\nFlattened contents:")
		for i, child := range flatTrack.Children() {
			childRange, _ := flatTrack.RangeOfChildAtIndex(i)
			switch c := child.(type) {
			case *opentimelineio.Clip:
				fmt.Printf("  [%d] %.2fs - %.2fs: %s\n",
					i,
					childRange.StartTime().ToSeconds(),
					childRange.EndTimeExclusive().ToSeconds(),
					c.Name())
			case *opentimelineio.Gap:
				fmt.Printf("  [%d] %.2fs - %.2fs: (gap)\n",
					i,
					childRange.StartTime().ToSeconds(),
					childRange.EndTimeExclusive().ToSeconds())
			case *opentimelineio.Transition:
				fmt.Printf("  [%d] transition: %s\n", i, c.TransitionType())
			}
		}
	}

	// Write flattened timeline
	if err := opentimelineio.ToJSONFile(flattened, outputPath, "  "); err != nil {
		log.Fatalf("Failed to write %s: %v", outputPath, err)
	}

	fmt.Printf("\nSaved flattened timeline to: %s\n", outputPath)
}
