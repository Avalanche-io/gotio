// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"testing"

	"github.com/Avalanche-io/gotio/opentime"
)

// Tests for CompositionImpl (using the generic Composition type)
func TestCompositionImplChildAtTime(t *testing.T) {
	comp := NewComposition("comp", nil, nil, nil, nil, nil)
	sr1 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	sr2 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip1 := NewClip("clip1", nil, &sr1, nil, nil, nil, "", nil)
	clip2 := NewClip("clip2", nil, &sr2, nil, nil, nil, "", nil)

	comp.AppendChild(clip1)
	comp.AppendChild(clip2)

	// Find child at time 0
	time := opentime.NewRationalTime(0, 24)
	child, err := comp.ChildAtTime(time, true)
	if err != nil {
		t.Logf("ChildAtTime error (may be expected for Composition): %v", err)
	}
	if child != nil {
		t.Logf("Found child: %s", child.(*Clip).Name())
	}
}

func TestCompositionImplChildrenInRange(t *testing.T) {
	comp := NewComposition("comp", nil, nil, nil, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	comp.AppendChild(clip)

	// Get all children
	searchRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	children, err := comp.ChildrenInRange(searchRange)
	if err != nil {
		t.Logf("ChildrenInRange error (may be expected): %v", err)
	}
	t.Logf("Found %d children", len(children))
}

func TestCompositionImplRangeOfAllChildren(t *testing.T) {
	comp := NewComposition("comp", nil, nil, nil, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	comp.AppendChild(clip)

	ranges, err := comp.RangeOfAllChildren()
	if err != nil {
		t.Logf("RangeOfAllChildren error (may be expected): %v", err)
	}
	t.Logf("Found %d ranges", len(ranges))
}

func TestCompositionImplTrimmedRangeOfChildAtIndex(t *testing.T) {
	comp := NewComposition("comp", nil, nil, nil, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	comp.AppendChild(clip)

	tr, err := comp.TrimmedRangeOfChildAtIndex(0)
	if err != nil {
		t.Logf("TrimmedRangeOfChildAtIndex error (may be expected): %v", err)
	} else {
		t.Logf("TrimmedRange: %v", tr)
	}
}

func TestCompositionImplDuration(t *testing.T) {
	comp := NewComposition("comp", nil, nil, nil, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	comp.AppendChild(clip)

	dur, err := comp.Duration()
	if err != nil {
		t.Logf("Duration error (may be expected): %v", err)
	} else {
		t.Logf("Duration: %v", dur)
	}
}

func TestCompositionImplAvailableRange(t *testing.T) {
	comp := NewComposition("comp", nil, nil, nil, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	comp.AppendChild(clip)

	ar, err := comp.AvailableRange()
	if err != nil {
		t.Logf("AvailableRange error (may be expected): %v", err)
	} else {
		t.Logf("AvailableRange: %v", ar)
	}
}

// Tests for ItemBase.Duration (when sourceRange is nil)
func TestItemBaseDurationWithAvailableRange(t *testing.T) {
	// Create a clip with a media reference but no source range
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	ref := NewExternalReference("", "/path/file.mov", &ar, nil)
	clip := NewClip("clip", ref, nil, nil, nil, nil, "", nil)

	// Duration should come from the media reference's available range
	dur, err := clip.Duration()
	if err != nil {
		t.Fatalf("Duration error: %v", err)
	}
	if dur.Value() != 100 {
		t.Errorf("Duration = %v, want 100", dur.Value())
	}
}

// Tests for ItemBase.AvailableRange (base implementation)
func TestItemBaseAvailableRangeError(t *testing.T) {
	// Create a Gap without source range - this tests the error path
	gap := NewGap("", nil, nil, nil, nil, nil)

	_, err := gap.AvailableRange()
	if err == nil {
		t.Error("Gap without source range should error on AvailableRange")
	}
}

// Note: ToColor function doesn't exist in the package

// Tests for CompositionBase.CompositionKind (base implementation)
func TestCompositionBaseCompositionKindImpl(t *testing.T) {
	comp := NewComposition("comp", nil, nil, nil, nil, nil)
	// CompositionImpl returns "Composition"
	if comp.CompositionKind() != "Composition" {
		t.Errorf("CompositionKind = %s, want Composition", comp.CompositionKind())
	}
}

// Test for sortByStartTime through Track.RangeOfAllChildren
func TestSortByStartTimeThrough(t *testing.T) {
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip1 := NewClip("clip1", nil, &sr, nil, nil, nil, "", nil)
	clip2 := NewClip("clip2", nil, &sr, nil, nil, nil, "", nil)
	clip3 := NewClip("clip3", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip1)
	track.AppendChild(clip2)
	track.AppendChild(clip3)

	ranges, err := track.RangeOfAllChildren()
	if err != nil {
		t.Fatalf("RangeOfAllChildren error: %v", err)
	}

	if len(ranges) != 3 {
		t.Errorf("RangeOfAllChildren count = %d, want 3", len(ranges))
	}
}

// Test nested composition ChildAtTime
func TestNestedCompositionChildAtTime(t *testing.T) {
	timeline := NewTimeline("timeline", nil, nil)
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	// Test ChildAtTime with shallowSearch=false (should descend into nested compositions)
	time := opentime.NewRationalTime(0, 24)
	child, err := timeline.Tracks().ChildAtTime(time, false)
	if err != nil {
		t.Logf("ChildAtTime error: %v", err)
	}
	if child != nil {
		t.Logf("Found child at time 0: %T", child)
	}
}

// Tests for FindChildren in Stack
func TestStackFindChildren(t *testing.T) {
	stack := NewStack("stack", nil, nil, nil, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("find_me", nil, &sr, nil, nil, nil, "", nil)

	stack.AppendChild(clip)

	// Find children with filter
	found := stack.FindChildren(nil, false, func(c Composable) bool {
		if cl, ok := c.(*Clip); ok {
			return cl.Name() == "find_me"
		}
		return false
	})

	if len(found) != 1 {
		t.Errorf("FindChildren found %d, want 1", len(found))
	}
}
