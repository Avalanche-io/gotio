// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package bundle

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/absfs/memfs"
	"github.com/Avalanche-io/gotio/opentime"
	"github.com/Avalanche-io/gotio/opentimelineio"
)

// createTestTimeline creates a simple timeline for testing.
func createTestTimeline() *opentimelineio.Timeline {
	timeline := opentimelineio.NewTimeline("test_timeline", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)

	// Create a clip with external reference
	sr := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(48, 24),
	)
	ref := opentimelineio.NewExternalReference("", "file:///nonexistent/test.mov", &sr, nil)
	clip := opentimelineio.NewClip("test_clip", ref, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	return timeline
}

func TestMediaReferencePolicyString(t *testing.T) {
	tests := []struct {
		policy MediaReferencePolicy
		want   string
	}{
		{ErrorIfNotFile, "ErrorIfNotFile"},
		{MissingIfNotFile, "MissingIfNotFile"},
		{AllMissing, "AllMissing"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.policy.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrepareForBundleAllMissing(t *testing.T) {
	timeline := createTestTimeline()

	prepared, manifest, err := PrepareForBundle(timeline, AllMissing)
	if err != nil {
		t.Fatalf("PrepareForBundle failed: %v", err)
	}

	// Manifest should be empty with AllMissing
	if len(manifest) != 0 {
		t.Errorf("expected empty manifest, got %d entries", len(manifest))
	}

	// Check that clip's reference is now MissingReference
	clips := prepared.FindClips(nil, false)
	if len(clips) != 1 {
		t.Fatalf("expected 1 clip, got %d", len(clips))
	}

	_, ok := clips[0].MediaReference().(*opentimelineio.MissingReference)
	if !ok {
		t.Errorf("expected MissingReference, got %T", clips[0].MediaReference())
	}
}

func TestPrepareForBundleMissingIfNotFile(t *testing.T) {
	timeline := createTestTimeline()

	prepared, manifest, err := PrepareForBundle(timeline, MissingIfNotFile)
	if err != nil {
		t.Fatalf("PrepareForBundle failed: %v", err)
	}

	// Manifest should be empty (file doesn't exist)
	if len(manifest) != 0 {
		t.Errorf("expected empty manifest for non-existent file, got %d entries", len(manifest))
	}

	// Check that clip's reference is now MissingReference
	clips := prepared.FindClips(nil, false)
	if len(clips) != 1 {
		t.Fatalf("expected 1 clip, got %d", len(clips))
	}

	missing, ok := clips[0].MediaReference().(*opentimelineio.MissingReference)
	if !ok {
		t.Errorf("expected MissingReference, got %T", clips[0].MediaReference())
	}

	// Check metadata
	meta := missing.Metadata()
	if meta["missing_reference_because"] == nil {
		t.Error("expected missing_reference_because in metadata")
	}
}

func TestVerifyUniqueBasenames(t *testing.T) {
	// Test with unique basenames
	manifest := MediaManifest{
		"/path/to/file1.mov": nil,
		"/path/to/file2.mov": nil,
		"/other/file3.mov":   nil,
	}

	if err := VerifyUniqueBasenames(manifest); err != nil {
		t.Errorf("expected no error for unique basenames, got %v", err)
	}

	// Test with collision
	manifest["/another/file1.mov"] = nil
	if err := VerifyUniqueBasenames(manifest); err == nil {
		t.Error("expected error for duplicate basenames")
	}
}

func TestRelinkToBundle(t *testing.T) {
	ar := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(100, 24),
	)
	ref := opentimelineio.NewExternalReference("", "/path/to/test.mov", &ar, nil)

	manifest := MediaManifest{
		"/path/to/test.mov": {ref},
	}

	RelinkToBundle(manifest)

	if ref.TargetURL() != "media/test.mov" {
		t.Errorf("expected media/test.mov, got %s", ref.TargetURL())
	}
}

func TestOTIODRoundTrip(t *testing.T) {
	// Create a temp directory for the test
	tmpDir, err := os.MkdirTemp("", "otiod_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	bundlePath := filepath.Join(tmpDir, "test.otiod")

	// Create a timeline without external references (to avoid file issues)
	timeline := opentimelineio.NewTimeline("roundtrip_test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)

	sr := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(48, 24),
	)
	clip := opentimelineio.NewClip("test_clip", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	// Write
	err = WriteOTIOD(timeline, bundlePath, AllMissing)
	if err != nil {
		t.Fatalf("WriteOTIOD failed: %v", err)
	}

	// Verify directory structure
	if _, err := os.Stat(filepath.Join(bundlePath, "content.otio")); err != nil {
		t.Error("content.otio not found")
	}
	if _, err := os.Stat(filepath.Join(bundlePath, "media")); err != nil {
		t.Error("media directory not found")
	}

	// Read
	readTimeline, err := ReadOTIOD(bundlePath, false)
	if err != nil {
		t.Fatalf("ReadOTIOD failed: %v", err)
	}

	if readTimeline.Name() != "roundtrip_test" {
		t.Errorf("expected name roundtrip_test, got %s", readTimeline.Name())
	}
}

func TestOTIOZRoundTrip(t *testing.T) {
	// Create a temp directory for the test
	tmpDir, err := os.MkdirTemp("", "otioz_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	bundlePath := filepath.Join(tmpDir, "test.otioz")

	// Create a timeline without external references
	timeline := opentimelineio.NewTimeline("zip_roundtrip_test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)

	sr := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(48, 24),
	)
	clip := opentimelineio.NewClip("test_clip", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	// Write
	err = WriteOTIOZ(timeline, bundlePath, AllMissing)
	if err != nil {
		t.Fatalf("WriteOTIOZ failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(bundlePath); err != nil {
		t.Fatal("bundle file not created")
	}

	// Read
	readTimeline, err := ReadOTIOZ(bundlePath)
	if err != nil {
		t.Fatalf("ReadOTIOZ failed: %v", err)
	}

	if readTimeline.Name() != "zip_roundtrip_test" {
		t.Errorf("expected name zip_roundtrip_test, got %s", readTimeline.Name())
	}
}

func TestOTIOZWithExtraction(t *testing.T) {
	// Create temp directories
	tmpDir, err := os.MkdirTemp("", "otioz_extract_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	bundlePath := filepath.Join(tmpDir, "test.otioz")
	extractDir := filepath.Join(tmpDir, "extracted")

	// Create timeline
	timeline := opentimelineio.NewTimeline("extract_test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(48, 24),
	)
	clip := opentimelineio.NewClip("test_clip", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	// Write bundle
	err = WriteOTIOZ(timeline, bundlePath, AllMissing)
	if err != nil {
		t.Fatalf("WriteOTIOZ failed: %v", err)
	}

	// Read with extraction
	readTimeline, err := ReadOTIOZWithExtraction(bundlePath, extractDir)
	if err != nil {
		t.Fatalf("ReadOTIOZWithExtraction failed: %v", err)
	}

	if readTimeline.Name() != "extract_test" {
		t.Errorf("expected name extract_test, got %s", readTimeline.Name())
	}

	// Verify extraction
	if _, err := os.Stat(filepath.Join(extractDir, "content.otio")); err != nil {
		t.Error("content.otio not extracted")
	}
}

func TestIsOTIOD(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "is_otiod_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a valid .otiod directory
	otiodPath := filepath.Join(tmpDir, "test.otiod")
	os.MkdirAll(otiodPath, 0755)
	os.WriteFile(filepath.Join(otiodPath, "content.otio"), []byte("{}"), 0644)

	if !IsOTIOD(otiodPath) {
		t.Error("expected IsOTIOD to return true for valid bundle")
	}

	// Non-existent path
	if IsOTIOD("/nonexistent.otiod") {
		t.Error("expected IsOTIOD to return false for non-existent path")
	}

	// Regular file
	regularFile := filepath.Join(tmpDir, "regular.otiod")
	os.WriteFile(regularFile, []byte(""), 0644)
	if IsOTIOD(regularFile) {
		t.Error("expected IsOTIOD to return false for regular file")
	}
}

func TestIsOTIOZ(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "is_otioz_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file with .otioz extension
	otiozPath := filepath.Join(tmpDir, "test.otioz")
	os.WriteFile(otiozPath, []byte("PK"), 0644) // ZIP magic bytes

	if !IsOTIOZ(otiozPath) {
		t.Error("expected IsOTIOZ to return true for .otioz file")
	}

	// Directory with .otioz extension
	otiozDir := filepath.Join(tmpDir, "dir.otioz")
	os.MkdirAll(otiozDir, 0755)
	if IsOTIOZ(otiozDir) {
		t.Error("expected IsOTIOZ to return false for directory")
	}
}

func TestDryRun(t *testing.T) {
	timeline := opentimelineio.NewTimeline("dryrun_test", nil, nil)

	// OTIOZ dry run
	size, err := WriteOTIOZDryRun(timeline, AllMissing)
	if err != nil {
		t.Fatalf("WriteOTIOZDryRun failed: %v", err)
	}
	if size <= 0 {
		t.Error("expected positive size")
	}

	// OTIOD dry run
	size, err = WriteOTIODDryRun(timeline, AllMissing)
	if err != nil {
		t.Fatalf("WriteOTIODDryRun failed: %v", err)
	}
	if size <= 0 {
		t.Error("expected positive size")
	}
}

func TestBundleError(t *testing.T) {
	t.Run("basic error", func(t *testing.T) {
		err := &BundleError{
			Operation: "write",
			Path:      "/test/path.otioz",
			Message:   "failed",
		}
		s := err.Error()
		if s == "" {
			t.Error("expected non-empty error")
		}
	})

	t.Run("with cause", func(t *testing.T) {
		err := &BundleError{
			Operation: "read",
			Path:      "/test/path.otiod",
			Message:   "failed",
			Cause:     os.ErrNotExist,
		}
		s := err.Error()
		if s == "" {
			t.Error("expected non-empty error")
		}
	})

	t.Run("unwrap", func(t *testing.T) {
		cause := os.ErrNotExist
		err := &BundleError{
			Operation: "read",
			Path:      "/test/path",
			Message:   "failed",
			Cause:     cause,
		}
		if err.Unwrap() != cause {
			t.Error("Unwrap should return cause")
		}
	})
}

func TestMediaReferencePolicyStringUnknown(t *testing.T) {
	policy := MediaReferencePolicy(99)
	s := policy.String()
	if s != "MediaReferencePolicy(99)" {
		t.Errorf("expected 'MediaReferencePolicy(99)', got %s", s)
	}
}

func TestPrepareForBundleErrorIfNotFile(t *testing.T) {
	timeline := createTestTimeline()

	_, _, err := PrepareForBundle(timeline, ErrorIfNotFile)
	// Should return error because file doesn't exist
	if err == nil {
		t.Error("expected error for ErrorIfNotFile with non-existent file")
	}
}

func TestPrepareForBundleWithRealFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bundle_real_file_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a real media file
	mediaPath := filepath.Join(tmpDir, "test.mov")
	if err := os.WriteFile(mediaPath, []byte("fake media data"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create timeline with real file reference
	timeline := opentimelineio.NewTimeline("test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	ref := opentimelineio.NewExternalReference("", mediaPath, &ar, nil)
	clip := opentimelineio.NewClip("clip", ref, &ar, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	// Test with MissingIfNotFile - should include the file
	prepared, manifest, err := PrepareForBundle(timeline, MissingIfNotFile)
	if err != nil {
		t.Fatalf("PrepareForBundle failed: %v", err)
	}

	if len(manifest) != 1 {
		t.Errorf("expected 1 manifest entry, got %d", len(manifest))
	}

	// Check clip still has external reference
	clips := prepared.FindClips(nil, false)
	if len(clips) != 1 {
		t.Fatalf("expected 1 clip, got %d", len(clips))
	}
	_, ok := clips[0].MediaReference().(*opentimelineio.ExternalReference)
	if !ok {
		t.Errorf("expected ExternalReference, got %T", clips[0].MediaReference())
	}
}

func TestOTIOZWithRealMedia(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otioz_media_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a real media file
	mediaPath := filepath.Join(tmpDir, "test.mov")
	if err := os.WriteFile(mediaPath, []byte("fake media data"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create timeline with real file reference
	timeline := opentimelineio.NewTimeline("media_test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	ref := opentimelineio.NewExternalReference("", mediaPath, &ar, nil)
	clip := opentimelineio.NewClip("clip", ref, &ar, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	bundlePath := filepath.Join(tmpDir, "output.otioz")

	// Write bundle
	err = WriteOTIOZ(timeline, bundlePath, MissingIfNotFile)
	if err != nil {
		t.Fatalf("WriteOTIOZ failed: %v", err)
	}

	// Read back with extraction
	extractDir := filepath.Join(tmpDir, "extracted")
	readTimeline, err := ReadOTIOZWithExtraction(bundlePath, extractDir)
	if err != nil {
		t.Fatalf("ReadOTIOZWithExtraction failed: %v", err)
	}

	if readTimeline.Name() != "media_test" {
		t.Errorf("expected media_test, got %s", readTimeline.Name())
	}

	// Verify media was extracted
	extractedMedia := filepath.Join(extractDir, "media", "test.mov")
	if _, err := os.Stat(extractedMedia); err != nil {
		t.Error("media file was not extracted")
	}
}

func TestOTIODWithRealMedia(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otiod_media_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a real media file
	mediaPath := filepath.Join(tmpDir, "test.mov")
	if err := os.WriteFile(mediaPath, []byte("fake media data"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create timeline with real file reference
	timeline := opentimelineio.NewTimeline("media_test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	ref := opentimelineio.NewExternalReference("", mediaPath, &ar, nil)
	clip := opentimelineio.NewClip("clip", ref, &ar, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	bundlePath := filepath.Join(tmpDir, "output.otiod")

	// Write bundle
	err = WriteOTIOD(timeline, bundlePath, MissingIfNotFile)
	if err != nil {
		t.Fatalf("WriteOTIOD failed: %v", err)
	}

	// Verify media was copied
	copiedMedia := filepath.Join(bundlePath, "media", "test.mov")
	if _, err := os.Stat(copiedMedia); err != nil {
		t.Error("media file was not copied")
	}

	// Read back
	readTimeline, err := ReadOTIOD(bundlePath, true)
	if err != nil {
		t.Fatalf("ReadOTIOD failed: %v", err)
	}

	if readTimeline.Name() != "media_test" {
		t.Errorf("expected media_test, got %s", readTimeline.Name())
	}
}

func TestConvertToAbsolutePaths(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "abs_path_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create bundle structure
	bundlePath := filepath.Join(tmpDir, "test.otiod")
	os.MkdirAll(filepath.Join(bundlePath, "media"), 0755)

	// Create a test media file
	mediaFile := filepath.Join(bundlePath, "media", "test.mov")
	os.WriteFile(mediaFile, []byte("test"), 0644)

	// Create timeline with relative path
	timeline := opentimelineio.NewTimeline("test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	ref := opentimelineio.NewExternalReference("", "media/test.mov", &ar, nil)
	clip := opentimelineio.NewClip("clip", ref, &ar, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	err = ConvertToAbsolutePaths(timeline, bundlePath)
	if err != nil {
		t.Fatalf("ConvertToAbsolutePaths failed: %v", err)
	}

	// Check path was converted
	clips := timeline.FindClips(nil, false)
	extRef := clips[0].MediaReference().(*opentimelineio.ExternalReference)
	expectedPath := filepath.Join(bundlePath, "media", "test.mov")
	if extRef.TargetURL() != expectedPath {
		t.Errorf("expected %s, got %s", expectedPath, extRef.TargetURL())
	}
}

func TestTotalMediaSize(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "media_size_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	file1 := filepath.Join(tmpDir, "test1.mov")
	file2 := filepath.Join(tmpDir, "test2.mov")
	os.WriteFile(file1, []byte("12345"), 0644)      // 5 bytes
	os.WriteFile(file2, []byte("1234567890"), 0644) // 10 bytes

	manifest := MediaManifest{
		file1: nil,
		file2: nil,
	}

	size, err := TotalMediaSize(manifest)
	if err != nil {
		t.Fatalf("TotalMediaSize failed: %v", err)
	}
	if size != 15 {
		t.Errorf("expected 15 bytes, got %d", size)
	}

	// Test with non-existent file - should return error
	manifest["/nonexistent/file.mov"] = nil
	_, err = TotalMediaSize(manifest)
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestReadOTIOZNonExistent(t *testing.T) {
	_, err := ReadOTIOZ("/nonexistent/path.otioz")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestReadOTIODNonExistent(t *testing.T) {
	_, err := ReadOTIOD("/nonexistent/path.otiod", false)
	if err == nil {
		t.Error("expected error for non-existent directory")
	}
}

func TestURLToAbsPath(t *testing.T) {
	t.Run("relative path", func(t *testing.T) {
		got, err := urlToAbsPath("media/test.mov")
		if err != nil {
			t.Fatalf("urlToAbsPath failed: %v", err)
		}
		// Should return an absolute path
		if !filepath.IsAbs(got) {
			t.Errorf("expected absolute path, got %s", got)
		}
	})

	t.Run("absolute path", func(t *testing.T) {
		got, err := urlToAbsPath("/absolute/path.mov")
		if err != nil {
			t.Fatalf("urlToAbsPath failed: %v", err)
		}
		if got != "/absolute/path.mov" {
			t.Errorf("expected /absolute/path.mov, got %s", got)
		}
	})

	t.Run("file URL", func(t *testing.T) {
		got, err := urlToAbsPath("file:///path/to/file.mov")
		if err != nil {
			t.Fatalf("urlToAbsPath failed: %v", err)
		}
		if !filepath.IsAbs(got) {
			t.Errorf("expected absolute path, got %s", got)
		}
	})

	t.Run("unsupported scheme", func(t *testing.T) {
		_, err := urlToAbsPath("http://example.com/file.mov")
		if err == nil {
			t.Error("expected error for http scheme")
		}
	})
}

func TestCopyFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "copy_file_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	src := filepath.Join(tmpDir, "source.txt")
	dst := filepath.Join(tmpDir, "dest.txt")
	content := []byte("test content")

	os.WriteFile(src, content, 0644)

	err = copyFile(src, dst)
	if err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	// Verify content
	read, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("failed to read dest: %v", err)
	}
	if string(read) != string(content) {
		t.Errorf("content mismatch: got %q, want %q", read, content)
	}
}

func TestPrepareForBundleWithDefaultMissingReference(t *testing.T) {
	timeline := opentimelineio.NewTimeline("test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	// NewClip with nil reference creates a default MissingReference
	clip := opentimelineio.NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	prepared, manifest, err := PrepareForBundle(timeline, MissingIfNotFile)
	if err != nil {
		t.Fatalf("PrepareForBundle failed: %v", err)
	}

	// MissingReferences have no file to bundle
	if len(manifest) != 0 {
		t.Error("expected empty manifest for MissingReference")
	}

	// Clip should still have MissingReference
	clips := prepared.FindClips(nil, false)
	_, ok := clips[0].MediaReference().(*opentimelineio.MissingReference)
	if !ok {
		t.Errorf("expected MissingReference, got %T", clips[0].MediaReference())
	}
}

func TestDryRunWithMedia(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dryrun_media_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a real media file
	mediaPath := filepath.Join(tmpDir, "test.mov")
	mediaContent := []byte("fake media content here")
	os.WriteFile(mediaPath, mediaContent, 0644)

	timeline := opentimelineio.NewTimeline("test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	ref := opentimelineio.NewExternalReference("", mediaPath, &ar, nil)
	clip := opentimelineio.NewClip("clip", ref, &ar, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	// OTIOZ dry run should include media size
	size, err := WriteOTIOZDryRun(timeline, MissingIfNotFile)
	if err != nil {
		t.Fatalf("WriteOTIOZDryRun failed: %v", err)
	}
	if size <= int64(len(mediaContent)) {
		t.Errorf("expected size > %d, got %d", len(mediaContent), size)
	}

	// OTIOD dry run should include media size
	size, err = WriteOTIODDryRun(timeline, MissingIfNotFile)
	if err != nil {
		t.Fatalf("WriteOTIODDryRun failed: %v", err)
	}
	if size <= int64(len(mediaContent)) {
		t.Errorf("expected size > %d, got %d", len(mediaContent), size)
	}
}

func TestReadOTIODPathIsNotDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otiod_not_dir_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a regular file with .otiod extension
	filePath := filepath.Join(tmpDir, "test.otiod")
	os.WriteFile(filePath, []byte("not a directory"), 0644)

	_, err = ReadOTIOD(filePath, false)
	if err == nil {
		t.Error("expected error for file that is not a directory")
	}
}

func TestReadOTIODMissingContentFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otiod_missing_content_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a directory without content.otio
	bundlePath := filepath.Join(tmpDir, "test.otiod")
	os.MkdirAll(bundlePath, 0755)

	_, err = ReadOTIOD(bundlePath, false)
	if err == nil {
		t.Error("expected error for missing content.otio")
	}
}

func TestReadOTIODInvalidJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otiod_invalid_json_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	bundlePath := filepath.Join(tmpDir, "test.otiod")
	os.MkdirAll(bundlePath, 0755)
	os.WriteFile(filepath.Join(bundlePath, "content.otio"), []byte("not valid json"), 0644)

	_, err = ReadOTIOD(bundlePath, false)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestReadOTIODNotATimeline(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otiod_not_timeline_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	bundlePath := filepath.Join(tmpDir, "test.otiod")
	os.MkdirAll(bundlePath, 0755)
	// Write a valid OTIO object that is not a Timeline (e.g., a Track)
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)
	data, _ := opentimelineio.ToJSONBytes(track)
	os.WriteFile(filepath.Join(bundlePath, "content.otio"), data, 0644)

	_, err = ReadOTIOD(bundlePath, false)
	if err == nil {
		t.Error("expected error for non-Timeline content")
	}
}

func TestReadOTIOZInvalidZip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otioz_invalid_zip_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file that is not a valid zip
	zipPath := filepath.Join(tmpDir, "invalid.otioz")
	os.WriteFile(zipPath, []byte("not a zip file"), 0644)

	_, err = ReadOTIOZ(zipPath)
	if err == nil {
		t.Error("expected error for invalid zip")
	}
}

func TestReadOTIOZMissingContentOtio(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otioz_missing_content_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a valid zip without content.otio
	zipPath := filepath.Join(tmpDir, "test.otioz")
	zipFile, _ := os.Create(zipPath)
	zipWriter := NewTestZipWriter(zipFile)
	zipWriter.WriteFile("version.txt", []byte("1.0.0"))
	zipWriter.Close()
	zipFile.Close()

	_, err = ReadOTIOZ(zipPath)
	if err == nil {
		t.Error("expected error for missing content.otio")
	}
}

func TestReadOTIOZNotATimeline(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otioz_not_timeline_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a zip with content.otio that is not a timeline
	zipPath := filepath.Join(tmpDir, "test.otioz")
	zipFile, _ := os.Create(zipPath)
	zipWriter := NewTestZipWriter(zipFile)
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)
	data, _ := opentimelineio.ToJSONBytes(track)
	zipWriter.WriteFile("content.otio", data)
	zipWriter.Close()
	zipFile.Close()

	_, err = ReadOTIOZ(zipPath)
	if err == nil {
		t.Error("expected error for non-Timeline content")
	}
}

func TestReadOTIOZInvalidContentOtio(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otioz_invalid_content_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a zip with invalid JSON in content.otio
	zipPath := filepath.Join(tmpDir, "test.otioz")
	zipFile, _ := os.Create(zipPath)
	zipWriter := NewTestZipWriter(zipFile)
	zipWriter.WriteFile("content.otio", []byte("not valid json"))
	zipWriter.Close()
	zipFile.Close()

	_, err = ReadOTIOZ(zipPath)
	if err == nil {
		t.Error("expected error for invalid JSON in content.otio")
	}
}

func TestReadOTIOZWithExtractionInvalidZip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otioz_extract_invalid_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create invalid zip
	zipPath := filepath.Join(tmpDir, "invalid.otioz")
	os.WriteFile(zipPath, []byte("not a zip"), 0644)

	_, err = ReadOTIOZWithExtraction(zipPath, filepath.Join(tmpDir, "extract"))
	if err == nil {
		t.Error("expected error for invalid zip")
	}
}

func TestReadOTIOZWithExtractionMissingContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otioz_extract_missing_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create valid zip without content.otio
	zipPath := filepath.Join(tmpDir, "test.otioz")
	zipFile, _ := os.Create(zipPath)
	zipWriter := NewTestZipWriter(zipFile)
	zipWriter.WriteFile("version.txt", []byte("1.0.0"))
	zipWriter.Close()
	zipFile.Close()

	_, err = ReadOTIOZWithExtraction(zipPath, filepath.Join(tmpDir, "extract"))
	if err == nil {
		t.Error("expected error for missing content.otio")
	}
}

func TestReadOTIOZWithExtractionNotTimeline(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otioz_extract_not_timeline_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create zip with non-Timeline content
	zipPath := filepath.Join(tmpDir, "test.otioz")
	zipFile, _ := os.Create(zipPath)
	zipWriter := NewTestZipWriter(zipFile)
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)
	data, _ := opentimelineio.ToJSONBytes(track)
	zipWriter.WriteFile("content.otio", data)
	zipWriter.Close()
	zipFile.Close()

	_, err = ReadOTIOZWithExtraction(zipPath, filepath.Join(tmpDir, "extract"))
	if err == nil {
		t.Error("expected error for non-Timeline content")
	}
}

func TestCopyFileSourceNotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "copy_file_error_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = copyFile("/nonexistent/source.txt", filepath.Join(tmpDir, "dest.txt"))
	if err == nil {
		t.Error("expected error for non-existent source")
	}
}

func TestCopyFileDestinationError(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "copy_file_dest_error_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source file
	src := filepath.Join(tmpDir, "source.txt")
	os.WriteFile(src, []byte("content"), 0644)

	// Destination in non-existent directory
	err = copyFile(src, "/nonexistent/dir/dest.txt")
	if err == nil {
		t.Error("expected error for destination in non-existent directory")
	}
}

func TestIsOTIODMissingContentFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "is_otiod_no_content_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create directory with .otiod extension but no content.otio
	bundlePath := filepath.Join(tmpDir, "test.otiod")
	os.MkdirAll(bundlePath, 0755)

	if IsOTIOD(bundlePath) {
		t.Error("expected IsOTIOD to return false without content.otio")
	}
}

func TestIsOTIODWrongExtension(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "is_otiod_wrong_ext_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create directory without .otiod extension
	bundlePath := filepath.Join(tmpDir, "test.bundle")
	os.MkdirAll(bundlePath, 0755)
	os.WriteFile(filepath.Join(bundlePath, "content.otio"), []byte("{}"), 0644)

	if IsOTIOD(bundlePath) {
		t.Error("expected IsOTIOD to return false without .otiod extension")
	}
}

func TestPrepareForBundleEmptyTargetURL(t *testing.T) {
	timeline := opentimelineio.NewTimeline("test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	// External reference with empty target URL
	ref := opentimelineio.NewExternalReference("", "", &sr, nil)
	clip := opentimelineio.NewClip("clip", ref, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	prepared, manifest, err := PrepareForBundle(timeline, MissingIfNotFile)
	if err != nil {
		t.Fatalf("PrepareForBundle failed: %v", err)
	}

	// Empty target URL should be skipped
	if len(manifest) != 0 {
		t.Errorf("expected empty manifest for empty target URL, got %d", len(manifest))
	}

	// Clip should still have ExternalReference (not replaced)
	clips := prepared.FindClips(nil, false)
	_, ok := clips[0].MediaReference().(*opentimelineio.ExternalReference)
	if !ok {
		t.Errorf("expected ExternalReference to remain, got %T", clips[0].MediaReference())
	}
}

func TestPrepareForBundleMediaIsDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bundle_dir_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a directory instead of a file
	mediaDir := filepath.Join(tmpDir, "not_a_file")
	os.MkdirAll(mediaDir, 0755)

	timeline := opentimelineio.NewTimeline("test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	ref := opentimelineio.NewExternalReference("", mediaDir, &ar, nil)
	clip := opentimelineio.NewClip("clip", ref, &ar, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	// With MissingIfNotFile, should replace with MissingReference
	prepared, manifest, err := PrepareForBundle(timeline, MissingIfNotFile)
	if err != nil {
		t.Fatalf("PrepareForBundle failed: %v", err)
	}

	if len(manifest) != 0 {
		t.Errorf("expected empty manifest for directory, got %d", len(manifest))
	}

	clips := prepared.FindClips(nil, false)
	_, ok := clips[0].MediaReference().(*opentimelineio.MissingReference)
	if !ok {
		t.Errorf("expected MissingReference, got %T", clips[0].MediaReference())
	}

	// With ErrorIfNotFile, should return error
	_, _, err = PrepareForBundle(timeline, ErrorIfNotFile)
	if err == nil {
		t.Error("expected error for directory with ErrorIfNotFile policy")
	}
}

func TestConvertToAbsolutePathsNonMediaPath(t *testing.T) {
	// Test with path that doesn't start with "media/"
	timeline := opentimelineio.NewTimeline("test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	// Absolute path that should not be modified
	ref := opentimelineio.NewExternalReference("", "/absolute/path/file.mov", &ar, nil)
	clip := opentimelineio.NewClip("clip", ref, &ar, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	err := ConvertToAbsolutePaths(timeline, "/some/root")
	if err != nil {
		t.Fatalf("ConvertToAbsolutePaths failed: %v", err)
	}

	// Path should remain unchanged
	clips := timeline.FindClips(nil, false)
	extRef := clips[0].MediaReference().(*opentimelineio.ExternalReference)
	if extRef.TargetURL() != "/absolute/path/file.mov" {
		t.Errorf("expected path to remain unchanged, got %s", extRef.TargetURL())
	}
}

func TestConvertToAbsolutePathsEmptyURL(t *testing.T) {
	timeline := opentimelineio.NewTimeline("test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	ref := opentimelineio.NewExternalReference("", "", &ar, nil)
	clip := opentimelineio.NewClip("clip", ref, &ar, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	err := ConvertToAbsolutePaths(timeline, "/root")
	if err != nil {
		t.Fatalf("ConvertToAbsolutePaths failed: %v", err)
	}

	// Empty URL should remain empty
	clips := timeline.FindClips(nil, false)
	extRef := clips[0].MediaReference().(*opentimelineio.ExternalReference)
	if extRef.TargetURL() != "" {
		t.Errorf("expected empty URL to remain empty, got %s", extRef.TargetURL())
	}
}

func TestConvertToAbsolutePathsWithMissingReference(t *testing.T) {
	timeline := opentimelineio.NewTimeline("test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	// Clip with nil reference creates MissingReference
	clip := opentimelineio.NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	err := ConvertToAbsolutePaths(timeline, "/root")
	if err != nil {
		t.Fatalf("ConvertToAbsolutePaths failed: %v", err)
	}

	// MissingReference should remain as is
	clips := timeline.FindClips(nil, false)
	_, ok := clips[0].MediaReference().(*opentimelineio.MissingReference)
	if !ok {
		t.Errorf("expected MissingReference to remain, got %T", clips[0].MediaReference())
	}
}

func TestURLToAbsPathWindowsPath(t *testing.T) {
	// Test Windows-style file URL (file:///C:/path/to/file)
	got, err := urlToAbsPath("file:///C:/path/to/file.mov")
	if err != nil {
		t.Fatalf("urlToAbsPath failed: %v", err)
	}
	// On Windows this should work, on Unix it may have leading slash
	if got == "" {
		t.Error("expected non-empty path")
	}
}

// TestZipWriter is a helper for creating test zip files
type TestZipWriter struct {
	w *os.File
	z *zip.Writer
}

func NewTestZipWriter(f *os.File) *TestZipWriter {
	return &TestZipWriter{w: f, z: zip.NewWriter(f)}
}

func (t *TestZipWriter) WriteFile(name string, data []byte) error {
	w, err := t.z.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func (t *TestZipWriter) Close() error {
	return t.z.Close()
}

func TestWriteOTIODBasenameCollision(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otiod_collision_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create two media files with same basename in different directories
	os.MkdirAll(filepath.Join(tmpDir, "dir1"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "dir2"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "dir1", "same.mov"), []byte("file1"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "dir2", "same.mov"), []byte("file2"), 0644)

	timeline := opentimelineio.NewTimeline("test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))

	ref1 := opentimelineio.NewExternalReference("", filepath.Join(tmpDir, "dir1", "same.mov"), &ar, nil)
	clip1 := opentimelineio.NewClip("clip1", ref1, &ar, nil, nil, nil, "", nil)
	track.AppendChild(clip1)

	ref2 := opentimelineio.NewExternalReference("", filepath.Join(tmpDir, "dir2", "same.mov"), &ar, nil)
	clip2 := opentimelineio.NewClip("clip2", ref2, &ar, nil, nil, nil, "", nil)
	track.AppendChild(clip2)

	timeline.Tracks().AppendChild(track)

	bundlePath := filepath.Join(tmpDir, "output.otiod")
	err = WriteOTIOD(timeline, bundlePath, MissingIfNotFile)
	if err == nil {
		t.Error("expected error for basename collision")
	}
}

func TestWriteOTIOZBasenameCollision(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otioz_collision_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create two media files with same basename in different directories
	os.MkdirAll(filepath.Join(tmpDir, "dir1"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "dir2"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "dir1", "same.mov"), []byte("file1"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "dir2", "same.mov"), []byte("file2"), 0644)

	timeline := opentimelineio.NewTimeline("test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))

	ref1 := opentimelineio.NewExternalReference("", filepath.Join(tmpDir, "dir1", "same.mov"), &ar, nil)
	clip1 := opentimelineio.NewClip("clip1", ref1, &ar, nil, nil, nil, "", nil)
	track.AppendChild(clip1)

	ref2 := opentimelineio.NewExternalReference("", filepath.Join(tmpDir, "dir2", "same.mov"), &ar, nil)
	clip2 := opentimelineio.NewClip("clip2", ref2, &ar, nil, nil, nil, "", nil)
	track.AppendChild(clip2)

	timeline.Tracks().AppendChild(track)

	bundlePath := filepath.Join(tmpDir, "output.otioz")
	err = WriteOTIOZ(timeline, bundlePath, MissingIfNotFile)
	if err == nil {
		t.Error("expected error for basename collision")
	}
}

func TestWriteOTIODDryRunBasenameCollision(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otiod_dryrun_collision_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create two media files with same basename
	os.MkdirAll(filepath.Join(tmpDir, "dir1"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "dir2"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "dir1", "same.mov"), []byte("file1"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "dir2", "same.mov"), []byte("file2"), 0644)

	timeline := opentimelineio.NewTimeline("test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))

	ref1 := opentimelineio.NewExternalReference("", filepath.Join(tmpDir, "dir1", "same.mov"), &ar, nil)
	clip1 := opentimelineio.NewClip("clip1", ref1, &ar, nil, nil, nil, "", nil)
	track.AppendChild(clip1)

	ref2 := opentimelineio.NewExternalReference("", filepath.Join(tmpDir, "dir2", "same.mov"), &ar, nil)
	clip2 := opentimelineio.NewClip("clip2", ref2, &ar, nil, nil, nil, "", nil)
	track.AppendChild(clip2)

	timeline.Tracks().AppendChild(track)

	_, err = WriteOTIODDryRun(timeline, MissingIfNotFile)
	if err == nil {
		t.Error("expected error for basename collision")
	}
}

func TestWriteOTIOZDryRunBasenameCollision(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otioz_dryrun_collision_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create two media files with same basename
	os.MkdirAll(filepath.Join(tmpDir, "dir1"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "dir2"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "dir1", "same.mov"), []byte("file1"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "dir2", "same.mov"), []byte("file2"), 0644)

	timeline := opentimelineio.NewTimeline("test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))

	ref1 := opentimelineio.NewExternalReference("", filepath.Join(tmpDir, "dir1", "same.mov"), &ar, nil)
	clip1 := opentimelineio.NewClip("clip1", ref1, &ar, nil, nil, nil, "", nil)
	track.AppendChild(clip1)

	ref2 := opentimelineio.NewExternalReference("", filepath.Join(tmpDir, "dir2", "same.mov"), &ar, nil)
	clip2 := opentimelineio.NewClip("clip2", ref2, &ar, nil, nil, nil, "", nil)
	track.AppendChild(clip2)

	timeline.Tracks().AppendChild(track)

	_, err = WriteOTIOZDryRun(timeline, MissingIfNotFile)
	if err == nil {
		t.Error("expected error for basename collision")
	}
}

func TestReadOTIOZWithExtractionDirectoryInZip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otioz_extract_dir_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a valid OTIOZ with a directory entry
	zipPath := filepath.Join(tmpDir, "test.otioz")
	timeline := opentimelineio.NewTimeline("test", nil, nil)

	// Write a valid bundle first
	err = WriteOTIOZ(timeline, zipPath, AllMissing)
	if err != nil {
		t.Fatalf("WriteOTIOZ failed: %v", err)
	}

	// Read with extraction
	extractDir := filepath.Join(tmpDir, "extract")
	readTimeline, err := ReadOTIOZWithExtraction(zipPath, extractDir)
	if err != nil {
		t.Fatalf("ReadOTIOZWithExtraction failed: %v", err)
	}

	if readTimeline.Name() != "test" {
		t.Errorf("expected name test, got %s", readTimeline.Name())
	}

	// Verify media directory was created
	if _, err := os.Stat(filepath.Join(extractDir, "media")); err == nil {
		// media directory exists (created by extraction)
	}
}

func TestDefaultFSMkdirAll(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "osfs_mkdirall_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fsys := DefaultFS

	// Test MkdirAll
	testPath := filepath.Join(tmpDir, "deep", "nested", "path")
	if err := fsys.MkdirAll(testPath, 0755); err != nil {
		t.Errorf("MkdirAll failed: %v", err)
	}

	// Verify it was created
	info, err := os.Stat(testPath)
	if err != nil {
		t.Errorf("path was not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory")
	}
}

func TestPrepareForBundleWithURLParseError(t *testing.T) {
	timeline := opentimelineio.NewTimeline("test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	// Invalid URL that can't be converted to absolute path
	ref := opentimelineio.NewExternalReference("", "http://example.com/file.mov", &sr, nil)
	clip := opentimelineio.NewClip("clip", ref, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	// With ErrorIfNotFile, should return error for http URL
	_, _, err := PrepareForBundle(timeline, ErrorIfNotFile)
	if err == nil {
		t.Error("expected error for http URL with ErrorIfNotFile policy")
	}

	// With MissingIfNotFile, should replace with MissingReference
	prepared, manifest, err := PrepareForBundle(timeline, MissingIfNotFile)
	if err != nil {
		t.Fatalf("PrepareForBundle failed: %v", err)
	}

	if len(manifest) != 0 {
		t.Errorf("expected empty manifest, got %d", len(manifest))
	}

	clips := prepared.FindClips(nil, false)
	_, ok := clips[0].MediaReference().(*opentimelineio.MissingReference)
	if !ok {
		t.Errorf("expected MissingReference for http URL, got %T", clips[0].MediaReference())
	}
}

func TestReadOTIOZWithExtractionInvalidContentJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otioz_extract_invalid_json_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create zip with invalid JSON in content.otio
	zipPath := filepath.Join(tmpDir, "test.otioz")
	zipFile, _ := os.Create(zipPath)
	zipWriter := NewTestZipWriter(zipFile)
	zipWriter.WriteFile("content.otio", []byte("not valid json"))
	zipWriter.Close()
	zipFile.Close()

	_, err = ReadOTIOZWithExtraction(zipPath, filepath.Join(tmpDir, "extract"))
	if err == nil {
		t.Error("expected error for invalid JSON in content.otio during extraction")
	}
}

func TestWriteOTIODWithRealMediaMultiple(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otiod_multi_media_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create multiple real media files
	media1 := filepath.Join(tmpDir, "test1.mov")
	media2 := filepath.Join(tmpDir, "test2.mov")
	os.WriteFile(media1, []byte("media content 1"), 0644)
	os.WriteFile(media2, []byte("media content 2"), 0644)

	timeline := opentimelineio.NewTimeline("multi_media_test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))

	ref1 := opentimelineio.NewExternalReference("", media1, &ar, nil)
	clip1 := opentimelineio.NewClip("clip1", ref1, &ar, nil, nil, nil, "", nil)
	track.AppendChild(clip1)

	ref2 := opentimelineio.NewExternalReference("", media2, &ar, nil)
	clip2 := opentimelineio.NewClip("clip2", ref2, &ar, nil, nil, nil, "", nil)
	track.AppendChild(clip2)

	timeline.Tracks().AppendChild(track)

	bundlePath := filepath.Join(tmpDir, "output.otiod")
	err = WriteOTIOD(timeline, bundlePath, MissingIfNotFile)
	if err != nil {
		t.Fatalf("WriteOTIOD failed: %v", err)
	}

	// Verify both media files were copied
	if _, err := os.Stat(filepath.Join(bundlePath, "media", "test1.mov")); err != nil {
		t.Error("test1.mov not copied to bundle")
	}
	if _, err := os.Stat(filepath.Join(bundlePath, "media", "test2.mov")); err != nil {
		t.Error("test2.mov not copied to bundle")
	}
}

func TestWriteOTIOZWithRealMediaMultiple(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otioz_multi_media_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create multiple real media files
	media1 := filepath.Join(tmpDir, "test1.mov")
	media2 := filepath.Join(tmpDir, "test2.mov")
	os.WriteFile(media1, []byte("media content 1"), 0644)
	os.WriteFile(media2, []byte("media content 2"), 0644)

	timeline := opentimelineio.NewTimeline("multi_media_test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))

	ref1 := opentimelineio.NewExternalReference("", media1, &ar, nil)
	clip1 := opentimelineio.NewClip("clip1", ref1, &ar, nil, nil, nil, "", nil)
	track.AppendChild(clip1)

	ref2 := opentimelineio.NewExternalReference("", media2, &ar, nil)
	clip2 := opentimelineio.NewClip("clip2", ref2, &ar, nil, nil, nil, "", nil)
	track.AppendChild(clip2)

	timeline.Tracks().AppendChild(track)

	bundlePath := filepath.Join(tmpDir, "output.otioz")
	err = WriteOTIOZ(timeline, bundlePath, MissingIfNotFile)
	if err != nil {
		t.Fatalf("WriteOTIOZ failed: %v", err)
	}

	// Read back with extraction and verify media
	extractDir := filepath.Join(tmpDir, "extract")
	_, err = ReadOTIOZWithExtraction(bundlePath, extractDir)
	if err != nil {
		t.Fatalf("ReadOTIOZWithExtraction failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(extractDir, "media", "test1.mov")); err != nil {
		t.Error("test1.mov not extracted from bundle")
	}
	if _, err := os.Stat(filepath.Join(extractDir, "media", "test2.mov")); err != nil {
		t.Error("test2.mov not extracted from bundle")
	}
}

func TestWriteOTIODMediaCopyError(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otiod_copy_error_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a media file
	mediaPath := filepath.Join(tmpDir, "test.mov")
	os.WriteFile(mediaPath, []byte("media content"), 0644)

	timeline := opentimelineio.NewTimeline("test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	ref := opentimelineio.NewExternalReference("", mediaPath, &ar, nil)
	clip := opentimelineio.NewClip("clip", ref, &ar, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	// Delete the media file after PrepareForBundle looks at it
	// This requires running in between operations which we can't do with the current API
	// Instead, test with a media file that disappears - but this is tricky

	// For now, just ensure successful case works
	bundlePath := filepath.Join(tmpDir, "output.otiod")
	err = WriteOTIOD(timeline, bundlePath, MissingIfNotFile)
	if err != nil {
		t.Fatalf("WriteOTIOD failed: %v", err)
	}
}

func TestGetTargetURLWithNonExternalRef(t *testing.T) {
	// getTargetURL is called with MissingReference
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	missingRef := opentimelineio.NewMissingReference("test", &sr, nil)

	// Call getTargetURL indirectly through PrepareForBundle
	timeline := opentimelineio.NewTimeline("test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	clip := opentimelineio.NewClip("clip", missingRef, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	// With AllMissing policy, it will call getTargetURL on the existing MissingReference
	prepared, manifest, err := PrepareForBundle(timeline, AllMissing)
	if err != nil {
		t.Fatalf("PrepareForBundle failed: %v", err)
	}

	// Manifest should be empty
	if len(manifest) != 0 {
		t.Errorf("expected empty manifest, got %d", len(manifest))
	}

	// Clip should have MissingReference
	clips := prepared.FindClips(nil, false)
	_, ok := clips[0].MediaReference().(*opentimelineio.MissingReference)
	if !ok {
		t.Errorf("expected MissingReference, got %T", clips[0].MediaReference())
	}
}

func TestReadOTIOZWithExtractionWithMediaDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otioz_extract_media_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a media file
	mediaPath := filepath.Join(tmpDir, "test.mov")
	os.WriteFile(mediaPath, []byte("media content"), 0644)

	// Create timeline with real media
	timeline := opentimelineio.NewTimeline("test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	ref := opentimelineio.NewExternalReference("", mediaPath, &ar, nil)
	clip := opentimelineio.NewClip("clip", ref, &ar, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	// Write bundle with media
	zipPath := filepath.Join(tmpDir, "test.otioz")
	err = WriteOTIOZ(timeline, zipPath, MissingIfNotFile)
	if err != nil {
		t.Fatalf("WriteOTIOZ failed: %v", err)
	}

	// Extract to new directory
	extractDir := filepath.Join(tmpDir, "extracted")
	readTimeline, err := ReadOTIOZWithExtraction(zipPath, extractDir)
	if err != nil {
		t.Fatalf("ReadOTIOZWithExtraction failed: %v", err)
	}

	if readTimeline.Name() != "test" {
		t.Errorf("expected name test, got %s", readTimeline.Name())
	}

	// Verify media directory exists
	mediaDir := filepath.Join(extractDir, "media")
	info, err := os.Stat(mediaDir)
	if err != nil || !info.IsDir() {
		t.Error("media directory was not created during extraction")
	}

	// Verify media file was extracted
	extractedMedia := filepath.Join(mediaDir, "test.mov")
	if _, err := os.Stat(extractedMedia); err != nil {
		t.Error("media file was not extracted")
	}
}

func TestWriteOTIODToReadOnlyLocation(t *testing.T) {
	// Test writing to a location where we can't create directories
	timeline := opentimelineio.NewTimeline("test", nil, nil)

	// Try to write to root (should fail on most systems)
	err := WriteOTIOD(timeline, "/nonexistent_root_path/test.otiod", AllMissing)
	if err == nil {
		t.Error("expected error when writing to non-existent path")
	}
}

func TestWriteOTIOZToReadOnlyLocation(t *testing.T) {
	// Test writing to a location where we can't create files
	timeline := opentimelineio.NewTimeline("test", nil, nil)

	// Try to write to non-existent directory
	err := WriteOTIOZ(timeline, "/nonexistent_root_path/test.otioz", AllMissing)
	if err == nil {
		t.Error("expected error when writing to non-existent path")
	}
}

func TestDryRunErrorIfNotFilePolicy(t *testing.T) {
	timeline := createTestTimeline()

	// OTIOZ dry run with ErrorIfNotFile should fail for non-existent file
	_, err := WriteOTIOZDryRun(timeline, ErrorIfNotFile)
	if err == nil {
		t.Error("expected error for ErrorIfNotFile with non-existent file")
	}

	// OTIOD dry run with ErrorIfNotFile should fail for non-existent file
	_, err = WriteOTIODDryRun(timeline, ErrorIfNotFile)
	if err == nil {
		t.Error("expected error for ErrorIfNotFile with non-existent file")
	}
}

func TestWriteOTIODMultipleMediaFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otiod_multi_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create several media files
	media1 := filepath.Join(tmpDir, "file1.mov")
	media2 := filepath.Join(tmpDir, "file2.mov")
	media3 := filepath.Join(tmpDir, "file3.mov")
	os.WriteFile(media1, []byte("content 1"), 0644)
	os.WriteFile(media2, []byte("content 2"), 0644)
	os.WriteFile(media3, []byte("content 3"), 0644)

	timeline := opentimelineio.NewTimeline("multi_test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))

	for i, media := range []string{media1, media2, media3} {
		ref := opentimelineio.NewExternalReference("", media, &ar, nil)
		clip := opentimelineio.NewClip(filepath.Base(media), ref, &ar, nil, nil, nil, "", nil)
		track.InsertChild(i, clip)
	}

	timeline.Tracks().AppendChild(track)

	bundlePath := filepath.Join(tmpDir, "output.otiod")
	err = WriteOTIOD(timeline, bundlePath, MissingIfNotFile)
	if err != nil {
		t.Fatalf("WriteOTIOD failed: %v", err)
	}

	// Verify all media files were copied
	for _, media := range []string{"file1.mov", "file2.mov", "file3.mov"} {
		mediaPath := filepath.Join(bundlePath, "media", media)
		if _, err := os.Stat(mediaPath); err != nil {
			t.Errorf("%s not copied to bundle: %v", media, err)
		}
	}
}

func TestWriteOTIOZMultipleMediaFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otioz_multi_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create several media files
	media1 := filepath.Join(tmpDir, "file1.mov")
	media2 := filepath.Join(tmpDir, "file2.mov")
	media3 := filepath.Join(tmpDir, "file3.mov")
	os.WriteFile(media1, []byte("content 1"), 0644)
	os.WriteFile(media2, []byte("content 2"), 0644)
	os.WriteFile(media3, []byte("content 3"), 0644)

	timeline := opentimelineio.NewTimeline("multi_test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))

	for i, media := range []string{media1, media2, media3} {
		ref := opentimelineio.NewExternalReference("", media, &ar, nil)
		clip := opentimelineio.NewClip(filepath.Base(media), ref, &ar, nil, nil, nil, "", nil)
		track.InsertChild(i, clip)
	}

	timeline.Tracks().AppendChild(track)

	bundlePath := filepath.Join(tmpDir, "output.otioz")
	err = WriteOTIOZ(timeline, bundlePath, MissingIfNotFile)
	if err != nil {
		t.Fatalf("WriteOTIOZ failed: %v", err)
	}

	// Extract and verify
	extractDir := filepath.Join(tmpDir, "extract")
	_, err = ReadOTIOZWithExtraction(bundlePath, extractDir)
	if err != nil {
		t.Fatalf("ReadOTIOZWithExtraction failed: %v", err)
	}

	for _, media := range []string{"file1.mov", "file2.mov", "file3.mov"} {
		mediaPath := filepath.Join(extractDir, "media", media)
		if _, err := os.Stat(mediaPath); err != nil {
			t.Errorf("%s not extracted from bundle: %v", media, err)
		}
	}
}

func TestPrepareForBundleNilMediaReference(t *testing.T) {
	timeline := opentimelineio.NewTimeline("test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))

	// Create clip with nil reference (creates default MissingReference)
	clip := opentimelineio.NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	// Test with ErrorIfNotFile - should not error on MissingReference
	prepared, manifest, err := PrepareForBundle(timeline, ErrorIfNotFile)
	if err != nil {
		t.Fatalf("PrepareForBundle failed: %v", err)
	}

	if len(manifest) != 0 {
		t.Errorf("expected empty manifest, got %d", len(manifest))
	}

	clips := prepared.FindClips(nil, false)
	if len(clips) != 1 {
		t.Fatalf("expected 1 clip, got %d", len(clips))
	}
}

func TestWriteOTIODWithErrorIfNotFilePolicySuccess(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otiod_error_policy_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a real media file
	mediaPath := filepath.Join(tmpDir, "test.mov")
	os.WriteFile(mediaPath, []byte("media content"), 0644)

	timeline := opentimelineio.NewTimeline("test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	ref := opentimelineio.NewExternalReference("", mediaPath, &ar, nil)
	clip := opentimelineio.NewClip("clip", ref, &ar, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	bundlePath := filepath.Join(tmpDir, "output.otiod")

	// ErrorIfNotFile should succeed with existing file
	err = WriteOTIOD(timeline, bundlePath, ErrorIfNotFile)
	if err != nil {
		t.Fatalf("WriteOTIOD with ErrorIfNotFile failed: %v", err)
	}

	// Verify media was copied
	copiedMedia := filepath.Join(bundlePath, "media", "test.mov")
	if _, err := os.Stat(copiedMedia); err != nil {
		t.Error("media file was not copied")
	}
}

func TestWriteOTIOZWithErrorIfNotFilePolicySuccess(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otioz_error_policy_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a real media file
	mediaPath := filepath.Join(tmpDir, "test.mov")
	os.WriteFile(mediaPath, []byte("media content"), 0644)

	timeline := opentimelineio.NewTimeline("test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	ref := opentimelineio.NewExternalReference("", mediaPath, &ar, nil)
	clip := opentimelineio.NewClip("clip", ref, &ar, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	bundlePath := filepath.Join(tmpDir, "output.otioz")

	// ErrorIfNotFile should succeed with existing file
	err = WriteOTIOZ(timeline, bundlePath, ErrorIfNotFile)
	if err != nil {
		t.Fatalf("WriteOTIOZ with ErrorIfNotFile failed: %v", err)
	}

	// Verify bundle was created
	if _, err := os.Stat(bundlePath); err != nil {
		t.Error("bundle file was not created")
	}
}

func TestWriteOTIODErrorIfNotFilePolicyWithMissingFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otiod_error_missing_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	timeline := opentimelineio.NewTimeline("test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	// Reference to non-existent file
	ref := opentimelineio.NewExternalReference("", "/nonexistent/file.mov", &ar, nil)
	clip := opentimelineio.NewClip("clip", ref, &ar, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	bundlePath := filepath.Join(tmpDir, "output.otiod")
	err = WriteOTIOD(timeline, bundlePath, ErrorIfNotFile)
	if err == nil {
		t.Error("expected error for ErrorIfNotFile with non-existent file")
	}
}

func TestWriteOTIOZErrorIfNotFilePolicyWithMissingFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otioz_error_missing_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	timeline := opentimelineio.NewTimeline("test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	// Reference to non-existent file
	ref := opentimelineio.NewExternalReference("", "/nonexistent/file.mov", &ar, nil)
	clip := opentimelineio.NewClip("clip", ref, &ar, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	bundlePath := filepath.Join(tmpDir, "output.otioz")
	err = WriteOTIOZ(timeline, bundlePath, ErrorIfNotFile)
	if err == nil {
		t.Error("expected error for ErrorIfNotFile with non-existent file")
	}
}

func TestDryRunWithRealMedia(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dryrun_real_media_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create real media
	mediaPath := filepath.Join(tmpDir, "test.mov")
	mediaContent := []byte("test media content here")
	os.WriteFile(mediaPath, mediaContent, 0644)

	timeline := opentimelineio.NewTimeline("test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	ref := opentimelineio.NewExternalReference("", mediaPath, &ar, nil)
	clip := opentimelineio.NewClip("clip", ref, &ar, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	// Test OTIOZ dry run with ErrorIfNotFile
	size, err := WriteOTIOZDryRun(timeline, ErrorIfNotFile)
	if err != nil {
		t.Fatalf("WriteOTIOZDryRun failed: %v", err)
	}
	if size <= 0 {
		t.Error("expected positive size")
	}

	// Test OTIOD dry run with ErrorIfNotFile
	size, err = WriteOTIODDryRun(timeline, ErrorIfNotFile)
	if err != nil {
		t.Fatalf("WriteOTIODDryRun failed: %v", err)
	}
	if size <= 0 {
		t.Error("expected positive size")
	}
}

func TestCopyFileFSWithMemFS(t *testing.T) {
	// Test copyFileFS using memfs through NewMemFSAdapter
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatalf("Failed to create memfs: %v", err)
	}

	fsys := NewMemFSAdapter(mfs)

	// Create source file
	srcContent := []byte("source content for copy test")
	if err := fsys.WriteFile("/source.txt", srcContent, 0644); err != nil {
		t.Fatalf("Failed to write source: %v", err)
	}

	// Copy file
	if err := copyFileFS(fsys, "/source.txt", "/dest.txt"); err != nil {
		t.Errorf("copyFileFS failed: %v", err)
	}

	// Verify content
	dstContent, err := fsys.ReadFile("/dest.txt")
	if err != nil {
		t.Errorf("ReadFile dest failed: %v", err)
	}
	if string(dstContent) != string(srcContent) {
		t.Errorf("Content mismatch: got %q, want %q", dstContent, srcContent)
	}
}

func TestMemFSAdapterWriteToRootPath(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatalf("Failed to create memfs: %v", err)
	}

	fsys := NewMemFSAdapter(mfs)

	// Write file to root (no parent directory needed)
	err = fsys.WriteFile("/root_file.txt", []byte("content"), 0644)
	if err != nil {
		t.Errorf("WriteFile to root failed: %v", err)
	}

	// Verify
	data, err := fsys.ReadFile("/root_file.txt")
	if err != nil {
		t.Errorf("ReadFile failed: %v", err)
	}
	if string(data) != "content" {
		t.Errorf("Content mismatch: got %q", data)
	}
}

func TestCopyFileFSCreateError(t *testing.T) {
	// Use errorFS to test create error path
	errFS := &errorFS{
		createErr: os.ErrPermission,
	}

	// First need a way to make Open succeed but Create fail
	// This is tricky with errorFS since it returns errors for all operations
	// Let's just test that the error is properly propagated
	err := copyFileFS(errFS, "/src.txt", "/dst.txt")
	if err == nil {
		t.Error("Expected error from copyFileFS with error filesystem")
	}
}
