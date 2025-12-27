// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

// Package opentime provides time representation types for OpenTimelineIO.
// It includes RationalTime for representing a moment in time with a rate,
// TimeRange for representing a range of time, and TimeTransform for
// transforming time values.
package opentime

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// IsDropFrameRate specifies options for drop frame timecode.
type IsDropFrameRate int

const (
	// InferFromRate automatically determines drop frame based on rate.
	InferFromRate IsDropFrameRate = -1
	// ForceNo forces non-drop frame timecode.
	ForceNo IsDropFrameRate = 0
	// ForceYes forces drop frame timecode.
	ForceYes IsDropFrameRate = 1
)

// RationalTime represents a measure of time defined by a value and rate.
// The time is value/rate seconds.
type RationalTime struct {
	value float64
	rate  float64
}

// NewRationalTime creates a new RationalTime with the given value and rate.
func NewRationalTime(value, rate float64) RationalTime {
	return RationalTime{value: value, rate: rate}
}

// Value returns the time value (number of ticks at the given rate).
func (rt RationalTime) Value() float64 {
	return rt.value
}

// Rate returns the time rate (ticks per second).
func (rt RationalTime) Rate() float64 {
	return rt.rate
}

// IsInvalidTime returns true if the time is invalid.
// The time is considered invalid if the value or rate are NaN,
// or if the rate is less than or equal to zero.
func (rt RationalTime) IsInvalidTime() bool {
	return math.IsNaN(rt.rate) || math.IsNaN(rt.value) || rt.rate <= 0
}

// IsValidTime returns true if the time is valid.
// The time is considered valid if the value and rate are not NaN,
// and the rate is greater than zero.
func (rt RationalTime) IsValidTime() bool {
	return !math.IsNaN(rt.rate) && !math.IsNaN(rt.value) && rt.rate > 0
}

// RescaledTo returns the time converted to a new rate.
func (rt RationalTime) RescaledTo(newRate float64) RationalTime {
	return RationalTime{
		value: rt.ValueRescaledTo(newRate),
		rate:  newRate,
	}
}

// RescaledToRate returns the time converted to match another RationalTime's rate.
func (rt RationalTime) RescaledToRate(other RationalTime) RationalTime {
	return rt.RescaledTo(other.rate)
}

// ValueRescaledTo returns the time value converted to a new rate.
func (rt RationalTime) ValueRescaledTo(newRate float64) float64 {
	if newRate == rt.rate {
		return rt.value
	}
	return (rt.value * newRate) / rt.rate
}

// ValueRescaledToRate returns the time value converted to match another RationalTime's rate.
func (rt RationalTime) ValueRescaledToRate(other RationalTime) float64 {
	return rt.ValueRescaledTo(other.rate)
}

// AlmostEqual returns whether this time is almost equal to another time.
func (rt RationalTime) AlmostEqual(other RationalTime, delta float64) bool {
	return math.Abs(rt.ValueRescaledTo(other.rate)-other.value) <= delta
}

// StrictlyEqual returns whether this value and rate are exactly equal to another time.
// Note that this is different from the equality comparison that rescales before comparing.
func (rt RationalTime) StrictlyEqual(other RationalTime) bool {
	return rt.value == other.value && rt.rate == other.rate
}

// Floor returns a time with the largest integer value not greater than this value.
func (rt RationalTime) Floor() RationalTime {
	return RationalTime{value: math.Floor(rt.value), rate: rt.rate}
}

// Ceil returns a time with the smallest integer value not less than this value.
func (rt RationalTime) Ceil() RationalTime {
	return RationalTime{value: math.Ceil(rt.value), rate: rt.rate}
}

// Round returns a time with the nearest integer value to this value.
func (rt RationalTime) Round() RationalTime {
	return RationalTime{value: math.Round(rt.value), rate: rt.rate}
}

// DurationFromStartEndTime computes the duration of samples from first to last (excluding last).
// For example, the duration of a clip from frame 10 to frame 15 is 5 frames.
// The result will be in the rate of the start time.
func DurationFromStartEndTime(startTime, endTimeExclusive RationalTime) RationalTime {
	if startTime.rate == endTimeExclusive.rate {
		return RationalTime{
			value: endTimeExclusive.value - startTime.value,
			rate:  startTime.rate,
		}
	}
	return RationalTime{
		value: endTimeExclusive.ValueRescaledTo(startTime.rate) - startTime.value,
		rate:  startTime.rate,
	}
}

// DurationFromStartEndTimeInclusive computes the duration of samples from first to last (including last).
// For example, the duration of a clip from frame 10 to frame 15 is 6 frames.
// The result will be in the rate of the start time.
func DurationFromStartEndTimeInclusive(startTime, endTimeInclusive RationalTime) RationalTime {
	if startTime.rate == endTimeInclusive.rate {
		return RationalTime{
			value: endTimeInclusive.value - startTime.value + 1,
			rate:  startTime.rate,
		}
	}
	return RationalTime{
		value: endTimeInclusive.ValueRescaledTo(startTime.rate) - startTime.value + 1,
		rate:  startTime.rate,
	}
}

// smptTimecodeRates contains valid SMPTE timecode rates.
var smpteTimecodeRates = []float64{
	23.976, 24, 25, 29.97, 30, 50, 59.94, 60,
}

// IsSMPTETimecodeRate returns true if the rate is supported by SMPTE timecode.
func IsSMPTETimecodeRate(rate float64) bool {
	for _, r := range smpteTimecodeRates {
		if math.Abs(rate-r) < 0.01 {
			return true
		}
	}
	return false
}

// NearestSMPTETimecodeRate returns the SMPTE timecode rate nearest to the given rate.
func NearestSMPTETimecodeRate(rate float64) float64 {
	nearest := smpteTimecodeRates[0]
	minDiff := math.Abs(rate - nearest)
	for _, r := range smpteTimecodeRates[1:] {
		diff := math.Abs(rate - r)
		if diff < minDiff {
			minDiff = diff
			nearest = r
		}
	}
	return nearest
}

// FromFrames converts a frame number and rate into a time.
func FromFrames(frame, rate float64) RationalTime {
	return RationalTime{value: math.Trunc(frame), rate: rate}
}

// FromSeconds converts a value in seconds and rate into a time.
func FromSeconds(seconds, rate float64) RationalTime {
	return RationalTime{value: seconds, rate: 1}.RescaledTo(rate)
}

// FromSecondsFloat converts a value in seconds into a time with rate 1.
func FromSecondsFloat(seconds float64) RationalTime {
	return RationalTime{value: seconds, rate: 1}
}

// ToFrames returns the frame number based on the current rate.
func (rt RationalTime) ToFrames() int {
	return int(rt.value)
}

// ToFramesAtRate returns the frame number based on the given rate.
func (rt RationalTime) ToFramesAtRate(rate float64) int {
	return int(rt.ValueRescaledTo(rate))
}

// ToSeconds returns the value in seconds.
func (rt RationalTime) ToSeconds() float64 {
	return rt.ValueRescaledTo(1)
}

// isDropFrameRate determines if a rate uses drop frame timecode.
func isDropFrameRate(rate float64) bool {
	// 29.97 and 59.94 use drop frame
	return math.Abs(rate-29.97) < 0.01 || math.Abs(rate-59.94) < 0.01
}

// ToTimecode converts to timecode (e.g., "HH:MM:SS;FRAME").
func (rt RationalTime) ToTimecode(rate float64, dropFrame IsDropFrameRate) (string, error) {
	if rt.IsInvalidTime() {
		return "", fmt.Errorf("invalid time")
	}

	useDropFrame := false
	if dropFrame == ForceYes {
		useDropFrame = true
	} else if dropFrame == InferFromRate {
		useDropFrame = isDropFrameRate(rate)
	}

	// Rescale to the target rate
	rescaled := rt.RescaledTo(rate)
	totalFrames := int64(math.Round(rescaled.value))

	if totalFrames < 0 {
		return "", fmt.Errorf("negative timecode not supported")
	}

	nominalRate := int64(math.Round(rate))
	if useDropFrame {
		// Drop frame calculation
		// For 29.97, drop 2 frames every minute except every 10th minute
		// For 59.94, drop 4 frames every minute except every 10th minute
		var dropFrames int64 = 2
		if nominalRate >= 60 {
			dropFrames = 4
		}

		framesPerMinute := nominalRate*60 - dropFrames
		framesPer10Minutes := framesPerMinute*10 + dropFrames

		d := totalFrames / framesPer10Minutes
		m := totalFrames % framesPer10Minutes

		if m < dropFrames {
			m += dropFrames
		}

		frameCount := d*framesPer10Minutes + (m-dropFrames)/framesPerMinute*(framesPerMinute+dropFrames) +
			(m-dropFrames)%framesPerMinute + dropFrames

		// Recalculate with adjusted frame count
		frames := int(frameCount % nominalRate)
		seconds := int((frameCount / nominalRate) % 60)
		minutes := int((frameCount / nominalRate / 60) % 60)
		hours := int(frameCount / nominalRate / 3600)

		return fmt.Sprintf("%02d:%02d:%02d;%02d", hours, minutes, seconds, frames), nil
	}

	// Non-drop frame
	frames := int(totalFrames % nominalRate)
	seconds := int((totalFrames / nominalRate) % 60)
	minutes := int((totalFrames / nominalRate / 60) % 60)
	hours := int(totalFrames / nominalRate / 3600)

	return fmt.Sprintf("%02d:%02d:%02d:%02d", hours, minutes, seconds, frames), nil
}

// ToTimecodeAuto converts to timecode using the current rate.
func (rt RationalTime) ToTimecodeAuto() (string, error) {
	return rt.ToTimecode(rt.rate, InferFromRate)
}

// ToNearestTimecode converts to the nearest timecode.
func (rt RationalTime) ToNearestTimecode(rate float64, dropFrame IsDropFrameRate) (string, error) {
	return rt.Round().ToTimecode(rate, dropFrame)
}

// ToTimeString returns a string in the form "hours:minutes:seconds".
// Seconds may have up to microsecond precision.
func (rt RationalTime) ToTimeString() string {
	totalSeconds := rt.ToSeconds()
	negative := totalSeconds < 0
	if negative {
		totalSeconds = -totalSeconds
	}

	hours := int(totalSeconds / 3600)
	minutes := int(math.Mod(totalSeconds/60, 60))
	seconds := math.Mod(totalSeconds, 60)

	prefix := ""
	if negative {
		prefix = "-"
	}

	// Format with microsecond precision, trimming trailing zeros
	// Ensure seconds has leading zero if needed
	intSeconds := int(seconds)
	fracSeconds := seconds - float64(intSeconds)

	var secondsStr string
	if fracSeconds == 0 {
		secondsStr = fmt.Sprintf("%02d.0", intSeconds)
	} else {
		fracStr := fmt.Sprintf("%.6f", fracSeconds)
		fracStr = strings.TrimPrefix(fracStr, "0")
		fracStr = strings.TrimRight(fracStr, "0")
		secondsStr = fmt.Sprintf("%02d%s", intSeconds, fracStr)
	}

	return fmt.Sprintf("%s%02d:%02d:%s", prefix, hours, minutes, secondsStr)
}

// timecodeRegex matches timecode strings.
var timecodeRegex = regexp.MustCompile(`^(-?)(\d{1,2}):(\d{2}):(\d{2})([;:])?(\d{2,})$`)

// FromTimecode converts a timecode string ("HH:MM:SS;FRAME") into a time.
func FromTimecode(timecode string, rate float64) (RationalTime, error) {
	matches := timecodeRegex.FindStringSubmatch(timecode)
	if matches == nil {
		return RationalTime{}, fmt.Errorf("invalid timecode format: %s", timecode)
	}

	negative := matches[1] == "-"
	hours, _ := strconv.Atoi(matches[2])
	minutes, _ := strconv.Atoi(matches[3])
	seconds, _ := strconv.Atoi(matches[4])
	separator := matches[5]
	frames, _ := strconv.Atoi(matches[6])

	useDropFrame := separator == ";"

	nominalRate := int(math.Round(rate))

	var totalFrames int64
	if useDropFrame {
		// Drop frame calculation
		dropFrames := 2
		if nominalRate >= 60 {
			dropFrames = 4
		}

		framesPerMinute := nominalRate*60 - dropFrames
		framesPer10Minutes := framesPerMinute*10 + dropFrames

		totalMinutes := int64(hours)*60 + int64(minutes)
		totalFrames = int64(framesPer10Minutes)*(totalMinutes/10) +
			int64(framesPerMinute)*(totalMinutes%10) +
			int64(seconds*nominalRate) + int64(frames) -
			int64(dropFrames)*(totalMinutes-totalMinutes/10)
	} else {
		totalFrames = int64(hours)*3600*int64(nominalRate) +
			int64(minutes)*60*int64(nominalRate) +
			int64(seconds)*int64(nominalRate) +
			int64(frames)
	}

	if negative {
		totalFrames = -totalFrames
	}

	return RationalTime{value: float64(totalFrames), rate: rate}, nil
}

// timeStringRegex matches time strings.
var timeStringRegex = regexp.MustCompile(`^(-?)(\d+):(\d{2}):(\d+(?:\.\d+)?)$`)

// FromTimeString parses a string in the form "hours:minutes:seconds".
func FromTimeString(timeString string, rate float64) (RationalTime, error) {
	matches := timeStringRegex.FindStringSubmatch(timeString)
	if matches == nil {
		return RationalTime{}, fmt.Errorf("invalid time string format: %s", timeString)
	}

	negative := matches[1] == "-"
	hours, _ := strconv.Atoi(matches[2])
	minutes, _ := strconv.Atoi(matches[3])
	seconds, _ := strconv.ParseFloat(matches[4], 64)

	totalSeconds := float64(hours)*3600 + float64(minutes)*60 + seconds
	if negative {
		totalSeconds = -totalSeconds
	}

	return FromSeconds(totalSeconds, rate), nil
}

// Add returns the sum of two times.
func (rt RationalTime) Add(other RationalTime) RationalTime {
	// Handle zero-rate (invalid) times by returning the other time
	if rt.rate <= 0 {
		return other
	}
	if other.rate <= 0 {
		return rt
	}

	if rt.rate < other.rate {
		return RationalTime{
			value: rt.ValueRescaledTo(other.rate) + other.value,
			rate:  other.rate,
		}
	}
	return RationalTime{
		value: rt.value + other.ValueRescaledTo(rt.rate),
		rate:  rt.rate,
	}
}

// Sub returns the difference of two times.
func (rt RationalTime) Sub(other RationalTime) RationalTime {
	if rt.rate < other.rate {
		return RationalTime{
			value: rt.ValueRescaledTo(other.rate) - other.value,
			rate:  other.rate,
		}
	}
	return RationalTime{
		value: rt.value - other.ValueRescaledTo(rt.rate),
		rate:  rt.rate,
	}
}

// Neg returns the negation of this time.
func (rt RationalTime) Neg() RationalTime {
	return RationalTime{value: -rt.value, rate: rt.rate}
}

// Cmp compares two RationalTime values.
// Returns -1 if rt < other, 0 if rt == other, 1 if rt > other.
func (rt RationalTime) Cmp(other RationalTime) int {
	thisSeconds := rt.value / rt.rate
	otherSeconds := other.value / other.rate
	if thisSeconds < otherSeconds {
		return -1
	}
	if thisSeconds > otherSeconds {
		return 1
	}
	return 0
}

// Equal returns whether two times are equal (rescaled comparison).
func (rt RationalTime) Equal(other RationalTime) bool {
	return rt.ValueRescaledTo(other.rate) == other.value
}

// String returns a string representation of the RationalTime.
func (rt RationalTime) String() string {
	return fmt.Sprintf("RationalTime(%g, %g)", rt.value, rt.rate)
}
