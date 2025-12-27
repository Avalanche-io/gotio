// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package algorithms

import (
	"github.com/mrjoshuak/gotio/opentime"
	"github.com/mrjoshuak/gotio/opentimelineio"
)

// InsertConfig holds configuration for the Insert operation.
type InsertConfig struct {
	RemoveTransitions bool
	FillTemplate      opentimelineio.Item
}

// InsertOption is a functional option for Insert.
type InsertOption func(*InsertConfig)

// WithInsertRemoveTransitions sets whether to remove transitions at the insert point.
func WithInsertRemoveTransitions(remove bool) InsertOption {
	return func(c *InsertConfig) {
		c.RemoveTransitions = remove
	}
}

// WithInsertFillTemplate sets the template item to use for filling gaps.
func WithInsertFillTemplate(template opentimelineio.Item) InsertOption {
	return func(c *InsertConfig) {
		c.FillTemplate = template
	}
}

// Insert inserts an item at a specific time, growing the composition.
// The composition is modified in place.
//
// Behavior:
//   - If time >= composition end: appends (with gap fill if needed)
//   - If time <= 0: prepends
//   - Otherwise: splits item at time, inserts between halves
//
// Parameters:
//   - item: The item to insert (will be cloned)
//   - composition: The composition to modify (usually a Track)
//   - time: The time at which to insert
//   - opts: Optional configuration
func Insert(
	item opentimelineio.Item,
	composition opentimelineio.Composition,
	time opentime.RationalTime,
	opts ...InsertOption,
) error {
	// Apply options
	config := &InsertConfig{
		RemoveTransitions: true,
	}
	for _, opt := range opts {
		opt(config)
	}

	// Clone the item
	clonedItem := item.Clone().(opentimelineio.Item)

	// Get composition duration
	compDuration, err := compositionDuration(composition)
	if err != nil {
		return err
	}

	// Handle empty composition
	if len(composition.Children()) == 0 || compDuration.Value() == 0 {
		// If time > 0, create a gap first
		if time.Value() > 0 {
			gap := createFillGap(time, config.FillTemplate)
			if err := composition.AppendChild(gap); err != nil {
				return err
			}
		}
		return composition.AppendChild(clonedItem)
	}

	zero := opentime.NewRationalTime(0, time.Rate())

	// Case 1: Prepend (time <= 0)
	if time.Cmp(zero) <= 0 {
		return composition.InsertChild(0, clonedItem)
	}

	// Case 2: Append (time >= end)
	if time.Cmp(compDuration) >= 0 {
		// Create fill gap if needed
		gapDuration := time.Sub(compDuration)
		if gapDuration.Value() > 0 {
			gap := createFillGap(gapDuration, config.FillTemplate)
			if err := composition.AppendChild(gap); err != nil {
				return err
			}
		}
		return composition.AppendChild(clonedItem)
	}

	// Case 3: Insert in middle - need to split an item
	// Remove transitions at insert point if requested
	if config.RemoveTransitions {
		insertRange := opentime.NewTimeRange(time, opentime.NewRationalTime(0, time.Rate()))
		removeTransitionsInRange(composition, insertRange)
	}

	// Find the item at the insert time
	itemAtInsert, itemIndex, itemRange, err := itemAtTime(composition, time)
	if err != nil {
		return err
	}

	if itemAtInsert == nil {
		// No item at time - just append
		return composition.AppendChild(clonedItem)
	}

	// Check if insert is exactly at item boundary
	if time.Equal(itemRange.StartTime()) {
		// Insert before this item
		return composition.InsertChild(itemIndex, clonedItem)
	}

	// Split the item
	firstPart, secondPart, err := splitItemAtTime(composition, itemAtInsert, itemIndex, itemRange, time)
	if err != nil {
		return err
	}

	// Remove the original item
	if err := composition.RemoveChild(itemIndex); err != nil {
		return err
	}

	// Insert: firstPart, newItem, secondPart
	insertIdx := itemIndex

	if firstPart != nil {
		if err := composition.InsertChild(insertIdx, firstPart); err != nil {
			return err
		}
		insertIdx++
	}

	if err := composition.InsertChild(insertIdx, clonedItem); err != nil {
		return err
	}
	insertIdx++

	if secondPart != nil {
		if err := composition.InsertChild(insertIdx, secondPart); err != nil {
			return err
		}
	}

	return nil
}
