// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package gotio

import (
	"testing"

	"github.com/Avalanche-io/gotio/opentime"
)

func TestTransformedTime_NilTarget(t *testing.T) {
	// When target is nil, time should be returned unchanged
	clip := NewClip("test", nil, nil, nil, nil, nil, "", nil)
	time := opentime.NewRationalTime(100, 24)

	result, err := clip.TransformedTime(time, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Equal(time) {
		t.Errorf("expected %v, got %v", time, result)
	}
}

func TestTransformedTime_SameItem(t *testing.T) {
	// When source and target are the same, time should be unchanged
	clip := NewClip("test", nil, nil, nil, nil, nil, "", nil)
	time := opentime.NewRationalTime(100, 24)

	result, err := clip.TransformedTime(time, clip)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Equal(time) {
		t.Errorf("expected %v, got %v", time, result)
	}
}

func TestTransformedTime_ClipToTrack(t *testing.T) {
	// Create a track with two clips
	// Clip1: 0-100 frames (source range starts at frame 10)
	// Clip2: 100-200 frames
	track := NewTrack("V1", nil, TrackKindVideo, nil, nil)

	// Clip 1: source range 10-110 (100 frames), placed at track time 0-100
	sr1 := opentime.NewTimeRange(
		opentime.NewRationalTime(10, 24),
		opentime.NewRationalTime(100, 24),
	)
	clip1 := NewClip("clip1", nil, &sr1, nil, nil, nil, "", nil)

	// Clip 2: source range 0-100 (100 frames), placed at track time 100-200
	sr2 := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(100, 24),
	)
	clip2 := NewClip("clip2", nil, &sr2, nil, nil, nil, "", nil)

	track.AppendChild(clip1)
	track.AppendChild(clip2)

	// Test: Transform time from clip1 to track
	// Clip1's internal time 50 (which is frame 60 in source = 10 + 50)
	// should map to track time 50 (frame 50 in track)
	clipTime := opentime.NewRationalTime(50, 24)
	result, err := clip1.TransformedTime(clipTime, track)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// In clip1's coordinate space, time 50 is 50 frames from start of trimmed range
	// The clip starts at track position 0, so track time should be:
	// (50 - clip1.trimmedRange.startTime) + rangeInParent.startTime
	// = (50 - 10) + 0 = 40
	expected := opentime.NewRationalTime(40, 24)
	if !result.Equal(expected) {
		t.Errorf("clip1 time 50 -> track: expected %v, got %v", expected, result)
	}
}

func TestTransformedTime_TrackToClip(t *testing.T) {
	track := NewTrack("V1", nil, TrackKindVideo, nil, nil)

	// Clip 1: source range 10-110 (100 frames)
	sr1 := opentime.NewTimeRange(
		opentime.NewRationalTime(10, 24),
		opentime.NewRationalTime(100, 24),
	)
	clip1 := NewClip("clip1", nil, &sr1, nil, nil, nil, "", nil)

	// Clip 2: source range 0-100 (100 frames)
	sr2 := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(100, 24),
	)
	clip2 := NewClip("clip2", nil, &sr2, nil, nil, nil, "", nil)

	track.AppendChild(clip1)
	track.AppendChild(clip2)

	// Test: Transform time from track to clip2
	// Track time 150 should map to clip2's time 50
	trackTime := opentime.NewRationalTime(150, 24)
	result, err := track.TransformedTime(trackTime, clip2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Track time 150 is in clip2's range (100-200 in track space)
	// clip2's rangeInParent starts at 100
	// clip2's trimmedRange starts at 0
	// So: (150 - 100) + 0 = 50
	expected := opentime.NewRationalTime(50, 24)
	if !result.Equal(expected) {
		t.Errorf("track time 150 -> clip2: expected %v, got %v", expected, result)
	}
}

func TestTransformedTime_ClipToClip(t *testing.T) {
	track := NewTrack("V1", nil, TrackKindVideo, nil, nil)

	// Clip 1: source range 10-110 (100 frames), at track 0-100
	sr1 := opentime.NewTimeRange(
		opentime.NewRationalTime(10, 24),
		opentime.NewRationalTime(100, 24),
	)
	clip1 := NewClip("clip1", nil, &sr1, nil, nil, nil, "", nil)

	// Clip 2: source range 20-120 (100 frames), at track 100-200
	sr2 := opentime.NewTimeRange(
		opentime.NewRationalTime(20, 24),
		opentime.NewRationalTime(100, 24),
	)
	clip2 := NewClip("clip2", nil, &sr2, nil, nil, nil, "", nil)

	track.AppendChild(clip1)
	track.AppendChild(clip2)

	// Test: Transform time from clip1 to clip2
	// Clip1's time 60 (frame 60 in its source, which is 50 frames from trimmed start)
	// Maps to track time 50 (since clip1 starts at 0 in track)
	// Then maps to clip2's time: (50 - 100) + 20 = -30 (out of clip2's range)
	// But the algorithm doesn't clip, it just transforms
	clipTime := opentime.NewRationalTime(60, 24)
	result, err := clip1.TransformedTime(clipTime, clip2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Clip1 time 60 -> track: (60 - 10) + 0 = 50
	// Track 50 -> clip2: (50 - 100) + 20 = -30
	expected := opentime.NewRationalTime(-30, 24)
	if !result.Equal(expected) {
		t.Errorf("clip1 time 60 -> clip2: expected %v, got %v", expected, result)
	}
}

func TestTransformedTime_NestedStack(t *testing.T) {
	// Create a nested hierarchy: track -> stack -> track -> clip
	outerTrack := NewTrack("V1", nil, TrackKindVideo, nil, nil)

	// Create a stack that will be placed in the outer track
	stack := NewStack("stack1", nil, nil, nil, nil, nil)

	// Create an inner track inside the stack
	innerTrack := NewTrack("inner", nil, TrackKindVideo, nil, nil)

	// Create a clip with source range 100-200
	sr := opentime.NewTimeRange(
		opentime.NewRationalTime(100, 24),
		opentime.NewRationalTime(100, 24),
	)
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	// Build hierarchy
	innerTrack.AppendChild(clip)
	stack.AppendChild(innerTrack)
	outerTrack.AppendChild(stack)

	// Test transform from clip to outerTrack
	// Clip time 150 (50 frames from start of clip's source range)
	clipTime := opentime.NewRationalTime(150, 24)
	result, err := clip.TransformedTime(clipTime, outerTrack)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// clip time 150 -> innerTrack: (150 - 100) + 0 = 50
	// innerTrack time 50 -> stack: (50 - 0) + 0 = 50 (stack children start at 0)
	// stack time 50 -> outerTrack: (50 - 0) + 0 = 50
	expected := opentime.NewRationalTime(50, 24)
	if !result.Equal(expected) {
		t.Errorf("nested clip time 150 -> outerTrack: expected %v, got %v", expected, result)
	}
}

func TestTransformedTimeRange(t *testing.T) {
	track := NewTrack("V1", nil, TrackKindVideo, nil, nil)

	sr := opentime.NewTimeRange(
		opentime.NewRationalTime(10, 24),
		opentime.NewRationalTime(100, 24),
	)
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)

	// Transform a time range from clip to track
	clipRange := opentime.NewTimeRange(
		opentime.NewRationalTime(30, 24),
		opentime.NewRationalTime(50, 24),
	)

	result, err := clip.TransformedTimeRange(clipRange, track)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Start time: (30 - 10) + 0 = 20
	// Duration should be preserved: 50
	expectedStart := opentime.NewRationalTime(20, 24)
	expectedDuration := opentime.NewRationalTime(50, 24)

	if !result.StartTime().Equal(expectedStart) {
		t.Errorf("transformed range start: expected %v, got %v", expectedStart, result.StartTime())
	}
	if !result.Duration().Equal(expectedDuration) {
		t.Errorf("transformed range duration: expected %v, got %v", expectedDuration, result.Duration())
	}
}

func TestHighestAncestor(t *testing.T) {
	// Test that highestAncestor returns the root
	track := NewTrack("V1", nil, TrackKindVideo, nil, nil)
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)
	track.AppendChild(clip)

	// Clip's highest ancestor should be track (since track has no parent)
	clipRoot := clip.highestAncestor()
	if clipRoot != track {
		t.Errorf("clip's highestAncestor should be track, got %T", clipRoot)
	}

	// Track's highest ancestor should be itself
	trackRoot := track.highestAncestor()
	if trackRoot != track {
		t.Errorf("track's highestAncestor should be track, got %T", trackRoot)
	}
}

func TestTransformedTime_NoParent(t *testing.T) {
	// When item has no parent, transformation should work (item is its own root)
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)
	time := opentime.NewRationalTime(100, 24)

	result, err := clip.TransformedTime(time, clip)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Equal(time) {
		t.Errorf("expected %v, got %v", time, result)
	}
}

func TestTransformedTime_MultipleClipsInTrack(t *testing.T) {
	// Test transformations with gaps between clips
	track := NewTrack("V1", nil, TrackKindVideo, nil, nil)

	// Add a gap at the beginning
	gap := NewGapWithDuration(opentime.NewRationalTime(50, 24))
	track.AppendChild(gap)

	// Clip with source range 0-100
	sr := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(100, 24),
	)
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)

	// Clip time 25 should map to track time 75 (50 from gap + 25 from clip)
	clipTime := opentime.NewRationalTime(25, 24)
	result, err := clip.TransformedTime(clipTime, track)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := opentime.NewRationalTime(75, 24)
	if !result.Equal(expected) {
		t.Errorf("clip time 25 -> track (after gap): expected %v, got %v", expected, result)
	}
}
