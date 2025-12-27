// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

// Package medialinker provides media linking/resolution for OpenTimelineIO.
// Media linkers resolve MissingReferences to actual media files based on
// custom logic (path templates, directory searches, etc.).
package medialinker

import (
	"github.com/Avalanche-io/gotio/opentimelineio"
)

// MediaLinker resolves media references for clips.
// Implementations can use various strategies to find media files:
// - Path templates (e.g., "/media/{name}.mov")
// - Directory searches
// - Database lookups
// - Custom studio-specific logic
type MediaLinker interface {
	// Name returns the unique name of this linker.
	Name() string

	// LinkMediaReference resolves a clip's media reference.
	// The returned MediaReference replaces the clip's current reference.
	// Returns nil to leave the reference unchanged.
	// The args map can contain linker-specific configuration.
	LinkMediaReference(
		clip *opentimelineio.Clip,
		args map[string]any,
	) (opentimelineio.MediaReference, error)
}

// LinkingPolicy determines when to apply media linking.
type LinkingPolicy int

const (
	// DoNotLink skips media linking entirely.
	DoNotLink LinkingPolicy = iota
	// UseDefaultLinker uses the registered default linker.
	UseDefaultLinker
	// UseNamedLinker uses a specific linker by name (requires linker name).
	UseNamedLinker
)

// String returns the string representation of a LinkingPolicy.
func (p LinkingPolicy) String() string {
	switch p {
	case DoNotLink:
		return "DoNotLink"
	case UseDefaultLinker:
		return "UseDefaultLinker"
	case UseNamedLinker:
		return "UseNamedLinker"
	default:
		return "Unknown"
	}
}

// LinkerError represents an error from a media linker.
type LinkerError struct {
	LinkerName string
	ClipName   string
	Message    string
	Cause      error
}

// Error implements the error interface.
func (e *LinkerError) Error() string {
	msg := "media linker"
	if e.LinkerName != "" {
		msg += " '" + e.LinkerName + "'"
	}
	msg += ": " + e.Message
	if e.ClipName != "" {
		msg += " (clip: " + e.ClipName + ")"
	}
	if e.Cause != nil {
		msg += ": " + e.Cause.Error()
	}
	return msg
}

// Unwrap returns the underlying error.
func (e *LinkerError) Unwrap() error {
	return e.Cause
}
