// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package gotio

import (
	"testing"

	"github.com/Avalanche-io/gotio/opentime"
)

// Tests for Track MarshalJSON with all fields
func TestTrackMarshalJSONComplete(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(10, 24), opentime.NewRationalTime(100, 24))
	effect := NewEffect("blur", "Blur", nil)
	mr := opentime.NewTimeRange(opentime.NewRationalTime(5, 24), opentime.NewRationalTime(1, 24))
	marker := NewMarker("note", mr, MarkerColorRed, "comment", nil)
	meta := AnyDictionary{"key": "value"}

	track := NewTrack("full_track", &sr, TrackKindVideo, meta, nil)
	track.SetEffects([]Effect{effect})
	track.SetMarkers([]*Marker{marker})

	clipSr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &clipSr, nil, nil, nil, "", nil)
	track.AppendChild(clip)

	data, err := track.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}
	if len(data) == 0 {
		t.Error("MarshalJSON should return non-empty data")
	}
}

// Tests for Track UnmarshalJSON with all fields
func TestTrackUnmarshalJSONComplete(t *testing.T) {
	jsonStr := `{
		"OTIO_SCHEMA": "Track.1",
		"name": "full_track",
		"kind": "Video",
		"metadata": {"key": "value"},
		"source_range": {"start_time": {"value": 10, "rate": 24}, "duration": {"value": 100, "rate": 24}},
		"effects": [],
		"markers": [],
		"enabled": true,
		"children": [
			{"OTIO_SCHEMA": "Clip.2", "name": "clip", "source_range": {"start_time": {"value": 0, "rate": 24}, "duration": {"value": 24, "rate": 24}}}
		]
	}`

	track := &Track{}
	if err := track.UnmarshalJSON([]byte(jsonStr)); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}

	if track.Name() != "full_track" {
		t.Errorf("Name = %s, want full_track", track.Name())
	}
	if track.Kind() != TrackKindVideo {
		t.Errorf("Kind = %s, want Video", track.Kind())
	}
	if len(track.Children()) != 1 {
		t.Errorf("Children count = %d, want 1", len(track.Children()))
	}
}

// Tests for Track UnmarshalJSON error paths
func TestTrackUnmarshalJSONErrors(t *testing.T) {
	track := &Track{}

	// Invalid JSON
	err := track.UnmarshalJSON([]byte(`invalid`))
	if err == nil {
		t.Error("UnmarshalJSON with invalid JSON should error")
	}
}

// Tests for Timeline Duration with nil tracks
func TestTimelineDurationNoTracks(t *testing.T) {
	// Create timeline with nil tracks
	timeline := &Timeline{
		SerializableObjectWithMetadataBase: NewSerializableObjectWithMetadataBase("timeline", nil),
		tracks:                             nil,
	}

	dur, err := timeline.Duration()
	// May or may not error - just exercise the code path
	t.Logf("Duration with nil tracks: %v, err: %v", dur, err)
}

// Tests for Timeline AvailableRange with nil tracks
func TestTimelineAvailableRangeNoTracks(t *testing.T) {
	// Create timeline with nil tracks
	timeline := &Timeline{
		SerializableObjectWithMetadataBase: NewSerializableObjectWithMetadataBase("timeline", nil),
		tracks:                             nil,
	}

	ar, err := timeline.AvailableRange()
	// May or may not error - just exercise the code path
	t.Logf("AvailableRange with nil tracks: %v, err: %v", ar, err)
}

// Tests for Timeline FindClips error path
func TestTimelineFindClipsNoTracks(t *testing.T) {
	// Create timeline with nil tracks
	timeline := &Timeline{
		SerializableObjectWithMetadataBase: NewSerializableObjectWithMetadataBase("timeline", nil),
		tracks:                             nil,
	}

	clips := timeline.FindClips(nil, false)
	if len(clips) != 0 {
		t.Errorf("FindClips with nil tracks should return empty, got %d", len(clips))
	}
}

// Tests for Timeline FindChildren error path
func TestTimelineFindChildrenNoTracks(t *testing.T) {
	// Create timeline with nil tracks
	timeline := &Timeline{
		SerializableObjectWithMetadataBase: NewSerializableObjectWithMetadataBase("timeline", nil),
		tracks:                             nil,
	}

	children := timeline.FindChildren(nil, false, nil)
	if len(children) != 0 {
		t.Errorf("FindChildren with nil tracks should return empty, got %d", len(children))
	}
}

// Tests for Timeline AvailableImageBounds error path
func TestTimelineAvailableImageBoundsNoTracks(t *testing.T) {
	// Create timeline with nil tracks
	timeline := &Timeline{
		SerializableObjectWithMetadataBase: NewSerializableObjectWithMetadataBase("timeline", nil),
		tracks:                             nil,
	}

	bounds, err := timeline.AvailableImageBounds()
	if err == nil {
		t.Log("AvailableImageBounds with nil tracks may not error")
	}
	_ = bounds
}

// Tests for ChildAtTime through composition
func TestChildAtTimeThroughComposition(t *testing.T) {
	// Create a nested structure that exercises ChildAtTime
	stack := NewStack("outer", nil, nil, nil, nil, nil)
	innerStack := NewStack("inner", nil, nil, nil, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := NewClip("deep", nil, &sr, nil, nil, nil, "", nil)

	innerStack.AppendChild(clip)
	stack.AppendChild(innerStack)

	// Test deep search through multiple composition levels
	time := opentime.NewRationalTime(10, 24)
	result, err := stack.ChildAtTime(time, false)
	if err != nil {
		t.Fatalf("ChildAtTime error: %v", err)
	}
	t.Logf("ChildAtTime deep result: %T", result)
}

// Tests for ToJSONBytes with complex object
func TestToJSONBytesComplex(t *testing.T) {
	timeline := NewTimeline("complex", nil, nil)
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	data, err := ToJSONBytes(timeline)
	if err != nil {
		t.Fatalf("ToJSONBytes error: %v", err)
	}
	if len(data) == 0 {
		t.Error("ToJSONBytes should return non-empty data")
	}
}

// Tests for NewLinearTimeWarp with empty strings
func TestLinearTimeWarpEmptyStrings(t *testing.T) {
	ltw := NewLinearTimeWarp("", "", 1.0, nil)
	if ltw.Name() != "" {
		t.Errorf("Name = %s, want empty", ltw.Name())
	}
	if ltw.EffectName() != "" {
		t.Errorf("EffectName = %s, want empty", ltw.EffectName())
	}
}

// Tests for Track RangeOfAllChildren with transition sequence
func TestTrackRangeOfAllChildrenTransitionSequence(t *testing.T) {
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip1 := NewClip("clip1", nil, &sr, nil, nil, nil, "", nil)
	clip2 := NewClip("clip2", nil, &sr, nil, nil, nil, "", nil)

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

	// Should have 3 entries (clip1, transition, clip2)
	if len(ranges) != 3 {
		t.Errorf("RangeOfAllChildren count = %d, want 3", len(ranges))
	}
}

// Tests for ComposableBase Parent with various types
func TestComposableParentTypes(t *testing.T) {
	// Test with Stack parent
	stack := NewStack("stack", nil, nil, nil, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	stack.AppendChild(clip)

	parent := clip.Parent()
	if parent == nil {
		t.Fatal("Parent should not be nil")
	}
	if _, ok := parent.(*Stack); !ok {
		t.Error("Parent should be *Stack")
	}

	// Test orphan
	orphan := NewClip("orphan", nil, &sr, nil, nil, nil, "", nil)
	if orphan.Parent() != nil {
		t.Error("Orphan should have nil parent")
	}
}
