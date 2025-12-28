// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package gotio

import (
	"testing"
)

func TestIndexError(t *testing.T) {
	err := &IndexError{Index: 5, Size: 3}

	expected := "index 5 out of bounds for size 3"
	if err.Error() != expected {
		t.Errorf("IndexError.Error() = %q, want %q", err.Error(), expected)
	}
}

func TestSchemaError(t *testing.T) {
	err := &SchemaError{Schema: "Unknown.1", Message: "not registered"}

	expected := "schema Unknown.1: not registered"
	if err.Error() != expected {
		t.Errorf("SchemaError.Error() = %q, want %q", err.Error(), expected)
	}
}

func TestTypeMismatchError(t *testing.T) {
	err := &TypeMismatchError{Expected: "Clip", Got: "Gap"}

	expected := "expected Clip, got Gap"
	if err.Error() != expected {
		t.Errorf("TypeMismatchError.Error() = %q, want %q", err.Error(), expected)
	}
}

func TestErrNotFound(t *testing.T) {
	if ErrNotFound == nil {
		t.Error("ErrNotFound should not be nil")
	}
}
