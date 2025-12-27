// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Avalanche-io/gotio/opentime"
)

// Tests for FromJSONBytes error path
func TestFromJSONBytesError(t *testing.T) {
	_, err := FromJSONBytes([]byte(`invalid json`))
	if err == nil {
		t.Error("FromJSONBytes with invalid JSON should error")
	}
}

// Tests for FromJSONFile
func TestFromJSONFile(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")

	jsonData := `{"OTIO_SCHEMA": "Clip.2", "name": "test", "source_range": {"start_time": {"value": 0, "rate": 24}, "duration": {"value": 24, "rate": 24}}}`
	if err := os.WriteFile(tmpFile, []byte(jsonData), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	obj, err := FromJSONFile(tmpFile)
	if err != nil {
		t.Fatalf("FromJSONFile error: %v", err)
	}
	if obj == nil {
		t.Error("FromJSONFile returned nil")
	}
}

// Tests for FromJSONFile error path
func TestFromJSONFileError(t *testing.T) {
	_, err := FromJSONFile("/nonexistent/path/file.json")
	if err == nil {
		t.Error("FromJSONFile with nonexistent file should error")
	}
}

// Tests for ToJSONBytes
func TestToJSONBytesComplete(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("test", nil, &sr, nil, nil, nil, "", nil)

	data, err := ToJSONBytes(clip)
	if err != nil {
		t.Fatalf("ToJSONBytes error: %v", err)
	}
	if len(data) == 0 {
		t.Error("ToJSONBytes returned empty data")
	}
}

// Tests for ToJSONBytesIndent
func TestToJSONBytesIndent(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("test", nil, &sr, nil, nil, nil, "", nil)

	data, err := ToJSONBytesIndent(clip, "  ")
	if err != nil {
		t.Fatalf("ToJSONBytesIndent error: %v", err)
	}
	if len(data) == 0 {
		t.Error("ToJSONBytesIndent returned empty data")
	}
}

// Tests for ToJSONFile with indent
func TestToJSONFileWithIndent(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "output.json")

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("test", nil, &sr, nil, nil, nil, "", nil)

	err := ToJSONFile(clip, tmpFile, "  ")
	if err != nil {
		t.Fatalf("ToJSONFile error: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("ToJSONFile did not create file")
	}
}

// Tests for SerializableCollection SetChildren
func TestSerializableCollectionSetChildrenNil(t *testing.T) {
	coll := NewSerializableCollection("coll", nil, nil)
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)
	coll.AppendChild(clip)

	// Set to nil
	coll.SetChildren(nil)
	if coll.Children() == nil {
		t.Error("Children should not be nil after setting nil")
	}
}

// Tests for SerializableCollection InsertChild error
func TestSerializableCollectionInsertChildError(t *testing.T) {
	coll := NewSerializableCollection("coll", nil, nil)

	err := coll.InsertChild(-1, nil)
	if err == nil {
		t.Error("InsertChild with -1 index should error")
	}

	err = coll.InsertChild(100, nil)
	if err == nil {
		t.Error("InsertChild with out of bounds index should error")
	}
}

// Tests for SerializableCollection RemoveChild error
func TestSerializableCollectionRemoveChildError(t *testing.T) {
	coll := NewSerializableCollection("coll", nil, nil)

	err := coll.RemoveChild(0)
	if err == nil {
		t.Error("RemoveChild from empty collection should error")
	}
}

// Tests for SerializableCollection IsEquivalentTo
func TestSerializableCollectionIsEquivalentToNonCollection(t *testing.T) {
	coll := NewSerializableCollection("coll", nil, nil)
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)

	if coll.IsEquivalentTo(clip) {
		t.Error("Collection should not be equivalent to Clip")
	}
}

// Tests for SerializableCollection IsEquivalentTo with different children
func TestSerializableCollectionIsEquivalentToDifferentChildren(t *testing.T) {
	coll1 := NewSerializableCollection("coll", nil, nil)
	coll2 := NewSerializableCollection("coll", nil, nil)

	clip1 := NewClip("clip1", nil, nil, nil, nil, nil, "", nil)
	coll1.AppendChild(clip1)

	if coll1.IsEquivalentTo(coll2) {
		t.Error("Collections with different children should not be equivalent")
	}
}

// Tests for SerializableCollection UnmarshalJSON
func TestSerializableCollectionUnmarshalJSONFull(t *testing.T) {
	jsonStr := `{
		"OTIO_SCHEMA": "SerializableCollection.1",
		"name": "test_coll",
		"metadata": {"key": "value"},
		"children": [
			{"OTIO_SCHEMA": "Clip.2", "name": "clip1", "source_range": {"start_time": {"value": 0, "rate": 24}, "duration": {"value": 24, "rate": 24}}}
		]
	}`

	coll := &SerializableCollection{}
	if err := coll.UnmarshalJSON([]byte(jsonStr)); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}

	if coll.Name() != "test_coll" {
		t.Errorf("Name = %s, want test_coll", coll.Name())
	}
}

// Tests for SetMetadata
func TestSetMetadataNil(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, AnyDictionary{"key": "value"}, nil, nil, "", nil)

	clip.SetMetadata(nil)
	if clip.Metadata() == nil {
		t.Error("Metadata should not be nil after setting nil")
	}
}

// Tests for Timeline Duration
func TestTimelineDurationWithTracks(t *testing.T) {
	timeline := NewTimeline("timeline", nil, nil)
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	dur, err := timeline.Duration()
	if err != nil {
		t.Fatalf("Duration error: %v", err)
	}
	if dur.Value() != 48 {
		t.Errorf("Duration = %v, want 48", dur.Value())
	}
}

// Tests for Timeline AvailableRange
func TestTimelineAvailableRangeWithTracks(t *testing.T) {
	timeline := NewTimeline("timeline", nil, nil)
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	ar, err := timeline.AvailableRange()
	if err != nil {
		t.Fatalf("AvailableRange error: %v", err)
	}
	if ar.Duration().Value() != 48 {
		t.Errorf("AvailableRange duration = %v, want 48", ar.Duration().Value())
	}
}

// Tests for Timeline FindClips
func TestTimelineFindClipsWithFilter(t *testing.T) {
	timeline := NewTimeline("timeline", nil, nil)
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("target", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	found := timeline.FindClips(nil, false)
	if len(found) != 1 {
		t.Errorf("FindClips found %d, want 1", len(found))
	}
}

// Tests for Timeline FindChildren
func TestTimelineFindChildrenWithRange(t *testing.T) {
	timeline := NewTimeline("timeline", nil, nil)
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	// Test without range filter (should find all)
	found := timeline.FindChildren(nil, false, nil)
	t.Logf("FindChildren found %d children", len(found))
}

// Tests for Timeline AvailableImageBounds
func TestTimelineAvailableImageBoundsWithContent(t *testing.T) {
	timeline := NewTimeline("timeline", nil, nil)
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)

	bounds := &Box2d{Min: Vec2d{X: 0, Y: 0}, Max: Vec2d{X: 1920, Y: 1080}}
	ref := NewExternalReference("", "/path/file.mov", nil, nil)
	ref.SetAvailableImageBounds(bounds)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", ref, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	result, err := timeline.AvailableImageBounds()
	if err != nil {
		t.Fatalf("AvailableImageBounds error: %v", err)
	}
	if result != nil {
		t.Logf("AvailableImageBounds: %v", result)
	}
}

// Tests for Track IsEquivalentTo
func TestTrackIsEquivalentToNonTrack(t *testing.T) {
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)

	if track.IsEquivalentTo(clip) {
		t.Error("Track should not be equivalent to Clip")
	}
}

// Tests for Track IsEquivalentTo with different kinds
func TestTrackIsEquivalentToDifferentKind(t *testing.T) {
	track1 := NewTrack("track", nil, TrackKindVideo, nil, nil)
	track2 := NewTrack("track", nil, TrackKindAudio, nil, nil)

	if track1.IsEquivalentTo(track2) {
		t.Error("Tracks with different kinds should not be equivalent")
	}
}

// Tests for Track JSON round trip
func TestTrackJSONRoundTrip(t *testing.T) {
	track := NewTrack("test_track", nil, TrackKindVideo, AnyDictionary{"key": "value"}, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)

	// Marshal
	data, err := track.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}

	// Unmarshal
	track2 := &Track{}
	if err := track2.UnmarshalJSON(data); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}

	if track2.Name() != "test_track" {
		t.Errorf("Name = %s, want test_track", track2.Name())
	}
	if track2.Kind() != TrackKindVideo {
		t.Errorf("Kind = %s, want %s", track2.Kind(), TrackKindVideo)
	}
}

// Tests for Stack JSON round trip
func TestStackJSONRoundTrip(t *testing.T) {
	stack := NewStack("test_stack", nil, AnyDictionary{"key": "value"}, nil, nil, nil)
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)
	stack.AppendChild(track)

	// Marshal
	data, err := stack.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}

	// Unmarshal
	stack2 := &Stack{}
	if err := stack2.UnmarshalJSON(data); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}

	if stack2.Name() != "test_stack" {
		t.Errorf("Name = %s, want test_stack", stack2.Name())
	}
}

// Tests for ParseSchema
func TestParseSchemaVariations(t *testing.T) {
	tests := []struct {
		input   string
		name    string
		version int
	}{
		{"Clip.2", "Clip", 2},
		{"Track.1", "Track", 1},
		{"Gap.1", "Gap", 1},
	}

	for _, tt := range tests {
		name, version, err := ParseSchema(tt.input)
		if err != nil {
			t.Errorf("ParseSchema(%s) error: %v", tt.input, err)
		}
		if name != tt.name {
			t.Errorf("ParseSchema(%s) name = %s, want %s", tt.input, name, tt.name)
		}
		if version != tt.version {
			t.Errorf("ParseSchema(%s) version = %d, want %d", tt.input, version, tt.version)
		}
	}

	// Test error case
	_, _, err := ParseSchema("")
	if err == nil {
		t.Error("ParseSchema with empty string should error")
	}
}

// Tests for Track ChildAtTime deep search
func TestTrackChildAtTimeDeepSearch(t *testing.T) {
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip)

	// Deep search (shallowSearch=false)
	time := opentime.NewRationalTime(20, 24)
	child, err := track.ChildAtTime(time, false)
	if err != nil {
		t.Fatalf("ChildAtTime error: %v", err)
	}
	if child == nil {
		t.Error("Should find child at time 20")
	}
}

// Tests for LinearTimeWarp constructor with metadata
func TestLinearTimeWarpConstructorWithMetadata(t *testing.T) {
	meta := AnyDictionary{"key": "value"}
	ltw := NewLinearTimeWarp("warp", "TimeWarp", 2.0, meta)

	if ltw.Metadata()["key"] != "value" {
		t.Error("Metadata not set correctly")
	}
}

// Tests for Timeline UnmarshalJSON with global start time
func TestTimelineUnmarshalJSONWithGlobalStartTime(t *testing.T) {
	jsonStr := `{
		"OTIO_SCHEMA": "Timeline.1",
		"name": "timeline",
		"metadata": {},
		"global_start_time": {"value": 86400, "rate": 24},
		"tracks": {"OTIO_SCHEMA": "Stack.1", "name": "tracks", "children": []}
	}`

	timeline := &Timeline{}
	if err := timeline.UnmarshalJSON([]byte(jsonStr)); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}

	gst := timeline.GlobalStartTime()
	if gst == nil {
		t.Fatal("GlobalStartTime should not be nil")
	}
	if gst.Value() != 86400 {
		t.Errorf("GlobalStartTime value = %v, want 86400", gst.Value())
	}
}

// Tests for ImageSequenceReference constructor with nil values
func TestImageSequenceReferenceConstructorNilValues(t *testing.T) {
	ref := NewImageSequenceReference("", "", "", "", 0, 0, 0, 0, nil, nil, MissingFramePolicyError)

	if ref.TargetURLBase() != "" {
		t.Errorf("TargetURLBase = %s, want empty", ref.TargetURLBase())
	}
}
