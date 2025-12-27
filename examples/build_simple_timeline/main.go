// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

// build_simple_timeline demonstrates creating a simple timeline from scratch.
//
// Usage:
//
//	go run main.go output.otio
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mrjoshuak/gotio/opentime"
	"github.com/mrjoshuak/gotio/opentimelineio"
)

// MediaFile represents a source media file
type MediaFile struct {
	Name           string
	Path           string
	AvailableRange opentime.TimeRange
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <output.otio>")
		os.Exit(1)
	}
	outputPath := os.Args[1]

	// Define our source media files
	mediaFiles := []MediaFile{
		{
			Name: "Opening Shot",
			Path: "file:///media/opening.mov",
			AvailableRange: opentime.NewTimeRange(
				opentime.NewRationalTime(0, 24),
				opentime.NewRationalTime(120, 24), // 5 seconds
			),
		},
		{
			Name: "Interview A",
			Path: "file:///media/interview_a.mov",
			AvailableRange: opentime.NewTimeRange(
				opentime.NewRationalTime(86400, 24), // starts at 1 hour (typical camera timecode)
				opentime.NewRationalTime(7200, 24),  // 5 minutes of footage
			),
		},
		{
			Name: "B-Roll Forest",
			Path: "file:///media/broll_forest.mov",
			AvailableRange: opentime.NewTimeRange(
				opentime.NewRationalTime(0, 24),
				opentime.NewRationalTime(480, 24), // 20 seconds
			),
		},
		{
			Name: "Closing Shot",
			Path: "file:///media/closing.mov",
			AvailableRange: opentime.NewTimeRange(
				opentime.NewRationalTime(0, 24),
				opentime.NewRationalTime(144, 24), // 6 seconds
			),
		},
	}

	// Create the timeline with metadata
	timeline := opentimelineio.NewTimeline(
		"Documentary Edit v1",
		nil, // no global start time
		opentimelineio.AnyDictionary{
			"project":  "My Documentary",
			"editor":   "Jane Doe",
			"revision": 1,
		},
	)

	// Create a video track
	videoTrack := opentimelineio.NewTrack(
		"V1",
		nil,
		opentimelineio.TrackKindVideo,
		nil,
		nil,
	)
	timeline.Tracks().AppendChild(videoTrack)

	// Add clips to the track
	for i, media := range mediaFiles {
		// Create external reference to the media file
		ref := opentimelineio.NewExternalReference(
			media.Name,
			media.Path,
			&media.AvailableRange,
			nil,
		)

		// Define which portion of the media to use
		// We'll use a trimmed portion that's slightly shorter than available
		sourceStart := media.AvailableRange.StartTime().Add(
			opentime.NewRationalTime(12, 24), // skip first 12 frames
		)
		sourceDuration := media.AvailableRange.Duration().Sub(
			opentime.NewRationalTime(24, 24), // trim 24 frames total
		)
		sourceRange := opentime.NewTimeRange(sourceStart, sourceDuration)

		// Create the clip
		clip := opentimelineio.NewClip(
			fmt.Sprintf("Clip %d - %s", i+1, media.Name),
			ref,
			&sourceRange,
			opentimelineio.AnyDictionary{
				"source_file": media.Path,
				"clip_index":  i,
			},
			nil, // no effects
			nil, // no markers
			"",  // default media reference key
			nil, // no color
		)

		// Add a marker to the first clip
		if i == 0 {
			markerRange := opentime.NewTimeRange(
				opentime.NewRationalTime(24, 24), // 1 second into clip
				opentime.NewRationalTime(0, 24),  // point marker
			)
			marker := opentimelineio.NewMarker(
				"Title Card",
				markerRange,
				opentimelineio.MarkerColorGreen,
				"Add title overlay here",
				nil,
			)
			clip.SetMarkers(append(clip.Markers(), marker))
		}

		// Add the clip to the track
		if err := videoTrack.AppendChild(clip); err != nil {
			log.Fatalf("Failed to add clip: %v", err)
		}

		fmt.Printf("Added: %s (%.2f seconds)\n",
			clip.Name(),
			sourceDuration.ToSeconds())
	}

	// Print timeline summary
	fmt.Println("\n--- Timeline Summary ---")
	fmt.Printf("Name: %s\n", timeline.Name())

	duration, err := timeline.Duration()
	if err != nil {
		log.Fatalf("Failed to get duration: %v", err)
	}
	fmt.Printf("Duration: %.2f seconds (%.0f frames @ 24fps)\n",
		duration.ToSeconds(),
		duration.Value())

	clips := timeline.FindClips(nil, false)
	fmt.Printf("Total clips: %d\n", len(clips))

	// Write to file
	if err := opentimelineio.ToJSONFile(timeline, outputPath, "  "); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
	fmt.Printf("\nSaved to: %s\n", outputPath)
}
