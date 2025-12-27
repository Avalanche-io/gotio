// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package algorithms

import (
	"github.com/mrjoshuak/gotio/opentime"
	"github.com/mrjoshuak/gotio/opentimelineio"
)

// itemSourceRange returns the source range for an item.
// If the item has a source range set, it returns that.
// Otherwise it falls back to the item's available range.
func itemSourceRange(item opentimelineio.Item) (opentime.TimeRange, error) {
	if sr := item.SourceRange(); sr != nil {
		return *sr, nil
	}
	return item.AvailableRange()
}

// itemAtTime finds the item at a specific time in a composition.
// Returns the item, its index, and the item's range in the composition.
// Returns nil, -1, zero range if no item is found at the time.
func itemAtTime(comp opentimelineio.Composition, time opentime.RationalTime) (opentimelineio.Item, int, opentime.TimeRange, error) {
	children := comp.Children()
	for i, child := range children {
		childRange, err := comp.RangeOfChildAtIndex(i)
		if err != nil {
			continue
		}
		if childRange.Contains(time) {
			if item, ok := child.(opentimelineio.Item); ok {
				return item, i, childRange, nil
			}
		}
	}
	return nil, -1, opentime.TimeRange{}, nil
}

// itemsInRange finds all items that intersect a time range.
// Returns the items, their indices, and their ranges.
func itemsInRange(comp opentimelineio.Composition, timeRange opentime.TimeRange) ([]opentimelineio.Item, []int, []opentime.TimeRange, error) {
	var items []opentimelineio.Item
	var indices []int
	var ranges []opentime.TimeRange

	children := comp.Children()
	for i, child := range children {
		childRange, err := comp.RangeOfChildAtIndex(i)
		if err != nil {
			continue
		}
		if timeRange.Intersects(childRange, opentime.DefaultEpsilon) {
			if item, ok := child.(opentimelineio.Item); ok {
				items = append(items, item)
				indices = append(indices, i)
				ranges = append(ranges, childRange)
			}
		}
	}
	return items, indices, ranges, nil
}

// transitionsInRange finds all transitions that intersect a time range.
func transitionsInRange(comp opentimelineio.Composition, timeRange opentime.TimeRange) ([]*opentimelineio.Transition, []int, error) {
	var transitions []*opentimelineio.Transition
	var indices []int

	children := comp.Children()
	for i, child := range children {
		tr, ok := child.(*opentimelineio.Transition)
		if !ok {
			continue
		}
		childRange, err := comp.RangeOfChildAtIndex(i)
		if err != nil {
			continue
		}
		if timeRange.Intersects(childRange, opentime.DefaultEpsilon) {
			transitions = append(transitions, tr)
			indices = append(indices, i)
		}
	}
	return transitions, indices, nil
}

// removeTransitionsInRange removes all transitions that intersect a time range.
// Returns true if any transitions were removed.
func removeTransitionsInRange(comp opentimelineio.Composition, timeRange opentime.TimeRange) (bool, error) {
	transitions, indices, err := transitionsInRange(comp, timeRange)
	if err != nil {
		return false, err
	}
	if len(transitions) == 0 {
		return false, nil
	}

	// Remove in reverse order to maintain valid indices
	for i := len(indices) - 1; i >= 0; i-- {
		if err := comp.RemoveChild(indices[i]); err != nil {
			return false, err
		}
	}
	return true, nil
}

// splitItemAtTime splits an item at a specific time, creating two items.
// The time is in composition coordinates.
// Returns the two resulting items (before and after the split point).
// If the time is at the item's start or end, returns the original item and nil.
func splitItemAtTime(
	comp opentimelineio.Composition,
	item opentimelineio.Item,
	itemIndex int,
	itemRange opentime.TimeRange,
	splitTime opentime.RationalTime,
) (opentimelineio.Item, opentimelineio.Item, error) {
	// Check if split is at boundaries
	if splitTime.Cmp(itemRange.StartTime()) <= 0 {
		return nil, item, nil
	}
	if splitTime.Cmp(itemRange.EndTimeExclusive()) >= 0 {
		return item, nil, nil
	}

	// Calculate the offset into the item where we split
	offsetInItem := splitTime.Sub(itemRange.StartTime())

	// Get the item's source range
	sourceRange, err := itemSourceRange(item)
	if err != nil {
		return nil, nil, err
	}

	// Calculate split point in source coordinates
	sourceOffset := offsetInItem.RescaledTo(sourceRange.StartTime().Rate())

	// Create the first part (before split)
	firstPart := item.Clone().(opentimelineio.Item)
	firstRange := opentime.NewTimeRange(
		sourceRange.StartTime(),
		sourceOffset,
	)
	firstPart.SetSourceRange(&firstRange)

	// Create the second part (after split)
	secondPart := item.Clone().(opentimelineio.Item)
	secondStart := sourceRange.StartTime().Add(sourceOffset)
	secondDuration := sourceRange.Duration().Sub(sourceOffset)
	secondRange := opentime.NewTimeRange(secondStart, secondDuration)
	secondPart.SetSourceRange(&secondRange)

	return firstPart, secondPart, nil
}

// clampToAvailableRange clamps a source range to the item's available range.
// Returns the clamped range, or the original range if no available range exists.
func clampToAvailableRange(item opentimelineio.Item, sourceRange opentime.TimeRange) opentime.TimeRange {
	availableRange, err := item.AvailableRange()
	if err != nil {
		return sourceRange
	}

	start := sourceRange.StartTime()
	end := sourceRange.EndTimeExclusive()

	// Clamp start
	if start.Cmp(availableRange.StartTime()) < 0 {
		start = availableRange.StartTime()
	}

	// Clamp end
	if end.Cmp(availableRange.EndTimeExclusive()) > 0 {
		end = availableRange.EndTimeExclusive()
	}

	// Ensure start doesn't exceed end
	if start.Cmp(end) >= 0 {
		return opentime.NewTimeRange(start, opentime.NewRationalTime(0, start.Rate()))
	}

	return opentime.RangeFromStartEndTime(start, end)
}

// compositionDuration returns the duration of a composition.
func compositionDuration(comp opentimelineio.Composition) (opentime.RationalTime, error) {
	children := comp.Children()
	if len(children) == 0 {
		return opentime.RationalTime{}, nil
	}

	var total opentime.RationalTime
	for _, child := range children {
		if !child.Visible() {
			continue
		}
		dur, err := child.Duration()
		if err != nil {
			return opentime.RationalTime{}, err
		}
		if total.Rate() <= 0 {
			total = dur
		} else {
			total = total.Add(dur)
		}
	}
	return total, nil
}

// compositionEndTime returns the end time of a composition.
func compositionEndTime(comp opentimelineio.Composition) (opentime.RationalTime, error) {
	dur, err := compositionDuration(comp)
	if err != nil {
		return opentime.RationalTime{}, err
	}
	return dur, nil
}

// createFillGap creates a gap with the specified duration.
// If fillTemplate is provided and is a Gap, it is cloned and its duration is set.
// Otherwise, a new Gap is created.
func createFillGap(duration opentime.RationalTime, fillTemplate opentimelineio.Item) *opentimelineio.Gap {
	if fillTemplate != nil {
		if gap, ok := fillTemplate.(*opentimelineio.Gap); ok {
			cloned := gap.Clone().(*opentimelineio.Gap)
			sr := opentime.NewTimeRange(opentime.RationalTime{}, duration)
			cloned.SetSourceRange(&sr)
			return cloned
		}
	}
	return opentimelineio.NewGapWithDuration(duration)
}

// maxRationalTime returns the maximum of two RationalTimes.
func maxRationalTime(a, b opentime.RationalTime) opentime.RationalTime {
	if a.Cmp(b) > 0 {
		return a
	}
	return b
}

// minRationalTime returns the minimum of two RationalTimes.
func minRationalTime(a, b opentime.RationalTime) opentime.RationalTime {
	if a.Cmp(b) < 0 {
		return a
	}
	return b
}

// isZeroOrNegative checks if a RationalTime is zero or negative.
func isZeroOrNegative(t opentime.RationalTime) bool {
	return t.Value() <= 0
}

// isPositive checks if a RationalTime is positive.
func isPositive(t opentime.RationalTime) bool {
	return t.Value() > 0
}

// getPreviousItem returns the item before the given index, or nil if none exists.
func getPreviousItem(comp opentimelineio.Composition, index int) opentimelineio.Item {
	if index <= 0 {
		return nil
	}
	children := comp.Children()
	if index > len(children) {
		return nil
	}
	if item, ok := children[index-1].(opentimelineio.Item); ok {
		return item
	}
	return nil
}

// getNextItem returns the item after the given index, or nil if none exists.
func getNextItem(comp opentimelineio.Composition, index int) opentimelineio.Item {
	children := comp.Children()
	if index < 0 || index >= len(children)-1 {
		return nil
	}
	if item, ok := children[index+1].(opentimelineio.Item); ok {
		return item
	}
	return nil
}

// adjustItemDuration adjusts an item's duration by a delta.
// Returns an error if the result would be negative.
func adjustItemDuration(item opentimelineio.Item, delta opentime.RationalTime) error {
	var sourceRange opentime.TimeRange
	if sr := item.SourceRange(); sr != nil {
		sourceRange = *sr
	} else {
		ar, err := item.AvailableRange()
		if err != nil {
			return err
		}
		sourceRange = ar
	}

	newDuration := sourceRange.Duration().Add(delta)
	if isZeroOrNegative(newDuration) {
		return ErrNegativeDuration
	}

	newRange := opentime.NewTimeRange(sourceRange.StartTime(), newDuration)
	item.SetSourceRange(&newRange)
	return nil
}

// adjustItemStartTime adjusts an item's source start time by a delta.
// This changes which part of the source media is shown.
func adjustItemStartTime(item opentimelineio.Item, delta opentime.RationalTime) error {
	var sourceRange opentime.TimeRange
	if sr := item.SourceRange(); sr != nil {
		sourceRange = *sr
	} else {
		ar, err := item.AvailableRange()
		if err != nil {
			return err
		}
		sourceRange = ar
	}

	newStart := sourceRange.StartTime().Add(delta)
	newDuration := sourceRange.Duration().Sub(delta)

	if isZeroOrNegative(newDuration) {
		return ErrNegativeDuration
	}

	newRange := opentime.NewTimeRange(newStart, newDuration)
	item.SetSourceRange(&newRange)
	return nil
}
