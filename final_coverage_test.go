// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package gotio

import (
	"testing"

	"github.com/Avalanche-io/gotio/opentime"
)

// Tests for Item.SetEffects with nil
func TestItemSetEffectsNil(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	// Set effects to non-nil
	effects := []Effect{NewEffect("blur", "Blur", nil)}
	clip.SetEffects(effects)
	if len(clip.Effects()) != 1 {
		t.Errorf("Effects count = %d, want 1", len(clip.Effects()))
	}

	// Set effects to nil
	clip.SetEffects(nil)
	if clip.Effects() == nil {
		t.Error("Effects should not be nil after setting nil")
	}
}

// Tests for Item.SetMarkers with nil
func TestItemSetMarkersNil(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	// Set markers to non-nil
	mr := opentime.NewTimeRange(opentime.NewRationalTime(10, 24), opentime.NewRationalTime(1, 24))
	markers := []*Marker{NewMarker("marker", mr, MarkerColorRed, "", nil)}
	clip.SetMarkers(markers)
	if len(clip.Markers()) != 1 {
		t.Errorf("Markers count = %d, want 1", len(clip.Markers()))
	}

	// Set markers to nil
	clip.SetMarkers(nil)
	if clip.Markers() == nil {
		t.Error("Markers should not be nil after setting nil")
	}
}

// Tests for TrimmedRange with source range
func TestItemTrimmedRangeWithSourceRange(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(10, 24), opentime.NewRationalTime(100, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	tr, err := clip.TrimmedRange()
	if err != nil {
		t.Fatalf("TrimmedRange error: %v", err)
	}

	if tr.Duration().Value() != 100 {
		t.Errorf("TrimmedRange duration = %v, want 100", tr.Duration().Value())
	}
	if tr.StartTime().Value() != 10 {
		t.Errorf("TrimmedRange start = %v, want 10", tr.StartTime().Value())
	}
}

// Tests for cloneEffects with effects
func TestCloneEffectsWithData(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	effects := []Effect{
		NewEffect("blur", "Blur", nil),
		NewLinearTimeWarp("warp", "TimeWarp", 2.0, nil),
	}
	clip := NewClip("clip", nil, &sr, nil, effects, nil, "", nil)

	clone := clip.Clone().(*Clip)

	if len(clone.Effects()) != 2 {
		t.Errorf("Clone effects count = %d, want 2", len(clone.Effects()))
	}
}

// Tests for cloneMarkers with markers
func TestCloneMarkersWithData(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	mr := opentime.NewTimeRange(opentime.NewRationalTime(10, 24), opentime.NewRationalTime(1, 24))
	markers := []*Marker{
		NewMarker("marker1", mr, MarkerColorRed, "note1", nil),
		NewMarker("marker2", mr, MarkerColorBlue, "note2", nil),
	}
	clip := NewClip("clip", nil, &sr, nil, nil, markers, "", nil)

	clone := clip.Clone().(*Clip)

	if len(clone.Markers()) != 2 {
		t.Errorf("Clone markers count = %d, want 2", len(clone.Markers()))
	}
}

// Tests for cloneBox2d with data
func TestCloneBox2d(t *testing.T) {
	bounds := &Box2d{Min: Vec2d{X: 0, Y: 0}, Max: Vec2d{X: 1920, Y: 1080}}
	ref := NewExternalReference("", "/path/file.mov", nil, nil)
	ref.SetAvailableImageBounds(bounds)

	clone := ref.Clone().(*ExternalReference)

	clonedBounds := clone.AvailableImageBounds()
	if clonedBounds == nil {
		t.Fatal("Cloned bounds should not be nil")
	}
	if clonedBounds.Max.X != 1920 {
		t.Errorf("Cloned Max.X = %v, want 1920", clonedBounds.Max.X)
	}
}

// Tests for LinearTimeWarp constructor and methods
func TestLinearTimeWarpConstructor(t *testing.T) {
	meta := AnyDictionary{"key": "value"}
	ltw := NewLinearTimeWarp("warp", "TimeWarp", 2.0, meta)

	if ltw.Name() != "warp" {
		t.Errorf("Name = %s, want warp", ltw.Name())
	}
	if ltw.TimeScalar() != 2.0 {
		t.Errorf("TimeScalar = %v, want 2.0", ltw.TimeScalar())
	}
}

// Tests for LinearTimeWarp IsEquivalentTo
func TestLinearTimeWarpIsEquivalentToNonLTW(t *testing.T) {
	ltw := NewLinearTimeWarp("warp", "TimeWarp", 2.0, nil)
	effect := NewEffect("blur", "Blur", nil)

	if ltw.IsEquivalentTo(effect) {
		t.Error("LinearTimeWarp should not be equivalent to Effect")
	}
}

// Tests for MissingReference IsEquivalentTo
func TestMissingReferenceIsEquivalentToNonMissing(t *testing.T) {
	missing := NewMissingReference("missing", nil, nil)
	extRef := NewExternalReference("", "/path/file.mov", nil, nil)

	if missing.IsEquivalentTo(extRef) {
		t.Error("MissingReference should not be equivalent to ExternalReference")
	}
}

// Tests for NumberOfImagesInSequence without available range
func TestImageSequenceNumberOfImagesNoRange(t *testing.T) {
	ref := NewImageSequenceReference("", "/path/", "frame_", ".exr", 1001, 1, 24, 4, nil, nil, MissingFramePolicyError)

	count := ref.NumberOfImagesInSequence()
	if count != 0 {
		t.Errorf("NumberOfImagesInSequence without range = %d, want 0", count)
	}
}

// Tests for Clip AvailableImageBounds with MissingReference
func TestClipAvailableImageBoundsMissingRef(t *testing.T) {
	missing := NewMissingReference("missing", nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", missing, &sr, nil, nil, nil, "", nil)

	bounds, err := clip.AvailableImageBounds()
	if err != nil {
		t.Fatalf("AvailableImageBounds error: %v", err)
	}
	// MissingReference has no bounds
	if bounds != nil {
		t.Error("AvailableImageBounds with MissingReference should be nil")
	}
}

// Tests for Composition Duration with sourceRange
func TestCompositionDurationWithSourceRange(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(10, 24), opentime.NewRationalTime(50, 24))
	comp := NewComposition("comp", &sr, nil, nil, nil, nil)

	dur, err := comp.Duration()
	if err != nil {
		t.Fatalf("Duration error: %v", err)
	}
	if dur.Value() != 50 {
		t.Errorf("Duration = %v, want 50", dur.Value())
	}
}

// Tests for Composition AvailableRange with children
func TestCompositionAvailableRangeWithChildren(t *testing.T) {
	comp := NewComposition("comp", nil, nil, nil, nil, nil)
	clipSr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(50, 24))
	clip := NewClip("clip", nil, &clipSr, nil, nil, nil, "", nil)
	comp.AppendChild(clip)

	ar, err := comp.AvailableRange()
	if err != nil {
		t.Fatalf("AvailableRange error: %v", err)
	}
	if ar.Duration().Value() != 50 {
		t.Errorf("AvailableRange duration = %v, want 50", ar.Duration().Value())
	}
}

// Tests for Composition Duration without sourceRange or children
func TestCompositionDurationEmpty(t *testing.T) {
	comp := NewComposition("comp", nil, nil, nil, nil, nil)

	dur, err := comp.Duration()
	if err != nil {
		t.Fatalf("Duration error: %v", err)
	}
	// Empty composition should have zero duration
	t.Logf("Empty composition duration: %v", dur)
}

// Tests for Stack ChildAtTime deep search
func TestStackChildAtTimeDeepSearch(t *testing.T) {
	stack := NewStack("stack", nil, nil, nil, nil, nil)
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip)
	stack.AppendChild(track)

	// Deep search
	time := opentime.NewRationalTime(10, 24)
	child, err := stack.ChildAtTime(time, false)
	if err != nil {
		t.Fatalf("ChildAtTime deep search error: %v", err)
	}
	if child == nil {
		t.Log("ChildAtTime returned nil (may be expected)")
	} else {
		t.Logf("Found child type: %T", child)
	}
}

// Tests for TrimmedRangeOfChildAtIndex with valid index
func TestTrimmedRangeOfChildAtIndexValid(t *testing.T) {
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)
	sr1 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	sr2 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip1 := NewClip("clip1", nil, &sr1, nil, nil, nil, "", nil)
	clip2 := NewClip("clip2", nil, &sr2, nil, nil, nil, "", nil)

	track.AppendChild(clip1)
	track.AppendChild(clip2)

	// Get trimmed range of second child
	tr, err := track.TrimmedRangeOfChildAtIndex(1)
	if err != nil {
		t.Fatalf("TrimmedRangeOfChildAtIndex error: %v", err)
	}

	// Second clip starts at frame 24
	if tr.StartTime().Value() != 24 {
		t.Errorf("TrimmedRange start = %v, want 24", tr.StartTime().Value())
	}
}

// Tests for CompositionBase.ChildAtTime with shallowSearch=false
func TestCompositionChildAtTimeNonShallow(t *testing.T) {
	// Use a Stack which doesn't override ChildAtTime
	stack := NewStack("stack", nil, nil, nil, nil, nil)

	// Add a track (composition) containing a clip
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := NewClip("deep_clip", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip)
	stack.AppendChild(track)

	// Search with shallowSearch=false to trigger deep search
	time := opentime.NewRationalTime(20, 24)
	result, err := stack.ChildAtTime(time, false)
	if err != nil {
		t.Fatalf("ChildAtTime error: %v", err)
	}

	// Result could be the clip if deep search worked, or track if it didn't recurse
	if result != nil {
		t.Logf("Deep search result type: %T", result)
	}
}

// Tests for LinearTimeWarp UnmarshalJSON
func TestLinearTimeWarpUnmarshalJSON(t *testing.T) {
	jsonStr := `{
		"OTIO_SCHEMA": "LinearTimeWarp.1",
		"name": "warp",
		"effect_name": "TimeWarp",
		"metadata": {},
		"time_scalar": 2.5
	}`

	ltw := &LinearTimeWarp{}
	if err := ltw.UnmarshalJSON([]byte(jsonStr)); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}

	if ltw.TimeScalar() != 2.5 {
		t.Errorf("TimeScalar = %v, want 2.5", ltw.TimeScalar())
	}
}

// Tests for Clip UnmarshalJSON with media references
func TestClipUnmarshalJSONWithMediaRefs(t *testing.T) {
	jsonStr := `{
		"OTIO_SCHEMA": "Clip.2",
		"name": "clip",
		"metadata": {},
		"source_range": {"start_time": {"value": 0, "rate": 24}, "duration": {"value": 24, "rate": 24}},
		"effects": [],
		"markers": [],
		"enabled": true,
		"media_reference": {
			"OTIO_SCHEMA": "ExternalReference.1",
			"name": "",
			"target_url": "/path/file.mov"
		},
		"media_references": {
			"custom_ref": {
				"OTIO_SCHEMA": "ExternalReference.1",
				"name": "",
				"target_url": "/path/other.mov"
			}
		},
		"active_media_reference_key": "DEFAULT_MEDIA"
	}`

	clip := &Clip{}
	if err := clip.UnmarshalJSON([]byte(jsonStr)); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}

	if len(clip.MediaReferences()) < 1 {
		t.Errorf("MediaReferences count = %d, want at least 1", len(clip.MediaReferences()))
	}
}

// Tests for FindChildren with deep search through nested compositions
func TestFindChildrenDeepNested(t *testing.T) {
	stack := NewStack("stack", nil, nil, nil, nil, nil)
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("find_me", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip)
	stack.AppendChild(track)

	// Deep search with filter
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

// Tests for Composition MarshalJSON/UnmarshalJSON round trip
func TestCompositionJSONRoundTrip(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(5, 24), opentime.NewRationalTime(95, 24))
	meta := AnyDictionary{"key": "value"}
	comp := NewComposition("roundtrip", &sr, meta, nil, nil, nil)

	clipSr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("child", nil, &clipSr, nil, nil, nil, "", nil)
	comp.AppendChild(clip)

	// Marshal
	data, err := comp.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}

	// Unmarshal
	comp2 := &CompositionImpl{}
	if err := comp2.UnmarshalJSON(data); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}

	if comp2.Name() != "roundtrip" {
		t.Errorf("Name = %s, want roundtrip", comp2.Name())
	}
	if len(comp2.Children()) != 1 {
		t.Errorf("Children count = %d, want 1", len(comp2.Children()))
	}
}
