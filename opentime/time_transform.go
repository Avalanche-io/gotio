// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentime

import "fmt"

// TimeTransform represents a one-dimensional time transform.
// It consists of an offset, scale, and optional rate.
type TimeTransform struct {
	offset RationalTime
	scale  float64
	rate   float64
}

// NewTimeTransform creates a new TimeTransform with the given offset, scale, and rate.
// A rate of -1 or less means no rate override (use the input's rate).
func NewTimeTransform(offset RationalTime, scale, rate float64) TimeTransform {
	return TimeTransform{offset: offset, scale: scale, rate: rate}
}

// NewTimeTransformDefault creates a default TimeTransform (no offset, scale=1, no rate override).
func NewTimeTransformDefault() TimeTransform {
	return TimeTransform{offset: RationalTime{value: 0, rate: 1}, scale: 1, rate: -1}
}

// Offset returns the offset.
func (tt TimeTransform) Offset() RationalTime {
	return tt.offset
}

// Scale returns the scale.
func (tt TimeTransform) Scale() float64 {
	return tt.scale
}

// Rate returns the rate. A value <= 0 means no rate override.
func (tt TimeTransform) Rate() float64 {
	return tt.rate
}

// AppliedToTime applies the transform to a RationalTime and returns the transformed time.
func (tt TimeTransform) AppliedToTime(other RationalTime) RationalTime {
	result := RationalTime{
		value: other.value * tt.scale,
		rate:  other.rate,
	}.Add(tt.offset)

	targetRate := tt.rate
	if targetRate <= 0 {
		targetRate = other.rate
	}

	if targetRate > 0 {
		return result.RescaledTo(targetRate)
	}
	return result
}

// AppliedToRange applies the transform to a TimeRange and returns the transformed range.
func (tt TimeTransform) AppliedToRange(other TimeRange) TimeRange {
	return RangeFromStartEndTime(
		tt.AppliedToTime(other.startTime),
		tt.AppliedToTime(other.EndTimeExclusive()),
	)
}

// AppliedToTransform applies this transform to another TimeTransform and returns the combined transform.
func (tt TimeTransform) AppliedToTransform(other TimeTransform) TimeTransform {
	rate := tt.rate
	if rate <= 0 {
		rate = other.rate
	}
	return TimeTransform{
		offset: tt.offset.Add(other.offset),
		scale:  tt.scale * other.scale,
		rate:   rate,
	}
}

// Equal returns whether two transforms are equal.
func (tt TimeTransform) Equal(other TimeTransform) bool {
	return tt.offset.Equal(other.offset) &&
		tt.scale == other.scale &&
		tt.rate == other.rate
}

// String returns a string representation of the TimeTransform.
func (tt TimeTransform) String() string {
	return fmt.Sprintf("TimeTransform(%s, %g, %g)", tt.offset.String(), tt.scale, tt.rate)
}
