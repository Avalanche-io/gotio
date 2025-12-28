// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package bundle

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Avalanche-io/gotio"
)

// ReadOTIOZ reads a .otioz bundle and returns the timeline.
// This only reads the content.otio file; media files are not extracted.
func ReadOTIOZ(path string) (*gotio.Timeline, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, &BundleError{
			Operation: "read",
			Path:      path,
			Message:   "failed to open zip",
			Cause:     err,
		}
	}
	defer r.Close()

	// Find content.otio
	var contentFile *zip.File
	for _, f := range r.File {
		if f.Name == "content.otio" {
			contentFile = f
			break
		}
	}

	if contentFile == nil {
		return nil, &BundleError{
			Operation: "read",
			Path:      path,
			Message:   "missing content.otio",
		}
	}

	// Read content.otio
	rc, err := contentFile.Open()
	if err != nil {
		return nil, &BundleError{
			Operation: "read",
			Path:      path,
			Message:   "failed to open content.otio",
			Cause:     err,
		}
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, &BundleError{
			Operation: "read",
			Path:      path,
			Message:   "failed to read content.otio",
			Cause:     err,
		}
	}

	// Parse OTIO
	obj, err := gotio.FromJSONBytes(data)
	if err != nil {
		return nil, &BundleError{
			Operation: "read",
			Path:      path,
			Message:   "failed to parse content.otio",
			Cause:     err,
		}
	}

	timeline, ok := obj.(*gotio.Timeline)
	if !ok {
		return nil, &BundleError{
			Operation: "read",
			Path:      path,
			Message:   "content.otio does not contain a Timeline",
		}
	}

	return timeline, nil
}

// ReadOTIOZWithExtraction reads a .otioz bundle and extracts all contents to a directory.
// Returns the timeline with media references pointing to extracted files.
func ReadOTIOZWithExtraction(bundlePath, extractDir string) (*gotio.Timeline, error) {
	r, err := zip.OpenReader(bundlePath)
	if err != nil {
		return nil, &BundleError{
			Operation: "extract",
			Path:      bundlePath,
			Message:   "failed to open zip",
			Cause:     err,
		}
	}
	defer r.Close()

	// Create extraction directory
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return nil, &BundleError{
			Operation: "extract",
			Path:      extractDir,
			Message:   "failed to create extraction directory",
			Cause:     err,
		}
	}

	var timeline *gotio.Timeline

	// Extract all files
	for _, f := range r.File {
		destPath := filepath.Join(extractDir, f.Name)

		// Create directories
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return nil, err
			}
			continue
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return nil, err
		}

		// Extract file
		if err := extractZipFile(f, destPath); err != nil {
			return nil, err
		}

		// Parse content.otio
		if f.Name == "content.otio" {
			data, err := os.ReadFile(destPath)
			if err != nil {
				return nil, err
			}

			obj, err := gotio.FromJSONBytes(data)
			if err != nil {
				return nil, err
			}

			var ok bool
			timeline, ok = obj.(*gotio.Timeline)
			if !ok {
				return nil, &BundleError{
					Operation: "extract",
					Path:      bundlePath,
					Message:   "content.otio does not contain a Timeline",
				}
			}
		}
	}

	if timeline == nil {
		return nil, &BundleError{
			Operation: "extract",
			Path:      bundlePath,
			Message:   "missing content.otio",
		}
	}

	// Convert relative paths to absolute
	ConvertToAbsolutePaths(timeline, extractDir)

	return timeline, nil
}

// WriteOTIOZ writes a timeline and its media to a .otioz bundle.
func WriteOTIOZ(
	timeline *gotio.Timeline,
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

	// Create output file
	f, err := os.Create(path)
	if err != nil {
		return &BundleError{
			Operation: "write",
			Path:      path,
			Message:   "failed to create file",
			Cause:     err,
		}
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	// Write version.txt (deflated)
	versionWriter, err := w.Create("version.txt")
	if err != nil {
		return err
	}
	if _, err := versionWriter.Write([]byte(BundleVersion)); err != nil {
		return err
	}

	// Write content.otio (deflated)
	contentData, err := gotio.ToJSONBytesIndent(prepared, "    ")
	if err != nil {
		return &BundleError{
			Operation: "write",
			Path:      path,
			Message:   "failed to serialize timeline",
			Cause:     err,
		}
	}

	contentWriter, err := w.Create("content.otio")
	if err != nil {
		return err
	}
	if _, err := contentWriter.Write(contentData); err != nil {
		return err
	}

	// Write media files (stored, no compression)
	for sourcePath := range manifest {
		basename := filepath.Base(sourcePath)
		bundlePath := "media/" + basename
		// Use forward slashes
		bundlePath = strings.ReplaceAll(bundlePath, "\\", "/")

		// Create file header with STORE method (no compression)
		header := &zip.FileHeader{
			Name:   bundlePath,
			Method: zip.Store,
		}

		mediaWriter, err := w.CreateHeader(header)
		if err != nil {
			return err
		}

		// Copy media file
		mediaFile, err := os.Open(sourcePath)
		if err != nil {
			return &BundleError{
				Operation: "write",
				Path:      sourcePath,
				Message:   "failed to open media file",
				Cause:     err,
			}
		}

		_, copyErr := io.Copy(mediaWriter, mediaFile)
		mediaFile.Close()
		if copyErr != nil {
			return &BundleError{
				Operation: "write",
				Path:      sourcePath,
				Message:   "failed to copy media file",
				Cause:     copyErr,
			}
		}
	}

	return nil
}

// WriteOTIOZDryRun calculates the total size of a .otioz bundle without writing.
func WriteOTIOZDryRun(
	timeline *gotio.Timeline,
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
	contentData, err := gotio.ToJSONBytesIndent(prepared, "    ")
	if err != nil {
		return 0, err
	}
	total += int64(len(contentData))

	// Size of version.txt
	total += int64(len(BundleVersion))

	// Size of media files
	mediaSize, err := TotalMediaSize(manifest)
	if err != nil {
		return 0, err
	}
	total += mediaSize

	// Add overhead for ZIP headers (rough estimate)
	total += int64(len(manifest)*100 + 200)

	return total, nil
}

// extractZipFile extracts a single file from a zip archive.
func extractZipFile(f *zip.File, destPath string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, rc)
	return err
}

// IsOTIOZ checks if a path is a valid .otioz bundle file.
func IsOTIOZ(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}

	return strings.HasSuffix(path, ".otioz")
}
