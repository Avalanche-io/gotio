// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package bundle

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/Avalanche-io/gotio"
)

// MediaManifest maps absolute source paths to the external references that point to them.
type MediaManifest map[string][]*gotio.ExternalReference

// PrepareForBundle processes a timeline for bundling according to the media policy.
// It returns a cloned timeline with adjusted media references and a manifest of media files to include.
func PrepareForBundle(
	timeline *gotio.Timeline,
	policy MediaReferencePolicy,
) (*gotio.Timeline, MediaManifest, error) {
	// Clone the timeline to avoid modifying the original
	cloned := timeline.Clone().(*gotio.Timeline)
	manifest := make(MediaManifest)

	// Find all clips
	clips := cloned.FindClips(nil, false)

	for _, clip := range clips {
		ref := clip.MediaReference()
		if ref == nil {
			continue
		}

		// Handle AllMissing policy
		if policy == AllMissing {
			replaceMissingRef := gotio.NewMissingReference(
				ref.Name(),
				ref.AvailableRange(),
				gotio.AnyDictionary{
					"original_target_url":       getTargetURL(ref),
					"missing_reference_because": "AllMissing policy",
				},
			)
			clip.SetMediaReference(replaceMissingRef)
			continue
		}

		// Only process ExternalReferences
		extRef, ok := ref.(*gotio.ExternalReference)
		if !ok {
			continue
		}

		targetURL := extRef.TargetURL()
		if targetURL == "" {
			continue
		}

		// Parse URL
		absPath, err := urlToAbsPath(targetURL)
		if err != nil {
			if policy == ErrorIfNotFile {
				return nil, nil, &BundleError{
					Operation: "prepare",
					Path:      targetURL,
					Message:   "invalid media URL",
					Cause:     err,
				}
			}
			// MissingIfNotFile - replace with missing reference
			replaceMissing(clip, ref, "invalid URL: "+err.Error())
			continue
		}

		// Check if file exists
		info, err := os.Stat(absPath)
		if err != nil || info.IsDir() {
			if policy == ErrorIfNotFile {
				return nil, nil, &BundleError{
					Operation: "prepare",
					Path:      absPath,
					Message:   "media file not found or is directory",
					Cause:     err,
				}
			}
			replaceMissing(clip, ref, "file not found")
			continue
		}

		// Add to manifest
		manifest[absPath] = append(manifest[absPath], extRef)
	}

	return cloned, manifest, nil
}

// VerifyUniqueBasenames checks that all files in the manifest have unique basenames.
func VerifyUniqueBasenames(manifest MediaManifest) error {
	basenames := make(map[string]string) // basename -> first full path

	for path := range manifest {
		base := filepath.Base(path)
		if existing, ok := basenames[base]; ok {
			return &BundleError{
				Operation: "verify",
				Path:      path,
				Message:   "basename collision with " + existing,
			}
		}
		basenames[base] = path
	}

	return nil
}

// RelinkToBundle updates all external references in the manifest to point to bundle paths.
func RelinkToBundle(manifest MediaManifest) {
	for absPath, refs := range manifest {
		basename := filepath.Base(absPath)
		bundlePath := "media/" + basename
		// Use forward slashes for cross-platform compatibility
		bundlePath = strings.ReplaceAll(bundlePath, "\\", "/")

		for _, ref := range refs {
			ref.SetTargetURL(bundlePath)
		}
	}
}

// ConvertToAbsolutePaths converts relative bundle paths to absolute paths.
func ConvertToAbsolutePaths(timeline *gotio.Timeline, bundleRoot string) error {
	clips := timeline.FindClips(nil, false)

	for _, clip := range clips {
		ref := clip.MediaReference()
		if ref == nil {
			continue
		}

		extRef, ok := ref.(*gotio.ExternalReference)
		if !ok {
			continue
		}

		targetURL := extRef.TargetURL()
		if targetURL == "" {
			continue
		}

		// Check if it's a relative path (starts with "media/")
		if strings.HasPrefix(targetURL, "media/") {
			absPath := filepath.Join(bundleRoot, targetURL)
			extRef.SetTargetURL(absPath)
		}
	}

	return nil
}

// urlToAbsPath converts a file URL or relative path to an absolute path.
func urlToAbsPath(rawURL string) (string, error) {
	// Try to parse as URL
	u, err := url.Parse(rawURL)
	if err != nil {
		// Not a valid URL, treat as path
		return filepath.Abs(rawURL)
	}

	// Handle file:// URLs
	if u.Scheme == "file" {
		path := u.Path
		// Handle Windows paths in file URLs
		if len(path) > 2 && path[0] == '/' && path[2] == ':' {
			path = path[1:] // Remove leading slash for Windows paths
		}
		return filepath.Abs(path)
	}

	// Empty scheme means relative path
	if u.Scheme == "" {
		return filepath.Abs(rawURL)
	}

	// Other schemes (http, https, etc.) are not supported
	return "", &BundleError{
		Message: "unsupported URL scheme: " + u.Scheme,
	}
}

// getTargetURL extracts the target URL from a media reference if it has one.
func getTargetURL(ref gotio.MediaReference) string {
	if extRef, ok := ref.(*gotio.ExternalReference); ok {
		return extRef.TargetURL()
	}
	return ""
}

// replaceMissing replaces a clip's media reference with a MissingReference.
func replaceMissing(clip *gotio.Clip, original gotio.MediaReference, reason string) {
	missing := gotio.NewMissingReference(
		original.Name(),
		original.AvailableRange(),
		gotio.AnyDictionary{
			"original_target_url":       getTargetURL(original),
			"missing_reference_because": reason,
		},
	)
	clip.SetMediaReference(missing)
}

// TotalMediaSize calculates the total size of all media files in the manifest.
func TotalMediaSize(manifest MediaManifest) (int64, error) {
	var total int64

	for path := range manifest {
		info, err := os.Stat(path)
		if err != nil {
			return 0, err
		}
		total += info.Size()
	}

	return total, nil
}
