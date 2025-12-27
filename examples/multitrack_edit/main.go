// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

// multitrack_edit demonstrates creating a complex multi-track timeline
// with video, audio, transitions, effects, and markers.
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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <output.otio>")
		os.Exit(1)
	}
	outputPath := os.Args[1]

	// Create the timeline
	globalStart := opentime.NewRationalTime(86400, 24) // 1:00:00:00
	timeline := opentimelineio.NewTimeline(
		"Multi-Track Example",
		&globalStart,
		opentimelineio.AnyDictionary{
			"project":     "Demo Project",
			"frame_rate":  24.0,
			"resolution":  "1920x1080",
			"color_space": "rec709",
		},
	)

	// === Video Track 1 (Main) ===
	v1 := opentimelineio.NewTrack("V1 - Main", nil, opentimelineio.TrackKindVideo, nil, nil)
	timeline.Tracks().AppendChild(v1)

	// Add clips to V1
	addClipToTrack(v1, "Interview_Wide", "interview_wide.mov", 0, 240)
	addTransition(v1, 12) // 12 frame dissolve
	addClipToTrack(v1, "Interview_CU", "interview_cu.mov", 0, 180)
	addTransition(v1, 12)
	addClipToTrack(v1, "Interview_Wide_2", "interview_wide.mov", 300, 150)

	// === Video Track 2 (B-Roll overlay) ===
	v2 := opentimelineio.NewTrack("V2 - B-Roll", nil, opentimelineio.TrackKindVideo, nil, nil)
	timeline.Tracks().AppendChild(v2)

	// Add gap to offset the B-roll
	addGap(v2, 48) // Start at 2 seconds

	// Add B-roll clips
	brollClip := addClipToTrack(v2, "BRoll_City", "broll_city.mov", 0, 72)

	// Add a speed effect to B-roll (slow motion)
	slowMo := opentimelineio.NewLinearTimeWarp(
		"slow_motion",
		"LinearTimeWarp",
		0.5, // 50% speed
		nil,
	)
	brollClip.SetEffects(append(brollClip.Effects(), slowMo))

	addGap(v2, 48) // Gap between B-roll
	addClipToTrack(v2, "BRoll_Nature", "broll_nature.mov", 24, 96)

	// === Video Track 3 (Graphics) ===
	v3 := opentimelineio.NewTrack("V3 - Graphics", nil, opentimelineio.TrackKindVideo, nil, nil)
	timeline.Tracks().AppendChild(v3)

	// Title card at the beginning
	titleClip := addClipToTrack(v3, "Title_Card", "title.png", 0, 72)

	// Add marker for title
	titleMarker := opentimelineio.NewMarker(
		"Title Start",
		opentime.NewTimeRange(
			opentime.NewRationalTime(0, 24),
			opentime.NewRationalTime(0, 24),
		),
		opentimelineio.MarkerColorGreen,
		"Title card begins here",
		nil,
	)
	titleClip.SetMarkers(append(titleClip.Markers(), titleMarker))

	// Add gap, then lower third
	addGap(v3, 96)
	addClipToTrack(v3, "Lower_Third", "lower_third.mov", 0, 72)

	// === Audio Track 1 (Dialog) ===
	a1 := opentimelineio.NewTrack("A1 - Dialog", nil, opentimelineio.TrackKindAudio, nil, nil)
	timeline.Tracks().AppendChild(a1)

	// Audio follows video
	addClipToTrack(a1, "Dialog_1", "dialog_01.wav", 0, 240)
	addClipToTrack(a1, "Dialog_2", "dialog_02.wav", 0, 180)
	addClipToTrack(a1, "Dialog_3", "dialog_03.wav", 0, 150)

	// === Audio Track 2 (Music) ===
	a2 := opentimelineio.NewTrack("A2 - Music", nil, opentimelineio.TrackKindAudio, nil, nil)
	timeline.Tracks().AppendChild(a2)

	// Music runs under everything
	musicClip := addClipToTrack(a2, "Background_Music", "music_bed.wav", 0, 600)

	// Mark music cue points
	musicMarker1 := opentimelineio.NewMarker(
		"Music Build",
		opentime.NewTimeRange(
			opentime.NewRationalTime(96, 24),
			opentime.NewRationalTime(0, 24),
		),
		opentimelineio.MarkerColorYellow,
		"Music intensity increases",
		nil,
	)
	musicClip.SetMarkers(append(musicClip.Markers(), musicMarker1))

	// === Audio Track 3 (SFX) ===
	a3 := opentimelineio.NewTrack("A3 - SFX", nil, opentimelineio.TrackKindAudio, nil, nil)
	timeline.Tracks().AppendChild(a3)

	// Spot SFX
	addGap(a3, 48)
	addClipToTrack(a3, "Ambience_City", "sfx_city_amb.wav", 0, 72)
	addGap(a3, 48)
	addClipToTrack(a3, "Ambience_Nature", "sfx_nature_amb.wav", 0, 96)

	// Print summary
	printTimelineSummary(timeline)

	// Write to file
	if err := opentimelineio.ToJSONFile(timeline, outputPath, "  "); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
	fmt.Printf("\nSaved to: %s\n", outputPath)
}

func addClipToTrack(track *opentimelineio.Track, name, file string, startFrame, duration float64) *opentimelineio.Clip {
	sourceRange := opentime.NewTimeRange(
		opentime.NewRationalTime(startFrame, 24),
		opentime.NewRationalTime(duration, 24),
	)

	availableRange := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(duration+startFrame+100, 24), // More available than used
	)

	ref := opentimelineio.NewExternalReference(
		name,
		"file:///media/"+file,
		&availableRange,
		nil,
	)

	clip := opentimelineio.NewClip(
		name,
		ref,
		&sourceRange,
		nil, nil, nil, "", nil,
	)

	if err := track.AppendChild(clip); err != nil {
		log.Fatalf("Failed to add clip %s: %v", name, err)
	}

	return clip
}

func addGap(track *opentimelineio.Track, frames float64) {
	gap := opentimelineio.NewGapWithDuration(
		opentime.NewRationalTime(frames, 24),
	)
	if err := track.AppendChild(gap); err != nil {
		log.Fatalf("Failed to add gap: %v", err)
	}
}

func addTransition(track *opentimelineio.Track, frames float64) {
	transition := opentimelineio.NewTransition(
		"Dissolve",
		opentimelineio.TransitionTypeSMPTEDissolve,
		opentime.NewRationalTime(frames, 24),
		opentime.NewRationalTime(frames, 24),
		nil,
	)
	if err := track.AppendChild(transition); err != nil {
		log.Fatalf("Failed to add transition: %v", err)
	}
}

func printTimelineSummary(timeline *opentimelineio.Timeline) {
	fmt.Println("\n=== Timeline Summary ===")
	fmt.Printf("Name: %s\n", timeline.Name())

	if gst := timeline.GlobalStartTime(); gst != nil {
		tc, _ := gst.ToTimecode(24, opentime.ForceNo)
		fmt.Printf("Global Start: %s\n", tc)
	}

	dur, _ := timeline.Duration()
	tc, _ := dur.ToTimecode(24, opentime.ForceNo)
	fmt.Printf("Duration: %s (%.2f seconds)\n", tc, dur.ToSeconds())

	fmt.Printf("\nVideo Tracks: %d\n", len(timeline.VideoTracks()))
	for i, track := range timeline.VideoTracks() {
		trackDur, _ := track.Duration()
		fmt.Printf("  V%d: %s (%.2fs, %d items)\n",
			i+1, track.Name(), trackDur.ToSeconds(), len(track.Children()))
	}

	fmt.Printf("\nAudio Tracks: %d\n", len(timeline.AudioTracks()))
	for i, track := range timeline.AudioTracks() {
		trackDur, _ := track.Duration()
		fmt.Printf("  A%d: %s (%.2fs, %d items)\n",
			i+1, track.Name(), trackDur.ToSeconds(), len(track.Children()))
	}

	clips := timeline.FindClips(nil, false)
	fmt.Printf("\nTotal Clips: %d\n", len(clips))

	// Count markers
	markerCount := 0
	for _, clip := range clips {
		markerCount += len(clip.Markers())
	}
	fmt.Printf("Total Markers: %d\n", markerCount)

	// Count effects
	effectCount := 0
	for _, clip := range clips {
		effectCount += len(clip.Effects())
	}
	fmt.Printf("Total Effects: %d\n", effectCount)
}
