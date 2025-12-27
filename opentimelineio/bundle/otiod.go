// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package bundle

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mrjoshuak/gotio/opentimelineio"
)

// ReadOTIOD reads a .otiod bundle directory and returns the timeline.
func ReadOTIOD(path string, absolutePaths bool) (*opentimelineio.Timeline, error) {
	// Check if directory exists
	info, err := os.Stat(path)
	if err != nil {
		return nil, &BundleError{
			Operation: "read",
			Path:      path,
			Message:   "bundle directory not found",
			Cause:     err,
		}
	}
	if !info.IsDir() {
		return nil, &BundleError{
			Operation: "read",
			Path:      path,
			Message:   "path is not a directory",
		}
	}

	// Read content.otio
	contentPath := filepath.Join(path, "content.otio")
	data, err := os.ReadFile(contentPath)
	if err != nil {
		return nil, &BundleError{
			Operation: "read",
			Path:      contentPath,
			Message:   "failed to read content.otio",
			Cause:     err,
		}
	}

	// Parse OTIO
	obj, err := opentimelineio.FromJSONBytes(data)
	if err != nil {
		return nil, &BundleError{
			Operation: "read",
			Path:      contentPath,
			Message:   "failed to parse content.otio",
			Cause:     err,
		}
	}

	timeline, ok := obj.(*opentimelineio.Timeline)
	if !ok {
		return nil, &BundleError{
			Operation: "read",
			Path:      path,
			Message:   "content.otio does not contain a Timeline",
		}
	}

	// Convert to absolute paths if requested
	if absolutePaths {
		ConvertToAbsolutePaths(timeline, path)
	}

	return timeline, nil
}

// WriteOTIOD writes a timeline and its media to a .otiod bundle directory.
func WriteOTIOD(
	timeline *opentimelineio.Timeline,
	path string,
	policy MediaReferencePolicy,
) error {
	// Prepare timeline and manifest
	prepared, manifest, err := PrepareForBundle(timeline, policy)
	if err != nil {
		return err
	}

	// Verify unique basenames
	if err := VerifyUniqueBasenames(manifest); err != nil {
		return err
	}

	// Relink to bundle paths
	RelinkToBundle(manifest)

	// Create bundle directory
	if err := os.MkdirAll(path, 0755); err != nil {
		return &BundleError{
			Operation: "write",
			Path:      path,
			Message:   "failed to create bundle directory",
			Cause:     err,
		}
	}

	// Create media directory
	mediaDir := filepath.Join(path, "media")
	if err := os.MkdirAll(mediaDir, 0755); err != nil {
		return &BundleError{
			Operation: "write",
			Path:      mediaDir,
			Message:   "failed to create media directory",
			Cause:     err,
		}
	}

	// Write content.otio
	contentData, err := opentimelineio.ToJSONBytesIndent(prepared, "    ")
	if err != nil {
		return &BundleError{
			Operation: "write",
			Path:      path,
			Message:   "failed to serialize timeline",
			Cause:     err,
		}
	}

	contentPath := filepath.Join(path, "content.otio")
	if err := os.WriteFile(contentPath, contentData, 0644); err != nil {
		return &BundleError{
			Operation: "write",
			Path:      contentPath,
			Message:   "failed to write content.otio",
			Cause:     err,
		}
	}

	// Copy media files
	for sourcePath := range manifest {
		basename := filepath.Base(sourcePath)
		destPath := filepath.Join(mediaDir, basename)

		if err := copyFile(sourcePath, destPath); err != nil {
			return &BundleError{
				Operation: "write",
				Path:      sourcePath,
				Message:   "failed to copy media file",
				Cause:     err,
			}
		}
	}

	return nil
}

// WriteOTIODDryRun calculates the total size of a .otiod bundle without writing.
func WriteOTIODDryRun(
	timeline *opentimelineio.Timeline,
	policy MediaReferencePolicy,
) (int64, error) {
	// Prepare timeline and manifest
	prepared, manifest, err := PrepareForBundle(timeline, policy)
	if err != nil {
		return 0, err
	}

	// Verify unique basenames
	if err := VerifyUniqueBasenames(manifest); err != nil {
		return 0, err
	}

	var total int64

	// Size of content.otio
	contentData, err := opentimelineio.ToJSONBytesIndent(prepared, "    ")
	if err != nil {
		return 0, err
	}
	total += int64(len(contentData))

	// Size of media files
	mediaSize, err := TotalMediaSize(manifest)
	if err != nil {
		return 0, err
	}
	total += mediaSize

	return total, nil
}

// IsOTIOD checks if a path is a valid .otiod bundle directory.
func IsOTIOD(path string) bool {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return false
	}

	// Check for .otiod extension
	if !strings.HasSuffix(path, ".otiod") {
		return false
	}

	// Check for content.otio
	contentPath := filepath.Join(path, "content.otio")
	if _, err := os.Stat(contentPath); err != nil {
		return false
	}

	return true
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
