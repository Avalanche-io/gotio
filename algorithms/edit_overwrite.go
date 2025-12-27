// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package algorithms

import (
	"github.com/Avalanche-io/gotio/opentime"
	"github.com/Avalanche-io/gotio/opentimelineio"
)

// OverwriteConfig holds configuration for the Overwrite operation.
type OverwriteConfig struct {
	RemoveTransitions bool
	FillTemplate      opentimelineio.Item
}

// OverwriteOption is a functional option for Overwrite.
type OverwriteOption func(*OverwriteConfig)

// WithRemoveTransitions sets whether to remove transitions that intersect the range.
func WithRemoveTransitions(remove bool) OverwriteOption {
	return func(c *OverwriteConfig) {
		c.RemoveTransitions = remove
	}
}

// WithFillTemplate sets the template item to use for filling gaps.
func WithFillTemplate(template opentimelineio.Item) OverwriteOption {
	return func(c *OverwriteConfig) {
		c.FillTemplate = template
	}
}

// Overwrite replaces content in a time range with a new item.
// The composition is modified in place.
//
// Behavior:
//   - If range starts after composition end: creates gap, appends item
//   - If range ends before composition start: inserts item at beginning
//   - Otherwise: splits items at boundaries, removes items in range, inserts new item
//
// Parameters:
//   - item: The item to insert (will be cloned)
//   - composition: The composition to modify (usually a Track)
//   - timeRange: The time range to overwrite
//   - opts: Optional configuration (remove transitions, fill template)
func Overwrite(
	item opentimelineio.Item,
	composition opentimelineio.Composition,
	timeRange opentime.TimeRange,
	opts ...OverwriteOption,
) error {
	// Apply options
	config := &OverwriteConfig{
		RemoveTransitions: true,
	}
	for _, opt := range opts {
		opt(config)
	}

	// Clone the item to avoid modifying the original
	clonedItem := item.Clone().(opentimelineio.Item)

	// Set the item's source range to match the overwrite duration if needed
	if sr := clonedItem.SourceRange(); sr == nil {
		ar, err := clonedItem.AvailableRange()
		if err == nil {
			// Trim to overwrite duration
			newRange := opentime.NewTimeRange(
				ar.StartTime(),
				timeRange.Duration(),
			)
			clonedItem.SetSourceRange(&newRange)
		}
	}

	// Get composition duration
	compDuration, err := compositionDuration(composition)
	if err != nil {
		return err
	}

	// Handle empty composition
	if len(composition.Children()) == 0 || compDuration.Value() == 0 {
		return handleEmptyComposition(clonedItem, composition, timeRange, config)
	}

	// Remove transitions if requested
	if config.RemoveTransitions {
		removeTransitionsInRange(composition, timeRange)
	}

	// Recalculate duration after removing transitions
	compDuration, _ = compositionDuration(composition)
	compEnd := compDuration

	rangeStart := timeRange.StartTime()
	rangeEnd := timeRange.EndTimeExclusive()

	// Case 1: Append after end
	if rangeStart.Cmp(compEnd) >= 0 {
		return handleAppendAfterEnd(clonedItem, composition, timeRange, compEnd, config)
	}

	// Case 2: Insert before start
	if rangeEnd.Cmp(opentime.NewRationalTime(0, rangeEnd.Rate())) <= 0 {
		return handleInsertBeforeStart(clonedItem, composition, timeRange, config)
	}

	// Case 3: Overwrite within composition
	return handleOverwriteMiddle(clonedItem, composition, timeRange, config)
}

// handleEmptyComposition handles overwrite on an empty composition.
func handleEmptyComposition(
	item opentimelineio.Item,
	comp opentimelineio.Composition,
	timeRange opentime.TimeRange,
	config *OverwriteConfig,
) error {
	rangeStart := timeRange.StartTime()

	// If range doesn't start at 0, create a fill gap
	if rangeStart.Value() > 0 {
		gap := createFillGap(rangeStart, config.FillTemplate)
		if err := comp.AppendChild(gap); err != nil {
			return err
		}
	}

	return comp.AppendChild(item)
}

// handleAppendAfterEnd handles overwrite that starts after composition end.
func handleAppendAfterEnd(
	item opentimelineio.Item,
	comp opentimelineio.Composition,
	timeRange opentime.TimeRange,
	compEnd opentime.RationalTime,
	config *OverwriteConfig,
) error {
	rangeStart := timeRange.StartTime()

	// Create fill gap between composition end and range start
	gapDuration := rangeStart.Sub(compEnd)
	if gapDuration.Value() > 0 {
		gap := createFillGap(gapDuration, config.FillTemplate)
		if err := comp.AppendChild(gap); err != nil {
			return err
		}
	}

	return comp.AppendChild(item)
}

// handleInsertBeforeStart handles overwrite that ends before composition start.
func handleInsertBeforeStart(
	item opentimelineio.Item,
	comp opentimelineio.Composition,
	timeRange opentime.TimeRange,
	config *OverwriteConfig,
) error {
	rangeEnd := timeRange.EndTimeExclusive()
	zero := opentime.NewRationalTime(0, rangeEnd.Rate())

	// Create fill gap between range end and composition start
	gapDuration := zero.Sub(rangeEnd)
	if gapDuration.Value() > 0 {
		if err := comp.InsertChild(0, item); err != nil {
			return err
		}
		gap := createFillGap(gapDuration, config.FillTemplate)
		return comp.InsertChild(1, gap)
	}

	return comp.InsertChild(0, item)
}

// handleOverwriteMiddle handles overwrite within the composition.
func handleOverwriteMiddle(
	item opentimelineio.Item,
	comp opentimelineio.Composition,
	timeRange opentime.TimeRange,
	config *OverwriteConfig,
) error {
	rangeStart := timeRange.StartTime()
	rangeEnd := timeRange.EndTimeExclusive()

	// Find all items in the range
	items, indices, ranges, err := itemsInRange(comp, timeRange)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		// No items in range - this shouldn't happen if composition has items
		return comp.AppendChild(item)
	}

	firstIndex := indices[0]
	firstRange := ranges[0]
	lastRange := ranges[len(ranges)-1]

	// Determine what parts to keep
	var leadingPart, trailingPart opentimelineio.Item

	// Check if we need to keep the beginning of the first item
	if rangeStart.Cmp(firstRange.StartTime()) > 0 {
		firstPart, _, splitErr := splitItemAtTime(comp, items[0], indices[0], firstRange, rangeStart)
		if splitErr != nil {
			return splitErr
		}
		leadingPart = firstPart
	}

	// Check if we need to keep the end of the last item
	if rangeEnd.Cmp(lastRange.EndTimeExclusive()) < 0 {
		_, secondPart, splitErr := splitItemAtTime(comp, items[len(items)-1], indices[len(indices)-1], lastRange, rangeEnd)
		if splitErr != nil {
			return splitErr
		}
		trailingPart = secondPart
	}

	// Remove all items that intersect the range (in reverse order to maintain indices)
	for i := len(indices) - 1; i >= 0; i-- {
		if err := comp.RemoveChild(indices[i]); err != nil {
			return err
		}
	}

	// Determine insert index
	insertIndex := firstIndex

	// Insert leading part if we have one
	if leadingPart != nil {
		if err := comp.InsertChild(insertIndex, leadingPart); err != nil {
			return err
		}
		insertIndex++
	}

	// Insert the new item
	if err := comp.InsertChild(insertIndex, item); err != nil {
		return err
	}
	insertIndex++

	// Insert trailing part if we have one
	if trailingPart != nil {
		if err := comp.InsertChild(insertIndex, trailingPart); err != nil {
			return err
		}
	}

	return nil
}
