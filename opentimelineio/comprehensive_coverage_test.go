// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"encoding/json"
	"testing"

	"github.com/mrjoshuak/gotio/opentime"
)

// Tests for CompositionBase.ChildAtTime error paths
func TestCompositionBaseChildAtTimeErrors(t *testing.T) {
	comp := NewComposition("comp", nil, nil, nil, nil, nil)

	// Test with empty composition
	time := opentime.NewRationalTime(10, 24)
	child, err := comp.ChildAtTime(time, true)
	if err != nil {
		t.Fatalf("ChildAtTime on empty composition error: %v", err)
	}
	if child != nil {
		t.Error("ChildAtTime on empty composition should return nil")
	}
}

// Tests for nested composition deep search
func TestCompositionBaseChildAtTimeDeepSearchNested(t *testing.T) {
	// Create outer composition
	outer := NewComposition("outer", nil, nil, nil, nil, nil)

	// Create inner composition
	inner := NewComposition("inner", nil, nil, nil, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := NewClip("deep_clip", nil, &sr, nil, nil, nil, "", nil)

	inner.AppendChild(clip)
	outer.AppendChild(inner)

	// Deep search at time 10
	time := opentime.NewRationalTime(10, 24)
	child, err := outer.ChildAtTime(time, false)
	if err != nil {
		t.Fatalf("ChildAtTime deep search error: %v", err)
	}
	// Should find something in the nested structure
	t.Logf("Found child: %v", child)
}

// Tests for CompositionImpl JSON marshal/unmarshal
func TestCompositionImplJSONComplete(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(5, 24), opentime.NewRationalTime(30, 24))
	effect := NewEffect("blur", "Blur", nil)
	mr := opentime.NewTimeRange(opentime.NewRationalTime(10, 24), opentime.NewRationalTime(1, 24))
	marker := NewMarker("note", mr, MarkerColorBlue, "test note", nil)
	meta := AnyDictionary{"key": "value", "num": 42}

	comp := NewComposition("complex_comp", &sr, meta, []Effect{effect}, []*Marker{marker}, nil)

	// Add children
	clipSr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("child_clip", nil, &clipSr, nil, nil, nil, "", nil)
	comp.AppendChild(clip)

	// Marshal
	data, err := comp.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}

	// Verify it's valid JSON
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	// Unmarshal
	comp2 := &CompositionImpl{}
	if err := comp2.UnmarshalJSON(data); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}

	if comp2.Name() != "complex_comp" {
		t.Errorf("Name = %s, want complex_comp", comp2.Name())
	}
}

// Tests for CompositionImpl.IsEquivalentTo with non-Composition
func TestCompositionImplIsEquivalentToNonComposition(t *testing.T) {
	comp := NewComposition("comp", nil, nil, nil, nil, nil)
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)

	if comp.IsEquivalentTo(clip) {
		t.Error("Composition should not be equivalent to Clip")
	}
}

// Tests for CompositionImpl.IsEquivalentTo with different children
func TestCompositionImplIsEquivalentToDifferentChildren(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))

	comp1 := NewComposition("comp", nil, nil, nil, nil, nil)
	comp2 := NewComposition("comp", nil, nil, nil, nil, nil)

	clip1 := NewClip("clip1", nil, &sr, nil, nil, nil, "", nil)
	clip2 := NewClip("clip2", nil, &sr, nil, nil, nil, "", nil)

	comp1.AppendChild(clip1)
	comp2.AppendChild(clip2)

	if comp1.IsEquivalentTo(comp2) {
		t.Error("Compositions with different children should not be equivalent")
	}
}

// Tests for CompositionImpl.Clone with children
func TestCompositionImplCloneWithChildren(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	comp := NewComposition("original", nil, nil, nil, nil, nil)
	clip := NewClip("child", nil, &sr, nil, nil, nil, "", nil)
	comp.AppendChild(clip)

	clone := comp.Clone().(*CompositionImpl)

	if len(clone.Children()) != 1 {
		t.Errorf("Clone children count = %d, want 1", len(clone.Children()))
	}
}

// Tests for Duration with composition
func TestCompositionDuration(t *testing.T) {
	comp := NewComposition("comp", nil, nil, nil, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	comp.AppendChild(clip)

	dur, err := comp.Duration()
	if err != nil {
		t.Fatalf("Duration error: %v", err)
	}
	t.Logf("Composition duration: %v", dur)
}

// Tests for AvailableRange with composition
func TestCompositionAvailableRange(t *testing.T) {
	comp := NewComposition("comp", nil, nil, nil, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	comp.AppendChild(clip)

	ar, err := comp.AvailableRange()
	if err != nil {
		t.Fatalf("AvailableRange error: %v", err)
	}
	t.Logf("Composition available range: %v", ar)
}

// Tests for TrimmedRangeOfChildAtIndex error paths
func TestTrimmedRangeOfChildAtIndexErrors(t *testing.T) {
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)

	// Test with invalid index
	_, err := track.TrimmedRangeOfChildAtIndex(-1)
	if err == nil {
		t.Error("TrimmedRangeOfChildAtIndex with -1 should error")
	}

	_, err = track.TrimmedRangeOfChildAtIndex(100)
	if err == nil {
		t.Error("TrimmedRangeOfChildAtIndex with 100 should error")
	}
}

// Tests for Effect IsEquivalentTo
func TestEffectIsEquivalentToNonEffect(t *testing.T) {
	effect := NewEffect("blur", "Blur", nil)
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)

	if effect.IsEquivalentTo(clip) {
		t.Error("Effect should not be equivalent to Clip")
	}
}

// Tests for FreezeFrame IsEquivalentTo variations
func TestFreezeFrameIsEquivalentToVariations(t *testing.T) {
	ff1 := NewFreezeFrame("freeze", nil)
	ff2 := NewFreezeFrame("freeze", nil)
	ff3 := NewFreezeFrame("different", nil)

	if !ff1.IsEquivalentTo(ff2) {
		t.Error("Identical FreezeFrames should be equivalent")
	}
	if ff1.IsEquivalentTo(ff3) {
		t.Error("Different FreezeFrames should not be equivalent")
	}

	// Test with non-FreezeFrame
	effect := NewEffect("blur", "Blur", nil)
	if ff1.IsEquivalentTo(effect) {
		t.Error("FreezeFrame should not be equivalent to Effect")
	}
}

// Tests for FreezeFrame JSON with metadata
func TestFreezeFrameJSONWithMetadata(t *testing.T) {
	ff := NewFreezeFrame("freeze", AnyDictionary{"key": "value"})

	data, err := ff.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}

	ff2 := &FreezeFrame{}
	if err := ff2.UnmarshalJSON(data); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}

	if ff2.Name() != "freeze" {
		t.Errorf("Name = %s, want freeze", ff2.Name())
	}
}

// Tests for GeneratorReference SetParameters
func TestGeneratorReferenceSetParameters(t *testing.T) {
	gen := NewGeneratorReference("gen", "SMPTEBars", nil, nil, nil)

	// Test setting parameters
	params := AnyDictionary{"width": 1920, "height": 1080}
	gen.SetParameters(params)

	if gen.Parameters()["width"] != 1920 {
		t.Errorf("Parameters[width] = %v, want 1920", gen.Parameters()["width"])
	}

	// Test setting nil parameters (should create empty map)
	gen.SetParameters(nil)
	if gen.Parameters() == nil {
		t.Error("Parameters should not be nil after setting nil")
	}
}

// Tests for GeneratorReference IsEquivalentTo
func TestGeneratorReferenceIsEquivalentToNonGenerator(t *testing.T) {
	gen := NewGeneratorReference("gen", "SMPTEBars", nil, nil, nil)
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)

	if gen.IsEquivalentTo(clip) {
		t.Error("GeneratorReference should not be equivalent to Clip")
	}
}

// Tests for ImageSequenceReference NumberOfImagesInSequence
func TestImageSequenceReferenceNumberOfImages(t *testing.T) {
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	ref := NewImageSequenceReference("", "/path/", "frame_", ".exr", 1001, 1, 24, 4, &ar, nil, MissingFramePolicyError)

	count := ref.NumberOfImagesInSequence()
	if count != 100 {
		t.Errorf("NumberOfImagesInSequence = %d, want 100", count)
	}
}

// Tests for ImageSequenceReference EndFrame without available range
func TestImageSequenceReferenceEndFrameNoRange(t *testing.T) {
	ref := NewImageSequenceReference("", "/path/", "frame_", ".exr", 1001, 1, 24, 4, nil, nil, MissingFramePolicyError)

	endFrame := ref.EndFrame()
	t.Logf("EndFrame without available range: %d", endFrame)
}

// Tests for Clip SetActiveMediaReferenceKey
func TestClipSetActiveMediaReferenceKeyInvalid(t *testing.T) {
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)

	// Set an invalid key (not in references map)
	err := clip.SetActiveMediaReferenceKey("nonexistent")
	if err == nil {
		t.Error("SetActiveMediaReferenceKey with nonexistent key should error")
	}
}

// Tests for Clip AvailableImageBounds
func TestClipAvailableImageBoundsNoRef(t *testing.T) {
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)

	bounds, err := clip.AvailableImageBounds()
	if err != nil {
		t.Fatalf("AvailableImageBounds error: %v", err)
	}
	if bounds != nil {
		t.Error("AvailableImageBounds should be nil for clip without media reference")
	}
}

// Tests for Clip UnmarshalJSON with complex data
func TestClipUnmarshalJSONComplex(t *testing.T) {
	jsonStr := `{
		"OTIO_SCHEMA": "Clip.2",
		"name": "test_clip",
		"metadata": {"key": "value"},
		"source_range": {"start_time": {"value": 10, "rate": 24}, "duration": {"value": 100, "rate": 24}},
		"effects": [],
		"markers": [],
		"enabled": true,
		"media_reference": {
			"OTIO_SCHEMA": "ExternalReference.1",
			"name": "",
			"target_url": "/path/file.mov"
		},
		"media_references": {},
		"active_media_reference_key": "DEFAULT_MEDIA"
	}`

	clip := &Clip{}
	if err := clip.UnmarshalJSON([]byte(jsonStr)); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}

	if clip.Name() != "test_clip" {
		t.Errorf("Name = %s, want test_clip", clip.Name())
	}
}

// Tests for Effect UnmarshalJSON error handling
func TestEffectUnmarshalJSONInvalid(t *testing.T) {
	effect := &EffectImpl{}
	err := effect.UnmarshalJSON([]byte(`invalid json`))
	if err == nil {
		t.Error("UnmarshalJSON with invalid JSON should error")
	}
}

// Tests for MissingReference UnmarshalJSON
func TestMissingReferenceUnmarshalJSONInvalid(t *testing.T) {
	ref := &MissingReference{}
	err := ref.UnmarshalJSON([]byte(`invalid json`))
	if err == nil {
		t.Error("UnmarshalJSON with invalid JSON should error")
	}
}

// Tests for Gap UnmarshalJSON with full data
func TestGapUnmarshalJSONFull(t *testing.T) {
	jsonStr := `{
		"OTIO_SCHEMA": "Gap.1",
		"name": "gap",
		"metadata": {"key": "value"},
		"source_range": {"start_time": {"value": 0, "rate": 24}, "duration": {"value": 24, "rate": 24}},
		"effects": [],
		"markers": [],
		"enabled": true,
		"color": {"r": 1.0, "g": 0.5, "b": 0.25, "a": 1.0}
	}`

	gap := &Gap{}
	if err := gap.UnmarshalJSON([]byte(jsonStr)); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}

	if gap.Name() != "gap" {
		t.Errorf("Name = %s, want gap", gap.Name())
	}
	if gap.ItemColor() == nil {
		t.Fatal("Color should not be nil")
	}
}

// Tests for childrenAtTime error path
func TestChildrenAtTimeError(t *testing.T) {
	// This is difficult to trigger because RangeOfChildAtIndex rarely fails
	// But we can test the normal path
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip)

	time := opentime.NewRationalTime(10, 24)
	child, err := track.ChildAtTime(time, true)
	if err != nil {
		t.Fatalf("ChildAtTime error: %v", err)
	}
	if child == nil {
		t.Error("Should find child at time 10")
	}
}

// Tests for FindChildren with searchRange error handling
func TestFindChildrenError(t *testing.T) {
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)

	// Empty track
	searchRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	found := track.FindChildren(&searchRange, true, nil)
	if len(found) != 0 {
		t.Errorf("FindChildren on empty track found %d, want 0", len(found))
	}
}

// Tests for Composable Parent with nil handling
func TestComposableParentNilCast(t *testing.T) {
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)

	// Parent is nil
	parent := clip.Parent()
	if parent != nil {
		t.Error("Parent should be nil for orphan clip")
	}
}

// Tests for Timeline IsEquivalentTo nil tracks
func TestTimelineIsEquivalentToNilTracks(t *testing.T) {
	// Create timelines with nil tracks
	timeline1 := &Timeline{
		SerializableObjectWithMetadataBase: NewSerializableObjectWithMetadataBase("timeline", nil),
		tracks:                             nil,
	}
	timeline2 := &Timeline{
		SerializableObjectWithMetadataBase: NewSerializableObjectWithMetadataBase("timeline", nil),
		tracks:                             nil,
	}

	if !timeline1.IsEquivalentTo(timeline2) {
		t.Error("Timelines with nil tracks should be equivalent")
	}

	// One nil, one not
	timeline3 := NewTimeline("timeline", nil, nil)
	if timeline1.IsEquivalentTo(timeline3) {
		// This may or may not be equivalent depending on implementation
		t.Log("Timeline with nil tracks vs non-nil tracks comparison")
	}
}

// Tests for Track RangeOfAllChildren error path
func TestTrackRangeOfAllChildrenWithTransition(t *testing.T) {
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip1 := NewClip("clip1", nil, &sr, nil, nil, nil, "", nil)
	clip2 := NewClip("clip2", nil, &sr, nil, nil, nil, "", nil)

	// Add a transition between clips
	inOffset := opentime.NewRationalTime(5, 24)
	outOffset := opentime.NewRationalTime(5, 24)
	transition := NewTransition("dissolve", "SMPTE_Dissolve", inOffset, outOffset, nil)

	track.AppendChild(clip1)
	track.AppendChild(transition)
	track.AppendChild(clip2)

	ranges, err := track.RangeOfAllChildren()
	if err != nil {
		t.Fatalf("RangeOfAllChildren error: %v", err)
	}

	// Should have ranges for all children
	if len(ranges) != 3 {
		t.Errorf("RangeOfAllChildren count = %d, want 3", len(ranges))
	}
}

// Tests for init function coverage via FromJSONString
func TestSchemaRegistrationViaFromJSON(t *testing.T) {
	// Test that registered schemas work via FromJSONString
	tests := []struct {
		name string
		json string
	}{
		{
			name: "Gap",
			json: `{"OTIO_SCHEMA": "Gap.1", "name": "gap", "source_range": {"start_time": {"value": 0, "rate": 24}, "duration": {"value": 24, "rate": 24}}}`,
		},
		{
			name: "FreezeFrame",
			json: `{"OTIO_SCHEMA": "FreezeFrame.1", "name": "freeze"}`,
		},
		{
			name: "TimeEffect",
			json: `{"OTIO_SCHEMA": "TimeEffect.1", "name": "effect", "effect_name": ""}`,
		},
		{
			name: "Transition",
			json: `{"OTIO_SCHEMA": "Transition.1", "name": "transition", "transition_type": "SMPTE_Dissolve", "in_offset": {"value": 5, "rate": 24}, "out_offset": {"value": 5, "rate": 24}}`,
		},
		{
			name: "Marker",
			json: `{"OTIO_SCHEMA": "Marker.2", "name": "marker", "marked_range": {"start_time": {"value": 0, "rate": 24}, "duration": {"value": 1, "rate": 24}}, "color": "RED"}`,
		},
		{
			name: "GeneratorReference",
			json: `{"OTIO_SCHEMA": "GeneratorReference.1", "name": "gen", "generator_kind": "SMPTEBars"}`,
		},
		{
			name: "ImageSequenceReference",
			json: `{"OTIO_SCHEMA": "ImageSequenceReference.1", "name": "", "target_url_base": "/path/", "name_prefix": "frame_", "name_suffix": ".exr", "start_frame": 1001, "frame_step": 1, "rate": 24, "frame_zero_padding": 4, "missing_frame_policy": "error"}`,
		},
		{
			name: "Composition",
			json: `{"OTIO_SCHEMA": "Composition.1", "name": "comp", "children": []}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, err := FromJSONString(tt.json)
			if err != nil {
				t.Fatalf("FromJSONString error: %v", err)
			}
			if obj == nil {
				t.Error("FromJSONString returned nil")
			}
		})
	}
}
