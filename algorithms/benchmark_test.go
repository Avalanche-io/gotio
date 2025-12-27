// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package algorithms

import (
	"fmt"
	"testing"

	"github.com/Avalanche-io/gotio/opentime"
	"github.com/Avalanche-io/gotio/opentimelineio"
)

// =============================================================================
// Benchmark Helpers
// =============================================================================

// createBenchmarkTrackForAlgo creates a track with n clips for algorithm benchmarks.
func createBenchmarkTrackForAlgo(n int) *opentimelineio.Track {
	track := opentimelineio.NewTrack("bench_track", nil, opentimelineio.TrackKindVideo, nil, nil)
	for i := 0; i < n; i++ {
		sr := opentime.NewTimeRange(
			opentime.NewRationalTime(0, 24),
			opentime.NewRationalTime(24, 24),
		)
		clip := opentimelineio.NewClip(fmt.Sprintf("clip_%d", i), nil, &sr, nil, nil, nil, "", nil)
		track.AppendChild(clip)
	}
	return track
}

// createBenchmarkTracks creates a slice of tracks for FlattenTracks benchmarks.
func createBenchmarkTracks(numTracks, clipsPerTrack int) []*opentimelineio.Track {
	tracks := make([]*opentimelineio.Track, numTracks)
	for t := 0; t < numTracks; t++ {
		tracks[t] = createBenchmarkTrackForAlgo(clipsPerTrack)
		tracks[t].SetName(fmt.Sprintf("track_%d", t))
	}
	return tracks
}

// createBenchmarkStackForAlgo creates a stack with the specified number of tracks and clips.
func createBenchmarkStackForAlgo(tracks, clipsPerTrack int) *opentimelineio.Stack {
	stack := opentimelineio.NewStack("bench_stack", nil, nil, nil, nil, nil)
	for t := 0; t < tracks; t++ {
		track := createBenchmarkTrackForAlgo(clipsPerTrack)
		track.SetName(fmt.Sprintf("track_%d", t))
		stack.AppendChild(track)
	}
	return stack
}

// createBenchmarkTimelineForAlgo creates a timeline with video tracks.
func createBenchmarkTimelineForAlgo(tracks, clipsPerTrack int) *opentimelineio.Timeline {
	timeline := opentimelineio.NewTimeline("bench_timeline", nil, nil)
	stack := timeline.Tracks()
	for t := 0; t < tracks; t++ {
		track := createBenchmarkTrackForAlgo(clipsPerTrack)
		track.SetName(fmt.Sprintf("V%d", t+1))
		track.SetKind(opentimelineio.TrackKindVideo)
		stack.AppendChild(track)
	}
	return timeline
}

// =============================================================================
// FlattenStack Benchmarks - O(n^2 * m) complexity
// =============================================================================

func BenchmarkFlattenStack(b *testing.B) {
	configs := []struct {
		tracks        int
		clipsPerTrack int
	}{
		{2, 10},
		{3, 25},
		{5, 50},
		{10, 50},
	}

	for _, c := range configs {
		name := fmt.Sprintf("tracks=%d_clips=%d", c.tracks, c.clipsPerTrack)
		b.Run(name, func(b *testing.B) {
			stack := createBenchmarkStackForAlgo(c.tracks, c.clipsPerTrack)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = FlattenStack(stack)
			}
		})
	}
}

func BenchmarkFlattenTracks(b *testing.B) {
	configs := []struct {
		tracks        int
		clipsPerTrack int
	}{
		{2, 25},
		{3, 50},
		{5, 100},
	}

	for _, c := range configs {
		name := fmt.Sprintf("tracks=%d_clips=%d", c.tracks, c.clipsPerTrack)
		b.Run(name, func(b *testing.B) {
			tracks := createBenchmarkTracks(c.tracks, c.clipsPerTrack)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = FlattenTracks(tracks)
			}
		})
	}
}

func BenchmarkFlattenTimelineVideoTracks(b *testing.B) {
	configs := []struct {
		tracks        int
		clipsPerTrack int
	}{
		{2, 25},
		{5, 50},
	}

	for _, c := range configs {
		name := fmt.Sprintf("tracks=%d_clips=%d", c.tracks, c.clipsPerTrack)
		b.Run(name, func(b *testing.B) {
			timeline := createBenchmarkTimelineForAlgo(c.tracks, c.clipsPerTrack)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = FlattenTimelineVideoTracks(timeline)
			}
		})
	}
}

// =============================================================================
// itemAtTime Benchmarks - O(n^2) due to RangeOfChildAtIndex
// =============================================================================

func BenchmarkItemAtTime(b *testing.B) {
	scales := []int{100, 500, 1000}
	for _, n := range scales {
		b.Run(fmt.Sprintf("clips=%d", n), func(b *testing.B) {
			track := createBenchmarkTrackForAlgo(n)
			midTime := opentime.NewRationalTime(float64(n)*24/2, 24)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _, _, _ = itemAtTime(track, midTime)
			}
		})
	}
}

func BenchmarkItemAtTime_ByPosition(b *testing.B) {
	n := 1000
	track := createBenchmarkTrackForAlgo(n)
	totalDuration := float64(n) * 24

	positions := []struct {
		name string
		time float64
	}{
		{"start", 12},
		{"quarter", totalDuration * 0.25},
		{"middle", totalDuration * 0.5},
		{"three_quarter", totalDuration * 0.75},
		{"end", totalDuration - 12},
	}

	for _, pos := range positions {
		b.Run(pos.name, func(b *testing.B) {
			t := opentime.NewRationalTime(pos.time, 24)
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _, _, _ = itemAtTime(track, t)
			}
		})
	}
}

// =============================================================================
// itemsInRange Benchmarks - O(n^2) due to RangeOfChildAtIndex
// =============================================================================

func BenchmarkItemsInRange(b *testing.B) {
	scales := []int{100, 500, 1000}
	for _, n := range scales {
		b.Run(fmt.Sprintf("clips=%d", n), func(b *testing.B) {
			track := createBenchmarkTrackForAlgo(n)
			start := opentime.NewRationalTime(float64(n)*24*0.25, 24)
			dur := opentime.NewRationalTime(float64(n)*24*0.5, 24)
			searchRange := opentime.NewTimeRange(start, dur)

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _, _, _ = itemsInRange(track, searchRange)
			}
		})
	}
}

func BenchmarkItemsInRange_Coverage(b *testing.B) {
	n := 500
	track := createBenchmarkTrackForAlgo(n)
	totalDuration := float64(n) * 24

	coverages := []struct {
		name    string
		percent float64
	}{
		{"10pct", 0.10},
		{"25pct", 0.25},
		{"50pct", 0.50},
		{"75pct", 0.75},
		{"100pct", 1.00},
	}

	for _, cov := range coverages {
		b.Run(cov.name, func(b *testing.B) {
			start := opentime.NewRationalTime(0, 24)
			dur := opentime.NewRationalTime(totalDuration*cov.percent, 24)
			searchRange := opentime.NewTimeRange(start, dur)
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _, _, _ = itemsInRange(track, searchRange)
			}
		})
	}
}

// =============================================================================
// FilteredComposition Benchmarks
// =============================================================================

func BenchmarkFilteredComposition_TypeFilter(b *testing.B) {
	configs := []struct {
		tracks        int
		clipsPerTrack int
	}{
		{3, 50},
		{5, 100},
	}

	for _, c := range configs {
		name := fmt.Sprintf("tracks=%d_clips=%d", c.tracks, c.clipsPerTrack)
		b.Run(name, func(b *testing.B) {
			stack := createBenchmarkStackForAlgo(c.tracks, c.clipsPerTrack)
			filter := func(comp opentimelineio.Composable) []opentimelineio.Composable {
				if _, ok := comp.(*opentimelineio.Clip); ok {
					return []opentimelineio.Composable{comp}
				}
				return nil
			}
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = FilteredComposition(stack, filter, nil)
			}
		})
	}
}

func BenchmarkFilteredComposition_KeepAll(b *testing.B) {
	configs := []struct {
		tracks        int
		clipsPerTrack int
	}{
		{3, 50},
		{5, 100},
	}

	for _, c := range configs {
		name := fmt.Sprintf("tracks=%d_clips=%d", c.tracks, c.clipsPerTrack)
		b.Run(name, func(b *testing.B) {
			stack := createBenchmarkStackForAlgo(c.tracks, c.clipsPerTrack)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = FilteredComposition(stack, KeepFilter, nil)
			}
		})
	}
}

// =============================================================================
// TrackTrimmedToRange Benchmarks
// =============================================================================

func BenchmarkTrackTrimmedToRange(b *testing.B) {
	scales := []int{100, 500, 1000}
	for _, n := range scales {
		b.Run(fmt.Sprintf("clips=%d", n), func(b *testing.B) {
			track := createBenchmarkTrackForAlgo(n)
			start := opentime.NewRationalTime(float64(n)*24*0.25, 24)
			dur := opentime.NewRationalTime(float64(n)*24*0.5, 24)
			trimRange := opentime.NewTimeRange(start, dur)

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = TrackTrimmedToRange(track, trimRange)
			}
		})
	}
}

// =============================================================================
// TopClipAtTime Benchmarks
// =============================================================================

func BenchmarkTopClipAtTime(b *testing.B) {
	configs := []struct {
		tracks        int
		clipsPerTrack int
	}{
		{3, 50},
		{5, 100},
		{10, 100},
	}

	for _, c := range configs {
		name := fmt.Sprintf("tracks=%d_clips=%d", c.tracks, c.clipsPerTrack)
		b.Run(name, func(b *testing.B) {
			stack := createBenchmarkStackForAlgo(c.tracks, c.clipsPerTrack)
			midTime := opentime.NewRationalTime(float64(c.clipsPerTrack)*24/2, 24)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = TopClipAtTime(stack, midTime)
			}
		})
	}
}

// =============================================================================
// Edit Operation Benchmarks
// =============================================================================

func BenchmarkInsert(b *testing.B) {
	scales := []int{100, 500}
	for _, n := range scales {
		b.Run(fmt.Sprintf("clips=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				track := createBenchmarkTrackForAlgo(n)
				sr := opentime.NewTimeRange(
					opentime.NewRationalTime(0, 24),
					opentime.NewRationalTime(24, 24),
				)
				clip := opentimelineio.NewClip("insert", nil, &sr, nil, nil, nil, "", nil)
				midTime := opentime.NewRationalTime(float64(n)*24/2, 24)
				b.StartTimer()
				_ = Insert(clip, track, midTime)
			}
		})
	}
}

func BenchmarkRemoveRange(b *testing.B) {
	scales := []int{100, 500}
	for _, n := range scales {
		b.Run(fmt.Sprintf("clips=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				track := createBenchmarkTrackForAlgo(n)
				start := opentime.NewRationalTime(float64(n)*24*0.25, 24)
				dur := opentime.NewRationalTime(float64(n)*24*0.25, 24)
				removeRange := opentime.NewTimeRange(start, dur)
				b.StartTimer()
				_ = RemoveRange(track, removeRange)
			}
		})
	}
}

func BenchmarkSlice(b *testing.B) {
	scales := []int{100, 500}
	for _, n := range scales {
		b.Run(fmt.Sprintf("clips=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				track := createBenchmarkTrackForAlgo(n)
				// Slice in the middle of a clip
				sliceTime := opentime.NewRationalTime(float64(n/2)*24+12, 24)
				b.StartTimer()
				_ = Slice(track, sliceTime)
			}
		})
	}
}

func BenchmarkOverwrite(b *testing.B) {
	scales := []int{100, 500}
	for _, n := range scales {
		b.Run(fmt.Sprintf("clips=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				track := createBenchmarkTrackForAlgo(n)
				sr := opentime.NewTimeRange(
					opentime.NewRationalTime(0, 24),
					opentime.NewRationalTime(48, 24),
				)
				clip := opentimelineio.NewClip("overwrite", nil, &sr, nil, nil, nil, "", nil)
				// Create a time range for the overwrite (start at middle, 48 frames duration)
				overwriteRange := opentime.NewTimeRange(
					opentime.NewRationalTime(float64(n)*24/2, 24),
					opentime.NewRationalTime(48, 24),
				)
				b.StartTimer()
				_ = Overwrite(clip, track, overwriteRange)
			}
		})
	}
}

func BenchmarkFill(b *testing.B) {
	// Create a track with gaps
	createTrackWithGaps := func(n int) *opentimelineio.Track {
		track := opentimelineio.NewTrack("gapped", nil, opentimelineio.TrackKindVideo, nil, nil)
		for i := 0; i < n; i++ {
			if i%2 == 0 {
				// Add clip
				sr := opentime.NewTimeRange(
					opentime.NewRationalTime(0, 24),
					opentime.NewRationalTime(24, 24),
				)
				clip := opentimelineio.NewClip(fmt.Sprintf("clip_%d", i), nil, &sr, nil, nil, nil, "", nil)
				track.AppendChild(clip)
			} else {
				// Add gap
				gap := opentimelineio.NewGapWithDuration(opentime.NewRationalTime(24, 24))
				track.AppendChild(gap)
			}
		}
		return track
	}

	scales := []int{50, 100}
	for _, n := range scales {
		b.Run(fmt.Sprintf("items=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				track := createTrackWithGaps(n)
				sr := opentime.NewTimeRange(
					opentime.NewRationalTime(0, 24),
					opentime.NewRationalTime(24, 24),
				)
				fillClip := opentimelineio.NewClip("fill", nil, &sr, nil, nil, nil, "", nil)
				// Fill at the first gap (index 1, so time = 24 frames)
				gapTime := opentime.NewRationalTime(24, 24)
				b.StartTimer()
				_ = Fill(fillClip, track, gapTime, ReferencePointSequence)
			}
		})
	}
}

// =============================================================================
// compositionDuration Benchmark
// =============================================================================

func BenchmarkCompositionDuration(b *testing.B) {
	scales := []int{100, 500, 1000}
	for _, n := range scales {
		b.Run(fmt.Sprintf("clips=%d", n), func(b *testing.B) {
			track := createBenchmarkTrackForAlgo(n)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = compositionDuration(track)
			}
		})
	}
}
