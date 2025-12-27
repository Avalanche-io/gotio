// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentime

import (
	"math"
	"testing"
)

func TestNewRationalTime(t *testing.T) {
	rt := NewRationalTime(24, 24)
	if rt.Value() != 24 {
		t.Errorf("Expected value 24, got %g", rt.Value())
	}
	if rt.Rate() != 24 {
		t.Errorf("Expected rate 24, got %g", rt.Rate())
	}
}

func TestRationalTimeIsValid(t *testing.T) {
	tests := []struct {
		name    string
		rt      RationalTime
		isValid bool
	}{
		{"valid", NewRationalTime(10, 24), true},
		{"zero rate", NewRationalTime(10, 0), false},
		{"negative rate", NewRationalTime(10, -1), false},
		{"nan value", NewRationalTime(math.NaN(), 24), false},
		{"nan rate", NewRationalTime(10, math.NaN()), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.rt.IsValidTime() != tt.isValid {
				t.Errorf("Expected IsValidTime() = %v, got %v", tt.isValid, tt.rt.IsValidTime())
			}
			if tt.rt.IsInvalidTime() != !tt.isValid {
				t.Errorf("Expected IsInvalidTime() = %v, got %v", !tt.isValid, tt.rt.IsInvalidTime())
			}
		})
	}
}

func TestRationalTimeRescale(t *testing.T) {
	rt := NewRationalTime(24, 24)
	rescaled := rt.RescaledTo(48)

	if rescaled.Value() != 48 {
		t.Errorf("Expected value 48, got %g", rescaled.Value())
	}
	if rescaled.Rate() != 48 {
		t.Errorf("Expected rate 48, got %g", rescaled.Rate())
	}
}

func TestRationalTimeValueRescale(t *testing.T) {
	rt := NewRationalTime(24, 24)
	value := rt.ValueRescaledTo(48)

	if value != 48 {
		t.Errorf("Expected 48, got %g", value)
	}

	// Same rate should return same value
	sameValue := rt.ValueRescaledTo(24)
	if sameValue != 24 {
		t.Errorf("Expected 24, got %g", sameValue)
	}
}

func TestRationalTimeAlmostEqual(t *testing.T) {
	rt1 := NewRationalTime(24, 24)
	rt2 := NewRationalTime(24.001, 24)
	rt3 := NewRationalTime(25, 24)

	if !rt1.AlmostEqual(rt2, 0.01) {
		t.Error("Expected times to be almost equal")
	}
	if rt1.AlmostEqual(rt3, 0.01) {
		t.Error("Expected times to not be almost equal")
	}
}

func TestRationalTimeStrictlyEqual(t *testing.T) {
	rt1 := NewRationalTime(24, 24)
	rt2 := NewRationalTime(24, 24)
	rt3 := NewRationalTime(48, 48) // Same time, different rate

	if !rt1.StrictlyEqual(rt2) {
		t.Error("Expected times to be strictly equal")
	}
	if rt1.StrictlyEqual(rt3) {
		t.Error("Expected times to not be strictly equal")
	}
}

func TestRationalTimeFloorCeilRound(t *testing.T) {
	rt := NewRationalTime(24.5, 24)

	floor := rt.Floor()
	if floor.Value() != 24 {
		t.Errorf("Expected floor 24, got %g", floor.Value())
	}

	ceil := rt.Ceil()
	if ceil.Value() != 25 {
		t.Errorf("Expected ceil 25, got %g", ceil.Value())
	}

	round := rt.Round()
	if round.Value() != 25 { // 24.5 rounds to 25
		t.Errorf("Expected round 25, got %g", round.Value())
	}
}

func TestDurationFromStartEndTime(t *testing.T) {
	start := NewRationalTime(10, 24)
	end := NewRationalTime(20, 24)

	duration := DurationFromStartEndTime(start, end)
	if duration.Value() != 10 {
		t.Errorf("Expected duration 10, got %g", duration.Value())
	}
	if duration.Rate() != 24 {
		t.Errorf("Expected rate 24, got %g", duration.Rate())
	}
}

func TestDurationFromStartEndTimeInclusive(t *testing.T) {
	start := NewRationalTime(10, 24)
	end := NewRationalTime(20, 24)

	duration := DurationFromStartEndTimeInclusive(start, end)
	if duration.Value() != 11 {
		t.Errorf("Expected duration 11, got %g", duration.Value())
	}
}

func TestDurationFromStartEndTimeDifferentRates(t *testing.T) {
	start := NewRationalTime(10, 24)
	end := NewRationalTime(40, 48) // 20 frames at 24fps

	duration := DurationFromStartEndTime(start, end)
	if duration.Value() != 10 {
		t.Errorf("Expected duration 10, got %g", duration.Value())
	}
	if duration.Rate() != 24 {
		t.Errorf("Expected rate 24, got %g", duration.Rate())
	}
}

func TestIsSMPTETimecodeRate(t *testing.T) {
	validRates := []float64{23.976, 24, 25, 29.97, 30, 50, 59.94, 60}
	for _, rate := range validRates {
		if !IsSMPTETimecodeRate(rate) {
			t.Errorf("Expected %g to be valid SMPTE rate", rate)
		}
	}

	invalidRates := []float64{12, 15, 48, 72, 120}
	for _, rate := range invalidRates {
		if IsSMPTETimecodeRate(rate) {
			t.Errorf("Expected %g to not be valid SMPTE rate", rate)
		}
	}
}

func TestNearestSMPTETimecodeRate(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{23, 23.976},
		{24, 24},
		{26, 25},
		{29, 29.97},
		{31, 30},
		{52, 50},      // 52 is closer to 50 than 59.94
		{58, 59.94},
		{65, 60},
	}

	for _, tt := range tests {
		result := NearestSMPTETimecodeRate(tt.input)
		if math.Abs(result-tt.expected) > 0.01 {
			t.Errorf("NearestSMPTETimecodeRate(%g) = %g, want %g", tt.input, result, tt.expected)
		}
	}
}

func TestFromFrames(t *testing.T) {
	rt := FromFrames(24, 24)
	if rt.Value() != 24 {
		t.Errorf("Expected value 24, got %g", rt.Value())
	}
	if rt.Rate() != 24 {
		t.Errorf("Expected rate 24, got %g", rt.Rate())
	}

	// Test truncation
	rt = FromFrames(24.9, 24)
	if rt.Value() != 24 {
		t.Errorf("Expected value 24, got %g", rt.Value())
	}
}

func TestFromSeconds(t *testing.T) {
	rt := FromSeconds(1.0, 24)
	if rt.Value() != 24 {
		t.Errorf("Expected value 24, got %g", rt.Value())
	}
	if rt.Rate() != 24 {
		t.Errorf("Expected rate 24, got %g", rt.Rate())
	}
}

func TestFromSecondsFloat(t *testing.T) {
	rt := FromSecondsFloat(1.5)
	if rt.Value() != 1.5 {
		t.Errorf("Expected value 1.5, got %g", rt.Value())
	}
	if rt.Rate() != 1 {
		t.Errorf("Expected rate 1, got %g", rt.Rate())
	}
}

func TestToFrames(t *testing.T) {
	rt := NewRationalTime(24, 24)
	if rt.ToFrames() != 24 {
		t.Errorf("Expected 24, got %d", rt.ToFrames())
	}

	frames := rt.ToFramesAtRate(48)
	if frames != 48 {
		t.Errorf("Expected 48, got %d", frames)
	}
}

func TestToSeconds(t *testing.T) {
	rt := NewRationalTime(24, 24)
	if rt.ToSeconds() != 1.0 {
		t.Errorf("Expected 1.0, got %g", rt.ToSeconds())
	}
}

func TestToTimecodeNDF(t *testing.T) {
	// 1 hour at 24fps = 86400 frames
	rt := NewRationalTime(86400, 24)
	tc, err := rt.ToTimecode(24, ForceNo)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if tc != "01:00:00:00" {
		t.Errorf("Expected 01:00:00:00, got %s", tc)
	}

	// 1 second at 24fps = 24 frames
	rt = NewRationalTime(24, 24)
	tc, err = rt.ToTimecode(24, ForceNo)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if tc != "00:00:01:00" {
		t.Errorf("Expected 00:00:01:00, got %s", tc)
	}
}

func TestToTimeString(t *testing.T) {
	rt := NewRationalTime(3661.5, 1) // 1 hour, 1 minute, 1.5 seconds
	timeStr := rt.ToTimeString()
	if timeStr != "01:01:01.5" {
		t.Errorf("Expected 01:01:01.5, got %s", timeStr)
	}
}

func TestFromTimecode(t *testing.T) {
	tests := []struct {
		timecode string
		rate     float64
		expected float64
	}{
		{"01:00:00:00", 24, 86400},
		{"00:00:01:00", 24, 24},
		{"00:00:00:12", 24, 12},
	}

	for _, tt := range tests {
		rt, err := FromTimecode(tt.timecode, tt.rate)
		if err != nil {
			t.Errorf("FromTimecode(%s, %g) error: %v", tt.timecode, tt.rate, err)
			continue
		}
		if rt.Value() != tt.expected {
			t.Errorf("FromTimecode(%s, %g) = %g, want %g", tt.timecode, tt.rate, rt.Value(), tt.expected)
		}
	}
}

func TestFromTimeString(t *testing.T) {
	rt, err := FromTimeString("01:00:00.0", 24)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if rt.ToSeconds() != 3600 {
		t.Errorf("Expected 3600 seconds, got %g", rt.ToSeconds())
	}
}

func TestRationalTimeArithmetic(t *testing.T) {
	rt1 := NewRationalTime(10, 24)
	rt2 := NewRationalTime(5, 24)

	// Add
	sum := rt1.Add(rt2)
	if sum.Value() != 15 {
		t.Errorf("Expected sum 15, got %g", sum.Value())
	}

	// Sub
	diff := rt1.Sub(rt2)
	if diff.Value() != 5 {
		t.Errorf("Expected diff 5, got %g", diff.Value())
	}

	// Neg
	neg := rt1.Neg()
	if neg.Value() != -10 {
		t.Errorf("Expected neg -10, got %g", neg.Value())
	}
}

func TestRationalTimeArithmeticDifferentRates(t *testing.T) {
	rt1 := NewRationalTime(24, 24) // 1 second
	rt2 := NewRationalTime(48, 48) // 1 second

	sum := rt1.Add(rt2)
	// Should use higher rate (48)
	if sum.Rate() != 48 {
		t.Errorf("Expected rate 48, got %g", sum.Rate())
	}
	if sum.ToSeconds() != 2.0 {
		t.Errorf("Expected 2.0 seconds, got %g", sum.ToSeconds())
	}
}

func TestRationalTimeCmp(t *testing.T) {
	rt1 := NewRationalTime(10, 24)
	rt2 := NewRationalTime(20, 24)
	rt3 := NewRationalTime(10, 24)

	if rt1.Cmp(rt2) != -1 {
		t.Errorf("Expected rt1 < rt2")
	}
	if rt2.Cmp(rt1) != 1 {
		t.Errorf("Expected rt2 > rt1")
	}
	if rt1.Cmp(rt3) != 0 {
		t.Errorf("Expected rt1 == rt3")
	}
}

func TestRationalTimeEqual(t *testing.T) {
	rt1 := NewRationalTime(24, 24)
	rt2 := NewRationalTime(48, 48) // Same time, different representation

	if !rt1.Equal(rt2) {
		t.Error("Expected times to be equal")
	}
}

func TestRationalTimeString(t *testing.T) {
	rt := NewRationalTime(24, 24)
	str := rt.String()
	if str != "RationalTime(24, 24)" {
		t.Errorf("Expected 'RationalTime(24, 24)', got '%s'", str)
	}
}
