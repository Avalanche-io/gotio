// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package algorithms

import (
	"github.com/mrjoshuak/gotio/opentime"
	"github.com/mrjoshuak/gotio/opentimelineio"
)

// Slide moves an item's position by adjusting the previous item's duration.
// Both items are modified in place.
//
// Behavior:
//   - Positive delta: expands previous item, moves this item right
//   - Negative delta: contracts previous item, moves this item left
//   - Clamps to prevent negative durations
//   - First item cannot be slid
//
// Parameters:
//   - item: The item to slide
//   - composition: The composition containing the item
//   - delta: Amount to slide (positive = right, negative = left)
func Slide(
	item opentimelineio.Item,
	composition opentimelineio.Composition,
	delta opentime.RationalTime,
) error {
	if delta.Value() == 0 {
		return nil
	}

	// Find the item's index
	itemIndex, err := composition.IndexOfChild(item)
	if err != nil {
		return newEditErrorForItem("slide", "item not in composition", item)
	}

	// Can't slide the first item
	if itemIndex == 0 {
		return nil
	}

	// Get the previous item
	prevItem := getPreviousItem(composition, itemIndex)
	if prevItem == nil {
		return nil
	}

	// Get previous item's source range
	var prevRange opentime.TimeRange
	if sr := prevItem.SourceRange(); sr != nil {
		prevRange = *sr
	} else {
		ar, err := prevItem.AvailableRange()
		if err != nil {
			return err
		}
		prevRange = ar
	}

	// Calculate new duration for previous item
	newPrevDuration := prevRange.Duration().Add(delta)

	// Clamp to prevent negative duration
	if newPrevDuration.Value() <= 0 {
		// Can only slide left by previous item's duration
		delta = prevRange.Duration().Neg()
		newPrevDuration = opentime.NewRationalTime(0, prevRange.Duration().Rate())
	}

	// Check available range of previous item for expansion
	if delta.Value() > 0 {
		prevAvail, err := prevItem.AvailableRange()
		if err == nil {
			maxDuration := prevAvail.Duration()
			if newPrevDuration.Cmp(maxDuration) > 0 {
				// Clamp to available
				delta = maxDuration.Sub(prevRange.Duration())
				newPrevDuration = maxDuration
			}
		}
	}

	// Update previous item's duration
	newPrevRange := opentime.NewTimeRange(prevRange.StartTime(), newPrevDuration)
	prevItem.SetSourceRange(&newPrevRange)

	return nil
}
