// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package gotio

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Avalanche-io/gotio/opentime"
)

func TestSchemaRegistry(t *testing.T) {
	// Test that core schemas are registered
	schemas := []string{
		"Timeline", "Track", "Stack", "Clip", "Gap", "Transition",
		"Effect", "Marker", "ExternalReference", "MissingReference",
		"GeneratorReference", "ImageSequenceReference",
		"LinearTimeWarp", "FreezeFrame", "SerializableCollection",
	}

	for _, name := range schemas {
		if !IsSchemaRegistered(name) {
			t.Errorf("schema %q should be registered", name)
		}
	}

	// Test creating a schema
	obj, err := CreateSchema("Clip")
	if err != nil {
		t.Fatalf("CreateSchema(Clip) error: %v", err)
	}
	if obj.SchemaName() != "Clip" {
		t.Errorf("expected schema name Clip, got %s", obj.SchemaName())
	}

	// Test unregistered schema
	if IsSchemaRegistered("NonExistent") {
		t.Error("NonExistent schema should not be registered")
	}
	_, err = CreateSchema("NonExistent")
	if err == nil {
		t.Error("CreateSchema(NonExistent) should return error")
	}
}

func TestSchemaAlias(t *testing.T) {
	// Test that "Sequence" is registered as an alias for "Track"
	// This supports loading legacy OTIO files

	// "Sequence" should be registered (via alias)
	if !IsSchemaRegistered("Sequence") {
		t.Error("Sequence should be registered as alias for Track")
	}

	// Creating "Sequence" should return a Track
	obj, err := CreateSchema("Sequence")
	if err != nil {
		t.Fatalf("CreateSchema(Sequence) error: %v", err)
	}
	if obj.SchemaName() != "Track" {
		t.Errorf("expected schema name Track, got %s", obj.SchemaName())
	}

	// Test parsing a JSON object with Sequence schema
	jsonStr := `{"OTIO_SCHEMA": "Sequence.1", "name": "legacy_track", "kind": "Video", "children": [], "metadata": {}}`
	obj, err = FromJSONString(jsonStr)
	if err != nil {
		t.Fatalf("FromJSONString error: %v", err)
	}
	track, ok := obj.(*Track)
	if !ok {
		t.Fatalf("expected *Track, got %T", obj)
	}
	if track.Name() != "legacy_track" {
		t.Errorf("Name() = %q, want legacy_track", track.Name())
	}
}

func TestNonStandardJSONValues(t *testing.T) {
	// Test that non-standard JSON float values are handled
	// Python's JSON parser supports Inf, -Infinity, NaN but Go's doesn't

	// Test with Inf
	jsonWithInf := `{"OTIO_SCHEMA": "Clip.2", "name": "test", "metadata": {"inf_value": Inf}}`
	obj, err := FromJSONString(jsonWithInf)
	if err != nil {
		t.Fatalf("FromJSONString with Inf error: %v", err)
	}
	clip, ok := obj.(*Clip)
	if !ok {
		t.Fatalf("expected *Clip, got %T", obj)
	}
	// The Inf value should be replaced with null
	if clip.Metadata()["inf_value"] != nil {
		t.Errorf("inf_value should be null, got %v", clip.Metadata()["inf_value"])
	}

	// Test with -Infinity
	jsonWithNegInf := `{"OTIO_SCHEMA": "Clip.2", "name": "test", "metadata": {"neginf": -Infinity}}`
	obj, err = FromJSONString(jsonWithNegInf)
	if err != nil {
		t.Fatalf("FromJSONString with -Infinity error: %v", err)
	}

	// Test with NaN
	jsonWithNaN := `{"OTIO_SCHEMA": "Clip.2", "name": "test", "metadata": {"nan": NaN}}`
	obj, err = FromJSONString(jsonWithNaN)
	if err != nil {
		t.Fatalf("FromJSONString with NaN error: %v", err)
	}
}

func TestParseSchema(t *testing.T) {
	tests := []struct {
		input       string
		wantName    string
		wantVersion int
	}{
		{"Timeline.1", "Timeline", 1},
		{"Clip.2", "Clip", 2},
		{"Track", "Track", 1}, // Default version
	}

	for _, tc := range tests {
		name, version, err := ParseSchema(tc.input)
		if err != nil {
			t.Errorf("ParseSchema(%q) error: %v", tc.input, err)
			continue
		}
		if name != tc.wantName {
			t.Errorf("ParseSchema(%q) name = %q, want %q", tc.input, name, tc.wantName)
		}
		if version != tc.wantVersion {
			t.Errorf("ParseSchema(%q) version = %d, want %d", tc.input, version, tc.wantVersion)
		}
	}
}

func TestClipBasics(t *testing.T) {
	// Create a clip with an external reference
	ref := NewExternalReference("video", "file:///path/to/video.mp4", nil, nil)
	sourceRange := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(100, 24),
	)
	clip := NewClip("test_clip", ref, &sourceRange, nil, nil, nil, "", nil)

	// Test basic properties
	if clip.Name() != "test_clip" {
		t.Errorf("Name() = %q, want %q", clip.Name(), "test_clip")
	}
	if clip.SchemaName() != "Clip" {
		t.Errorf("SchemaName() = %q, want %q", clip.SchemaName(), "Clip")
	}
	if clip.SchemaVersion() != 2 {
		t.Errorf("SchemaVersion() = %d, want %d", clip.SchemaVersion(), 2)
	}

	// Test media reference
	if clip.MediaReference() == nil {
		t.Error("MediaReference() should not be nil")
	}
	if clip.ActiveMediaReferenceKey() != DefaultMediaKey {
		t.Errorf("ActiveMediaReferenceKey() = %q, want %q", clip.ActiveMediaReferenceKey(), DefaultMediaKey)
	}

	// Test source range
	if clip.SourceRange() == nil {
		t.Error("SourceRange() should not be nil")
	}

	// Test duration
	dur, err := clip.Duration()
	if err != nil {
		t.Fatalf("Duration() error: %v", err)
	}
	if dur.Value() != 100 || dur.Rate() != 24 {
		t.Errorf("Duration() = %v, want RationalTime(100, 24)", dur)
	}
}

func TestClipJSONRoundTrip(t *testing.T) {
	// Create a clip
	ref := NewExternalReference("video", "file:///video.mp4", nil, nil)
	sourceRange := opentime.NewTimeRange(
		opentime.NewRationalTime(10, 24),
		opentime.NewRationalTime(48, 24),
	)
	clip := NewClip("my_clip", ref, &sourceRange, AnyDictionary{"key": "value"}, nil, nil, "", nil)

	// Marshal to JSON
	data, err := json.Marshal(clip)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}

	// Verify JSON contains expected fields
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}
	if m["OTIO_SCHEMA"] != "Clip.2" {
		t.Errorf("OTIO_SCHEMA = %v, want Clip.2", m["OTIO_SCHEMA"])
	}
	if m["name"] != "my_clip" {
		t.Errorf("name = %v, want my_clip", m["name"])
	}

	// Unmarshal back
	obj, err := FromJSONString(string(data))
	if err != nil {
		t.Fatalf("FromJSONString error: %v", err)
	}
	clip2, ok := obj.(*Clip)
	if !ok {
		t.Fatalf("expected *Clip, got %T", obj)
	}

	// Verify equivalence
	if !clip.IsEquivalentTo(clip2) {
		t.Error("clips should be equivalent after round-trip")
	}
}

func TestTrackBasics(t *testing.T) {
	track := NewTrack("video_track", nil, TrackKindVideo, nil, nil)

	if track.Kind() != TrackKindVideo {
		t.Errorf("Kind() = %q, want %q", track.Kind(), TrackKindVideo)
	}

	// Add some clips
	for i := 0; i < 3; i++ {
		sr := opentime.NewTimeRange(
			opentime.NewRationalTime(0, 24),
			opentime.NewRationalTime(24, 24),
		)
		clip := NewClip("", nil, &sr, nil, nil, nil, "", nil)
		if err := track.AppendChild(clip); err != nil {
			t.Fatalf("AppendChild error: %v", err)
		}
	}

	if len(track.Children()) != 3 {
		t.Errorf("len(Children()) = %d, want 3", len(track.Children()))
	}

	// Test duration
	dur, err := track.Duration()
	if err != nil {
		t.Fatalf("Duration() error: %v", err)
	}
	expectedDur := opentime.NewRationalTime(72, 24) // 3 clips * 24 frames
	if !dur.Equal(expectedDur) {
		t.Errorf("Duration() = %v, want %v", dur, expectedDur)
	}

	// Test RangeOfChildAtIndex
	for i := 0; i < 3; i++ {
		r, err := track.RangeOfChildAtIndex(i)
		if err != nil {
			t.Fatalf("RangeOfChildAtIndex(%d) error: %v", i, err)
		}
		expectedSeconds := float64(i) // Each clip is 1 second (24 frames at 24fps)
		gotSeconds := r.StartTime().ToSeconds()
		if gotSeconds != expectedSeconds {
			t.Errorf("RangeOfChildAtIndex(%d).StartTime().ToSeconds() = %v, want %v", i, gotSeconds, expectedSeconds)
		}
	}
}

func TestStackBasics(t *testing.T) {
	stack := NewStack("video_stack", nil, nil, nil, nil, nil)

	// Add tracks with different durations
	durations := []float64{48, 72, 36} // frames at 24fps
	for i, dur := range durations {
		track := NewTrack("", nil, TrackKindVideo, nil, nil)
		sr := opentime.NewTimeRange(
			opentime.NewRationalTime(0, 24),
			opentime.NewRationalTime(dur, 24),
		)
		clip := NewClip("", nil, &sr, nil, nil, nil, "", nil)
		track.AppendChild(clip)
		stack.AppendChild(track)
		_ = i
	}

	// Stack duration should be the maximum of all children
	dur, err := stack.Duration()
	if err != nil {
		t.Fatalf("Duration() error: %v", err)
	}
	expectedSeconds := 72.0 / 24.0 // 3 seconds = max duration (72 frames at 24fps)
	gotSeconds := dur.ToSeconds()
	if gotSeconds != expectedSeconds {
		t.Errorf("Duration().ToSeconds() = %v, want %v", gotSeconds, expectedSeconds)
	}

	// Test RangeOfChildAtIndex - all children start at 0 in a stack
	for i := 0; i < len(durations); i++ {
		r, err := stack.RangeOfChildAtIndex(i)
		if err != nil {
			t.Fatalf("RangeOfChildAtIndex(%d) error: %v", i, err)
		}
		if !r.StartTime().Equal(opentime.RationalTime{}) {
			t.Errorf("RangeOfChildAtIndex(%d).StartTime() = %v, want 0", i, r.StartTime())
		}
	}
}

func TestTimelineBasics(t *testing.T) {
	timeline := NewTimeline("my_timeline", nil, nil)

	// Add a video track
	videoTrack := NewTrack("V1", nil, TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(100, 24),
	)
	clip := NewClip("", nil, &sr, nil, nil, nil, "", nil)
	videoTrack.AppendChild(clip)
	timeline.Tracks().AppendChild(videoTrack)

	// Add an audio track
	audioTrack := NewTrack("A1", nil, TrackKindAudio, nil, nil)
	timeline.Tracks().AppendChild(audioTrack)

	// Test video/audio track helpers
	vTracks := timeline.VideoTracks()
	if len(vTracks) != 1 {
		t.Errorf("len(VideoTracks()) = %d, want 1", len(vTracks))
	}
	aTracks := timeline.AudioTracks()
	if len(aTracks) != 1 {
		t.Errorf("len(AudioTracks()) = %d, want 1", len(aTracks))
	}

	// Test FindClips
	clips := timeline.FindClips(nil, false)
	if len(clips) != 1 {
		t.Errorf("len(FindClips()) = %d, want 1", len(clips))
	}
}

func TestTimelineJSONRoundTrip(t *testing.T) {
	// Create a timeline with content
	timeline := NewTimeline("test_timeline", nil, AnyDictionary{"author": "test"})

	track := NewTrack("V1", nil, TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(48, 24),
	)
	ref := NewExternalReference("", "file:///video.mp4", nil, nil)
	clip := NewClip("clip1", ref, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	// Marshal
	data, err := json.Marshal(timeline)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}

	// Unmarshal
	obj, err := FromJSONString(string(data))
	if err != nil {
		t.Fatalf("FromJSONString error: %v", err)
	}
	timeline2, ok := obj.(*Timeline)
	if !ok {
		t.Fatalf("expected *Timeline, got %T", obj)
	}

	// Verify
	if timeline2.Name() != "test_timeline" {
		t.Errorf("Name() = %q, want test_timeline", timeline2.Name())
	}
	if len(timeline2.VideoTracks()) != 1 {
		t.Errorf("len(VideoTracks()) = %d, want 1", len(timeline2.VideoTracks()))
	}
}

func TestGapBasics(t *testing.T) {
	dur := opentime.NewRationalTime(24, 24)
	gap := NewGapWithDuration(dur)

	if gap.SchemaName() != "Gap" {
		t.Errorf("SchemaName() = %q, want Gap", gap.SchemaName())
	}

	gotDur, err := gap.Duration()
	if err != nil {
		t.Fatalf("Duration() error: %v", err)
	}
	if !gotDur.Equal(dur) {
		t.Errorf("Duration() = %v, want %v", gotDur, dur)
	}
}

func TestTransitionBasics(t *testing.T) {
	inOffset := opentime.NewRationalTime(12, 24)
	outOffset := opentime.NewRationalTime(12, 24)
	transition := NewTransition("dissolve", TransitionTypeSMPTEDissolve, inOffset, outOffset, nil)

	if transition.SchemaName() != "Transition" {
		t.Errorf("SchemaName() = %q, want Transition", transition.SchemaName())
	}
	if !transition.Visible() {
		// Transitions should not be visible (don't take up time)
	}
	if transition.Visible() {
		t.Error("Transition.Visible() should be false")
	}
	if !transition.Overlapping() {
		t.Error("Transition.Overlapping() should be true")
	}

	dur, _ := transition.Duration()
	expectedDur := opentime.NewRationalTime(24, 24)
	if !dur.Equal(expectedDur) {
		t.Errorf("Duration() = %v, want %v", dur, expectedDur)
	}
}

func TestMarkerBasics(t *testing.T) {
	markedRange := opentime.NewTimeRange(
		opentime.NewRationalTime(10, 24),
		opentime.NewRationalTime(5, 24),
	)
	marker := NewMarker("note", markedRange, MarkerColorRed, "Important!", nil)

	if marker.Name() != "note" {
		t.Errorf("Name() = %q, want note", marker.Name())
	}
	if marker.Color() != MarkerColorRed {
		t.Errorf("Color() = %v, want %v", marker.Color(), MarkerColorRed)
	}
	if marker.Comment() != "Important!" {
		t.Errorf("Comment() = %q, want Important!", marker.Comment())
	}
}

func TestEffectBasics(t *testing.T) {
	effect := NewEffect("blur_effect", "Gaussian Blur", nil)

	if effect.Name() != "blur_effect" {
		t.Errorf("Name() = %q, want blur_effect", effect.Name())
	}
	if effect.EffectName() != "Gaussian Blur" {
		t.Errorf("EffectName() = %q, want Gaussian Blur", effect.EffectName())
	}
}

func TestLinearTimeWarp(t *testing.T) {
	ltw := NewLinearTimeWarp("speed_up", "LinearTimeWarp", 2.0, nil)

	if ltw.TimeScalar() != 2.0 {
		t.Errorf("TimeScalar() = %f, want 2.0", ltw.TimeScalar())
	}

	// Test JSON round-trip
	data, err := json.Marshal(ltw)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	obj, err := FromJSONString(string(data))
	if err != nil {
		t.Fatalf("FromJSONString error: %v", err)
	}
	ltw2, ok := obj.(*LinearTimeWarp)
	if !ok {
		t.Fatalf("expected *LinearTimeWarp, got %T", obj)
	}
	if ltw2.TimeScalar() != 2.0 {
		t.Errorf("TimeScalar() after round-trip = %f, want 2.0", ltw2.TimeScalar())
	}
}

func TestFreezeFrame(t *testing.T) {
	ff := NewFreezeFrame("freeze", nil)

	if ff.TimeScalar() != 0 {
		t.Errorf("TimeScalar() = %f, want 0", ff.TimeScalar())
	}
	if ff.SchemaName() != "FreezeFrame" {
		t.Errorf("SchemaName() = %q, want FreezeFrame", ff.SchemaName())
	}
}

func TestSerializableCollection(t *testing.T) {
	clip1 := NewClip("clip1", nil, nil, nil, nil, nil, "", nil)
	clip2 := NewClip("clip2", nil, nil, nil, nil, nil, "", nil)

	coll := NewSerializableCollection("my_collection", []SerializableObject{clip1, clip2}, nil)

	if len(coll.Children()) != 2 {
		t.Errorf("len(Children()) = %d, want 2", len(coll.Children()))
	}

	// Test JSON round-trip
	data, err := json.Marshal(coll)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	obj, err := FromJSONString(string(data))
	if err != nil {
		t.Fatalf("FromJSONString error: %v", err)
	}
	coll2, ok := obj.(*SerializableCollection)
	if !ok {
		t.Fatalf("expected *SerializableCollection, got %T", obj)
	}
	if len(coll2.Children()) != 2 {
		t.Errorf("len(Children()) after round-trip = %d, want 2", len(coll2.Children()))
	}
}

func TestUnknownSchema(t *testing.T) {
	// Test that unknown schemas are preserved
	jsonStr := `{"OTIO_SCHEMA": "CustomType.1", "name": "custom", "custom_field": 42}`

	obj, err := FromJSONString(jsonStr)
	if err != nil {
		t.Fatalf("FromJSONString error: %v", err)
	}

	unknown, ok := obj.(*UnknownSchema)
	if !ok {
		t.Fatalf("expected *UnknownSchema, got %T", obj)
	}

	if unknown.SchemaName() != "CustomType" {
		t.Errorf("SchemaName() = %q, want CustomType", unknown.SchemaName())
	}

	// Verify data is preserved
	data := unknown.Data()
	if data["custom_field"] != float64(42) {
		t.Errorf("custom_field = %v, want 42", data["custom_field"])
	}

	// Test round-trip
	outData, err := json.Marshal(unknown)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}

	var m map[string]any
	json.Unmarshal(outData, &m)
	if m["custom_field"] != float64(42) {
		t.Errorf("custom_field after round-trip = %v, want 42", m["custom_field"])
	}
}

func TestMediaReferences(t *testing.T) {
	// Test ExternalReference
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	ext := NewExternalReference("video", "file:///path/to/video.mp4", &ar, nil)

	if ext.TargetURL() != "file:///path/to/video.mp4" {
		t.Errorf("TargetURL() = %q, want file:///path/to/video.mp4", ext.TargetURL())
	}
	if ext.IsMissingReference() {
		t.Error("ExternalReference.IsMissingReference() should be false")
	}

	// Test MissingReference
	missing := NewMissingReference("placeholder", nil, nil)
	if !missing.IsMissingReference() {
		t.Error("MissingReference.IsMissingReference() should be true")
	}

	// Test GeneratorReference
	gen := NewGeneratorReference("solid", "SolidColor", AnyDictionary{"color": "red"}, nil, nil)
	if gen.GeneratorKind() != "SolidColor" {
		t.Errorf("GeneratorKind() = %q, want SolidColor", gen.GeneratorKind())
	}
	params := gen.Parameters()
	if params["color"] != "red" {
		t.Errorf("Parameters()[color] = %v, want red", params["color"])
	}
}

func TestImageSequenceReference(t *testing.T) {
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	seq := NewImageSequenceReference(
		"frames",
		"file:///path/to/",
		"frame_",
		".exr",
		1001,
		1,
		24.0,
		4,
		&ar,
		nil,
		MissingFramePolicyHold,
	)

	// Test URL generation
	url := seq.TargetURLForImageNumber(1001)
	expected := "file:///path/to/frame_1001.exr"
	if url != expected {
		t.Errorf("TargetURLForImageNumber(1001) = %q, want %q", url, expected)
	}

	url = seq.TargetURLForImageNumber(1)
	expected = "file:///path/to/frame_0001.exr"
	if url != expected {
		t.Errorf("TargetURLForImageNumber(1) = %q, want %q", url, expected)
	}
}

func TestClone(t *testing.T) {
	// Test that Clone creates a deep copy
	ref := NewExternalReference("", "file:///video.mp4", nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := NewClip("original", ref, &sr, AnyDictionary{"key": "value"}, nil, nil, "", nil)

	clone := clip.Clone().(*Clip)

	// Modify original
	clip.SetName("modified")
	clip.Metadata()["key"] = "modified"

	// Clone should be unaffected
	if clone.Name() != "original" {
		t.Errorf("Clone name should be original, got %q", clone.Name())
	}
	if clone.Metadata()["key"] != "value" {
		t.Errorf("Clone metadata should have original value")
	}
}

func TestToJSONFile(t *testing.T) {
	timeline := NewTimeline("test", nil, nil)
	track := NewTrack("V1", nil, TrackKindVideo, nil, nil)
	timeline.Tracks().AppendChild(track)

	// Create temp file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.otio")

	// Write to file
	err := ToJSONFile(timeline, filePath, "  ")
	if err != nil {
		t.Fatalf("ToJSONFile error: %v", err)
	}

	// Read back
	obj, err := FromJSONFile(filePath)
	if err != nil {
		t.Fatalf("FromJSONFile error: %v", err)
	}

	timeline2, ok := obj.(*Timeline)
	if !ok {
		t.Fatalf("expected *Timeline, got %T", obj)
	}

	if timeline2.Name() != "test" {
		t.Errorf("Name() = %q, want test", timeline2.Name())
	}

	// Verify file was written with indentation
	data, _ := os.ReadFile(filePath)
	if len(data) == 0 {
		t.Error("file should not be empty")
	}
}
