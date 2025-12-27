// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentime

import (
	"testing"
)

func TestNewTimeTransform(t *testing.T) {
	offset := NewRationalTime(10, 24)
	tt := NewTimeTransform(offset, 2.0, 48)

	if !tt.Offset().StrictlyEqual(offset) {
		t.Error("Offset mismatch")
	}
	if tt.Scale() != 2.0 {
		t.Errorf("Expected scale 2.0, got %g", tt.Scale())
	}
	if tt.Rate() != 48 {
		t.Errorf("Expected rate 48, got %g", tt.Rate())
	}
}

func TestNewTimeTransformDefault(t *testing.T) {
	tt := NewTimeTransformDefault()

	if tt.Offset().Value() != 0 {
		t.Errorf("Expected zero offset, got %g", tt.Offset().Value())
	}
	if tt.Scale() != 1.0 {
		t.Errorf("Expected scale 1.0, got %g", tt.Scale())
	}
	if tt.Rate() != -1 {
		t.Errorf("Expected rate -1, got %g", tt.Rate())
	}
}

func TestTimeTransformAppliedToTime(t *testing.T) {
	// Scale by 2
	tt := NewTimeTransform(NewRationalTime(0, 24), 2.0, -1)
	input := NewRationalTime(10, 24)
	result := tt.AppliedToTime(input)

	if result.Value() != 20 {
		t.Errorf("Expected value 20, got %g", result.Value())
	}
	if result.Rate() != 24 {
		t.Errorf("Expected rate 24, got %g", result.Rate())
	}
}

func TestTimeTransformAppliedToTimeWithOffset(t *testing.T) {
	// Offset by 10 frames
	tt := NewTimeTransform(NewRationalTime(10, 24), 1.0, -1)
	input := NewRationalTime(5, 24)
	result := tt.AppliedToTime(input)

	if result.Value() != 15 {
		t.Errorf("Expected value 15, got %g", result.Value())
	}
}

func TestTimeTransformAppliedToTimeWithRateChange(t *testing.T) {
	// Change rate from 24 to 48
	tt := NewTimeTransform(NewRationalTime(0, 24), 1.0, 48)
	input := NewRationalTime(24, 24) // 1 second at 24fps

	result := tt.AppliedToTime(input)

	if result.Rate() != 48 {
		t.Errorf("Expected rate 48, got %g", result.Rate())
	}
	if result.Value() != 48 { // 1 second at 48fps
		t.Errorf("Expected value 48, got %g", result.Value())
	}
}

func TestTimeTransformAppliedToTimeComplex(t *testing.T) {
	// Scale by 2 and add 10 frames
	tt := NewTimeTransform(NewRationalTime(10, 24), 2.0, -1)
	input := NewRationalTime(10, 24)
	result := tt.AppliedToTime(input)

	// (10 * 2) + 10 = 30
	if result.Value() != 30 {
		t.Errorf("Expected value 30, got %g", result.Value())
	}
}

func TestTimeTransformAppliedToRange(t *testing.T) {
	// Scale by 2
	tt := NewTimeTransform(NewRationalTime(0, 24), 2.0, -1)
	input := NewTimeRangeFromValues(10, 20, 24) // 10-30

	result := tt.AppliedToRange(input)

	// Start: 10 * 2 = 20
	// End: 30 * 2 = 60
	// Duration: 60 - 20 = 40
	if result.StartTime().Value() != 20 {
		t.Errorf("Expected start 20, got %g", result.StartTime().Value())
	}
	if result.Duration().Value() != 40 {
		t.Errorf("Expected duration 40, got %g", result.Duration().Value())
	}
}

func TestTimeTransformAppliedToTransform(t *testing.T) {
	tt1 := NewTimeTransform(NewRationalTime(10, 24), 2.0, -1)
	tt2 := NewTimeTransform(NewRationalTime(5, 24), 3.0, -1)

	combined := tt1.AppliedToTransform(tt2)

	// Offsets add: 10 + 5 = 15
	if combined.Offset().Value() != 15 {
		t.Errorf("Expected offset 15, got %g", combined.Offset().Value())
	}

	// Scales multiply: 2 * 3 = 6
	if combined.Scale() != 6 {
		t.Errorf("Expected scale 6, got %g", combined.Scale())
	}
}

func TestTimeTransformAppliedToTransformWithRate(t *testing.T) {
	tt1 := NewTimeTransform(NewRationalTime(0, 24), 1.0, 48)
	tt2 := NewTimeTransform(NewRationalTime(0, 24), 1.0, -1)

	combined := tt1.AppliedToTransform(tt2)

	// First transform's rate should take precedence
	if combined.Rate() != 48 {
		t.Errorf("Expected rate 48, got %g", combined.Rate())
	}

	// When first has no rate, use second's
	tt1NoRate := NewTimeTransform(NewRationalTime(0, 24), 1.0, -1)
	tt2WithRate := NewTimeTransform(NewRationalTime(0, 24), 1.0, 96)

	combined2 := tt1NoRate.AppliedToTransform(tt2WithRate)
	if combined2.Rate() != 96 {
		t.Errorf("Expected rate 96, got %g", combined2.Rate())
	}
}

func TestTimeTransformEqual(t *testing.T) {
	tt1 := NewTimeTransform(NewRationalTime(10, 24), 2.0, 48)
	tt2 := NewTimeTransform(NewRationalTime(10, 24), 2.0, 48)
	tt3 := NewTimeTransform(NewRationalTime(10, 24), 3.0, 48)

	if !tt1.Equal(tt2) {
		t.Error("Expected equal transforms")
	}
	if tt1.Equal(tt3) {
		t.Error("Expected unequal transforms")
	}
}

func TestTimeTransformString(t *testing.T) {
	tt := NewTimeTransform(NewRationalTime(10, 24), 2, 48)
	str := tt.String()
	expected := "TimeTransform(RationalTime(10, 24), 2, 48)"
	if str != expected {
		t.Errorf("Expected '%s', got '%s'", expected, str)
	}
}

func TestTimeTransformIdentity(t *testing.T) {
	// Identity transform should not change values
	tt := NewTimeTransformDefault()
	input := NewRationalTime(42, 30)

	result := tt.AppliedToTime(input)

	if !result.StrictlyEqual(input) {
		t.Errorf("Identity transform changed value from %v to %v", input, result)
	}
}
