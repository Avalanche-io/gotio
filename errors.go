// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package gotio

import (
	"errors"
	"fmt"
)

// Error types for opentimelineio.
var (
	ErrNotFound                    = errors.New("not found")
	ErrMissingReference            = errors.New("missing reference")
	ErrMediaReferenceNotFound      = errors.New("media reference not found")
	ErrCannotComputeAvailableRange = errors.New("cannot compute available range")
	ErrInvalidTimecode             = errors.New("invalid timecode")
	ErrChildAlreadyHasParent       = errors.New("child already has a parent")
	ErrNotAChild                   = errors.New("item is not a child of a composition")
	ErrNoCommonAncestor            = errors.New("items do not share a common ancestor")
)

// IndexError indicates an index out of bounds.
type IndexError struct {
	Index int
	Size  int
}

func (e *IndexError) Error() string {
	return fmt.Sprintf("index %d out of bounds for size %d", e.Index, e.Size)
}

// TypeMismatchError indicates a type mismatch.
type TypeMismatchError struct {
	Expected string
	Got      string
}

func (e *TypeMismatchError) Error() string {
	return fmt.Sprintf("expected %s, got %s", e.Expected, e.Got)
}

// SchemaError indicates a schema error.
type SchemaError struct {
	Schema  string
	Message string
}

func (e *SchemaError) Error() string {
	return fmt.Sprintf("schema %s: %s", e.Schema, e.Message)
}

// JSONError indicates a JSON parsing error.
type JSONError struct {
	Message string
}

func (e *JSONError) Error() string {
	return fmt.Sprintf("JSON error: %s", e.Message)
}
