// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

// summarize_timing prints a timing summary of a timeline.
//
// This example shows:
// - Calculating total duration
// - Finding gaps and their total duration
// - Listing clips with their timing information
// - Converting between timecode and frames
//
// Usage:
//
//	go run main.go input.otio
package main

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/mrjoshuak/gotio/opentime"
	"github.com/mrjoshuak/gotio/opentimelineio"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <input.otio>")
		os.Exit(1)
	}
	inputPath := os.Args[1]

	// Load the timeline
	obj, err := opentimelineio.FromJSONFile(inputPath)
	if err != nil {
		log.Fatalf("Failed to load %s: %v", inputPath, err)
	}

	timeline, ok := obj.(*opentimelineio.Timeline)
	if !ok {
		log.Fatalf("Expected Timeline, got %T", obj)
	}

	fmt.Printf("Timeline: %s\n", timeline.Name())
	fmt.Println("=" + repeatString("=", len(timeline.Name())+9))

	// Overall timing
	duration, err := timeline.Duration()
	if err != nil {
		log.Fatalf("Failed to get duration: %v", err)
	}

	fmt.Printf("\nTotal Duration: %s (%.2f seconds)\n",
		formatTimecode(duration), duration.ToSeconds())

	// Analyze each track
	fmt.Println("\n--- Track Analysis ---")

	var totalClipDuration float64
	var totalGapDuration float64
	var clipCount int

	for _, child := range timeline.Tracks().Children() {
		track, ok := child.(*opentimelineio.Track)
		if !ok {
			continue
		}

		trackDur, _ := track.Duration()
		fmt.Printf("\nTrack: %s (%s)\n", track.Name(), track.Kind())
		fmt.Printf("  Duration: %s (%.2f seconds)\n",
			formatTimecode(trackDur), trackDur.ToSeconds())

		// Count clips and gaps
		trackClipDur := 0.0
		trackGapDur := 0.0
		trackClipCount := 0

		for i, item := range track.Children() {
			itemRange, _ := track.RangeOfChildAtIndex(i)

			switch item.(type) {
			case *opentimelineio.Clip:
				trackClipDur += itemRange.Duration().ToSeconds()
				trackClipCount++
			case *opentimelineio.Gap:
				trackGapDur += itemRange.Duration().ToSeconds()
			}
		}

		fmt.Printf("  Clips: %d (%.2fs content)\n", trackClipCount, trackClipDur)
		if trackGapDur > 0 {
			fmt.Printf("  Gaps: %.2fs\n", trackGapDur)
		}

		totalClipDuration += trackClipDur
		totalGapDuration += trackGapDur
		clipCount += trackClipCount
	}

	// Clip list
	fmt.Println("\n--- Clip List ---")
	fmt.Printf("%-4s %-20s %-15s %-15s %-10s\n",
		"#", "Name", "In", "Out", "Duration")
	fmt.Println(repeatString("-", 70))

	clips := timeline.FindClips(nil, false)
	for i, clip := range clips {
		dur, _ := clip.Duration()
		sr := clip.SourceRange()

		inTime := opentime.RationalTime{}
		outTime := opentime.RationalTime{}
		if sr != nil {
			inTime = sr.StartTime()
			outTime = sr.EndTimeExclusive()
		}

		name := truncateString(clip.Name(), 20)
		fmt.Printf("%-4d %-20s %-15s %-15s %.2fs\n",
			i+1, name,
			formatTimecode(inTime),
			formatTimecode(outTime),
			dur.ToSeconds())
	}

	// Media usage analysis
	fmt.Println("\n--- Media Usage ---")
	mediaUsage := analyzeMediaUsage(clips)

	// Sort by usage count
	type mediaInfo struct {
		url   string
		count int
		dur   float64
	}
	var mediaList []mediaInfo
	for url, info := range mediaUsage {
		mediaList = append(mediaList, mediaInfo{url, info.count, info.duration})
	}
	sort.Slice(mediaList, func(i, j int) bool {
		return mediaList[i].count > mediaList[j].count
	})

	for _, m := range mediaList {
		fmt.Printf("  %s\n    Used %d times, %.2fs total\n", m.url, m.count, m.dur)
	}

	// Summary
	fmt.Println("\n--- Summary ---")
	fmt.Printf("Total clips: %d\n", clipCount)
	fmt.Printf("Unique media files: %d\n", len(mediaUsage))
	fmt.Printf("Total content duration: %.2fs\n", totalClipDuration)
	if totalGapDuration > 0 {
		fmt.Printf("Total gap duration: %.2fs\n", totalGapDuration)
	}
}

type mediaStats struct {
	count    int
	duration float64
}

func analyzeMediaUsage(clips []*opentimelineio.Clip) map[string]mediaStats {
	usage := make(map[string]mediaStats)

	for _, clip := range clips {
		ref := clip.MediaReference()
		if ref == nil {
			continue
		}

		var url string
		switch r := ref.(type) {
		case *opentimelineio.ExternalReference:
			url = r.TargetURL()
		case *opentimelineio.MissingReference:
			url = "(missing)"
		case *opentimelineio.GeneratorReference:
			url = fmt.Sprintf("generator:%s", r.GeneratorKind())
		default:
			url = "(unknown)"
		}

		dur, _ := clip.Duration()
		stats := usage[url]
		stats.count++
		stats.duration += dur.ToSeconds()
		usage[url] = stats
	}

	return usage
}

func formatTimecode(t opentime.RationalTime) string {
	if t.Rate() <= 0 {
		return "00:00:00:00"
	}
	tc, err := t.ToTimecode(t.Rate(), opentime.ForceNo)
	if err != nil {
		return fmt.Sprintf("%.0f", t.Value())
	}
	return tc
}

func repeatString(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
