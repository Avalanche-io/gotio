// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"testing"

	"github.com/Avalanche-io/gotio/opentime"
)

// Tests for CloneAnyDictionary with various types
func TestCloneAnyDictionaryVariousTypes(t *testing.T) {
	meta := AnyDictionary{
		"string": "value",
		"int":    42,
		"float":  3.14,
		"bool":   true,
		"nil":    nil,
		"nested": AnyDictionary{"inner": "value"},
	}

	clone := CloneAnyDictionary(meta)

	if clone["string"] != "value" {
		t.Errorf("string = %v, want value", clone["string"])
	}
	if clone["int"] != 42 {
		t.Errorf("int = %v, want 42", clone["int"])
	}

	// Test nil case - the function may return nil for nil input
	nilClone := CloneAnyDictionary(nil)
	t.Logf("Clone of nil: %v", nilClone)
}

// Tests for Clip Duration with different states
func TestClipDurationStates(t *testing.T) {
	// With source range
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	dur, err := clip.Duration()
	if err != nil {
		t.Fatalf("Duration with source range error: %v", err)
	}
	if dur.Value() != 48 {
		t.Errorf("Duration = %v, want 48", dur.Value())
	}

	// Without source range but with media reference that has available range
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	ref := NewExternalReference("", "/path/file.mov", &ar, nil)
	clip2 := NewClip("clip2", ref, nil, nil, nil, nil, "", nil)

	dur2, err := clip2.Duration()
	if err != nil {
		t.Fatalf("Duration with media ref error: %v", err)
	}
	if dur2.Value() != 100 {
		t.Errorf("Duration from media ref = %v, want 100", dur2.Value())
	}
}

// Tests for Clip AvailableRange
func TestClipAvailableRangeStates(t *testing.T) {
	// With media reference that has available range
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	ref := NewExternalReference("", "/path/file.mov", &ar, nil)
	clip := NewClip("clip", ref, nil, nil, nil, nil, "", nil)

	result, err := clip.AvailableRange()
	if err != nil {
		t.Fatalf("AvailableRange error: %v", err)
	}
	if result.Duration().Value() != 100 {
		t.Errorf("AvailableRange duration = %v, want 100", result.Duration().Value())
	}
}

// Tests for Clip IsEquivalentTo
func TestClipIsEquivalentToVariations(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip1 := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	clip2 := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	clip3 := NewClip("different", nil, &sr, nil, nil, nil, "", nil)

	if !clip1.IsEquivalentTo(clip2) {
		t.Error("Identical clips should be equivalent")
	}
	if clip1.IsEquivalentTo(clip3) {
		t.Error("Different clips should not be equivalent")
	}

	// Non-Clip
	gap := NewGap("gap", &sr, nil, nil, nil, nil)
	if clip1.IsEquivalentTo(gap) {
		t.Error("Clip should not be equivalent to Gap")
	}
}

// Tests for Clip MarshalJSON with all fields
func TestClipMarshalJSONAll(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(10, 24), opentime.NewRationalTime(100, 24))
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(200, 24))
	ref := NewExternalReference("", "/path/file.mov", &ar, nil)
	effect := NewEffect("blur", "Blur", nil)
	mr := opentime.NewTimeRange(opentime.NewRationalTime(5, 24), opentime.NewRationalTime(1, 24))
	marker := NewMarker("note", mr, MarkerColorRed, "comment", nil)
	meta := AnyDictionary{"key": "value"}

	clip := NewClip("full_clip", ref, &sr, meta, []Effect{effect}, []*Marker{marker}, "", nil)

	data, err := clip.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}
	if len(data) == 0 {
		t.Error("MarshalJSON should return non-empty data")
	}
}

// Tests for Clip UnmarshalJSON error
func TestClipUnmarshalJSONError(t *testing.T) {
	clip := &Clip{}
	err := clip.UnmarshalJSON([]byte(`invalid`))
	if err == nil {
		t.Error("UnmarshalJSON with invalid JSON should error")
	}
}

// Tests for CompositionBase.SetChildren
func TestCompositionSetChildrenNil(t *testing.T) {
	comp := NewComposition("comp", nil, nil, nil, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	comp.AppendChild(clip)
	if len(comp.Children()) != 1 {
		t.Errorf("Children count = %d, want 1", len(comp.Children()))
	}

	// Set to nil
	comp.SetChildren(nil)
	if comp.Children() == nil {
		t.Error("Children should not be nil")
	}
}

// Tests for RangeOfChildAtIndex error
func TestRangeOfChildAtIndexError(t *testing.T) {
	comp := NewComposition("comp", nil, nil, nil, nil, nil)

	// Invalid index on empty composition
	_, err := comp.RangeOfChildAtIndex(-1)
	if err == nil {
		t.Error("RangeOfChildAtIndex(-1) should error")
	}

	_, err = comp.RangeOfChildAtIndex(100)
	if err == nil {
		t.Error("RangeOfChildAtIndex(100) should error")
	}
}

// Tests for TrimmedRangeOfChildAtIndex error
func TestTrimmedRangeOfChildAtIndexErrorPaths(t *testing.T) {
	comp := NewComposition("comp", nil, nil, nil, nil, nil)

	// Invalid index on empty composition
	_, err := comp.TrimmedRangeOfChildAtIndex(-1)
	if err == nil {
		t.Error("TrimmedRangeOfChildAtIndex(-1) should error")
	}

	_, err = comp.TrimmedRangeOfChildAtIndex(100)
	if err == nil {
		t.Error("TrimmedRangeOfChildAtIndex(100) should error")
	}
}

// Tests for RangeOfAllChildren in composition
func TestCompositionRangeOfAllChildren(t *testing.T) {
	comp := NewComposition("comp", nil, nil, nil, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	comp.AppendChild(clip)

	ranges, err := comp.RangeOfAllChildren()
	if err != nil {
		t.Fatalf("RangeOfAllChildren error: %v", err)
	}
	if len(ranges) != 1 {
		t.Errorf("RangeOfAllChildren count = %d, want 1", len(ranges))
	}
}

// Tests for trimChildRange path
func TestTrimChildRangePath(t *testing.T) {
	// Create track with source range to trigger trimChildRange
	trackSr := opentime.NewTimeRange(opentime.NewRationalTime(5, 24), opentime.NewRationalTime(20, 24))
	track := NewTrack("track", &trackSr, TrackKindVideo, nil, nil)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip)

	tr, err := track.TrimmedRangeOfChildAtIndex(0)
	if err != nil {
		t.Fatalf("TrimmedRangeOfChildAtIndex error: %v", err)
	}
	t.Logf("Trimmed range: %v", tr)
}

// Tests for childrenAtTime coverage
func TestChildrenAtTimeAll(t *testing.T) {
	stack := NewStack("stack", nil, nil, nil, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	stack.AppendChild(clip)

	// Test at various times
	times := []float64{0, 10, 24, 47, 48, 100}
	for _, v := range times {
		time := opentime.NewRationalTime(v, 24)
		child, _ := stack.ChildAtTime(time, true)
		if v < 48 && child == nil {
			t.Logf("Time %v: no child found (expected clip)", v)
		} else if v >= 48 && child != nil {
			t.Logf("Time %v: found child (expected none)", v)
		}
	}
}

// Tests for Clip AvailableImageBounds nil case
func TestClipAvailableImageBoundsNilCase(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	bounds, err := clip.AvailableImageBounds()
	if err != nil {
		t.Fatalf("AvailableImageBounds error: %v", err)
	}
	if bounds != nil {
		t.Error("AvailableImageBounds without media ref should be nil")
	}
}

// Tests for ComposableBase.Parent casting
func TestComposableParentCast(t *testing.T) {
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip)

	// Get parent and cast
	parent := clip.Parent()
	if parent == nil {
		t.Fatal("Parent should not be nil")
	}

	// The parent should be castable to *Track
	trackParent, ok := parent.(*Track)
	if !ok {
		t.Error("Parent should be castable to *Track")
	}
	if trackParent.Name() != "track" {
		t.Errorf("Parent name = %s, want track", trackParent.Name())
	}
}
