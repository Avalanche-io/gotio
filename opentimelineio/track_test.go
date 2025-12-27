// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"encoding/json"
	"testing"

	"github.com/Avalanche-io/gotio/opentime"
)

func TestTrackKind(t *testing.T) {
	// Default is Video
	track := NewTrack("V1", nil, "", nil, nil)
	if track.Kind() != TrackKindVideo {
		t.Errorf("Default Kind = %s, want %s", track.Kind(), TrackKindVideo)
	}

	// Set kind
	track.SetKind(TrackKindAudio)
	if track.Kind() != TrackKindAudio {
		t.Errorf("After SetKind = %s, want %s", track.Kind(), TrackKindAudio)
	}
}

func TestTrackCompositionKind(t *testing.T) {
	track := NewTrack("V1", nil, TrackKindVideo, nil, nil)

	if track.CompositionKind() != "Track" {
		t.Errorf("CompositionKind = %s, want Track", track.CompositionKind())
	}
}

func TestTrackSequentialRanges(t *testing.T) {
	track := NewTrack("V1", nil, TrackKindVideo, nil, nil)

	// Add 3 clips, each 1 second (24 frames)
	for i := 0; i < 3; i++ {
		sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
		clip := NewClip("", nil, &sr, nil, nil, nil, "", nil)
		track.AppendChild(clip)
	}

	// Check ranges - clips should be sequential
	r0, _ := track.RangeOfChildAtIndex(0)
	if r0.StartTime().Value() != 0 || r0.Duration().Value() != 24 {
		t.Errorf("Child 0 range: start=%v, duration=%v", r0.StartTime().Value(), r0.Duration().Value())
	}

	r1, _ := track.RangeOfChildAtIndex(1)
	if r1.StartTime().Value() != 24 || r1.Duration().Value() != 24 {
		t.Errorf("Child 1 range: start=%v, duration=%v", r1.StartTime().Value(), r1.Duration().Value())
	}

	r2, _ := track.RangeOfChildAtIndex(2)
	if r2.StartTime().Value() != 48 || r2.Duration().Value() != 24 {
		t.Errorf("Child 2 range: start=%v, duration=%v", r2.StartTime().Value(), r2.Duration().Value())
	}
}

func TestTrackTrimmedRangeOfChildAtIndex(t *testing.T) {
	// Create track with source range that trims children
	trackSr := opentime.NewTimeRange(opentime.NewRationalTime(12, 24), opentime.NewRationalTime(24, 24))
	track := NewTrack("V1", &trackSr, TrackKindVideo, nil, nil)

	// Add clip that starts at 0 with duration 48
	clipSr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := NewClip("clip", nil, &clipSr, nil, nil, nil, "", nil)
	track.AppendChild(clip)

	// Trimmed range should be intersection of child range and source range
	trimmed, err := track.TrimmedRangeOfChildAtIndex(0)
	if err != nil {
		t.Fatalf("TrimmedRangeOfChildAtIndex error: %v", err)
	}

	// Child is at 0-48, source range is 12-36, intersection is 12-36
	if trimmed.StartTime().Value() != 12 {
		t.Errorf("Trimmed start = %v, want 12", trimmed.StartTime().Value())
	}
	if trimmed.Duration().Value() != 24 {
		t.Errorf("Trimmed duration = %v, want 24", trimmed.Duration().Value())
	}
}

func TestTrackRangeOfAllChildren(t *testing.T) {
	track := NewTrack("V1", nil, TrackKindVideo, nil, nil)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip1 := NewClip("clip1", nil, &sr, nil, nil, nil, "", nil)
	clip2 := NewClip("clip2", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip1)
	track.AppendChild(clip2)

	ranges, err := track.RangeOfAllChildren()
	if err != nil {
		t.Fatalf("RangeOfAllChildren error: %v", err)
	}

	// First clip at 0, second at 24
	if ranges[clip1].StartTime().Value() != 0 {
		t.Errorf("clip1 start = %v, want 0", ranges[clip1].StartTime().Value())
	}
	if ranges[clip2].StartTime().Value() != 24 {
		t.Errorf("clip2 start = %v, want 24", ranges[clip2].StartTime().Value())
	}
}

func TestTrackHandlesOfChild(t *testing.T) {
	track := NewTrack("V1", nil, TrackKindVideo, nil, nil)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip1 := NewClip("clip1", nil, &sr, nil, nil, nil, "", nil)
	transition := NewTransition("", TransitionTypeSMPTEDissolve,
		opentime.NewRationalTime(6, 24),
		opentime.NewRationalTime(6, 24), nil)
	clip2 := NewClip("clip2", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip1)
	track.AppendChild(transition)
	track.AppendChild(clip2)

	// clip1 should have no inHandle but an outHandle
	inHandle, outHandle, err := track.HandlesOfChild(clip1)
	if err != nil {
		t.Fatalf("HandlesOfChild(clip1) error: %v", err)
	}
	if inHandle != nil {
		t.Error("clip1 should have no inHandle")
	}
	if outHandle == nil || outHandle.Value() != 6 {
		t.Errorf("clip1 outHandle = %v, want 6", outHandle)
	}

	// clip2 should have an inHandle but no outHandle
	inHandle, outHandle, err = track.HandlesOfChild(clip2)
	if err != nil {
		t.Fatalf("HandlesOfChild(clip2) error: %v", err)
	}
	if inHandle == nil || inHandle.Value() != 6 {
		t.Errorf("clip2 inHandle = %v, want 6", inHandle)
	}
	if outHandle != nil {
		t.Error("clip2 should have no outHandle")
	}
}

func TestTrackNeighborsOfWithGapPolicy(t *testing.T) {
	track := NewTrack("V1", nil, TrackKindVideo, nil, nil)

	transition := NewTransition("", TransitionTypeSMPTEDissolve,
		opentime.NewRationalTime(6, 24),
		opentime.NewRationalTime(6, 24), nil)
	track.AppendChild(transition)

	// With NeighborGapPolicyAroundTransitions, a lone transition should get gaps
	prev, next, err := track.NeighborsOf(transition, NeighborGapPolicyAroundTransitions)
	if err != nil {
		t.Fatalf("NeighborsOf error: %v", err)
	}
	if prev == nil {
		t.Error("prev should be a gap")
	}
	if _, ok := prev.(*Gap); !ok {
		t.Error("prev should be a Gap")
	}
	if next == nil {
		t.Error("next should be a gap")
	}
	if _, ok := next.(*Gap); !ok {
		t.Error("next should be a Gap")
	}
}

func TestTrackClone(t *testing.T) {
	track := NewTrack("V1", nil, TrackKindVideo, AnyDictionary{"key": "value"}, nil)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)

	clone := track.Clone().(*Track)

	if clone.Name() != "V1" {
		t.Errorf("Clone name = %s, want V1", clone.Name())
	}
	if clone.Kind() != TrackKindVideo {
		t.Errorf("Clone kind = %s, want %s", clone.Kind(), TrackKindVideo)
	}
	if len(clone.Children()) != 1 {
		t.Errorf("Clone children count = %d, want 1", len(clone.Children()))
	}

	// Verify deep copy
	clone.SetKind(TrackKindAudio)
	if track.Kind() == TrackKindAudio {
		t.Error("Modifying clone affected original")
	}
}

func TestTrackIsEquivalentTo(t *testing.T) {
	t1 := NewTrack("V1", nil, TrackKindVideo, nil, nil)
	t2 := NewTrack("V1", nil, TrackKindVideo, nil, nil)
	t3 := NewTrack("V1", nil, TrackKindAudio, nil, nil)

	if !t1.IsEquivalentTo(t2) {
		t.Error("Identical tracks should be equivalent")
	}
	if t1.IsEquivalentTo(t3) {
		t.Error("Tracks with different kinds should not be equivalent")
	}

	// Test with non-Track
	stack := NewStack("stack", nil, nil, nil, nil, nil)
	if t1.IsEquivalentTo(stack) {
		t.Error("Track should not be equivalent to Stack")
	}
}

func TestTrackJSONWithKind(t *testing.T) {
	track := NewTrack("A1", nil, TrackKindAudio, nil, nil)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 48000), opentime.NewRationalTime(48000, 48000))
	clip := NewClip("audio", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)

	data, err := json.Marshal(track)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	track2 := &Track{}
	if err := json.Unmarshal(data, track2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if track2.Name() != "A1" {
		t.Errorf("Name mismatch: got %s", track2.Name())
	}
	if track2.Kind() != TrackKindAudio {
		t.Errorf("Kind mismatch: got %s", track2.Kind())
	}
	if len(track2.Children()) != 1 {
		t.Errorf("Children count mismatch: got %d", len(track2.Children()))
	}
}

func TestTrackAvailableImageBounds(t *testing.T) {
	track := NewTrack("V1", nil, TrackKindVideo, nil, nil)

	// Test empty track
	bounds, err := track.AvailableImageBounds()
	if err != nil {
		t.Fatalf("AvailableImageBounds error: %v", err)
	}
	if bounds != nil {
		t.Error("Empty track should have nil bounds")
	}
}

func TestTrackSchema(t *testing.T) {
	track := NewTrack("V1", nil, TrackKindVideo, nil, nil)

	if track.SchemaName() != "Track" {
		t.Errorf("SchemaName = %s, want Track", track.SchemaName())
	}
	if track.SchemaVersion() != 1 {
		t.Errorf("SchemaVersion = %d, want 1", track.SchemaVersion())
	}
}
