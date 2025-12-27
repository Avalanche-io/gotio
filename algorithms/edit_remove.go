// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package algorithms

import (
	"github.com/mrjoshuak/gotio/opentime"
	"github.com/mrjoshuak/gotio/opentimelineio"
)

// RemoveConfig holds configuration for the Remove operation.
type RemoveConfig struct {
	Fill         bool
	FillTemplate opentimelineio.Item
}

// RemoveOption is a functional option for Remove.
type RemoveOption func(*RemoveConfig)

// WithFill sets whether to fill the removed space with a gap.
func WithFill(fill bool) RemoveOption {
	return func(c *RemoveConfig) {
		c.Fill = fill
	}
}

// WithRemoveFillTemplate sets the template item to use for filling.
func WithRemoveFillTemplate(template opentimelineio.Item) RemoveOption {
	return func(c *RemoveConfig) {
		c.FillTemplate = template
	}
}

// Remove removes an item at a specific time and optionally fills the space.
// The composition is modified in place.
//
// Behavior:
//   - If fill=true: inserts a gap (or template) in place of removed item
//   - If fill=false: adjacent items become adjacent (composition shrinks)
//
// Parameters:
//   - composition: The composition to modify (usually a Track)
//   - time: Time where the item to remove exists
//   - opts: Optional configuration (fill, template)
func Remove(
	composition opentimelineio.Composition,
	time opentime.RationalTime,
	opts ...RemoveOption,
) error {
	// Apply options
	config := &RemoveConfig{
		Fill: true,
	}
	for _, opt := range opts {
		opt(config)
	}

	// Find the item at time
	item, itemIndex, _, err := itemAtTime(composition, time)
	if err != nil {
		return err
	}

	if item == nil {
		return newEditErrorAt("remove", "no item at time", time)
	}

	// Get item's duration for filling
	itemDuration, err := item.Duration()
	if err != nil {
		return err
	}

	// Remove the item
	if err := composition.RemoveChild(itemIndex); err != nil {
		return err
	}

	// Fill if requested
	if config.Fill {
		gap := createFillGap(itemDuration, config.FillTemplate)
		return composition.InsertChild(itemIndex, gap)
	}

	return nil
}

// RemoveRange removes all items within a time range.
// The composition is modified in place.
//
// Behavior:
//   - Items completely within range are removed
//   - Items partially within range are trimmed
//   - If fill=true: a single gap fills the removed space
//   - If fill=false: composition shrinks
//
// Parameters:
//   - composition: The composition to modify (usually a Track)
//   - timeRange: Range of items to remove
//   - opts: Optional configuration (fill, template)
func RemoveRange(
	composition opentimelineio.Composition,
	timeRange opentime.TimeRange,
	opts ...RemoveOption,
) error {
	// Apply options
	config := &RemoveConfig{
		Fill: true,
	}
	for _, opt := range opts {
		opt(config)
	}

	// Find all items in range
	items, indices, ranges, err := itemsInRange(composition, timeRange)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		return nil
	}

	rangeStart := timeRange.StartTime()
	rangeEnd := timeRange.EndTimeExclusive()
	firstRange := ranges[0]
	lastRange := ranges[len(ranges)-1]

	// Determine what parts to keep
	var leadingPart, trailingPart opentimelineio.Item

	// Check if we need to keep the beginning of the first item
	if rangeStart.Cmp(firstRange.StartTime()) > 0 {
		firstPart, _, splitErr := splitItemAtTime(composition, items[0], indices[0], firstRange, rangeStart)
		if splitErr != nil {
			return splitErr
		}
		leadingPart = firstPart
	}

	// Check if we need to keep the end of the last item
	if rangeEnd.Cmp(lastRange.EndTimeExclusive()) < 0 {
		_, secondPart, splitErr := splitItemAtTime(composition, items[len(items)-1], indices[len(indices)-1], lastRange, rangeEnd)
		if splitErr != nil {
			return splitErr
		}
		trailingPart = secondPart
	}

	// Remove all items in range (reverse order)
	for i := len(indices) - 1; i >= 0; i-- {
		if err := composition.RemoveChild(indices[i]); err != nil {
			return err
		}
	}

	// Determine insert position
	insertIndex := indices[0]

	// Insert leading part if any
	if leadingPart != nil {
		if err := composition.InsertChild(insertIndex, leadingPart); err != nil {
			return err
		}
		insertIndex++
	}

	// Insert fill gap if requested
	if config.Fill {
		gap := createFillGap(timeRange.Duration(), config.FillTemplate)
		if err := composition.InsertChild(insertIndex, gap); err != nil {
			return err
		}
		insertIndex++
	}

	// Insert trailing part if any
	if trailingPart != nil {
		if err := composition.InsertChild(insertIndex, trailingPart); err != nil {
			return err
		}
	}

	return nil
}
