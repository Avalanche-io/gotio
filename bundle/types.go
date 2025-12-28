// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

// Package bundle provides support for OTIO file bundles (.otioz and .otiod).
// These formats package timelines with their associated media files for
// distribution, archiving, and interchange.
package bundle

import (
	"fmt"
)

// MediaReferencePolicy determines how media references are handled when writing bundles.
type MediaReferencePolicy int

const (
	// ErrorIfNotFile raises an error for non-file references or missing files.
	ErrorIfNotFile MediaReferencePolicy = iota
	// MissingIfNotFile replaces problematic references with MissingReference.
	MissingIfNotFile
	// AllMissing replaces all references with MissingReference (no media bundled).
	AllMissing
)

// String returns the string representation of a MediaReferencePolicy.
func (p MediaReferencePolicy) String() string {
	switch p {
	case ErrorIfNotFile:
		return "ErrorIfNotFile"
	case MissingIfNotFile:
		return "MissingIfNotFile"
	case AllMissing:
		return "AllMissing"
	default:
		return fmt.Sprintf("MediaReferencePolicy(%d)", p)
	}
}

// BundleVersion is the current version of the bundle format.
const BundleVersion = "1.0.0"

// BundleError represents an error that occurred during bundle operations.
type BundleError struct {
	Operation string
	Path      string
	Message   string
	Cause     error
}

// Error implements the error interface.
func (e *BundleError) Error() string {
	msg := fmt.Sprintf("bundle %s: %s", e.Operation, e.Message)
	if e.Path != "" {
		msg += fmt.Sprintf(" (path: %s)", e.Path)
	}
	if e.Cause != nil {
		msg += fmt.Sprintf(": %v", e.Cause)
	}
	return msg
}

// Unwrap returns the underlying error.
func (e *BundleError) Unwrap() error {
	return e.Cause
}
