// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"encoding/json"
	"testing"

	"github.com/mrjoshuak/gotio/opentime"
)

// Tests for SerializableCollection
func TestSerializableCollectionMethods(t *testing.T) {
	coll := NewSerializableCollection("test_coll", nil, nil)

	// Test initial state
	if len(coll.Children()) != 0 {
		t.Error("New collection should have no children")
	}

	// Test AppendChild
	clip1 := NewClip("clip1", nil, nil, nil, nil, nil, "", nil)
	clip2 := NewClip("clip2", nil, nil, nil, nil, nil, "", nil)
	coll.AppendChild(clip1)
	coll.AppendChild(clip2)

	if len(coll.Children()) != 2 {
		t.Errorf("Children count = %d, want 2", len(coll.Children()))
	}

	// Test InsertChild
	clip3 := NewClip("clip3", nil, nil, nil, nil, nil, "", nil)
	coll.InsertChild(1, clip3)

	if len(coll.Children()) != 3 {
		t.Errorf("Children count = %d, want 3", len(coll.Children()))
	}

	// Test RemoveChild
	coll.RemoveChild(0)
	if len(coll.Children()) != 2 {
		t.Errorf("Children count = %d, want 2", len(coll.Children()))
	}

	// Test SetChildren
	newChildren := []SerializableObject{
		NewClip("new1", nil, nil, nil, nil, nil, "", nil),
		NewClip("new2", nil, nil, nil, nil, nil, "", nil),
	}
	coll.SetChildren(newChildren)

	if len(coll.Children()) != 2 {
		t.Errorf("Children count = %d, want 2", len(coll.Children()))
	}

	// Test ClearChildren
	coll.ClearChildren()
	if len(coll.Children()) != 0 {
		t.Errorf("Children count after clear = %d, want 0", len(coll.Children()))
	}
}

func TestSerializableCollectionFindChildren(t *testing.T) {
	coll := NewSerializableCollection("test", nil, nil)
	clip := NewClip("found_clip", nil, nil, nil, nil, nil, "", nil)
	coll.AppendChild(clip)

	found := coll.FindChildren(func(obj SerializableObject) bool {
		if c, ok := obj.(*Clip); ok {
			return c.Name() == "found_clip"
		}
		return false
	})

	if len(found) != 1 {
		t.Errorf("FindChildren found %d, want 1", len(found))
	}
}

func TestSerializableCollectionSchema(t *testing.T) {
	coll := NewSerializableCollection("coll", nil, nil)

	if coll.SchemaName() != "SerializableCollection" {
		t.Errorf("SchemaName = %s, want SerializableCollection", coll.SchemaName())
	}
	if coll.SchemaVersion() != 1 {
		t.Errorf("SchemaVersion = %d, want 1", coll.SchemaVersion())
	}
}

func TestSerializableCollectionClone(t *testing.T) {
	coll := NewSerializableCollection("original", nil, nil)
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)
	coll.AppendChild(clip)

	clone := coll.Clone().(*SerializableCollection)

	if clone.Name() != "original" {
		t.Errorf("Clone name = %s, want original", clone.Name())
	}
	if len(clone.Children()) != 1 {
		t.Errorf("Clone children count = %d, want 1", len(clone.Children()))
	}
}

func TestSerializableCollectionIsEquivalentTo(t *testing.T) {
	coll1 := NewSerializableCollection("coll", nil, nil)
	coll2 := NewSerializableCollection("coll", nil, nil)
	coll3 := NewSerializableCollection("different", nil, nil)

	if !coll1.IsEquivalentTo(coll2) {
		t.Error("Identical collections should be equivalent")
	}
	if coll1.IsEquivalentTo(coll3) {
		t.Error("Different collections should not be equivalent")
	}

	// Test with non-collection
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)
	if coll1.IsEquivalentTo(clip) {
		t.Error("Collection should not be equivalent to Clip")
	}
}

// Tests for Color
func TestColorNewColor(t *testing.T) {
	// Test NewColor with RGBA values
	color := NewColor(1.0, 0.5, 0.25, 1.0)
	if color == nil {
		t.Fatal("NewColor returned nil")
	}
	if color.R != 1.0 {
		t.Errorf("R = %v, want 1.0", color.R)
	}
	if color.G != 0.5 {
		t.Errorf("G = %v, want 0.5", color.G)
	}
	if color.B != 0.25 {
		t.Errorf("B = %v, want 0.25", color.B)
	}
	if color.A != 1.0 {
		t.Errorf("A = %v, want 1.0", color.A)
	}
}

func TestColorRGB(t *testing.T) {
	color := NewColorRGB(128, 64, 255)

	if color.R != 128 {
		t.Errorf("R = %v, want 128", color.R)
	}
	if color.G != 64 {
		t.Errorf("G = %v, want 64", color.G)
	}
	if color.B != 255 {
		t.Errorf("B = %v, want 255", color.B)
	}
}

// Tests for JSONError
func TestJSONError(t *testing.T) {
	err := &JSONError{Message: "parsing failed"}
	expected := "JSON error: parsing failed"
	if err.Error() != expected {
		t.Errorf("JSONError.Error() = %q, want %q", err.Error(), expected)
	}
}

// Tests for ToJSONString
func TestToJSONString(t *testing.T) {
	clip := NewClip("test_clip", nil, nil, nil, nil, nil, "", nil)

	jsonStr, err := ToJSONString(clip, "")
	if err != nil {
		t.Fatalf("ToJSONString error: %v", err)
	}

	if jsonStr == "" {
		t.Error("ToJSONString should return non-empty string")
	}

	// Verify it's valid JSON
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
		t.Errorf("ToJSONString output is not valid JSON: %v", err)
	}
}

// Tests for Stack methods
func TestStackSetChildRemoveChild(t *testing.T) {
	stack := NewStack("stack", nil, nil, nil, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))

	clip1 := NewClip("clip1", nil, &sr, nil, nil, nil, "", nil)
	clip2 := NewClip("clip2", nil, &sr, nil, nil, nil, "", nil)
	clip3 := NewClip("clip3", nil, &sr, nil, nil, nil, "", nil)

	stack.AppendChild(clip1)
	stack.AppendChild(clip2)

	// Test SetChild
	err := stack.SetChild(1, clip3)
	if err != nil {
		t.Fatalf("SetChild error: %v", err)
	}

	if stack.Children()[1].(*Clip).Name() != "clip3" {
		t.Error("SetChild did not replace the child correctly")
	}

	// Test RemoveChild
	err = stack.RemoveChild(0)
	if err != nil {
		t.Fatalf("RemoveChild error: %v", err)
	}

	if len(stack.Children()) != 1 {
		t.Errorf("Children count after remove = %d, want 1", len(stack.Children()))
	}
}

// Tests for Track.SetChild
func TestTrackSetChild(t *testing.T) {
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))

	clip1 := NewClip("clip1", nil, &sr, nil, nil, nil, "", nil)
	clip2 := NewClip("clip2", nil, &sr, nil, nil, nil, "", nil)
	clip3 := NewClip("clip3", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip1)
	track.AppendChild(clip2)

	// Test SetChild
	err := track.SetChild(1, clip3)
	if err != nil {
		t.Fatalf("SetChild error: %v", err)
	}

	if track.Children()[1].(*Clip).Name() != "clip3" {
		t.Error("SetChild did not replace the child correctly")
	}
}

// Tests for Timeline methods
func TestTimelineFindChildren(t *testing.T) {
	timeline := NewTimeline("timeline", nil, nil)
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("find_me", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	found := timeline.FindChildren(nil, false, func(obj Composable) bool {
		if c, ok := obj.(*Clip); ok {
			return c.Name() == "find_me"
		}
		return false
	})

	if len(found) != 1 {
		t.Errorf("FindChildren found %d, want 1", len(found))
	}
}

func TestTimelineAvailableImageBounds(t *testing.T) {
	timeline := NewTimeline("timeline", nil, nil)

	bounds, err := timeline.AvailableImageBounds()
	// This may return nil if no clips with image bounds
	if err != nil {
		t.Logf("AvailableImageBounds returned error (expected for empty timeline): %v", err)
	}
	_ = bounds // May be nil
}

// Tests for IsParentOf
func TestCompositionIsParentOf(t *testing.T) {
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip)

	// Check parent relationship
	if !track.IsParentOf(clip) {
		t.Error("Track should be parent of clip")
	}

	// Remove and check again
	track.RemoveChild(0)
	if track.IsParentOf(clip) {
		t.Error("Track should not be parent after removal")
	}
}

// Tests for AvailableImageBounds setting
func TestMediaReferenceSetAvailableImageBounds(t *testing.T) {
	ref := NewExternalReference("", "/path/to/file.mov", nil, nil)

	bounds := &Box2d{
		Min: Vec2d{X: 0, Y: 0},
		Max: Vec2d{X: 1920, Y: 1080},
	}

	ref.SetAvailableImageBounds(bounds)

	got := ref.AvailableImageBounds()
	if got == nil {
		t.Fatal("AvailableImageBounds should not be nil")
	}
	if got.Max.X != 1920 {
		t.Errorf("AvailableImageBounds Max.X = %v, want 1920", got.Max.X)
	}
}

// Tests for LinearTimeWarp.IsEquivalentTo
func TestLinearTimeWarpIsEquivalentTo(t *testing.T) {
	ltw1 := NewLinearTimeWarp("warp", "TimeWarp", 2.0, nil)
	ltw2 := NewLinearTimeWarp("warp", "TimeWarp", 2.0, nil)
	ltw3 := NewLinearTimeWarp("warp", "TimeWarp", 1.5, nil)

	if !ltw1.IsEquivalentTo(ltw2) {
		t.Error("Identical LinearTimeWarps should be equivalent")
	}
	if ltw1.IsEquivalentTo(ltw3) {
		t.Error("Different LinearTimeWarps should not be equivalent")
	}
}

// Tests for ImageSequenceReference.IsEquivalentTo
func TestImageSequenceReferenceIsEquivalentTo(t *testing.T) {
	ref1 := NewImageSequenceReference("", "/path/", "frame_", ".exr", 1001, 1, 24, 4, nil, nil, MissingFramePolicyError)
	ref2 := NewImageSequenceReference("", "/path/", "frame_", ".exr", 1001, 1, 24, 4, nil, nil, MissingFramePolicyError)
	ref3 := NewImageSequenceReference("", "/different/", "frame_", ".exr", 1001, 1, 24, 4, nil, nil, MissingFramePolicyError)

	if !ref1.IsEquivalentTo(ref2) {
		t.Error("Identical ImageSequenceReferences should be equivalent")
	}
	if ref1.IsEquivalentTo(ref3) {
		t.Error("Different ImageSequenceReferences should not be equivalent")
	}

	// Test with non-ImageSequenceReference
	extRef := NewExternalReference("", "/path/file.mov", nil, nil)
	if ref1.IsEquivalentTo(extRef) {
		t.Error("ImageSequenceReference should not be equivalent to ExternalReference")
	}
}

// Tests for marker SetComment
func TestMarkerSetComment(t *testing.T) {
	mr := opentime.NewTimeRange(opentime.NewRationalTime(10, 24), opentime.NewRationalTime(1, 24))
	marker := NewMarker("marker", mr, MarkerColorRed, "original comment", nil)

	marker.SetComment("updated comment")

	if marker.Comment() != "updated comment" {
		t.Errorf("Comment = %s, want 'updated comment'", marker.Comment())
	}
}

// Tests for TrimmedRange
func TestItemTrimmedRange(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(10, 24), opentime.NewRationalTime(48, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	trimmed, err := clip.TrimmedRange()
	if err != nil {
		t.Fatalf("TrimmedRange error: %v", err)
	}

	if trimmed.StartTime().Value() != 10 {
		t.Errorf("TrimmedRange start = %v, want 10", trimmed.StartTime().Value())
	}
	if trimmed.Duration().Value() != 48 {
		t.Errorf("TrimmedRange duration = %v, want 48", trimmed.Duration().Value())
	}
}

// Tests for Stack schema
func TestStackSchema(t *testing.T) {
	stack := NewStack("stack", nil, nil, nil, nil, nil)

	if stack.SchemaName() != "Stack" {
		t.Errorf("SchemaName = %s, want Stack", stack.SchemaName())
	}
	if stack.SchemaVersion() != 1 {
		t.Errorf("SchemaVersion = %d, want 1", stack.SchemaVersion())
	}
}

// Tests for EffectImpl Schema
func TestEffectImplSchema(t *testing.T) {
	effect := NewEffect("effect", "EffectType", nil)

	if effect.SchemaName() != "Effect" {
		t.Errorf("SchemaName = %s, want Effect", effect.SchemaName())
	}
	if effect.SchemaVersion() != 1 {
		t.Errorf("SchemaVersion = %d, want 1", effect.SchemaVersion())
	}
}
