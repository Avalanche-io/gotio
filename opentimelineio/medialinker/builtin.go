// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package medialinker

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mrjoshuak/gotio/opentimelineio"
)

// PathTemplateLinker resolves media paths using a template.
// The template can contain placeholders:
//   - {name} - clip name
//   - {basename} - original file basename (without extension)
//   - {ext} - original file extension
//   - {dir} - original file directory
type PathTemplateLinker struct {
	template string
}

// NewPathTemplateLinker creates a new PathTemplateLinker.
func NewPathTemplateLinker(template string) *PathTemplateLinker {
	return &PathTemplateLinker{template: template}
}

// Name returns the linker name.
func (l *PathTemplateLinker) Name() string {
	return "path_template"
}

// LinkMediaReference resolves using the template.
func (l *PathTemplateLinker) LinkMediaReference(
	clip *opentimelineio.Clip,
	args map[string]any,
) (opentimelineio.MediaReference, error) {
	ref := clip.MediaReference()
	if ref == nil {
		return nil, nil
	}

	// Get original URL if available
	var originalURL string
	var originalExt string
	var originalBasename string
	var originalDir string

	if extRef, ok := ref.(*opentimelineio.ExternalReference); ok {
		originalURL = extRef.TargetURL()
		originalExt = filepath.Ext(originalURL)
		originalBasename = strings.TrimSuffix(filepath.Base(originalURL), originalExt)
		originalDir = filepath.Dir(originalURL)
	}

	// Check for missing reference with stored URL
	if missing, ok := ref.(*opentimelineio.MissingReference); ok {
		if meta := missing.Metadata(); meta != nil {
			if url, ok := meta["original_target_url"].(string); ok {
				originalURL = url
				originalExt = filepath.Ext(url)
				originalBasename = strings.TrimSuffix(filepath.Base(url), originalExt)
				originalDir = filepath.Dir(url)
			}
		}
	}

	// Apply template
	path := l.template
	path = strings.ReplaceAll(path, "{name}", clip.Name())
	path = strings.ReplaceAll(path, "{basename}", originalBasename)
	path = strings.ReplaceAll(path, "{ext}", originalExt)
	path = strings.ReplaceAll(path, "{dir}", originalDir)

	// Check if file exists
	if _, err := os.Stat(path); err != nil {
		return nil, nil // File doesn't exist, leave reference unchanged
	}

	// Create new external reference
	return opentimelineio.NewExternalReference(
		ref.Name(),
		path,
		ref.AvailableRange(),
		ref.Metadata(),
	), nil
}

// DirectoryLinker searches directories for matching media files.
type DirectoryLinker struct {
	searchPaths []string
	extensions  []string
}

// NewDirectoryLinker creates a new DirectoryLinker.
func NewDirectoryLinker(searchPaths, extensions []string) *DirectoryLinker {
	return &DirectoryLinker{
		searchPaths: searchPaths,
		extensions:  extensions,
	}
}

// Name returns the linker name.
func (l *DirectoryLinker) Name() string {
	return "directory"
}

// LinkMediaReference searches for matching files.
func (l *DirectoryLinker) LinkMediaReference(
	clip *opentimelineio.Clip,
	args map[string]any,
) (opentimelineio.MediaReference, error) {
	ref := clip.MediaReference()
	if ref == nil {
		return nil, nil
	}

	// Get search name
	searchName := clip.Name()
	if searchName == "" {
		// Try to get from original reference
		if extRef, ok := ref.(*opentimelineio.ExternalReference); ok {
			searchName = strings.TrimSuffix(
				filepath.Base(extRef.TargetURL()),
				filepath.Ext(extRef.TargetURL()),
			)
		}
	}

	if searchName == "" {
		return nil, nil
	}

	// Search directories
	for _, dir := range l.searchPaths {
		for _, ext := range l.extensions {
			path := filepath.Join(dir, searchName+ext)
			if _, err := os.Stat(path); err == nil {
				return opentimelineio.NewExternalReference(
					ref.Name(),
					path,
					ref.AvailableRange(),
					ref.Metadata(),
				), nil
			}
		}

		// Also try without extension modification
		path := filepath.Join(dir, searchName)
		if _, err := os.Stat(path); err == nil {
			return opentimelineio.NewExternalReference(
				ref.Name(),
				path,
				ref.AvailableRange(),
				ref.Metadata(),
			), nil
		}
	}

	return nil, nil // Not found
}

// NullLinker is a linker that does nothing (for testing).
type NullLinker struct {
	name string
}

// NewNullLinker creates a new NullLinker.
func NewNullLinker(name string) *NullLinker {
	return &NullLinker{name: name}
}

// Name returns the linker name.
func (l *NullLinker) Name() string {
	return l.name
}

// LinkMediaReference returns nil (no change).
func (l *NullLinker) LinkMediaReference(
	clip *opentimelineio.Clip,
	args map[string]any,
) (opentimelineio.MediaReference, error) {
	return nil, nil
}

// MetadataLinker sets metadata on MissingReferences to indicate linking was attempted.
type MetadataLinker struct {
	name string
}

// NewMetadataLinker creates a new MetadataLinker.
func NewMetadataLinker(name string) *MetadataLinker {
	return &MetadataLinker{name: name}
}

// Name returns the linker name.
func (l *MetadataLinker) Name() string {
	return l.name
}

// LinkMediaReference adds metadata to missing references.
func (l *MetadataLinker) LinkMediaReference(
	clip *opentimelineio.Clip,
	args map[string]any,
) (opentimelineio.MediaReference, error) {
	ref := clip.MediaReference()
	if ref == nil {
		return nil, nil
	}

	missing, ok := ref.(*opentimelineio.MissingReference)
	if !ok {
		return nil, nil // Only modify MissingReferences
	}

	// Clone and add metadata
	meta := opentimelineio.CloneAnyDictionary(missing.Metadata())
	meta["linked_by"] = l.name
	meta["linking_args"] = args

	return opentimelineio.NewMissingReference(
		missing.Name(),
		missing.AvailableRange(),
		meta,
	), nil
}
