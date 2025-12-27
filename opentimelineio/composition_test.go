// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"testing"

	"github.com/Avalanche-io/gotio/opentime"
)

func TestCompositionChildManagement(t *testing.T) {
	comp := NewComposition("test", nil, nil, nil, nil, nil)

	// Test AppendChild
	clip1 := NewClip("clip1", nil, nil, nil, nil, nil, "", nil)
	clip2 := NewClip("clip2", nil, nil, nil, nil, nil, "", nil)
	clip3 := NewClip("clip3", nil, nil, nil, nil, nil, "", nil)

	if err := comp.AppendChild(clip1); err != nil {
		t.Fatalf("AppendChild error: %v", err)
	}
	if err := comp.AppendChild(clip2); err != nil {
		t.Fatalf("AppendChild error: %v", err)
	}

	if len(comp.Children()) != 2 {
		t.Errorf("len(Children()) = %d, want 2", len(comp.Children()))
	}

	// Test InsertChild
	if err := comp.InsertChild(1, clip3); err != nil {
		t.Fatalf("InsertChild error: %v", err)
	}
	if len(comp.Children()) != 3 {
		t.Errorf("len(Children()) = %d, want 3", len(comp.Children()))
	}
	if comp.Children()[1] != clip3 {
		t.Error("InsertChild should insert at index 1")
	}

	// Test IndexOfChild
	idx, err := comp.IndexOfChild(clip3)
	if err != nil {
		t.Fatalf("IndexOfChild error: %v", err)
	}
	if idx != 1 {
		t.Errorf("IndexOfChild = %d, want 1", idx)
	}

	// Test HasChild
	if !comp.HasChild(clip1) {
		t.Error("HasChild(clip1) should be true")
	}

	clip4 := NewClip("clip4", nil, nil, nil, nil, nil, "", nil)
	if comp.HasChild(clip4) {
		t.Error("HasChild(clip4) should be false")
	}

	// Test SetChild
	if err := comp.SetChild(0, clip4); err != nil {
		t.Fatalf("SetChild error: %v", err)
	}
	if comp.Children()[0] != clip4 {
		t.Error("SetChild should replace child at index 0")
	}

	// Test RemoveChild
	if err := comp.RemoveChild(1); err != nil {
		t.Fatalf("RemoveChild error: %v", err)
	}
	if len(comp.Children()) != 2 {
		t.Errorf("len(Children()) = %d, want 2", len(comp.Children()))
	}

	// Test ClearChildren
	comp.ClearChildren()
	if len(comp.Children()) != 0 {
		t.Errorf("len(Children()) = %d, want 0", len(comp.Children()))
	}
}

func TestCompositionIndexErrors(t *testing.T) {
	comp := NewComposition("test", nil, nil, nil, nil, nil)
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)
	comp.AppendChild(clip)

	// Test invalid indices
	if err := comp.InsertChild(-1, clip); err == nil {
		t.Error("InsertChild(-1) should error")
	}
	if err := comp.InsertChild(10, clip); err == nil {
		t.Error("InsertChild(10) should error")
	}
	if err := comp.SetChild(-1, clip); err == nil {
		t.Error("SetChild(-1) should error")
	}
	if err := comp.SetChild(10, clip); err == nil {
		t.Error("SetChild(10) should error")
	}
	if err := comp.RemoveChild(-1); err == nil {
		t.Error("RemoveChild(-1) should error")
	}
	if err := comp.RemoveChild(10); err == nil {
		t.Error("RemoveChild(10) should error")
	}
	if _, err := comp.RangeOfChildAtIndex(-1); err == nil {
		t.Error("RangeOfChildAtIndex(-1) should error")
	}
	if _, err := comp.RangeOfChildAtIndex(10); err == nil {
		t.Error("RangeOfChildAtIndex(10) should error")
	}
}

func TestCompositionFindChildren(t *testing.T) {
	timeline := NewTimeline("test", nil, nil)
	track := NewTrack("V1", nil, TrackKindVideo, nil, nil)

	// Add clips and gaps
	for i := 0; i < 3; i++ {
		sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
		clip := NewClip("", nil, &sr, nil, nil, nil, "", nil)
		track.AppendChild(clip)

		if i < 2 {
			gap := NewGapWithDuration(opentime.NewRationalTime(12, 24))
			track.AppendChild(gap)
		}
	}
	timeline.Tracks().AppendChild(track)

	// Test FindClips
	clips := timeline.FindClips(nil, false)
	if len(clips) != 3 {
		t.Errorf("len(FindClips) = %d, want 3", len(clips))
	}

	// Test FindChildren with filter
	children := track.FindChildren(nil, true, func(c Composable) bool {
		_, ok := c.(*Gap)
		return ok
	})
	if len(children) != 2 {
		t.Errorf("len(FindChildren for gaps) = %d, want 2", len(children))
	}
}

func TestCompositionChildAtTime(t *testing.T) {
	track := NewTrack("test", nil, TrackKindVideo, nil, nil)

	// Add clips: each 1 second (24 frames at 24fps)
	for i := 0; i < 3; i++ {
		sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
		clip := NewClip("", nil, &sr, nil, nil, nil, "", nil)
		track.AppendChild(clip)
	}

	// Test ChildAtTime
	tests := []struct {
		time  float64
		index int
	}{
		{0.0, 0},  // First clip
		{0.5, 0},  // Still first clip
		{1.0, 1},  // Second clip
		{1.5, 1},  // Still second clip
		{2.0, 2},  // Third clip
	}

	for _, tc := range tests {
		searchTime := opentime.NewRationalTime(tc.time*24, 24)
		child, err := track.ChildAtTime(searchTime, true)
		if err != nil {
			t.Fatalf("ChildAtTime(%v) error: %v", tc.time, err)
		}
		idx, _ := track.IndexOfChild(child)
		if idx != tc.index {
			t.Errorf("ChildAtTime(%v) index = %d, want %d", tc.time, idx, tc.index)
		}
	}

	// Test time outside range
	child, _ := track.ChildAtTime(opentime.NewRationalTime(100, 24), true)
	if child != nil {
		t.Error("ChildAtTime outside range should return nil")
	}
}

func TestCompositionChildrenInRange(t *testing.T) {
	track := NewTrack("test", nil, TrackKindVideo, nil, nil)

	// Add 3 clips, each 1 second
	for i := 0; i < 3; i++ {
		sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
		clip := NewClip("", nil, &sr, nil, nil, nil, "", nil)
		track.AppendChild(clip)
	}

	// Search for clips in range 0.5 to 1.5 seconds (should overlap clips 0 and 1)
	searchRange := opentime.NewTimeRange(
		opentime.NewRationalTime(12, 24),
		opentime.NewRationalTime(24, 24),
	)
	children, err := track.ChildrenInRange(searchRange)
	if err != nil {
		t.Fatalf("ChildrenInRange error: %v", err)
	}
	if len(children) != 2 {
		t.Errorf("len(ChildrenInRange) = %d, want 2", len(children))
	}
}

func TestTrackNeighbors(t *testing.T) {
	track := NewTrack("test", nil, TrackKindVideo, nil, nil)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip1 := NewClip("clip1", nil, &sr, nil, nil, nil, "", nil)
	clip2 := NewClip("clip2", nil, &sr, nil, nil, nil, "", nil)
	clip3 := NewClip("clip3", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip1)
	track.AppendChild(clip2)
	track.AppendChild(clip3)

	// Test neighbors of middle clip
	prev, next, err := track.NeighborsOf(clip2, NeighborGapPolicyNever)
	if err != nil {
		t.Fatalf("NeighborsOf error: %v", err)
	}
	if prev != clip1 {
		t.Error("prev should be clip1")
	}
	if next != clip3 {
		t.Error("next should be clip3")
	}

	// Test neighbors of first clip
	prev, next, _ = track.NeighborsOf(clip1, NeighborGapPolicyNever)
	if prev != nil {
		t.Error("prev of first clip should be nil")
	}
	if next != clip2 {
		t.Error("next of first clip should be clip2")
	}
}

func TestTrackWithTransition(t *testing.T) {
	track := NewTrack("test", nil, TrackKindVideo, nil, nil)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip1 := NewClip("clip1", nil, &sr, nil, nil, nil, "", nil)
	clip2 := NewClip("clip2", nil, &sr, nil, nil, nil, "", nil)
	transition := NewTransition("dissolve", TransitionTypeSMPTEDissolve,
		opentime.NewRationalTime(12, 24),
		opentime.NewRationalTime(12, 24),
		nil)

	track.AppendChild(clip1)
	track.AppendChild(transition)
	track.AppendChild(clip2)

	// Transition should not be visible (doesn't take up time)
	if transition.Visible() {
		t.Error("Transition.Visible() should be false")
	}

	// Track duration should only count visible children
	dur, _ := track.Duration()
	expectedSeconds := 4.0 // 2 clips * 2 seconds each
	if dur.ToSeconds() != expectedSeconds {
		t.Errorf("Track Duration = %v seconds, want %v", dur.ToSeconds(), expectedSeconds)
	}

	// Test handles
	inHandle, outHandle, err := track.HandlesOfChild(clip2)
	if err != nil {
		t.Fatalf("HandlesOfChild error: %v", err)
	}
	if inHandle == nil {
		t.Error("clip2 should have inHandle from transition")
	} else if inHandle.Value() != 12 {
		t.Errorf("inHandle.Value() = %v, want 12", inHandle.Value())
	}
	if outHandle != nil {
		t.Error("clip2 should not have outHandle")
	}
}

func TestCompositionHasClips(t *testing.T) {
	timeline := NewTimeline("test", nil, nil)
	track := NewTrack("V1", nil, TrackKindVideo, nil, nil)
	timeline.Tracks().AppendChild(track)

	// Empty track has no clips
	if timeline.Tracks().HasClips() {
		t.Error("Empty track should not have clips")
	}

	// Add a gap - still no clips
	gap := NewGapWithDuration(opentime.NewRationalTime(24, 24))
	track.AppendChild(gap)
	if timeline.Tracks().HasClips() {
		t.Error("Track with only gap should not have clips")
	}

	// Add a clip
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	if !timeline.Tracks().HasClips() {
		t.Error("Track with clip should have clips")
	}
}

func TestCompositionSetChildren(t *testing.T) {
	comp := NewComposition("test", nil, nil, nil, nil, nil)

	clips := []Composable{
		NewClip("clip1", nil, nil, nil, nil, nil, "", nil),
		NewClip("clip2", nil, nil, nil, nil, nil, "", nil),
	}

	if err := comp.SetChildren(clips); err != nil {
		t.Fatalf("SetChildren error: %v", err)
	}

	if len(comp.Children()) != 2 {
		t.Errorf("len(Children()) = %d, want 2", len(comp.Children()))
	}

	// Set nil should clear
	if err := comp.SetChildren(nil); err != nil {
		t.Fatalf("SetChildren(nil) error: %v", err)
	}
	if len(comp.Children()) != 0 {
		t.Errorf("len(Children()) after nil = %d, want 0", len(comp.Children()))
	}
}

func TestCompositionSourceRange(t *testing.T) {
	sr := opentime.NewTimeRange(
		opentime.NewRationalTime(10, 24),
		opentime.NewRationalTime(24, 24),
	)
	track := NewTrack("test", &sr, TrackKindVideo, nil, nil)

	// Add clip longer than source range
	clipSr := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(100, 24),
	)
	clip := NewClip("clip", nil, &clipSr, nil, nil, nil, "", nil)
	track.AppendChild(clip)

	// Duration should be source range duration, not clip duration
	dur, _ := track.Duration()
	if dur.Value() != 24 {
		t.Errorf("Duration().Value() = %v, want 24", dur.Value())
	}

	// Trimmed range of child should be intersection with source range
	trimmed, err := track.TrimmedRangeOfChildAtIndex(0)
	if err != nil {
		t.Fatalf("TrimmedRangeOfChildAtIndex error: %v", err)
	}
	// The trimmed range should be within the source range
	if trimmed.Duration().Value() != 24 {
		t.Errorf("Trimmed duration = %v, want 24", trimmed.Duration().Value())
	}
}
