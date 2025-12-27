// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentime

import (
	"math"
	"testing"
)

func TestNewTimeRange(t *testing.T) {
	start := NewRationalTime(10, 24)
	dur := NewRationalTime(20, 24)
	tr := NewTimeRange(start, dur)

	if !tr.StartTime().StrictlyEqual(start) {
		t.Error("Start time mismatch")
	}
	if !tr.Duration().StrictlyEqual(dur) {
		t.Error("Duration mismatch")
	}
}

func TestNewTimeRangeFromStartTime(t *testing.T) {
	start := NewRationalTime(10, 24)
	tr := NewTimeRangeFromStartTime(start)

	if !tr.StartTime().StrictlyEqual(start) {
		t.Error("Start time mismatch")
	}
	if tr.Duration().Value() != 0 {
		t.Errorf("Expected zero duration, got %g", tr.Duration().Value())
	}
}

func TestNewTimeRangeFromValues(t *testing.T) {
	tr := NewTimeRangeFromValues(10, 20, 24)

	if tr.StartTime().Value() != 10 {
		t.Errorf("Expected start 10, got %g", tr.StartTime().Value())
	}
	if tr.Duration().Value() != 20 {
		t.Errorf("Expected duration 20, got %g", tr.Duration().Value())
	}
	if tr.StartTime().Rate() != 24 {
		t.Errorf("Expected rate 24, got %g", tr.StartTime().Rate())
	}
}

func TestTimeRangeIsValid(t *testing.T) {
	tests := []struct {
		name    string
		tr      TimeRange
		isValid bool
	}{
		{"valid", NewTimeRangeFromValues(10, 20, 24), true},
		{"zero duration", NewTimeRangeFromValues(10, 0, 24), true},
		{"negative duration", NewTimeRangeFromValues(10, -1, 24), false},
		{"invalid start", NewTimeRange(NewRationalTime(10, 0), NewRationalTime(20, 24)), false},
		{"invalid duration", NewTimeRange(NewRationalTime(10, 24), NewRationalTime(20, 0)), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.tr.IsValidRange() != tt.isValid {
				t.Errorf("Expected IsValidRange() = %v, got %v", tt.isValid, tt.tr.IsValidRange())
			}
			if tt.tr.IsInvalidRange() != !tt.isValid {
				t.Errorf("Expected IsInvalidRange() = %v, got %v", !tt.isValid, tt.tr.IsInvalidRange())
			}
		})
	}
}

func TestTimeRangeEndTime(t *testing.T) {
	tr := NewTimeRangeFromValues(10, 20, 24)

	endExclusive := tr.EndTimeExclusive()
	if endExclusive.Value() != 30 {
		t.Errorf("Expected end exclusive 30, got %g", endExclusive.Value())
	}

	endInclusive := tr.EndTimeInclusive()
	if endInclusive.Value() != 29 {
		t.Errorf("Expected end inclusive 29, got %g", endInclusive.Value())
	}
}

func TestTimeRangeDurationExtendedBy(t *testing.T) {
	tr := NewTimeRangeFromValues(10, 20, 24)
	extension := NewRationalTime(5, 24)

	extended := tr.DurationExtendedBy(extension)
	if extended.Duration().Value() != 25 {
		t.Errorf("Expected duration 25, got %g", extended.Duration().Value())
	}
	if extended.StartTime().Value() != 10 {
		t.Errorf("Start time should not change, got %g", extended.StartTime().Value())
	}
}

func TestTimeRangeExtendedBy(t *testing.T) {
	tr1 := NewTimeRangeFromValues(10, 20, 24) // 10-30
	tr2 := NewTimeRangeFromValues(25, 20, 24) // 25-45

	extended := tr1.ExtendedBy(tr2)
	if extended.StartTime().Value() != 10 {
		t.Errorf("Expected start 10, got %g", extended.StartTime().Value())
	}
	if extended.EndTimeExclusive().Value() != 45 {
		t.Errorf("Expected end 45, got %g", extended.EndTimeExclusive().Value())
	}
}

func TestTimeRangeClampedTime(t *testing.T) {
	tr := NewTimeRangeFromValues(10, 20, 24) // 10-30

	// Time before range
	before := NewRationalTime(5, 24)
	clamped := tr.ClampedTime(before)
	if clamped.Value() != 10 {
		t.Errorf("Expected clamped to start 10, got %g", clamped.Value())
	}

	// Time within range
	within := NewRationalTime(15, 24)
	clamped = tr.ClampedTime(within)
	if clamped.Value() != 15 {
		t.Errorf("Expected unchanged 15, got %g", clamped.Value())
	}

	// Time after range
	after := NewRationalTime(40, 24)
	clamped = tr.ClampedTime(after)
	if clamped.Value() != 29 { // end inclusive
		t.Errorf("Expected clamped to end 29, got %g", clamped.Value())
	}
}

func TestTimeRangeClampedRange(t *testing.T) {
	tr := NewTimeRangeFromValues(10, 20, 24) // 10-30
	other := NewTimeRangeFromValues(5, 30, 24) // 5-35

	clamped := tr.ClampedRange(other)
	if clamped.StartTime().Value() != 10 {
		t.Errorf("Expected start 10, got %g", clamped.StartTime().Value())
	}
	if clamped.EndTimeExclusive().Value() != 30 {
		t.Errorf("Expected end 30, got %g", clamped.EndTimeExclusive().Value())
	}
}

func TestTimeRangeContains(t *testing.T) {
	tr := NewTimeRangeFromValues(10, 20, 24) // 10-30

	tests := []struct {
		name     string
		time     RationalTime
		contains bool
	}{
		{"before", NewRationalTime(5, 24), false},
		{"at start", NewRationalTime(10, 24), true},
		{"within", NewRationalTime(20, 24), true},
		{"at end exclusive", NewRationalTime(30, 24), false},
		{"after", NewRationalTime(35, 24), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tr.Contains(tt.time) != tt.contains {
				t.Errorf("Contains(%s) = %v, want %v", tt.time, tr.Contains(tt.time), tt.contains)
			}
		})
	}
}

func TestTimeRangeContainsRange(t *testing.T) {
	tr := NewTimeRangeFromValues(10, 20, 24) // 10-30

	// Fully contained
	inner := NewTimeRangeFromValues(15, 10, 24) // 15-25
	if !tr.ContainsRange(inner, DefaultEpsilon) {
		t.Error("Expected to contain inner range")
	}

	// Not contained - extends beyond
	outer := NewTimeRangeFromValues(15, 20, 24) // 15-35
	if tr.ContainsRange(outer, DefaultEpsilon) {
		t.Error("Expected not to contain outer range")
	}
}

func TestTimeRangeOverlapsRange(t *testing.T) {
	tr := NewTimeRangeFromValues(10, 20, 24) // 10-30

	// Overlapping
	overlap := NewTimeRangeFromValues(20, 20, 24) // 20-40
	if !tr.OverlapsRange(overlap, DefaultEpsilon) {
		t.Error("Expected ranges to overlap")
	}

	// Not overlapping
	noOverlap := NewTimeRangeFromValues(40, 10, 24) // 40-50
	if tr.OverlapsRange(noOverlap, DefaultEpsilon) {
		t.Error("Expected ranges not to overlap")
	}
}

func TestTimeRangeBefore(t *testing.T) {
	tr := NewTimeRangeFromValues(10, 20, 24) // 10-30

	// Before
	after := NewTimeRangeFromValues(40, 10, 24) // 40-50
	if !tr.Before(after, DefaultEpsilon) {
		t.Error("Expected tr to be before after")
	}

	// Not before
	overlap := NewTimeRangeFromValues(20, 20, 24) // 20-40
	if tr.Before(overlap, DefaultEpsilon) {
		t.Error("Expected tr not to be before overlap")
	}
}

func TestTimeRangeMeets(t *testing.T) {
	tr := NewTimeRangeFromValues(10, 20, 24) // 10-30

	// Meets
	next := NewTimeRangeFromValues(30, 10, 24) // 30-40
	if !tr.Meets(next, DefaultEpsilon) {
		t.Error("Expected tr to meet next")
	}

	// Doesn't meet
	gap := NewTimeRangeFromValues(35, 10, 24) // 35-45
	if tr.Meets(gap, DefaultEpsilon) {
		t.Error("Expected tr not to meet gap")
	}
}

func TestTimeRangeBegins(t *testing.T) {
	tr := NewTimeRangeFromValues(10, 10, 24) // 10-20
	other := NewTimeRangeFromValues(10, 20, 24) // 10-30

	if !tr.Begins(other, DefaultEpsilon) {
		t.Error("Expected tr to begin in other")
	}

	// Different start
	different := NewTimeRangeFromValues(15, 10, 24) // 15-25
	if tr.Begins(different, DefaultEpsilon) {
		t.Error("Expected tr not to begin in different")
	}
}

func TestTimeRangeFinishes(t *testing.T) {
	tr := NewTimeRangeFromValues(20, 10, 24) // 20-30
	other := NewTimeRangeFromValues(10, 20, 24) // 10-30

	if !tr.Finishes(other, DefaultEpsilon) {
		t.Error("Expected tr to finish in other")
	}

	// Different end
	different := NewTimeRangeFromValues(15, 10, 24) // 15-25
	if tr.Finishes(different, DefaultEpsilon) {
		t.Error("Expected tr not to finish in different")
	}
}

func TestTimeRangeIntersects(t *testing.T) {
	tr := NewTimeRangeFromValues(10, 20, 24) // 10-30

	tests := []struct {
		name       string
		other      TimeRange
		intersects bool
	}{
		{"fully before", NewTimeRangeFromValues(0, 5, 24), false},
		{"touches start", NewTimeRangeFromValues(0, 10, 24), false},
		{"overlaps start", NewTimeRangeFromValues(5, 10, 24), true},
		{"contained", NewTimeRangeFromValues(15, 5, 24), true},
		{"overlaps end", NewTimeRangeFromValues(25, 10, 24), true},
		{"touches end", NewTimeRangeFromValues(30, 10, 24), false},
		{"fully after", NewTimeRangeFromValues(40, 10, 24), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tr.Intersects(tt.other, DefaultEpsilon) != tt.intersects {
				t.Errorf("Intersects(%s) = %v, want %v", tt.other, tr.Intersects(tt.other, DefaultEpsilon), tt.intersects)
			}
		})
	}
}

func TestTimeRangeEqual(t *testing.T) {
	tr1 := NewTimeRangeFromValues(10, 20, 24)
	tr2 := NewTimeRangeFromValues(10, 20, 24)
	tr3 := NewTimeRangeFromValues(10, 21, 24)

	if !tr1.Equal(tr2) {
		t.Error("Expected equal ranges")
	}
	if tr1.Equal(tr3) {
		t.Error("Expected unequal ranges")
	}
}

func TestRangeFromStartEndTime(t *testing.T) {
	start := NewRationalTime(10, 24)
	end := NewRationalTime(30, 24)

	tr := RangeFromStartEndTime(start, end)
	if tr.StartTime().Value() != 10 {
		t.Errorf("Expected start 10, got %g", tr.StartTime().Value())
	}
	if tr.Duration().Value() != 20 {
		t.Errorf("Expected duration 20, got %g", tr.Duration().Value())
	}
}

func TestRangeFromStartEndTimeInclusive(t *testing.T) {
	start := NewRationalTime(10, 24)
	end := NewRationalTime(29, 24)

	tr := RangeFromStartEndTimeInclusive(start, end)
	if tr.StartTime().Value() != 10 {
		t.Errorf("Expected start 10, got %g", tr.StartTime().Value())
	}
	if tr.Duration().Value() != 20 {
		t.Errorf("Expected duration 20, got %g", tr.Duration().Value())
	}
}

func TestTimeRangeString(t *testing.T) {
	tr := NewTimeRangeFromValues(10, 20, 24)
	str := tr.String()
	expected := "TimeRange(RationalTime(10, 24), RationalTime(20, 24))"
	if str != expected {
		t.Errorf("Expected '%s', got '%s'", expected, str)
	}
}

func TestTimeRangeEndTimeInclusiveFractional(t *testing.T) {
	// Test with fractional duration
	tr := NewTimeRange(
		NewRationalTime(0, 24),
		NewRationalTime(10.5, 24),
	)

	endInclusive := tr.EndTimeInclusive()
	if endInclusive.Value() != 10 { // floor of 10.5
		t.Errorf("Expected end inclusive 10, got %g", endInclusive.Value())
	}
}

func TestTimeRangeEndTimeInclusiveSmall(t *testing.T) {
	// Test with very small duration
	tr := NewTimeRange(
		NewRationalTime(10, 24),
		NewRationalTime(0.5, 24),
	)

	endInclusive := tr.EndTimeInclusive()
	// With duration < 1, should return start time
	if endInclusive.Value() != 10 {
		t.Errorf("Expected end inclusive 10 (start), got %g", endInclusive.Value())
	}
}

func TestDefaultEpsilon(t *testing.T) {
	expected := 1.0 / (2 * 192000.0)
	if math.Abs(DefaultEpsilon-expected) > 1e-15 {
		t.Errorf("DefaultEpsilon = %g, want %g", DefaultEpsilon, expected)
	}
}
