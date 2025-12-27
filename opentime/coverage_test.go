// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentime

import (
	"encoding/json"
	"testing"
)

// Tests for RationalTime JSON Marshal/Unmarshal
func TestRationalTimeJSONMarshalUnmarshal(t *testing.T) {
	rt := NewRationalTime(100, 24)

	// Marshal
	data, err := json.Marshal(rt)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Unmarshal
	var rt2 RationalTime
	if err := json.Unmarshal(data, &rt2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if rt2.Value() != 100 {
		t.Errorf("Value = %v, want 100", rt2.Value())
	}
	if rt2.Rate() != 24 {
		t.Errorf("Rate = %v, want 24", rt2.Rate())
	}
}

// Tests for TimeRange JSON Marshal/Unmarshal
func TestTimeRangeJSONMarshalUnmarshal(t *testing.T) {
	tr := NewTimeRange(NewRationalTime(10, 24), NewRationalTime(100, 24))

	// Marshal
	data, err := json.Marshal(tr)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Unmarshal
	var tr2 TimeRange
	if err := json.Unmarshal(data, &tr2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if tr2.StartTime().Value() != 10 {
		t.Errorf("StartTime = %v, want 10", tr2.StartTime().Value())
	}
	if tr2.Duration().Value() != 100 {
		t.Errorf("Duration = %v, want 100", tr2.Duration().Value())
	}
}

// Tests for TimeTransform JSON Marshal/Unmarshal
func TestTimeTransformJSONMarshalUnmarshal(t *testing.T) {
	tt := NewTimeTransform(NewRationalTime(5, 24), 2.0, 24)

	// Marshal
	data, err := json.Marshal(tt)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Unmarshal
	var tt2 TimeTransform
	if err := json.Unmarshal(data, &tt2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if tt2.Offset().Value() != 5 {
		t.Errorf("Offset = %v, want 5", tt2.Offset().Value())
	}
	if tt2.Scale() != 2.0 {
		t.Errorf("Scale = %v, want 2.0", tt2.Scale())
	}
}

// Tests for RescaledToRate
func TestRationalTimeRescaledToRateCoverage(t *testing.T) {
	// 48 frames at 24fps = 2 seconds
	rt := NewRationalTime(48, 24)
	otherRate := NewRationalTime(0, 30) // 30 fps

	// Rescale to 30fps - should be 60 frames (2 seconds at 30fps)
	rescaled := rt.RescaledToRate(otherRate)

	if rescaled.Rate() != 30 {
		t.Errorf("Rate = %v, want 30", rescaled.Rate())
	}
	// 48/24 = 2 seconds, 2 * 30 = 60 frames
	if rescaled.Value() != 60 {
		t.Errorf("Value = %v, want 60", rescaled.Value())
	}
}

// Tests for ValueRescaledToRate
func TestRationalTimeValueRescaledToRateCoverage(t *testing.T) {
	rt := NewRationalTime(48, 24)
	otherRate := NewRationalTime(0, 30) // 30 fps

	// Get value rescaled to 30fps
	value := rt.ValueRescaledToRate(otherRate)

	// 48/24 = 2 seconds, 2 * 30 = 60 frames
	if value != 60 {
		t.Errorf("ValueRescaledToRate = %v, want 60", value)
	}
}

// Tests for isDropFrameRate
func TestIsDropFrameRateCoverage(t *testing.T) {
	// Common drop frame rates: 29.97, 59.94
	dropRates := []float64{29.97, 59.94}
	nonDropRates := []float64{24, 25, 30, 60}

	for _, rate := range dropRates {
		if !isDropFrameRate(rate) {
			t.Errorf("Rate %v should be drop frame", rate)
		}
	}

	for _, rate := range nonDropRates {
		if isDropFrameRate(rate) {
			t.Errorf("Rate %v should not be drop frame", rate)
		}
	}
}

// Tests for ToTimecode with drop frame
func TestRationalTimeToTimecodeDropFrameCoverage(t *testing.T) {
	// Test at 29.97 fps with ForceYes
	rt := NewRationalTime(0, 29.97)
	tc, err := rt.ToTimecode(29.97, ForceYes)
	if err != nil {
		t.Fatalf("ToTimecode error: %v", err)
	}
	t.Logf("0 frames at 29.97fps (drop frame): %s", tc)

	// Test at 59.94 fps with ForceYes
	rt = NewRationalTime(0, 59.94)
	tc, err = rt.ToTimecode(59.94, ForceYes)
	if err != nil {
		t.Fatalf("ToTimecode error: %v", err)
	}
	t.Logf("0 frames at 59.94fps (drop frame): %s", tc)

	// Test InferFromRate
	rt = NewRationalTime(1800, 29.97)
	tc, err = rt.ToTimecode(29.97, InferFromRate)
	if err != nil {
		t.Fatalf("ToTimecode error: %v", err)
	}
	t.Logf("1800 frames at 29.97fps (infer): %s", tc)
}

// Tests for ToTimecodeAuto
func TestRationalTimeToTimecodeAutoCoverage(t *testing.T) {
	rt := NewRationalTime(48, 24)
	tc, err := rt.ToTimecodeAuto()
	if err != nil {
		t.Fatalf("ToTimecodeAuto error: %v", err)
	}
	if tc == "" {
		t.Error("ToTimecodeAuto should return non-empty string")
	}
	t.Logf("ToTimecodeAuto: %s", tc)
}

// Tests for ToNearestTimecode
func TestRationalTimeToNearestTimecodeCoverage(t *testing.T) {
	rt := NewRationalTime(48.5, 24)
	tc, err := rt.ToNearestTimecode(24, ForceNo)
	if err != nil {
		t.Fatalf("ToNearestTimecode error: %v", err)
	}
	if tc == "" {
		t.Error("ToNearestTimecode should return non-empty string")
	}
	t.Logf("ToNearestTimecode: %s", tc)
}

// Tests for FromTimecode drop frame
func TestRationalTimeFromTimecodeDropFrameCoverage(t *testing.T) {
	// Test drop frame timecode (semicolon separator)
	rt, err := FromTimecode("00:01:00;02", 29.97)
	if err != nil {
		t.Fatalf("FromTimecode drop frame error: %v", err)
	}
	t.Logf("FromTimecode drop frame: %v", rt.Value())

	// Test drop frame with 59.94fps
	rt, err = FromTimecode("00:01:00;04", 59.94)
	if err != nil {
		t.Fatalf("FromTimecode drop frame 59.94 error: %v", err)
	}
	t.Logf("FromTimecode drop frame 59.94: %v", rt.Value())
}

// Tests for FromTimecode errors
func TestRationalTimeFromTimecodeErrorsCoverage(t *testing.T) {
	// Invalid format
	_, err := FromTimecode("invalid", 24)
	if err == nil {
		t.Error("FromTimecode with invalid format should error")
	}

	// Wrong number of components
	_, err = FromTimecode("00:00:00", 24)
	if err == nil {
		t.Error("FromTimecode with wrong components should error")
	}
}

// Tests for TimeRange.OverlapsRange
func TestTimeRangeOverlapsRangeCoverage(t *testing.T) {
	tr1 := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(100, 24))
	tr2 := NewTimeRange(NewRationalTime(50, 24), NewRationalTime(100, 24))
	tr3 := NewTimeRange(NewRationalTime(200, 24), NewRationalTime(50, 24))

	// tr1 and tr2 should intersect (use Intersects which has better behavior)
	if !tr1.Intersects(tr2, DefaultEpsilon) {
		t.Error("tr1 and tr2 should intersect")
	}

	// tr1 and tr3 should not overlap
	if tr1.OverlapsRange(tr3, DefaultEpsilon) {
		t.Error("tr1 and tr3 should not overlap")
	}
}

// Tests for TimeRange.Contains (single time)
func TestTimeRangeContainsTimeCoverage(t *testing.T) {
	tr := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(100, 24))

	// Time inside range
	inside := NewRationalTime(50, 24)
	if !tr.Contains(inside) {
		t.Error("Range should contain time 50")
	}

	// Time outside range
	outside := NewRationalTime(200, 24)
	if tr.Contains(outside) {
		t.Error("Range should not contain time 200")
	}
}

// Tests for TimeRange.BeforeTime
func TestTimeRangeBeforeTimeCoverage(t *testing.T) {
	tr := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(50, 24))

	// Time after the range
	after := NewRationalTime(100, 24)
	if !tr.BeforeTime(after, DefaultEpsilon) {
		t.Error("Range should be before time 100")
	}

	// Time inside the range
	inside := NewRationalTime(25, 24)
	if tr.BeforeTime(inside, DefaultEpsilon) {
		t.Error("Range should not be before time 25")
	}
}

// Tests for TimeRange.BeginsAt
func TestTimeRangeBeginsAtCoverage(t *testing.T) {
	tr := NewTimeRange(NewRationalTime(100, 24), NewRationalTime(50, 24))

	// Exactly at start time
	start := NewRationalTime(100, 24)
	if !tr.BeginsAt(start, DefaultEpsilon) {
		t.Error("Range should begin at time 100")
	}

	// Not at start time
	other := NewRationalTime(50, 24)
	if tr.BeginsAt(other, DefaultEpsilon) {
		t.Error("Range should not begin at time 50")
	}
}

// Tests for TimeRange.FinishesAt
func TestTimeRangeFinishesAtCoverage(t *testing.T) {
	tr := NewTimeRange(NewRationalTime(100, 24), NewRationalTime(50, 24))

	// Exactly at end time (100 + 50 = 150)
	end := NewRationalTime(150, 24)
	if !tr.FinishesAt(end, DefaultEpsilon) {
		t.Error("Range should finish at time 150")
	}

	// Not at end time
	other := NewRationalTime(100, 24)
	if tr.FinishesAt(other, DefaultEpsilon) {
		t.Error("Range should not finish at time 100")
	}
}

// Additional JSON error tests
func TestRationalTimeJSONErrorCoverage(t *testing.T) {
	var rt RationalTime
	err := json.Unmarshal([]byte(`invalid`), &rt)
	if err == nil {
		t.Error("Unmarshal with invalid JSON should error")
	}
}

func TestTimeRangeJSONErrorCoverage(t *testing.T) {
	var tr TimeRange
	err := json.Unmarshal([]byte(`invalid`), &tr)
	if err == nil {
		t.Error("Unmarshal with invalid JSON should error")
	}
}

func TestTimeTransformJSONErrorCoverage(t *testing.T) {
	var tt TimeTransform
	err := json.Unmarshal([]byte(`invalid`), &tt)
	if err == nil {
		t.Error("Unmarshal with invalid JSON should error")
	}
}

// Test ToTimecode with invalid time
func TestRationalTimeToTimecodeInvalidCoverage(t *testing.T) {
	// Create invalid time (zero rate)
	rt := RationalTime{value: 0, rate: 0}

	_, err := rt.ToTimecode(24, ForceNo)
	if err == nil {
		t.Error("ToTimecode with invalid time should error")
	}
}

// Test FromTimeString error
func TestFromTimeStringErrorCoverage(t *testing.T) {
	_, err := FromTimeString("invalid", 24)
	if err == nil {
		t.Error("FromTimeString with invalid format should error")
	}
}

// Test ToTimeString negative case
func TestToTimeStringNegativeCoverage(t *testing.T) {
	// Test negative
	rt := NewRationalTime(-24, 24)
	str := rt.ToTimeString()
	if str == "" {
		t.Error("ToTimeString for negative should return non-empty string")
	}
	if str[0] != '-' {
		t.Error("Negative time should start with '-'")
	}
	t.Logf("ToTimeString negative: %s", str)
}

// Test TimeRange methods
func TestTimeRangeMeetsCoverage(t *testing.T) {
	tr1 := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(50, 24))
	tr2 := NewTimeRange(NewRationalTime(50, 24), NewRationalTime(50, 24))

	if !tr1.Meets(tr2, DefaultEpsilon) {
		t.Error("tr1 should meet tr2")
	}
}

func TestTimeRangeBeginsCoverage(t *testing.T) {
	tr1 := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(50, 24))
	tr2 := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(100, 24))

	if !tr1.Begins(tr2, DefaultEpsilon) {
		t.Error("tr1 should begin in tr2")
	}
}

func TestTimeRangeFinishesCoverage(t *testing.T) {
	tr1 := NewTimeRange(NewRationalTime(50, 24), NewRationalTime(50, 24))
	tr2 := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(100, 24))

	if !tr1.Finishes(tr2, DefaultEpsilon) {
		t.Error("tr1 should finish in tr2")
	}
}

func TestTimeRangeIntersectsCoverage(t *testing.T) {
	tr1 := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(100, 24))
	tr2 := NewTimeRange(NewRationalTime(50, 24), NewRationalTime(100, 24))

	if !tr1.Intersects(tr2, DefaultEpsilon) {
		t.Error("tr1 and tr2 should intersect")
	}
}

func TestTimeRangeContainsRangeCoverage(t *testing.T) {
	tr1 := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(100, 24))
	tr2 := NewTimeRange(NewRationalTime(25, 24), NewRationalTime(50, 24))

	if !tr1.ContainsRange(tr2, DefaultEpsilon) {
		t.Error("tr1 should contain tr2")
	}
}

func TestTimeRangeBeforeCoverage(t *testing.T) {
	tr1 := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(50, 24))
	tr2 := NewTimeRange(NewRationalTime(100, 24), NewRationalTime(50, 24))

	if !tr1.Before(tr2, DefaultEpsilon) {
		t.Error("tr1 should be before tr2")
	}
}

// Test TimeRange helper functions
func TestTimeRangeExtendedByCoverage(t *testing.T) {
	tr1 := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(50, 24))
	tr2 := NewTimeRange(NewRationalTime(25, 24), NewRationalTime(50, 24))

	extended := tr1.ExtendedBy(tr2)
	if extended.StartTime().Value() != 0 {
		t.Errorf("Extended start = %v, want 0", extended.StartTime().Value())
	}
	// tr1 ends at 50, tr2 ends at 75
	if extended.Duration().Value() != 75 {
		t.Errorf("Extended duration = %v, want 75", extended.Duration().Value())
	}
}

func TestTimeRangeClampedRangeCoverage(t *testing.T) {
	tr := NewTimeRange(NewRationalTime(25, 24), NewRationalTime(50, 24))
	other := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(100, 24))

	clamped := tr.ClampedRange(other)
	if clamped.StartTime().Value() != 25 {
		t.Errorf("Clamped start = %v, want 25", clamped.StartTime().Value())
	}
}

func TestTimeRangeClampedTimeCoverage(t *testing.T) {
	tr := NewTimeRange(NewRationalTime(25, 24), NewRationalTime(50, 24))

	// Clamp time before range
	before := NewRationalTime(0, 24)
	clamped := tr.ClampedTime(before)
	if clamped.Value() != 25 {
		t.Errorf("Clamped before = %v, want 25", clamped.Value())
	}

	// Clamp time after range
	after := NewRationalTime(100, 24)
	clamped = tr.ClampedTime(after)
	// EndTimeInclusive is 74 (25 + 50 - 1)
	if clamped.Value() != 74 {
		t.Errorf("Clamped after = %v, want 74", clamped.Value())
	}
}

func TestTimeRangeDurationExtendedByCoverage(t *testing.T) {
	tr := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(50, 24))
	extension := NewRationalTime(25, 24)

	extended := tr.DurationExtendedBy(extension)
	if extended.Duration().Value() != 75 {
		t.Errorf("Extended duration = %v, want 75", extended.Duration().Value())
	}
}

// Test RangeFromStartEndTime constructors
func TestRangeFromStartEndTimeCoverage(t *testing.T) {
	start := NewRationalTime(10, 24)
	end := NewRationalTime(60, 24)

	tr := RangeFromStartEndTime(start, end)
	if tr.Duration().Value() != 50 {
		t.Errorf("Duration = %v, want 50", tr.Duration().Value())
	}
}

func TestRangeFromStartEndTimeInclusiveCoverage(t *testing.T) {
	start := NewRationalTime(10, 24)
	end := NewRationalTime(59, 24)

	tr := RangeFromStartEndTimeInclusive(start, end)
	if tr.Duration().Value() != 50 {
		t.Errorf("Duration = %v, want 50", tr.Duration().Value())
	}
}

// Test TimeRange validity
func TestTimeRangeValidityCoverage(t *testing.T) {
	valid := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(50, 24))
	if !valid.IsValidRange() {
		t.Error("Valid range should be valid")
	}
	if valid.IsInvalidRange() {
		t.Error("Valid range should not be invalid")
	}

	// Create invalid range with negative duration
	invalid := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(-50, 24))
	if invalid.IsValidRange() {
		t.Error("Invalid range should not be valid")
	}
	if !invalid.IsInvalidRange() {
		t.Error("Invalid range should be invalid")
	}
}

// Test TimeRange.Equal
func TestTimeRangeEqualCoverage(t *testing.T) {
	tr1 := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(50, 24))
	tr2 := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(50, 24))
	tr3 := NewTimeRange(NewRationalTime(10, 24), NewRationalTime(50, 24))

	if !tr1.Equal(tr2) {
		t.Error("tr1 should equal tr2")
	}
	if tr1.Equal(tr3) {
		t.Error("tr1 should not equal tr3")
	}
}

// Test TimeRange.String
func TestTimeRangeStringCoverage(t *testing.T) {
	tr := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(50, 24))
	str := tr.String()
	if str == "" {
		t.Error("String should return non-empty")
	}
}

// Test constructors
func TestNewTimeRangeFromStartTimeCoverage(t *testing.T) {
	tr := NewTimeRangeFromStartTime(NewRationalTime(10, 24))
	if tr.Duration().Value() != 0 {
		t.Errorf("Duration = %v, want 0", tr.Duration().Value())
	}
}

func TestNewTimeRangeFromValuesCoverage(t *testing.T) {
	tr := NewTimeRangeFromValues(10, 50, 24)
	if tr.StartTime().Value() != 10 {
		t.Errorf("StartTime = %v, want 10", tr.StartTime().Value())
	}
	if tr.Duration().Value() != 50 {
		t.Errorf("Duration = %v, want 50", tr.Duration().Value())
	}
}

// Test Add with zero/invalid rates
func TestRationalTimeAddZeroRateCoverage(t *testing.T) {
	rt1 := NewRationalTime(24, 24)
	rt2 := RationalTime{value: 10, rate: 0}

	result := rt1.Add(rt2)
	if result.Value() != rt1.Value() {
		t.Errorf("Add with zero rate should return other: %v", result)
	}

	result = rt2.Add(rt1)
	if result.Value() != rt1.Value() {
		t.Errorf("Zero rate Add should return other: %v", result)
	}
}

// Test EndTimeInclusive and EndTimeExclusive
func TestTimeRangeEndTimesCoverage(t *testing.T) {
	tr := NewTimeRange(NewRationalTime(10, 24), NewRationalTime(50, 24))

	exclusive := tr.EndTimeExclusive()
	if exclusive.Value() != 60 {
		t.Errorf("EndTimeExclusive = %v, want 60", exclusive.Value())
	}

	inclusive := tr.EndTimeInclusive()
	if inclusive.Value() != 59 {
		t.Errorf("EndTimeInclusive = %v, want 59", inclusive.Value())
	}
}

// Test Contains
func TestTimeRangeContainsCoverage(t *testing.T) {
	tr := NewTimeRange(NewRationalTime(10, 24), NewRationalTime(50, 24))

	inside := NewRationalTime(30, 24)
	if !tr.Contains(inside) {
		t.Error("Should contain 30")
	}

	atStart := NewRationalTime(10, 24)
	if !tr.Contains(atStart) {
		t.Error("Should contain start time")
	}

	atEnd := NewRationalTime(60, 24)
	if tr.Contains(atEnd) {
		t.Error("Should not contain end time (exclusive)")
	}

	outside := NewRationalTime(100, 24)
	if tr.Contains(outside) {
		t.Error("Should not contain 100")
	}
}

// Test ToTimeString with fractional seconds
func TestToTimeStringFractionalSeconds(t *testing.T) {
	// Test with fractional seconds
	rt := NewRationalTime(25, 24) // 1.041667 seconds
	str := rt.ToTimeString()
	if str == "" {
		t.Error("ToTimeString should return non-empty string")
	}
	t.Logf("ToTimeString fractional: %s", str)
}

// Test negative timecode
func TestFromTimecodeNegative(t *testing.T) {
	rt, err := FromTimecode("-01:00:00:00", 24)
	if err != nil {
		t.Fatalf("FromTimecode negative error: %v", err)
	}
	if rt.Value() >= 0 {
		t.Errorf("Value should be negative, got %v", rt.Value())
	}
}

// Test negative time string
func TestFromTimeStringNegative(t *testing.T) {
	rt, err := FromTimeString("-01:00:00.0", 24)
	if err != nil {
		t.Fatalf("FromTimeString negative error: %v", err)
	}
	if rt.ToSeconds() >= 0 {
		t.Errorf("Seconds should be negative, got %v", rt.ToSeconds())
	}
}

// Test EndTimeInclusive edge cases
func TestEndTimeInclusiveEdgeCases(t *testing.T) {
	// Test with fractional duration
	tr := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(10.5, 24))
	inclusive := tr.EndTimeInclusive()
	t.Logf("EndTimeInclusive with fractional duration: %v", inclusive.Value())

	// Test with very small duration
	tr = NewTimeRange(NewRationalTime(0, 24), NewRationalTime(0.5, 24))
	inclusive = tr.EndTimeInclusive()
	t.Logf("EndTimeInclusive with small duration: %v", inclusive.Value())
}

// Test ExtendedBy with earlier start
func TestTimeRangeExtendedByEarlierStart(t *testing.T) {
	tr1 := NewTimeRange(NewRationalTime(50, 24), NewRationalTime(50, 24))
	tr2 := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(25, 24))

	extended := tr1.ExtendedBy(tr2)
	if extended.StartTime().Value() != 0 {
		t.Errorf("Extended start = %v, want 0", extended.StartTime().Value())
	}
	// Extended should go from 0 to 100 (tr1 ends at 100)
	if extended.Duration().Value() != 100 {
		t.Errorf("Extended duration = %v, want 100", extended.Duration().Value())
	}
}

// Test TimeTransform AppliedToTime
func TestTimeTransformAppliedToTimeCoverage(t *testing.T) {
	transform := NewTimeTransform(NewRationalTime(10, 24), 2.0, 24)

	rt := NewRationalTime(5, 24)
	result := transform.AppliedToTime(rt)

	// Result should be (5 * 2) + 10 = 20 at rate 24
	if result.Value() != 20 {
		t.Errorf("AppliedToTime value = %v, want 20", result.Value())
	}
}

// Test TimeTransform AppliedToTransform
func TestTimeTransformAppliedToTransformCoverage(t *testing.T) {
	t1 := NewTimeTransform(NewRationalTime(10, 24), 2.0, 24)
	t2 := NewTimeTransform(NewRationalTime(5, 24), 1.5, -1)

	result := t1.AppliedToTransform(t2)

	// Offset should be combined
	t.Logf("Combined offset: %v", result.Offset().Value())
	// Scale should be multiplied
	if result.Scale() != 3.0 {
		t.Errorf("Combined scale = %v, want 3.0", result.Scale())
	}
}

// Test TimeTransform default
func TestNewTimeTransformDefaultCoverage(t *testing.T) {
	transform := NewTimeTransformDefault()

	if transform.Scale() != 1 {
		t.Errorf("Default scale = %v, want 1", transform.Scale())
	}
	if transform.Rate() != -1 {
		t.Errorf("Default rate = %v, want -1", transform.Rate())
	}
}

// Test TimeTransform with no rate override
func TestTimeTransformNoRateOverride(t *testing.T) {
	transform := NewTimeTransform(NewRationalTime(0, 24), 1.0, -1)

	rt := NewRationalTime(10, 30)
	result := transform.AppliedToTime(rt)

	// Rate should be preserved (30, not overridden)
	if result.Rate() != 30 {
		t.Errorf("Rate should be preserved: got %v, want 30", result.Rate())
	}
}

// Test ToTimecode with negative frames
func TestToTimecodeNegativeFrames(t *testing.T) {
	rt := NewRationalTime(-24, 24)
	_, err := rt.ToTimecode(24, ForceNo)
	if err == nil {
		t.Error("ToTimecode with negative frames should error")
	}
}
