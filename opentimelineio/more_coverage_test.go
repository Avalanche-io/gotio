// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"encoding/json"
	"testing"

	"github.com/mrjoshuak/gotio/opentime"
)

// Tests for AvailableImageBounds with clips
func TestStackAvailableImageBoundsWithClip(t *testing.T) {
	stack := NewStack("stack", nil, nil, nil, nil, nil)

	// Test with empty stack
	bounds, err := stack.AvailableImageBounds()
	if err != nil {
		t.Logf("Empty stack AvailableImageBounds: %v", err)
	}
	if bounds != nil {
		t.Error("Empty stack should have nil bounds")
	}

	// Add a clip with bounds
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	ref := NewExternalReference("", "/path/file.mov", &ar, nil)
	ref.SetAvailableImageBounds(&Box2d{
		Min: Vec2d{X: 0, Y: 0},
		Max: Vec2d{X: 1920, Y: 1080},
	})

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := NewClip("clip", ref, &sr, nil, nil, nil, "", nil)

	stack.AppendChild(clip)

	bounds, err = stack.AvailableImageBounds()
	if err != nil {
		t.Fatalf("AvailableImageBounds error: %v", err)
	}
	if bounds == nil {
		t.Fatal("Should have bounds after adding clip with bounds")
	}
	if bounds.Max.X != 1920 {
		t.Errorf("Bounds Max.X = %v, want 1920", bounds.Max.X)
	}
}

func TestTrackAvailableImageBoundsWithClip(t *testing.T) {
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)

	// Test with empty track
	bounds, err := track.AvailableImageBounds()
	if err != nil {
		t.Logf("Empty track AvailableImageBounds: %v", err)
	}
	if bounds != nil {
		t.Error("Empty track should have nil bounds")
	}

	// Add a clip with bounds
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	ref := NewExternalReference("", "/path/file.mov", &ar, nil)
	ref.SetAvailableImageBounds(&Box2d{
		Min: Vec2d{X: 0, Y: 0},
		Max: Vec2d{X: 1920, Y: 1080},
	})

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := NewClip("clip", ref, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip)

	bounds, err = track.AvailableImageBounds()
	if err != nil {
		t.Fatalf("AvailableImageBounds error: %v", err)
	}
	if bounds == nil {
		t.Fatal("Should have bounds after adding clip with bounds")
	}
	if bounds.Max.X != 1920 {
		t.Errorf("Bounds Max.X = %v, want 1920", bounds.Max.X)
	}
}

// Tests for Gap JSON
func TestGapJSONComplete(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(10, 30), opentime.NewRationalTime(60, 30))
	gap := NewGap("test_gap", &sr, AnyDictionary{"key": "value"}, nil, nil, nil)

	// Marshal
	data, err := json.Marshal(gap)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Verify JSON contains expected fields
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal to map error: %v", err)
	}

	if m["name"] != "test_gap" {
		t.Errorf("name = %v, want test_gap", m["name"])
	}

	// Unmarshal to new Gap
	gap2 := &Gap{}
	if err := json.Unmarshal(data, gap2); err != nil {
		t.Fatalf("Unmarshal to Gap error: %v", err)
	}

	if gap2.Name() != "test_gap" {
		t.Errorf("Unmarshaled name = %s, want test_gap", gap2.Name())
	}
	if gap2.SourceRange() == nil {
		t.Error("Unmarshaled source range should not be nil")
	}
}

// Tests for areMetadataEqual (note: Clip.IsEquivalentTo doesn't compare metadata)
func TestMetadataEquality(t *testing.T) {
	// areMetadataEqual is used internally by some types
	// Test it directly via the function
	meta1 := AnyDictionary{"key": "value", "num": 42}
	meta2 := AnyDictionary{"key": "value", "num": 42}
	meta3 := AnyDictionary{"key": "different"}

	// Same metadata should be equal
	if !areMetadataEqual(meta1, meta2) {
		t.Error("Identical metadata should be equal")
	}

	// Different metadata should not be equal
	if areMetadataEqual(meta1, meta3) {
		t.Error("Different metadata should not be equal")
	}

	// Different sizes
	meta4 := AnyDictionary{"a": 1, "b": 2}
	if areMetadataEqual(meta1, meta4) {
		t.Error("Metadata with different sizes should not be equal")
	}

	// Empty vs non-empty
	empty := AnyDictionary{}
	if areMetadataEqual(meta1, empty) {
		t.Error("Non-empty and empty metadata should not be equal")
	}
}

// Tests for RangeOfChildAtIndex error cases
func TestRangeOfChildAtIndexErrors(t *testing.T) {
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)

	// Test with empty track
	_, err := track.RangeOfChildAtIndex(0)
	if err == nil {
		t.Error("RangeOfChildAtIndex(0) on empty track should error")
	}

	// Test with negative index
	_, err = track.RangeOfChildAtIndex(-1)
	if err == nil {
		t.Error("RangeOfChildAtIndex(-1) should error")
	}

	// Add a clip
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)

	// Test with out of bounds index
	_, err = track.RangeOfChildAtIndex(5)
	if err == nil {
		t.Error("RangeOfChildAtIndex(5) should error")
	}
}

// Tests for Clip methods with nil media reference
func TestClipWithNilRef(t *testing.T) {
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)

	// MediaReference should return MissingReference
	ref := clip.MediaReference()
	if ref == nil {
		t.Error("MediaReference should not be nil")
	}
	if !ref.IsMissingReference() {
		t.Error("MediaReference should be MissingReference for nil input")
	}

	// AvailableRange should error
	_, err := clip.AvailableRange()
	if err == nil {
		t.Error("AvailableRange should error for MissingReference")
	}
}

// Tests for SetEffects on an item
func TestItemSetEffects(t *testing.T) {
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)

	// Initially no effects
	if len(clip.Effects()) != 0 {
		t.Errorf("Initial effects = %d, want 0", len(clip.Effects()))
	}

	// Set effects
	effects := []Effect{
		NewEffect("blur", "BlurEffect", nil),
		NewEffect("sharpen", "SharpenEffect", nil),
	}
	clip.SetEffects(effects)

	if len(clip.Effects()) != 2 {
		t.Errorf("Effects after set = %d, want 2", len(clip.Effects()))
	}
}

// Tests for SetMarkers on an item
func TestItemSetMarkers(t *testing.T) {
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)

	// Initially no markers
	if len(clip.Markers()) != 0 {
		t.Errorf("Initial markers = %d, want 0", len(clip.Markers()))
	}

	// Set markers
	mr := opentime.NewTimeRange(opentime.NewRationalTime(10, 24), opentime.NewRationalTime(1, 24))
	markers := []*Marker{
		NewMarker("m1", mr, MarkerColorRed, "comment1", nil),
		NewMarker("m2", mr, MarkerColorBlue, "comment2", nil),
	}
	clip.SetMarkers(markers)

	if len(clip.Markers()) != 2 {
		t.Errorf("Markers after set = %d, want 2", len(clip.Markers()))
	}
}

// Tests for Transition with in/out offset
func TestTransitionOffsetsMore(t *testing.T) {
	inOffset := opentime.NewRationalTime(5, 24)
	outOffset := opentime.NewRationalTime(10, 24)

	trans := NewTransition("transition", "SMPTE_Dissolve", inOffset, outOffset, nil)

	if trans.InOffset().Value() != 5 {
		t.Errorf("InOffset = %v, want 5", trans.InOffset().Value())
	}

	if trans.OutOffset().Value() != 10 {
		t.Errorf("OutOffset = %v, want 10", trans.OutOffset().Value())
	}
}

// Tests for Clip with effects and markers in JSON
func TestClipFullJSON(t *testing.T) {
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	ref := NewExternalReference("", "/path/file.mov", &ar, nil)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(10, 24), opentime.NewRationalTime(50, 24))
	mr := opentime.NewTimeRange(opentime.NewRationalTime(20, 24), opentime.NewRationalTime(1, 24))

	effects := []Effect{NewEffect("blur", "BlurEffect", nil)}
	markers := []*Marker{NewMarker("m1", mr, MarkerColorRed, "marker comment", nil)}

	clip := NewClip("test_clip", ref, &sr, AnyDictionary{"key": "value"}, effects, markers, "", nil)

	// Marshal
	data, err := json.Marshal(clip)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Unmarshal
	clip2 := &Clip{}
	if err := json.Unmarshal(data, clip2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if clip2.Name() != "test_clip" {
		t.Errorf("Unmarshaled name = %s, want test_clip", clip2.Name())
	}
	if len(clip2.Effects()) != 1 {
		t.Errorf("Unmarshaled effects = %d, want 1", len(clip2.Effects()))
	}
	if len(clip2.Markers()) != 1 {
		t.Errorf("Unmarshaled markers = %d, want 1", len(clip2.Markers()))
	}
}

// Tests for color with alpha
func TestColorAlpha(t *testing.T) {
	color := NewColor(0.5, 0.6, 0.7, 0.8)

	if color.R != 0.5 {
		t.Errorf("R = %v, want 0.5", color.R)
	}
	if color.G != 0.6 {
		t.Errorf("G = %v, want 0.6", color.G)
	}
	if color.B != 0.7 {
		t.Errorf("B = %v, want 0.7", color.B)
	}
	if color.A != 0.8 {
		t.Errorf("A = %v, want 0.8", color.A)
	}
}

// Tests for cloneColor
func TestCloneColor(t *testing.T) {
	original := NewColorRGB(100, 150, 200)
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", original)

	clone := clip.Clone().(*Clip)

	if clone.ItemColor() == nil {
		t.Fatal("Clone should have item color")
	}
	if clone.ItemColor().R != 100 {
		t.Errorf("Clone color R = %v, want 100", clone.ItemColor().R)
	}

	// Verify clone is independent
	clone.SetItemColor(NewColorRGB(50, 50, 50))
	if clip.ItemColor().R != 100 {
		t.Error("Original color should not change when clone changes")
	}
}

// Tests for empty composition
func TestEmptyComposition(t *testing.T) {
	track := NewTrack("empty", nil, TrackKindVideo, nil, nil)

	// Test duration of empty track
	dur, err := track.Duration()
	if err != nil {
		// Expected for empty track without source range
		t.Logf("Empty track duration error (expected): %v", err)
	} else {
		if dur.Value() != 0 {
			t.Errorf("Empty track duration = %v, want 0", dur.Value())
		}
	}
}

// Tests for SerializableCollection JSON
func TestSerializableCollectionJSON(t *testing.T) {
	coll := NewSerializableCollection("test_coll", nil, AnyDictionary{"key": "value"})
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)
	coll.AppendChild(clip)

	// Marshal
	data, err := json.Marshal(coll)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Unmarshal
	coll2 := &SerializableCollection{}
	if err := json.Unmarshal(data, coll2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if coll2.Name() != "test_coll" {
		t.Errorf("Unmarshaled name = %s, want test_coll", coll2.Name())
	}
	if len(coll2.Children()) != 1 {
		t.Errorf("Unmarshaled children = %d, want 1", len(coll2.Children()))
	}
}

// Tests for Stack JSON complete roundtrip
func TestStackJSONComplete(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	stack := NewStack("test_stack", &sr, nil, nil, nil, nil)
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	stack.AppendChild(clip)

	// Marshal
	data, err := json.Marshal(stack)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Unmarshal
	stack2 := &Stack{}
	if err := json.Unmarshal(data, stack2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if stack2.Name() != "test_stack" {
		t.Errorf("Unmarshaled name = %s, want test_stack", stack2.Name())
	}
	if len(stack2.Children()) != 1 {
		t.Errorf("Unmarshaled children = %d, want 1", len(stack2.Children()))
	}
}

// Tests for Track JSON
func TestTrackJSONComplete(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	track := NewTrack("test_track", &sr, TrackKindAudio, nil, nil)
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)

	// Marshal
	data, err := json.Marshal(track)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Unmarshal
	track2 := &Track{}
	if err := json.Unmarshal(data, track2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if track2.Name() != "test_track" {
		t.Errorf("Unmarshaled name = %s, want test_track", track2.Name())
	}
	if track2.Kind() != TrackKindAudio {
		t.Errorf("Unmarshaled kind = %s, want Audio", track2.Kind())
	}
	if len(track2.Children()) != 1 {
		t.Errorf("Unmarshaled children = %d, want 1", len(track2.Children()))
	}
}

// Tests for Timeline JSON
func TestTimelineJSONComplete(t *testing.T) {
	startTime := opentime.NewRationalTime(0, 24)
	timeline := NewTimeline("test_timeline", &startTime, nil)

	track := NewTrack("track", nil, TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	// Marshal
	data, err := json.Marshal(timeline)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Unmarshal
	timeline2 := &Timeline{}
	if err := json.Unmarshal(data, timeline2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if timeline2.Name() != "test_timeline" {
		t.Errorf("Unmarshaled name = %s, want test_timeline", timeline2.Name())
	}
	if timeline2.GlobalStartTime() == nil {
		t.Error("Unmarshaled global start time should not be nil")
	}
}
