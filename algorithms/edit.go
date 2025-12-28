// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package algorithms

import (
	"fmt"

	"github.com/Avalanche-io/gotio/opentime"
	"github.com/Avalanche-io/gotio"
)

// ReferencePoint determines how fill operations place clips.
type ReferencePoint int

const (
	// ReferencePointSource uses clip's natural duration.
	ReferencePointSource ReferencePoint = iota
	// ReferencePointSequence trims clip to fit gap exactly.
	ReferencePointSequence
	// ReferencePointFit adds time warp to stretch clip to gap.
	ReferencePointFit
)

// String returns the string representation of a ReferencePoint.
func (r ReferencePoint) String() string {
	switch r {
	case ReferencePointSource:
		return "Source"
	case ReferencePointSequence:
		return "Sequence"
	case ReferencePointFit:
		return "Fit"
	default:
		return fmt.Sprintf("ReferencePoint(%d)", r)
	}
}

// EditError represents an error that occurred during an edit operation.
type EditError struct {
	Operation string
	Message   string
	Time      *opentime.RationalTime
	Item      gotio.Composable
}

// Error implements the error interface.
func (e *EditError) Error() string {
	msg := fmt.Sprintf("edit %s: %s", e.Operation, e.Message)
	if e.Time != nil {
		msg += fmt.Sprintf(" at time %v", *e.Time)
	}
	if e.Item != nil {
		msg += fmt.Sprintf(" (item: %s)", e.Item.Name())
	}
	return msg
}

// Common edit errors.
var (
	ErrNotAnItem           = &EditError{Message: "not an item at specified time"}
	ErrNotAGap             = &EditError{Message: "item at specified time is not a gap"}
	ErrNotAChildOf         = &EditError{Message: "item is not a child of composition"}
	ErrCannotTrimTransition = &EditError{Message: "cannot trim through a transition"}
	ErrNoChildren          = &EditError{Message: "composition has no children"}
	ErrInvalidTime         = &EditError{Message: "time is outside composition bounds"}
	ErrInvalidRange        = &EditError{Message: "range is outside composition bounds"}
	ErrNegativeDuration    = &EditError{Message: "operation would result in negative duration"}
)

// newEditError creates a new EditError for a specific operation.
func newEditError(operation, message string) *EditError {
	return &EditError{
		Operation: operation,
		Message:   message,
	}
}

// newEditErrorAt creates a new EditError with time context.
func newEditErrorAt(operation, message string, time opentime.RationalTime) *EditError {
	return &EditError{
		Operation: operation,
		Message:   message,
		Time:      &time,
	}
}

// newEditErrorForItem creates a new EditError with item context.
func newEditErrorForItem(operation, message string, item gotio.Composable) *EditError {
	return &EditError{
		Operation: operation,
		Message:   message,
		Item:      item,
	}
}
