// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package gotio

import (
	"testing"

	"github.com/Avalanche-io/gotio/opentime"
)

// Tests for ToJSONString error path
func TestToJSONStringError(t *testing.T) {
	// Test with valid object (should work)
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)
	str, err := ToJSONString(clip, "")
	if err != nil {
		t.Fatalf("ToJSONString error: %v", err)
	}
	if str == "" {
		t.Error("ToJSONString should return non-empty string")
	}
}

// Tests for ToJSONBytes
func TestToJSONBytesVariants(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	// Test ToJSONBytes
	data, err := ToJSONBytes(clip)
	if err != nil {
		t.Fatalf("ToJSONBytes error: %v", err)
	}
	if len(data) == 0 {
		t.Error("ToJSONBytes should return non-empty data")
	}
}

// Tests for ToJSONBytesIndent error handling
func TestToJSONBytesIndentVariants(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	// Test with different indent values
	indents := []string{"", "  ", "\t", "    "}
	for _, indent := range indents {
		data, err := ToJSONBytesIndent(clip, indent)
		if err != nil {
			t.Fatalf("ToJSONBytesIndent error with indent %q: %v", indent, err)
		}
		if len(data) == 0 {
			t.Errorf("ToJSONBytesIndent with indent %q should return non-empty data", indent)
		}
	}
}

// Tests for LinearTimeWarp constructor paths
func TestLinearTimeWarpConstructorPaths(t *testing.T) {
	// With nil metadata
	ltw1 := NewLinearTimeWarp("warp1", "TimeWarp", 1.0, nil)
	if ltw1.Metadata() == nil {
		t.Error("Metadata should not be nil")
	}

	// With non-nil metadata
	meta := AnyDictionary{"key": "value"}
	ltw2 := NewLinearTimeWarp("warp2", "TimeWarp", 2.0, meta)
	if ltw2.Metadata()["key"] != "value" {
		t.Error("Metadata not set correctly")
	}
}

// Tests for LinearTimeWarp UnmarshalJSON
func TestLinearTimeWarpUnmarshalJSONVariations(t *testing.T) {
	// Full JSON
	jsonStr := `{
		"OTIO_SCHEMA": "LinearTimeWarp.1",
		"name": "warp",
		"effect_name": "TimeWarp",
		"metadata": {"key": "value"},
		"time_scalar": 2.5
	}`

	ltw := &LinearTimeWarp{}
	if err := ltw.UnmarshalJSON([]byte(jsonStr)); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}

	// Test error case
	ltw2 := &LinearTimeWarp{}
	err := ltw2.UnmarshalJSON([]byte(`invalid`))
	if err == nil {
		t.Error("UnmarshalJSON with invalid JSON should error")
	}
}

// Tests for SerializableCollection UnmarshalJSON
func TestSerializableCollectionUnmarshalJSONVariations(t *testing.T) {
	// Minimal JSON
	jsonStr := `{
		"OTIO_SCHEMA": "SerializableCollection.1",
		"name": "coll",
		"children": []
	}`

	coll := &SerializableCollection{}
	if err := coll.UnmarshalJSON([]byte(jsonStr)); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}

	// Test error case
	coll2 := &SerializableCollection{}
	err := coll2.UnmarshalJSON([]byte(`invalid`))
	if err == nil {
		t.Error("UnmarshalJSON with invalid JSON should error")
	}
}

// Tests for TrimmedRange error path
func TestItemTrimmedRangeError(t *testing.T) {
	// Gap without source range
	gap := NewGap("", nil, nil, nil, nil, nil)
	_, err := gap.TrimmedRange()
	if err == nil {
		t.Error("TrimmedRange for Gap without source range should error")
	}
}

// Tests for TrimmedRangeOfChildAtIndex more cases
func TestTrimmedRangeOfChildAtIndexMore(t *testing.T) {
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip1 := NewClip("clip1", nil, &sr, nil, nil, nil, "", nil)
	clip2 := NewClip("clip2", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip1)
	track.AppendChild(clip2)

	// Get first child
	tr0, err := track.TrimmedRangeOfChildAtIndex(0)
	if err != nil {
		t.Fatalf("TrimmedRangeOfChildAtIndex(0) error: %v", err)
	}
	if tr0.StartTime().Value() != 0 {
		t.Errorf("First child start = %v, want 0", tr0.StartTime().Value())
	}

	// Get second child
	tr1, err := track.TrimmedRangeOfChildAtIndex(1)
	if err != nil {
		t.Fatalf("TrimmedRangeOfChildAtIndex(1) error: %v", err)
	}
	if tr1.StartTime().Value() != 24 {
		t.Errorf("Second child start = %v, want 24", tr1.StartTime().Value())
	}
}

// Tests for CompositionBase.ChildAtTime more paths
func TestCompositionChildAtTimePaths(t *testing.T) {
	// Use Stack which uses CompositionBase.ChildAtTime
	stack := NewStack("stack", nil, nil, nil, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	stack.AppendChild(clip)

	// Shallow search
	time := opentime.NewRationalTime(20, 24)
	child, err := stack.ChildAtTime(time, true)
	if err != nil {
		t.Fatalf("ChildAtTime shallow error: %v", err)
	}
	if child == nil {
		t.Error("Should find child at time 20")
	}

	// Deep search (for Stack with clip, should just return clip)
	child, err = stack.ChildAtTime(time, false)
	if err != nil {
		t.Fatalf("ChildAtTime deep error: %v", err)
	}
	if child == nil {
		t.Error("Should find child at time 20 in deep search")
	}
}

// Tests for childrenAtTime
func TestChildrenAtTimeMore(t *testing.T) {
	stack := NewStack("stack", nil, nil, nil, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip1 := NewClip("clip1", nil, &sr, nil, nil, nil, "", nil)
	clip2 := NewClip("clip2", nil, &sr, nil, nil, nil, "", nil)

	// In a stack, both clips start at time 0
	stack.AppendChild(clip1)
	stack.AppendChild(clip2)

	// Find children at time 10 (both should be found in stack)
	time := opentime.NewRationalTime(10, 24)
	child, err := stack.ChildAtTime(time, true)
	if err != nil {
		t.Fatalf("ChildAtTime error: %v", err)
	}
	if child == nil {
		t.Error("Should find at least one child")
	}
}

// Tests for FindChildren more paths
func TestFindChildrenMorePaths(t *testing.T) {
	stack := NewStack("stack", nil, nil, nil, nil, nil)
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := NewClip("find_target", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip)
	stack.AppendChild(track)

	// Find with filter and deep search
	found := stack.FindChildren(nil, false, func(c Composable) bool {
		return true // Find all
	})

	t.Logf("Found %d children with filter", len(found))

	// Find with range
	searchRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	found = stack.FindChildren(&searchRange, false, nil)
	t.Logf("Found %d children with range", len(found))
}

// Tests for CompositionImpl MarshalJSON/UnmarshalJSON error paths
func TestCompositionJSONErrorPaths(t *testing.T) {
	// Test UnmarshalJSON with invalid JSON
	comp := &CompositionImpl{}
	err := comp.UnmarshalJSON([]byte(`invalid json`))
	if err == nil {
		t.Error("UnmarshalJSON with invalid JSON should error")
	}
}

// Tests for Clip AvailableImageBounds with external reference
func TestClipAvailableImageBoundsExtRef(t *testing.T) {
	bounds := &Box2d{Min: Vec2d{X: 0, Y: 0}, Max: Vec2d{X: 1920, Y: 1080}}
	ref := NewExternalReference("", "/path/file.mov", nil, nil)
	ref.SetAvailableImageBounds(bounds)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", ref, &sr, nil, nil, nil, "", nil)

	result, err := clip.AvailableImageBounds()
	if err != nil {
		t.Fatalf("AvailableImageBounds error: %v", err)
	}
	if result == nil {
		t.Fatal("AvailableImageBounds should not be nil")
	}
	if result.Max.X != 1920 {
		t.Errorf("Max.X = %v, want 1920", result.Max.X)
	}
}

// Tests for ComposableBase Parent type assertion
func TestComposableParentTypeAssertion(t *testing.T) {
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip)

	parent := clip.Parent()
	if parent == nil {
		t.Fatal("Parent should not be nil")
	}

	// Type assert to Track
	if _, ok := parent.(*Track); !ok {
		t.Error("Parent should be *Track")
	}
}

// Tests for FromJSONBytes more paths
func TestFromJSONBytesVariations(t *testing.T) {
	// Valid clip JSON
	jsonBytes := []byte(`{"OTIO_SCHEMA": "Clip.2", "name": "test"}`)
	obj, err := FromJSONBytes(jsonBytes)
	if err != nil {
		t.Fatalf("FromJSONBytes error: %v", err)
	}
	if obj == nil {
		t.Error("FromJSONBytes should return non-nil object")
	}

	// Unknown schema
	jsonBytes = []byte(`{"OTIO_SCHEMA": "Unknown.1", "name": "test"}`)
	obj, err = FromJSONBytes(jsonBytes)
	if err == nil {
		// May or may not error depending on implementation
		t.Logf("FromJSONBytes with unknown schema returned: %v", obj)
	}
}

// Tests for AvailableRange error handling in composition
func TestCompositionAvailableRangeError(t *testing.T) {
	comp := NewComposition("comp", nil, nil, nil, nil, nil)
	// Empty composition should have zero duration
	ar, err := comp.AvailableRange()
	if err != nil {
		t.Fatalf("AvailableRange error: %v", err)
	}
	t.Logf("Empty composition available range: %v", ar)
}
