// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentime

import (
	"fmt"
	"math"
)

// DefaultEpsilon is the default epsilon value for time range comparisons.
// It is computed to be twice 192kHz, the fastest commonly used audio rate.
const DefaultEpsilon = 1.0 / (2 * 192000.0)

// TimeRange represents a time range defined by a start time and duration.
// The duration indicates a time range that is inclusive of the start time,
// and exclusive of the end time.
type TimeRange struct {
	startTime RationalTime
	duration  RationalTime
}

// NewTimeRange creates a new TimeRange with the given start time and duration.
func NewTimeRange(startTime, duration RationalTime) TimeRange {
	return TimeRange{startTime: startTime, duration: duration}
}

// NewTimeRangeFromStartTime creates a new TimeRange with the given start time and zero duration.
func NewTimeRangeFromStartTime(startTime RationalTime) TimeRange {
	return TimeRange{
		startTime: startTime,
		duration:  RationalTime{value: 0, rate: startTime.rate},
	}
}

// NewTimeRangeFromValues creates a new TimeRange from raw values.
func NewTimeRangeFromValues(startTime, duration, rate float64) TimeRange {
	return TimeRange{
		startTime: RationalTime{value: startTime, rate: rate},
		duration:  RationalTime{value: duration, rate: rate},
	}
}

// StartTime returns the start time.
func (tr TimeRange) StartTime() RationalTime {
	return tr.startTime
}

// Duration returns the duration.
func (tr TimeRange) Duration() RationalTime {
	return tr.duration
}

// IsInvalidRange returns true if the time range is invalid.
// The time range is considered invalid if either the start time or
// duration is invalid, or if the duration is less than zero.
func (tr TimeRange) IsInvalidRange() bool {
	return tr.startTime.IsInvalidTime() || tr.duration.IsInvalidTime() || tr.duration.value < 0
}

// IsValidRange returns true if the time range is valid.
// The time range is considered valid if both the start time and
// duration are valid, and the duration is greater than or equal to zero.
func (tr TimeRange) IsValidRange() bool {
	return tr.startTime.IsValidTime() && tr.duration.IsValidTime() && tr.duration.value >= 0
}

// EndTimeInclusive returns the inclusive end time (last sample in range).
func (tr TimeRange) EndTimeInclusive() RationalTime {
	et := tr.EndTimeExclusive()

	if (et.Sub(tr.startTime.RescaledTo(tr.duration.rate))).value > 1 {
		if tr.duration.value != math.Floor(tr.duration.value) {
			return et.Floor()
		}
		return et.Sub(RationalTime{value: 1, rate: tr.duration.rate})
	}
	return tr.startTime
}

// EndTimeExclusive returns the exclusive end time (first sample after range).
func (tr TimeRange) EndTimeExclusive() RationalTime {
	return tr.duration.Add(tr.startTime.RescaledTo(tr.duration.rate))
}

// DurationExtendedBy extends the duration by the given time.
func (tr TimeRange) DurationExtendedBy(other RationalTime) TimeRange {
	return TimeRange{
		startTime: tr.startTime,
		duration:  tr.duration.Add(other),
	}
}

// ExtendedBy extends the time range to encompass another time range.
func (tr TimeRange) ExtendedBy(other TimeRange) TimeRange {
	newStartTime := tr.startTime
	if other.startTime.Cmp(newStartTime) < 0 {
		newStartTime = other.startTime
	}

	thisEnd := tr.EndTimeExclusive()
	otherEnd := other.EndTimeExclusive()
	newEndTime := thisEnd
	if otherEnd.Cmp(newEndTime) > 0 {
		newEndTime = otherEnd
	}

	return TimeRange{
		startTime: newStartTime,
		duration:  DurationFromStartEndTime(newStartTime, newEndTime),
	}
}

// ClampedTime clamps a time to this time range.
func (tr TimeRange) ClampedTime(other RationalTime) RationalTime {
	// min(max(other, startTime), endTimeInclusive)
	result := other
	if result.Cmp(tr.startTime) < 0 {
		result = tr.startTime
	}
	endInclusive := tr.EndTimeInclusive()
	if result.Cmp(endInclusive) > 0 {
		result = endInclusive
	}
	return result
}

// ClampedRange clamps another time range to this time range.
func (tr TimeRange) ClampedRange(other TimeRange) TimeRange {
	newStartTime := other.startTime
	if tr.startTime.Cmp(newStartTime) > 0 {
		newStartTime = tr.startTime
	}

	r := TimeRange{startTime: newStartTime, duration: other.duration}
	rEnd := r.EndTimeExclusive()
	thisEnd := tr.EndTimeExclusive()

	endTime := rEnd
	if thisEnd.Cmp(rEnd) < 0 {
		endTime = thisEnd
	}

	return TimeRange{
		startTime: newStartTime,
		duration:  endTime.Sub(newStartTime),
	}
}

// Contains returns whether this time range contains the given time.
func (tr TimeRange) Contains(other RationalTime) bool {
	return tr.startTime.Cmp(other) <= 0 && other.Cmp(tr.EndTimeExclusive()) < 0
}

// ContainsRange returns whether this time range contains another time range.
func (tr TimeRange) ContainsRange(other TimeRange, epsilon float64) bool {
	thisStart := tr.startTime.ToSeconds()
	thisEnd := tr.EndTimeExclusive().ToSeconds()
	otherStart := other.startTime.ToSeconds()
	otherEnd := other.EndTimeExclusive().ToSeconds()

	return greaterThan(otherStart, thisStart, epsilon) && lessThan(otherEnd, thisEnd, epsilon)
}

// OverlapsRange returns whether this and another time range overlap.
func (tr TimeRange) OverlapsRange(other TimeRange, epsilon float64) bool {
	thisStart := tr.startTime.ToSeconds()
	thisEnd := tr.EndTimeExclusive().ToSeconds()
	otherStart := other.startTime.ToSeconds()
	otherEnd := other.EndTimeExclusive().ToSeconds()

	return lessThan(thisStart, otherStart, epsilon) &&
		greaterThan(thisEnd, otherStart, epsilon) &&
		greaterThan(otherEnd, thisEnd, epsilon)
}

// Before returns whether this time range precedes another time range.
func (tr TimeRange) Before(other TimeRange, epsilon float64) bool {
	thisEnd := tr.EndTimeExclusive().ToSeconds()
	otherStart := other.startTime.ToSeconds()
	return greaterThan(otherStart, thisEnd, epsilon)
}

// BeforeTime returns whether this time range precedes the given time.
func (tr TimeRange) BeforeTime(other RationalTime, epsilon float64) bool {
	thisEnd := tr.EndTimeExclusive().ToSeconds()
	otherTime := other.ToSeconds()
	return lessThan(thisEnd, otherTime, epsilon)
}

// Meets returns whether this time range meets another time range.
func (tr TimeRange) Meets(other TimeRange, epsilon float64) bool {
	thisEnd := tr.EndTimeExclusive().ToSeconds()
	otherStart := other.startTime.ToSeconds()
	return otherStart-thisEnd <= epsilon && otherStart-thisEnd >= 0
}

// Begins returns whether this time range begins in another time range.
func (tr TimeRange) Begins(other TimeRange, epsilon float64) bool {
	thisStart := tr.startTime.ToSeconds()
	thisEnd := tr.EndTimeExclusive().ToSeconds()
	otherStart := other.startTime.ToSeconds()
	otherEnd := other.EndTimeExclusive().ToSeconds()

	return math.Abs(otherStart-thisStart) <= epsilon && lessThan(thisEnd, otherEnd, epsilon)
}

// BeginsAt returns whether this range begins at the given time.
func (tr TimeRange) BeginsAt(other RationalTime, epsilon float64) bool {
	thisStart := tr.startTime.ToSeconds()
	otherStart := other.ToSeconds()
	return math.Abs(otherStart-thisStart) <= epsilon
}

// Finishes returns whether this time range finishes in another time range.
func (tr TimeRange) Finishes(other TimeRange, epsilon float64) bool {
	thisStart := tr.startTime.ToSeconds()
	thisEnd := tr.EndTimeExclusive().ToSeconds()
	otherStart := other.startTime.ToSeconds()
	otherEnd := other.EndTimeExclusive().ToSeconds()

	return math.Abs(thisEnd-otherEnd) <= epsilon && greaterThan(thisStart, otherStart, epsilon)
}

// FinishesAt returns whether this time range finishes at the given time.
func (tr TimeRange) FinishesAt(other RationalTime, epsilon float64) bool {
	thisEnd := tr.EndTimeExclusive().ToSeconds()
	otherEnd := other.ToSeconds()
	return math.Abs(thisEnd-otherEnd) <= epsilon
}

// Intersects returns whether this time range intersects another time range.
func (tr TimeRange) Intersects(other TimeRange, epsilon float64) bool {
	thisStart := tr.startTime.ToSeconds()
	thisEnd := tr.EndTimeExclusive().ToSeconds()
	otherStart := other.startTime.ToSeconds()
	otherEnd := other.EndTimeExclusive().ToSeconds()

	return lessThan(thisStart, otherEnd, epsilon) && greaterThan(thisEnd, otherStart, epsilon)
}

// Equal returns whether two time ranges are equal.
func (tr TimeRange) Equal(other TimeRange) bool {
	start := tr.startTime.Sub(other.startTime)
	duration := tr.duration.Sub(other.duration)
	return math.Abs(start.ToSeconds()) < DefaultEpsilon &&
		math.Abs(duration.ToSeconds()) < DefaultEpsilon
}

// RangeFromStartEndTime creates a time range from a start time and exclusive end time.
func RangeFromStartEndTime(startTime, endTimeExclusive RationalTime) TimeRange {
	return TimeRange{
		startTime: startTime,
		duration:  DurationFromStartEndTime(startTime, endTimeExclusive),
	}
}

// RangeFromStartEndTimeInclusive creates a time range from a start time and inclusive end time.
func RangeFromStartEndTimeInclusive(startTime, endTimeInclusive RationalTime) TimeRange {
	return TimeRange{
		startTime: startTime,
		duration:  DurationFromStartEndTimeInclusive(startTime, endTimeInclusive),
	}
}

// String returns a string representation of the TimeRange.
func (tr TimeRange) String() string {
	return fmt.Sprintf("TimeRange(%s, %s)", tr.startTime.String(), tr.duration.String())
}

// Helper functions for epsilon comparisons

func greaterThan(lhs, rhs, epsilon float64) bool {
	return lhs-rhs >= epsilon
}

func lessThan(lhs, rhs, epsilon float64) bool {
	return rhs-lhs >= epsilon
}
