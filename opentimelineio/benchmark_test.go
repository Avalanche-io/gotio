// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"fmt"
	"testing"

	"github.com/Avalanche-io/gotio/opentime"
)

// =============================================================================
// Benchmark Helpers
// =============================================================================

// createBenchmarkTrack creates a track with n clips, each 24 frames at 24fps.
func createBenchmarkTrack(n int) *Track {
	track := NewTrack("bench_track", nil, TrackKindVideo, nil, nil)
	for i := 0; i < n; i++ {
		sr := opentime.NewTimeRange(
			opentime.NewRationalTime(0, 24),
			opentime.NewRationalTime(24, 24),
		)
		clip := NewClip(fmt.Sprintf("clip_%d", i), nil, &sr, nil, nil, nil, "", nil)
		track.AppendChild(clip)
	}
	return track
}

// createSimpleClip creates a minimal clip for insertion tests.
func createSimpleClip(name string) *Clip {
	sr := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(24, 24),
	)
	return NewClip(name, nil, &sr, nil, nil, nil, "", nil)
}

// createBenchmarkStack creates a stack with the specified number of tracks and clips per track.
func createBenchmarkStack(tracks, clipsPerTrack int) *Stack {
	stack := NewStack("bench_stack", nil, nil, nil, nil, nil)
	for t := 0; t < tracks; t++ {
		track := createBenchmarkTrack(clipsPerTrack)
		track.SetName(fmt.Sprintf("track_%d", t))
		stack.AppendChild(track)
	}
	return stack
}

// createBenchmarkTimeline creates a timeline with video and audio tracks.
func createBenchmarkTimeline(videoTracks, audioTracks, clipsPerTrack int) *Timeline {
	timeline := NewTimeline("bench_timeline", nil, nil)
	stack := timeline.Tracks()

	for v := 0; v < videoTracks; v++ {
		track := createBenchmarkTrack(clipsPerTrack)
		track.SetName(fmt.Sprintf("V%d", v+1))
		track.SetKind(TrackKindVideo)
		stack.AppendChild(track)
	}

	for a := 0; a < audioTracks; a++ {
		track := createBenchmarkTrack(clipsPerTrack)
		track.SetName(fmt.Sprintf("A%d", a+1))
		track.SetKind(TrackKindAudio)
		stack.AppendChild(track)
	}

	return timeline
}

// createBenchmarkClipWithMetadata creates a clip with realistic metadata.
func createBenchmarkClipWithMetadata() *Clip {
	sr := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(240, 24),
	)
	ar := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(10000, 24),
	)
	ref := NewExternalReference("source", "file:///media/source.mov", &ar, nil)
	metadata := AnyDictionary{
		"author":      "benchmark",
		"version":     1,
		"description": "A clip with realistic metadata for benchmarking purposes",
		"tags":        []string{"test", "benchmark", "performance"},
	}
	return NewClip("clip_with_metadata", ref, &sr, metadata, nil, nil, "", nil)
}

// =============================================================================
// RangeOfChildAtIndex Benchmarks - O(n) per call hotspot
// =============================================================================

func BenchmarkTrack_RangeOfChildAtIndex(b *testing.B) {
	scales := []int{10, 100, 1000, 10000}
	for _, n := range scales {
		b.Run(fmt.Sprintf("clips=%d", n), func(b *testing.B) {
			track := createBenchmarkTrack(n)
			idx := n / 2 // Query middle element
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = track.RangeOfChildAtIndex(idx)
			}
		})
	}
}

func BenchmarkTrack_RangeOfChildAtIndex_ByPosition(b *testing.B) {
	n := 1000
	track := createBenchmarkTrack(n)
	positions := []struct {
		name  string
		index int
	}{
		{"first", 0},
		{"quarter", n / 4},
		{"middle", n / 2},
		{"three_quarter", 3 * n / 4},
		{"last", n - 1},
	}

	for _, pos := range positions {
		b.Run(pos.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = track.RangeOfChildAtIndex(pos.index)
			}
		})
	}
}

func BenchmarkTrack_RangeOfChildAtIndex_AllIndices(b *testing.B) {
	// Tests the O(n^2) pattern of iterating and calling RangeOfChildAtIndex
	scales := []int{10, 100, 500}
	for _, n := range scales {
		b.Run(fmt.Sprintf("clips=%d", n), func(b *testing.B) {
			track := createBenchmarkTrack(n)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				for j := 0; j < n; j++ {
					_, _ = track.RangeOfChildAtIndex(j)
				}
			}
		})
	}
}

func BenchmarkStack_RangeOfChildAtIndex(b *testing.B) {
	// Stack should be O(1) since all children start at time 0
	scales := []int{10, 100, 1000}
	for _, n := range scales {
		b.Run(fmt.Sprintf("tracks=%d", n), func(b *testing.B) {
			stack := createBenchmarkStack(n, 10)
			idx := n / 2
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = stack.RangeOfChildAtIndex(idx)
			}
		})
	}
}

// =============================================================================
// ChildrenInRange Benchmarks - O(n^2) due to RangeOfChildAtIndex calls
// =============================================================================

func BenchmarkTrack_ChildrenInRange(b *testing.B) {
	scales := []int{10, 100, 1000}
	for _, n := range scales {
		b.Run(fmt.Sprintf("clips=%d", n), func(b *testing.B) {
			track := createBenchmarkTrack(n)
			// Search range covering middle 50%
			start := opentime.NewRationalTime(float64(n)*24*0.25, 24)
			dur := opentime.NewRationalTime(float64(n)*24*0.5, 24)
			searchRange := opentime.NewTimeRange(start, dur)

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = track.ChildrenInRange(searchRange)
			}
		})
	}
}

func BenchmarkTrack_ChildrenInRange_Coverage(b *testing.B) {
	// Test different coverage amounts
	n := 1000
	track := createBenchmarkTrack(n)
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
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = track.ChildrenInRange(searchRange)
			}
		})
	}
}

// =============================================================================
// Clone Benchmarks - O(n*d) recursive deep copy
// =============================================================================

func BenchmarkClip_Clone(b *testing.B) {
	clip := createSimpleClip("bench")
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = clip.Clone()
	}
}

func BenchmarkClip_Clone_WithMetadata(b *testing.B) {
	clip := createBenchmarkClipWithMetadata()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = clip.Clone()
	}
}

func BenchmarkTrack_Clone(b *testing.B) {
	scales := []int{10, 100, 1000}
	for _, n := range scales {
		b.Run(fmt.Sprintf("clips=%d", n), func(b *testing.B) {
			track := createBenchmarkTrack(n)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = track.Clone()
			}
		})
	}
}

func BenchmarkStack_Clone(b *testing.B) {
	configs := []struct {
		tracks        int
		clipsPerTrack int
	}{
		{2, 10},
		{5, 50},
		{10, 100},
	}

	for _, c := range configs {
		name := fmt.Sprintf("tracks=%d_clips=%d", c.tracks, c.clipsPerTrack)
		b.Run(name, func(b *testing.B) {
			stack := createBenchmarkStack(c.tracks, c.clipsPerTrack)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = stack.Clone()
			}
		})
	}
}

func BenchmarkTimeline_Clone(b *testing.B) {
	configs := []struct {
		videoTracks   int
		audioTracks   int
		clipsPerTrack int
	}{
		{1, 1, 10},
		{3, 2, 50},
		{5, 5, 100},
	}

	for _, c := range configs {
		name := fmt.Sprintf("v=%d_a=%d_clips=%d", c.videoTracks, c.audioTracks, c.clipsPerTrack)
		b.Run(name, func(b *testing.B) {
			timeline := createBenchmarkTimeline(c.videoTracks, c.audioTracks, c.clipsPerTrack)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = timeline.Clone()
			}
		})
	}
}

// =============================================================================
// FindChildren Benchmarks - O(n*d) recursive with filter
// =============================================================================

func BenchmarkTrack_FindChildren_ShallowSearch(b *testing.B) {
	scales := []int{100, 1000}
	for _, n := range scales {
		b.Run(fmt.Sprintf("clips=%d", n), func(b *testing.B) {
			track := createBenchmarkTrack(n)
			filter := func(c Composable) bool {
				_, ok := c.(*Clip)
				return ok
			}
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = track.FindChildren(nil, true, filter)
			}
		})
	}
}

func BenchmarkTrack_FindChildren_NoFilter(b *testing.B) {
	scales := []int{100, 1000}
	for _, n := range scales {
		b.Run(fmt.Sprintf("clips=%d", n), func(b *testing.B) {
			track := createBenchmarkTrack(n)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = track.FindChildren(nil, true, nil)
			}
		})
	}
}

func BenchmarkStack_FindChildren_DeepSearch(b *testing.B) {
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
			stack := createBenchmarkStack(c.tracks, c.clipsPerTrack)
			filter := func(c Composable) bool {
				_, ok := c.(*Clip)
				return ok
			}
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = stack.FindChildren(nil, false, filter)
			}
		})
	}
}

func BenchmarkTimeline_FindClips(b *testing.B) {
	configs := []struct {
		tracks        int
		clipsPerTrack int
	}{
		{2, 50},
		{5, 100},
	}

	for _, c := range configs {
		name := fmt.Sprintf("tracks=%d_clips=%d", c.tracks, c.clipsPerTrack)
		b.Run(name, func(b *testing.B) {
			timeline := createBenchmarkTimeline(c.tracks, 0, c.clipsPerTrack)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = timeline.FindClips(nil, false)
			}
		})
	}
}

// =============================================================================
// InsertChild/RemoveChild Benchmarks - O(n) slice operations
// =============================================================================

func BenchmarkTrack_InsertChild(b *testing.B) {
	scales := []int{10, 100, 1000}
	positions := []string{"start", "middle", "end"}

	for _, n := range scales {
		for _, pos := range positions {
			name := fmt.Sprintf("n=%d_pos=%s", n, pos)
			b.Run(name, func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					track := createBenchmarkTrack(n)
					clip := createSimpleClip("insert")
					var idx int
					switch pos {
					case "start":
						idx = 0
					case "middle":
						idx = n / 2
					case "end":
						idx = n
					}
					b.StartTimer()
					_ = track.InsertChild(idx, clip)
				}
			})
		}
	}
}

func BenchmarkTrack_RemoveChild(b *testing.B) {
	scales := []int{10, 100, 1000}
	positions := []string{"start", "middle", "end"}

	for _, n := range scales {
		for _, pos := range positions {
			name := fmt.Sprintf("n=%d_pos=%s", n, pos)
			b.Run(name, func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					track := createBenchmarkTrack(n)
					var idx int
					switch pos {
					case "start":
						idx = 0
					case "middle":
						idx = n / 2
					case "end":
						idx = n - 1
					}
					b.StartTimer()
					_ = track.RemoveChild(idx)
				}
			})
		}
	}
}

func BenchmarkTrack_AppendChild(b *testing.B) {
	scales := []int{10, 100, 1000}
	for _, n := range scales {
		b.Run(fmt.Sprintf("existing=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				track := createBenchmarkTrack(n)
				clip := createSimpleClip("append")
				b.StartTimer()
				_ = track.AppendChild(clip)
			}
		})
	}
}

// =============================================================================
// JSON Serialization Benchmarks
// =============================================================================

func BenchmarkClip_MarshalJSON(b *testing.B) {
	clip := createBenchmarkClipWithMetadata()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = ToJSONBytes(clip)
	}
}

func BenchmarkTrack_MarshalJSON(b *testing.B) {
	scales := []int{10, 100, 500}
	for _, n := range scales {
		b.Run(fmt.Sprintf("clips=%d", n), func(b *testing.B) {
			track := createBenchmarkTrack(n)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = ToJSONBytes(track)
			}
		})
	}
}

func BenchmarkTimeline_MarshalJSON(b *testing.B) {
	configs := []struct {
		tracks        int
		clipsPerTrack int
	}{
		{2, 10},
		{5, 50},
		{10, 100},
	}

	for _, c := range configs {
		name := fmt.Sprintf("tracks=%d_clips=%d", c.tracks, c.clipsPerTrack)
		b.Run(name, func(b *testing.B) {
			timeline := createBenchmarkTimeline(c.tracks, 0, c.clipsPerTrack)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = ToJSONBytes(timeline)
			}
		})
	}
}

func BenchmarkClip_UnmarshalJSON(b *testing.B) {
	clip := createBenchmarkClipWithMetadata()
	data, _ := ToJSONBytes(clip)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = FromJSONBytes(data)
	}
}

func BenchmarkTrack_UnmarshalJSON(b *testing.B) {
	scales := []int{10, 100, 500}
	for _, n := range scales {
		b.Run(fmt.Sprintf("clips=%d", n), func(b *testing.B) {
			track := createBenchmarkTrack(n)
			data, _ := ToJSONBytes(track)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = FromJSONBytes(data)
			}
		})
	}
}

func BenchmarkTimeline_UnmarshalJSON(b *testing.B) {
	configs := []struct {
		tracks        int
		clipsPerTrack int
	}{
		{2, 10},
		{5, 50},
		{10, 100},
	}

	for _, c := range configs {
		name := fmt.Sprintf("tracks=%d_clips=%d", c.tracks, c.clipsPerTrack)
		b.Run(name, func(b *testing.B) {
			timeline := createBenchmarkTimeline(c.tracks, 0, c.clipsPerTrack)
			data, _ := ToJSONBytes(timeline)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = FromJSONBytes(data)
			}
		})
	}
}

// =============================================================================
// Fast JSON Serialization Benchmarks (for comparison)
// =============================================================================

func BenchmarkTimeline_MarshalJSON_Fast(b *testing.B) {
	configs := []struct {
		tracks        int
		clipsPerTrack int
	}{
		{2, 10},
		{5, 50},
		{10, 100},
	}

	for _, c := range configs {
		name := fmt.Sprintf("tracks=%d_clips=%d", c.tracks, c.clipsPerTrack)
		b.Run(name, func(b *testing.B) {
			timeline := createBenchmarkTimeline(c.tracks, 0, c.clipsPerTrack)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = ToJSONBytes(timeline)
			}
		})
	}
}

func BenchmarkTimeline_UnmarshalJSON_Fast(b *testing.B) {
	configs := []struct {
		tracks        int
		clipsPerTrack int
	}{
		{2, 10},
		{5, 50},
		{10, 100},
	}

	for _, c := range configs {
		name := fmt.Sprintf("tracks=%d_clips=%d", c.tracks, c.clipsPerTrack)
		b.Run(name, func(b *testing.B) {
			timeline := createBenchmarkTimeline(c.tracks, 0, c.clipsPerTrack)
			data, _ := ToJSONBytes(timeline)
			sizeMB := float64(len(data)) / (1024 * 1024)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = FromJSONBytes(data)
			}
			b.ReportMetric(sizeMB*float64(b.N)/b.Elapsed().Seconds(), "MB/s")
		})
	}
}

// =============================================================================
// Duration Benchmarks
// =============================================================================

func BenchmarkTrack_Duration(b *testing.B) {
	scales := []int{10, 100, 1000}
	for _, n := range scales {
		b.Run(fmt.Sprintf("clips=%d", n), func(b *testing.B) {
			track := createBenchmarkTrack(n)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = track.Duration()
			}
		})
	}
}

func BenchmarkStack_Duration(b *testing.B) {
	configs := []struct {
		tracks        int
		clipsPerTrack int
	}{
		{2, 10},
		{5, 50},
		{10, 100},
	}

	for _, c := range configs {
		name := fmt.Sprintf("tracks=%d_clips=%d", c.tracks, c.clipsPerTrack)
		b.Run(name, func(b *testing.B) {
			stack := createBenchmarkStack(c.tracks, c.clipsPerTrack)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = stack.Duration()
			}
		})
	}
}

// =============================================================================
// ChildAtTime Benchmarks
// =============================================================================

func BenchmarkTrack_ChildAtTime(b *testing.B) {
	scales := []int{100, 1000}
	for _, n := range scales {
		b.Run(fmt.Sprintf("clips=%d", n), func(b *testing.B) {
			track := createBenchmarkTrack(n)
			midTime := opentime.NewRationalTime(float64(n)*24/2, 24)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = track.ChildAtTime(midTime, true)
			}
		})
	}
}

func BenchmarkStack_ChildAtTime(b *testing.B) {
	configs := []struct {
		tracks        int
		clipsPerTrack int
	}{
		{5, 50},
		{10, 100},
	}

	for _, c := range configs {
		name := fmt.Sprintf("tracks=%d_clips=%d", c.tracks, c.clipsPerTrack)
		b.Run(name, func(b *testing.B) {
			stack := createBenchmarkStack(c.tracks, c.clipsPerTrack)
			midTime := opentime.NewRationalTime(float64(c.clipsPerTrack)*24/2, 24)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = stack.ChildAtTime(midTime, true)
			}
		})
	}
}
