// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package medialinker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Avalanche-io/gotio/opentime"
	"github.com/Avalanche-io/gotio/opentimelineio"
)

func createTestClip(name string, url string) *opentimelineio.Clip {
	sr := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(48, 24),
	)
	var ref opentimelineio.MediaReference
	if url != "" {
		ref = opentimelineio.NewExternalReference("", url, &sr, nil)
	}
	return opentimelineio.NewClip(name, ref, &sr, nil, nil, nil, "", nil)
}

func createTestTimeline(clips ...*opentimelineio.Clip) *opentimelineio.Timeline {
	timeline := opentimelineio.NewTimeline("test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)

	for _, clip := range clips {
		track.AppendChild(clip)
	}

	timeline.Tracks().AppendChild(track)
	return timeline
}

func TestRegistry(t *testing.T) {
	// Create a new registry for testing
	reg := NewRegistry()

	// Register a linker
	linker := NewNullLinker("test_linker")
	reg.Register(linker)

	// Get by name
	got, err := reg.Get("test_linker")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Name() != "test_linker" {
		t.Errorf("expected test_linker, got %s", got.Name())
	}

	// Get non-existent
	_, err = reg.Get("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent linker")
	}

	// Set default
	err = reg.SetDefault("test_linker")
	if err != nil {
		t.Fatalf("SetDefault failed: %v", err)
	}

	// Get default
	def, err := reg.Default()
	if err != nil {
		t.Fatalf("Default failed: %v", err)
	}
	if def.Name() != "test_linker" {
		t.Errorf("expected test_linker as default, got %s", def.Name())
	}

	// Available
	avail := reg.Available()
	if len(avail) != 1 || avail[0] != "test_linker" {
		t.Errorf("unexpected available: %v", avail)
	}

	// Unregister
	reg.Unregister("test_linker")
	_, err = reg.Get("test_linker")
	if err == nil {
		t.Error("expected error after unregister")
	}
}

func TestGlobalRegistry(t *testing.T) {
	// Clear the global registry first
	defaultRegistry.Clear()

	linker := NewNullLinker("global_test")
	Register(linker)

	got, err := Get("global_test")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Name() != "global_test" {
		t.Errorf("expected global_test, got %s", got.Name())
	}

	avail := Available()
	if len(avail) == 0 {
		t.Error("expected available linkers")
	}

	// Cleanup
	defaultRegistry.Clear()
}

func TestLinkMediaWithLinker(t *testing.T) {
	clip := createTestClip("test", "/path/to/media.mov")
	timeline := createTestTimeline(clip)

	// Use a metadata linker that marks clips as linked
	linker := NewMetadataLinker("test_metadata_linker")

	// First, replace with MissingReference to trigger linking
	missing := opentimelineio.NewMissingReference("", nil, nil)
	clip.SetMediaReference(missing)

	err := LinkMediaWithLinker(timeline, linker)
	if err != nil {
		t.Fatalf("LinkMediaWithLinker failed: %v", err)
	}

	// Check that metadata was added
	ref := clip.MediaReference()
	missingRef, ok := ref.(*opentimelineio.MissingReference)
	if !ok {
		t.Fatalf("expected MissingReference, got %T", ref)
	}

	meta := missingRef.Metadata()
	if meta["linked_by"] != "test_metadata_linker" {
		t.Error("expected linked_by metadata")
	}
}

func TestLinkClip(t *testing.T) {
	clip := createTestClip("test", "")
	clip.SetMediaReference(opentimelineio.NewMissingReference("", nil, nil))

	linker := NewMetadataLinker("clip_linker")
	err := LinkClip(clip, linker, map[string]any{"key": "value"})
	if err != nil {
		t.Fatalf("LinkClip failed: %v", err)
	}

	meta := clip.MediaReference().(*opentimelineio.MissingReference).Metadata()
	if meta["linked_by"] != "clip_linker" {
		t.Error("expected linked_by metadata")
	}
	args := meta["linking_args"].(map[string]any)
	if args["key"] != "value" {
		t.Error("expected linking_args to contain key")
	}
}

func TestNullLinker(t *testing.T) {
	linker := NewNullLinker("null")
	clip := createTestClip("test", "/path/to/media.mov")

	ref, err := linker.LinkMediaReference(clip, nil)
	if err != nil {
		t.Fatalf("LinkMediaReference failed: %v", err)
	}
	if ref != nil {
		t.Error("expected nil from NullLinker")
	}
}

func TestPathTemplateLinker(t *testing.T) {
	// Create a temp file for testing
	tmpDir, err := os.MkdirTemp("", "linker_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test media file
	testFile := filepath.Join(tmpDir, "test_clip.mov")
	if err := os.WriteFile(testFile, []byte{}, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create linker with template pointing to temp dir
	template := filepath.Join(tmpDir, "{name}.mov")
	linker := NewPathTemplateLinker(template)

	// Create clip with matching name
	clip := createTestClip("test_clip", "/original/path.mov")

	ref, err := linker.LinkMediaReference(clip, nil)
	if err != nil {
		t.Fatalf("LinkMediaReference failed: %v", err)
	}

	if ref == nil {
		t.Fatal("expected non-nil reference")
	}

	extRef, ok := ref.(*opentimelineio.ExternalReference)
	if !ok {
		t.Fatalf("expected ExternalReference, got %T", ref)
	}

	if extRef.TargetURL() != testFile {
		t.Errorf("expected %s, got %s", testFile, extRef.TargetURL())
	}
}

func TestDirectoryLinker(t *testing.T) {
	// Create temp dirs
	tmpDir, err := os.MkdirTemp("", "dir_linker_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mediaDir := filepath.Join(tmpDir, "media")
	os.MkdirAll(mediaDir, 0755)

	// Create test file
	testFile := filepath.Join(mediaDir, "test_clip.mp4")
	if err := os.WriteFile(testFile, []byte{}, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create linker
	linker := NewDirectoryLinker(
		[]string{mediaDir},
		[]string{".mov", ".mp4", ".mxf"},
	)

	// Create clip
	clip := createTestClip("test_clip", "/original/test_clip.mov")

	ref, err := linker.LinkMediaReference(clip, nil)
	if err != nil {
		t.Fatalf("LinkMediaReference failed: %v", err)
	}

	if ref == nil {
		t.Fatal("expected non-nil reference")
	}

	extRef := ref.(*opentimelineio.ExternalReference)
	if extRef.TargetURL() != testFile {
		t.Errorf("expected %s, got %s", testFile, extRef.TargetURL())
	}
}

func TestLinkMediaContinueOnError(t *testing.T) {
	// Create timeline with multiple clips
	clip1 := createTestClip("clip1", "")
	clip2 := createTestClip("clip2", "")
	clip3 := createTestClip("clip3", "")

	clip1.SetMediaReference(opentimelineio.NewMissingReference("", nil, nil))
	clip2.SetMediaReference(opentimelineio.NewMissingReference("", nil, nil))
	clip3.SetMediaReference(opentimelineio.NewMissingReference("", nil, nil))

	timeline := createTestTimeline(clip1, clip2, clip3)

	// Create a linker that marks clips
	linker := NewMetadataLinker("continue_test")

	// Should process all clips with ContinueOnError
	err := LinkMediaWithLinker(timeline, linker, WithContinueOnError(true))
	if err != nil {
		// There might be an error from one clip but it should continue
	}

	// Verify all clips were processed
	clips := timeline.FindClips(nil, false)
	for _, clip := range clips {
		ref := clip.MediaReference()
		if missing, ok := ref.(*opentimelineio.MissingReference); ok {
			meta := missing.Metadata()
			if meta["linked_by"] != "continue_test" {
				t.Errorf("clip %s was not linked", clip.Name())
			}
		}
	}
}

func TestLinkerError(t *testing.T) {
	err := &LinkerError{
		LinkerName: "test",
		ClipName:   "my_clip",
		Message:    "test error",
	}

	str := err.Error()
	if str == "" {
		t.Error("expected non-empty error string")
	}

	// With cause
	err.Cause = os.ErrNotExist
	str = err.Error()
	if str == "" {
		t.Error("expected non-empty error string with cause")
	}
}

func TestLinkerErrorUnwrap(t *testing.T) {
	cause := os.ErrNotExist
	err := &LinkerError{
		LinkerName: "test",
		ClipName:   "clip",
		Message:    "error",
		Cause:      cause,
	}

	if err.Unwrap() != cause {
		t.Error("Unwrap should return cause")
	}
}

func TestLinkingPolicyString(t *testing.T) {
	tests := []struct {
		policy   LinkingPolicy
		expected string
	}{
		{DoNotLink, "DoNotLink"},
		{UseDefaultLinker, "UseDefaultLinker"},
		{UseNamedLinker, "UseNamedLinker"},
		{LinkingPolicy(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.policy.String(); got != tt.expected {
			t.Errorf("LinkingPolicy(%d).String() = %s, want %s", tt.policy, got, tt.expected)
		}
	}
}

func TestGlobalSetDefault(t *testing.T) {
	// Clear and set up
	defaultRegistry.Clear()
	linker := NewNullLinker("default_test")
	Register(linker)

	// Set default
	err := SetDefault("default_test")
	if err != nil {
		t.Fatalf("SetDefault failed: %v", err)
	}

	// Get default
	def, err := Default()
	if err != nil {
		t.Fatalf("Default failed: %v", err)
	}
	if def.Name() != "default_test" {
		t.Errorf("expected default_test, got %s", def.Name())
	}

	// Cleanup
	defaultRegistry.Clear()
}

func TestRegistryDefaultName(t *testing.T) {
	reg := NewRegistry()
	linker := NewNullLinker("test")
	reg.Register(linker)
	reg.SetDefault("test")

	if name := reg.DefaultName(); name != "test" {
		t.Errorf("expected 'test', got %s", name)
	}
}

func TestLinkMedia(t *testing.T) {
	defaultRegistry.Clear()
	linker := NewMetadataLinker("link_media_test")
	Register(linker)

	clip := createTestClip("test", "")
	clip.SetMediaReference(opentimelineio.NewMissingReference("", nil, nil))
	timeline := createTestTimeline(clip)

	err := LinkMedia(timeline, "link_media_test")
	if err != nil {
		t.Fatalf("LinkMedia failed: %v", err)
	}

	// Verify linking happened
	clips := timeline.FindClips(nil, false)
	ref := clips[0].MediaReference()
	missing := ref.(*opentimelineio.MissingReference)
	if missing.Metadata()["linked_by"] != "link_media_test" {
		t.Error("expected linked_by metadata")
	}

	defaultRegistry.Clear()
}

func TestLinkMediaDefault(t *testing.T) {
	defaultRegistry.Clear()
	linker := NewMetadataLinker("default_linker")
	Register(linker)
	SetDefault("default_linker")

	clip := createTestClip("test", "")
	clip.SetMediaReference(opentimelineio.NewMissingReference("", nil, nil))
	timeline := createTestTimeline(clip)

	err := LinkMediaDefault(timeline)
	if err != nil {
		t.Fatalf("LinkMediaDefault failed: %v", err)
	}

	// Verify linking happened
	clips := timeline.FindClips(nil, false)
	ref := clips[0].MediaReference()
	missing := ref.(*opentimelineio.MissingReference)
	if missing.Metadata()["linked_by"] != "default_linker" {
		t.Error("expected linked_by metadata")
	}

	defaultRegistry.Clear()
}

func TestLinkMediaWithArgs(t *testing.T) {
	defaultRegistry.Clear()
	linker := NewMetadataLinker("args_test")
	Register(linker)

	clip := createTestClip("test", "")
	clip.SetMediaReference(opentimelineio.NewMissingReference("", nil, nil))
	timeline := createTestTimeline(clip)

	err := LinkMedia(timeline, "args_test", WithArgs(map[string]any{"custom": "value"}))
	if err != nil {
		t.Fatalf("LinkMedia with args failed: %v", err)
	}

	// Verify args were passed
	clips := timeline.FindClips(nil, false)
	ref := clips[0].MediaReference()
	missing := ref.(*opentimelineio.MissingReference)
	args := missing.Metadata()["linking_args"].(map[string]any)
	if args["custom"] != "value" {
		t.Error("expected custom arg")
	}

	defaultRegistry.Clear()
}

func TestPathTemplateLinkerName(t *testing.T) {
	linker := NewPathTemplateLinker("/path/{name}.mov")
	if name := linker.Name(); name != "path_template" {
		t.Errorf("expected 'path_template', got %s", name)
	}
}

func TestDirectoryLinkerName(t *testing.T) {
	linker := NewDirectoryLinker([]string{"/path"}, []string{".mov"})
	if name := linker.Name(); name != "directory" {
		t.Errorf("expected 'directory', got %s", name)
	}
}

func TestMetadataLinkerName(t *testing.T) {
	linker := NewMetadataLinker("custom_name")
	if name := linker.Name(); name != "custom_name" {
		t.Errorf("expected 'custom_name', got %s", name)
	}
}

func TestPathTemplateLinkerWithMissingReference(t *testing.T) {
	// Create a temp file
	tmpDir, err := os.MkdirTemp("", "template_missing_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test_clip.mov")
	os.WriteFile(testFile, []byte{}, 0644)

	template := filepath.Join(tmpDir, "{name}.mov")
	linker := NewPathTemplateLinker(template)

	// Create clip with MissingReference containing original_target_url
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	meta := map[string]any{"original_target_url": "/original/test_clip.mov"}
	missingRef := opentimelineio.NewMissingReference("", &sr, meta)
	clip := opentimelineio.NewClip("test_clip", missingRef, &sr, nil, nil, nil, "", nil)

	ref, err := linker.LinkMediaReference(clip, nil)
	if err != nil {
		t.Fatalf("LinkMediaReference failed: %v", err)
	}

	if ref == nil {
		t.Fatal("expected non-nil reference")
	}
}

func TestDirectoryLinkerNoName(t *testing.T) {
	linker := NewDirectoryLinker([]string{"/nonexistent"}, []string{".mov"})

	// Create clip with no name and nil reference
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := opentimelineio.NewClip("", nil, &sr, nil, nil, nil, "", nil)

	ref, err := linker.LinkMediaReference(clip, nil)
	if err != nil {
		t.Fatalf("LinkMediaReference failed: %v", err)
	}

	// Should return nil (no search name)
	if ref != nil {
		t.Error("expected nil reference for empty name")
	}
}

func TestDirectoryLinkerFromExternalRef(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dir_linker_ext_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mediaDir := filepath.Join(tmpDir, "media")
	os.MkdirAll(mediaDir, 0755)
	testFile := filepath.Join(mediaDir, "original_name.mov")
	os.WriteFile(testFile, []byte{}, 0644)

	linker := NewDirectoryLinker([]string{mediaDir}, []string{".mov"})

	// Create clip with empty name but ExternalReference with URL
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	extRef := opentimelineio.NewExternalReference("", "/other/path/original_name.mov", &sr, nil)
	clip := opentimelineio.NewClip("", extRef, &sr, nil, nil, nil, "", nil)

	ref, err := linker.LinkMediaReference(clip, nil)
	if err != nil {
		t.Fatalf("LinkMediaReference failed: %v", err)
	}

	if ref == nil {
		t.Fatal("expected non-nil reference")
	}
}

func TestMetadataLinkerNonMissing(t *testing.T) {
	linker := NewMetadataLinker("test")

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	extRef := opentimelineio.NewExternalReference("", "/path/file.mov", &sr, nil)
	clip := opentimelineio.NewClip("test", extRef, &sr, nil, nil, nil, "", nil)

	ref, err := linker.LinkMediaReference(clip, nil)
	if err != nil {
		t.Fatalf("LinkMediaReference failed: %v", err)
	}

	// Should return nil for non-MissingReference
	if ref != nil {
		t.Error("expected nil for non-MissingReference")
	}
}

func TestLinkClipWithNilArgs(t *testing.T) {
	clip := createTestClip("test", "")
	clip.SetMediaReference(opentimelineio.NewMissingReference("", nil, nil))

	linker := NewMetadataLinker("nil_args_test")

	// Pass nil args explicitly
	err := LinkClip(clip, linker, nil)
	if err != nil {
		t.Fatalf("LinkClip failed: %v", err)
	}

	// Verify it worked
	meta := clip.MediaReference().(*opentimelineio.MissingReference).Metadata()
	if meta["linked_by"] != "nil_args_test" {
		t.Error("expected linked_by metadata")
	}
}

func TestPathTemplateLinkerNilReference(t *testing.T) {
	linker := NewPathTemplateLinker("/path/{name}.mov")

	clip := opentimelineio.NewClip("test", nil, nil, nil, nil, nil, "", nil)

	ref, err := linker.LinkMediaReference(clip, nil)
	if err != nil {
		t.Fatalf("LinkMediaReference failed: %v", err)
	}

	// Should return nil for nil reference
	if ref != nil {
		t.Error("expected nil for nil reference")
	}
}

func TestDirectoryLinkerNilReference(t *testing.T) {
	linker := NewDirectoryLinker([]string{"/path"}, []string{".mov"})

	clip := opentimelineio.NewClip("test", nil, nil, nil, nil, nil, "", nil)

	ref, err := linker.LinkMediaReference(clip, nil)
	if err != nil {
		t.Fatalf("LinkMediaReference failed: %v", err)
	}

	// Should return nil for nil reference
	if ref != nil {
		t.Error("expected nil for nil reference")
	}
}

func TestMetadataLinkerDefaultMissingReference(t *testing.T) {
	linker := NewMetadataLinker("test")

	// NewClip with nil media reference creates a default MissingReference
	clip := opentimelineio.NewClip("test", nil, nil, nil, nil, nil, "", nil)

	ref, err := linker.LinkMediaReference(clip, nil)
	if err != nil {
		t.Fatalf("LinkMediaReference failed: %v", err)
	}

	// MetadataLinker modifies MissingReferences, so we get a new one back
	if ref == nil {
		t.Error("expected non-nil for default MissingReference")
	}

	missing, ok := ref.(*opentimelineio.MissingReference)
	if !ok {
		t.Fatalf("expected MissingReference, got %T", ref)
	}

	if missing.Metadata()["linked_by"] != "test" {
		t.Error("expected linked_by metadata")
	}
}

func TestLinkMediaNonExistentLinker(t *testing.T) {
	defaultRegistry.Clear()

	clip := createTestClip("test", "")
	timeline := createTestTimeline(clip)

	err := LinkMedia(timeline, "nonexistent")
	if err == nil {
		t.Error("expected error for non-existent linker")
	}
}

func TestLinkMediaDefaultNoDefault(t *testing.T) {
	defaultRegistry.Clear()

	clip := createTestClip("test", "")
	timeline := createTestTimeline(clip)

	err := LinkMediaDefault(timeline)
	if err == nil {
		t.Error("expected error when no default set")
	}
}
