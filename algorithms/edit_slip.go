// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package algorithms

import (
	"github.com/Avalanche-io/gotio/opentime"
	"github.com/Avalanche-io/gotio"
)

// Slip moves an item's playhead through source media without changing position or duration.
// The item is modified in place.
//
// Behavior:
//   - Adds delta to source_range.start_time
//   - Clamps to available_range if present
//   - Duration and position in composition unchanged
//
// Parameters:
//   - item: The item to slip
//   - delta: Amount to move source start (positive = forward in source)
func Slip(item gotio.Item, delta opentime.RationalTime) error {
	if delta.Value() == 0 {
		return nil
	}

	// Get current source range
	sourceRange, err := itemSourceRange(item)
	if err != nil {
		return err
	}

	// Calculate new start time
	newStart := sourceRange.StartTime().Add(delta)
	duration := sourceRange.Duration()

	// Try to get available range for clamping
	availableRange, err := item.AvailableRange()
	if err == nil {
		// Clamp to available range
		availStart := availableRange.StartTime()
		availEnd := availableRange.EndTimeExclusive()

		// Clamp start
		if newStart.Cmp(availStart) < 0 {
			newStart = availStart
		}

		// Ensure end doesn't exceed available
		newEnd := newStart.Add(duration)
		if newEnd.Cmp(availEnd) > 0 {
			// Shift start back to fit
			newStart = availEnd.Sub(duration)
			if newStart.Cmp(availStart) < 0 {
				newStart = availStart
			}
		}
	}

	// Set the new source range (duration unchanged)
	newRange := opentime.NewTimeRange(newStart, duration)
	item.SetSourceRange(&newRange)

	return nil
}
