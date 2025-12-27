// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package algorithms

import (
	"github.com/mrjoshuak/gotio/opentime"
	"github.com/mrjoshuak/gotio/opentimelineio"
)

// Ripple adjusts an item's source range with clamping to available media.
// Unlike Trim, Ripple does not affect adjacent items.
// The item is modified in place.
//
// Behavior:
//   - deltaIn adjusts source start with clamping
//   - deltaOut adjusts source end with clamping
//   - Available range bounds are checked
//   - No effect on adjacent items
//
// Parameters:
//   - item: The item to adjust
//   - deltaIn: Adjustment to source_range start
//   - deltaOut: Adjustment to source_range end (duration change)
func Ripple(
	item opentimelineio.Item,
	deltaIn opentime.RationalTime,
	deltaOut opentime.RationalTime,
) error {
	if deltaIn.Value() == 0 && deltaOut.Value() == 0 {
		return nil
	}

	// Get current source range
	sourceRange, err := itemSourceRange(item)
	if err != nil {
		return err
	}

	start := sourceRange.StartTime()
	end := sourceRange.EndTimeExclusive()

	// Get available range for clamping
	availRange, hasAvail := item.AvailableRange()
	if hasAvail != nil {
		// No available range - just apply deltas with basic validation
		availRange = opentime.TimeRange{}
	}

	// Apply deltaIn to start
	if deltaIn.Value() != 0 {
		newStart := start.Add(deltaIn)

		// Clamp to available range start
		if hasAvail == nil && newStart.Cmp(availRange.StartTime()) < 0 {
			newStart = availRange.StartTime()
		}

		// Ensure start doesn't exceed end
		if newStart.Cmp(end) >= 0 {
			newStart = end.Sub(opentime.NewRationalTime(1, end.Rate()))
		}

		start = newStart
	}

	// Apply deltaOut to end
	if deltaOut.Value() != 0 {
		newEnd := end.Add(deltaOut)

		// Clamp to available range end
		if hasAvail == nil && newEnd.Cmp(availRange.EndTimeExclusive()) > 0 {
			newEnd = availRange.EndTimeExclusive()
		}

		// Ensure end doesn't precede start
		if newEnd.Cmp(start) <= 0 {
			newEnd = start.Add(opentime.NewRationalTime(1, start.Rate()))
		}

		end = newEnd
	}

	// Calculate new range
	duration := end.Sub(start)
	if duration.Value() <= 0 {
		return ErrNegativeDuration
	}

	newRange := opentime.NewTimeRange(start, duration)
	item.SetSourceRange(&newRange)

	return nil
}
