// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package algorithms

import (
	"github.com/Avalanche-io/gotio/opentime"
	"github.com/Avalanche-io/gotio"
)

// Roll moves an edit point, adjusting both adjacent items.
// All affected items are modified in place.
//
// Behavior:
//   - deltaIn adjusts current item's head and previous item's tail
//   - deltaOut adjusts current item's tail and next item's head
//   - Edit point moves, but total duration is preserved
//
// Parameters:
//   - item: The item whose edit points to roll
//   - composition: The composition containing the item
//   - deltaIn: Amount to roll the in-point (positive = roll right)
//   - deltaOut: Amount to roll the out-point (positive = roll right)
func Roll(
	item gotio.Item,
	composition gotio.Composition,
	deltaIn opentime.RationalTime,
	deltaOut opentime.RationalTime,
) error {
	if deltaIn.Value() == 0 && deltaOut.Value() == 0 {
		return nil
	}

	// Find item's index
	itemIndex, err := composition.IndexOfChild(item)
	if err != nil {
		return newEditErrorForItem("roll", "item not in composition", item)
	}

	// Get current source range
	sourceRange, err := itemSourceRange(item)
	if err != nil {
		return err
	}

	// Handle deltaIn (roll in-point with previous item)
	if deltaIn.Value() != 0 {
		if err := rollInPoint(item, composition, itemIndex, sourceRange, deltaIn); err != nil {
			return err
		}
		// Update source range for deltaOut processing
		if sr := item.SourceRange(); sr != nil {
			sourceRange = *sr
		}
	}

	// Handle deltaOut (roll out-point with next item)
	if deltaOut.Value() != 0 {
		if err := rollOutPoint(item, composition, itemIndex, sourceRange, deltaOut); err != nil {
			return err
		}
	}

	return nil
}

// rollInPoint rolls the in-point between this item and the previous item.
func rollInPoint(
	item gotio.Item,
	composition gotio.Composition,
	itemIndex int,
	sourceRange opentime.TimeRange,
	deltaIn opentime.RationalTime,
) error {
	prevItem := getPreviousItem(composition, itemIndex)
	if prevItem == nil {
		// No previous item - can only roll if we're trimming head (positive delta)
		if deltaIn.Value() > 0 {
			// Trim head
			newStart := sourceRange.StartTime().Add(deltaIn)
			newDuration := sourceRange.Duration().Sub(deltaIn)
			if newDuration.Value() <= 0 {
				return ErrNegativeDuration
			}
			newRange := opentime.NewTimeRange(newStart, newDuration)
			item.SetSourceRange(&newRange)
		}
		return nil
	}

	// Get previous item's range
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

	// Clamp deltaIn based on constraints
	effectiveDelta := deltaIn

	// Can't roll left more than our start allows (from available range)
	if deltaIn.Value() < 0 {
		availRange, err := item.AvailableRange()
		if err == nil {
			minStart := availRange.StartTime()
			newStart := sourceRange.StartTime().Add(deltaIn)
			if newStart.Cmp(minStart) < 0 {
				effectiveDelta = minStart.Sub(sourceRange.StartTime())
			}
		}
	}

	// Can't roll right more than previous item's duration
	if effectiveDelta.Value() > 0 {
		if prevRange.Duration().Cmp(effectiveDelta) < 0 {
			effectiveDelta = prevRange.Duration()
		}
	}

	// Update current item: source start shifts, duration changes inversely
	newStart := sourceRange.StartTime().Add(effectiveDelta)
	newDuration := sourceRange.Duration().Sub(effectiveDelta)
	if newDuration.Value() <= 0 {
		return ErrNegativeDuration
	}
	newRange := opentime.NewTimeRange(newStart, newDuration)
	item.SetSourceRange(&newRange)

	// Update previous item: duration changes
	newPrevDuration := prevRange.Duration().Add(effectiveDelta)
	if newPrevDuration.Value() < 0 {
		newPrevDuration = opentime.NewRationalTime(0, prevRange.Duration().Rate())
	}
	newPrevRange := opentime.NewTimeRange(prevRange.StartTime(), newPrevDuration)
	prevItem.SetSourceRange(&newPrevRange)

	return nil
}

// rollOutPoint rolls the out-point between this item and the next item.
func rollOutPoint(
	item gotio.Item,
	composition gotio.Composition,
	itemIndex int,
	sourceRange opentime.TimeRange,
	deltaOut opentime.RationalTime,
) error {
	nextItem := getNextItem(composition, itemIndex)
	if nextItem == nil {
		// No next item - can only roll if we're extending tail (positive delta)
		if deltaOut.Value() > 0 {
			// Clamp to available range
			availRange, err := item.AvailableRange()
			var newDuration opentime.RationalTime
			if err == nil {
				maxEnd := availRange.EndTimeExclusive()
				newEnd := sourceRange.EndTimeExclusive().Add(deltaOut)
				if newEnd.Cmp(maxEnd) > 0 {
					newEnd = maxEnd
				}
				newDuration = newEnd.Sub(sourceRange.StartTime())
			} else {
				newDuration = sourceRange.Duration().Add(deltaOut)
			}

			newRange := opentime.NewTimeRange(sourceRange.StartTime(), newDuration)
			item.SetSourceRange(&newRange)
		}
		return nil
	}

	// Get next item's range
	var nextRange opentime.TimeRange
	if sr := nextItem.SourceRange(); sr != nil {
		nextRange = *sr
	} else {
		ar, err := nextItem.AvailableRange()
		if err != nil {
			return err
		}
		nextRange = ar
	}

	// Clamp deltaOut based on constraints
	effectiveDelta := deltaOut

	// Can't roll right more than next item's duration
	if effectiveDelta.Value() > 0 {
		if nextRange.Duration().Cmp(effectiveDelta) < 0 {
			effectiveDelta = nextRange.Duration()
		}
	}

	// Can't roll left more than our duration
	if effectiveDelta.Value() < 0 {
		if sourceRange.Duration().Add(effectiveDelta).Value() <= 0 {
			effectiveDelta = sourceRange.Duration().Neg().Add(opentime.NewRationalTime(1, sourceRange.Duration().Rate()))
		}
	}

	// Update current item: duration changes
	newDuration := sourceRange.Duration().Add(effectiveDelta)
	if newDuration.Value() <= 0 {
		return ErrNegativeDuration
	}
	newRange := opentime.NewTimeRange(sourceRange.StartTime(), newDuration)
	item.SetSourceRange(&newRange)

	// Update next item: source start shifts, duration changes inversely
	newNextStart := nextRange.StartTime().Add(effectiveDelta)
	newNextDuration := nextRange.Duration().Sub(effectiveDelta)
	if newNextDuration.Value() < 0 {
		newNextDuration = opentime.NewRationalTime(0, nextRange.Duration().Rate())
	}
	newNextRange := opentime.NewTimeRange(newNextStart, newNextDuration)
	nextItem.SetSourceRange(&newNextRange)

	return nil
}
