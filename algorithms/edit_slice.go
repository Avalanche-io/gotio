// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package algorithms

import (
	"github.com/mrjoshuak/gotio/opentime"
	"github.com/mrjoshuak/gotio/opentimelineio"
)

// SliceConfig holds configuration for the Slice operation.
type SliceConfig struct {
	RemoveTransitions bool
}

// SliceOption is a functional option for Slice.
type SliceOption func(*SliceConfig)

// WithSliceRemoveTransitions sets whether to remove transitions at the slice point.
func WithSliceRemoveTransitions(remove bool) SliceOption {
	return func(c *SliceConfig) {
		c.RemoveTransitions = remove
	}
}

// Slice cuts an item at a specific time, creating two items.
// The composition is modified in place.
//
// Behavior:
//   - If time is at item boundary: no-op
//   - If time is within an item: splits into two items with adjusted source ranges
//   - Does not change composition duration
//
// Parameters:
//   - composition: The composition to modify (usually a Track)
//   - time: The time at which to slice
//   - opts: Optional configuration
func Slice(
	composition opentimelineio.Composition,
	time opentime.RationalTime,
	opts ...SliceOption,
) error {
	// Apply options
	config := &SliceConfig{
		RemoveTransitions: true,
	}
	for _, opt := range opts {
		opt(config)
	}

	// Get composition duration
	compDuration, err := compositionDuration(composition)
	if err != nil {
		return err
	}

	zero := opentime.NewRationalTime(0, time.Rate())

	// Check if time is within composition bounds
	if time.Cmp(zero) <= 0 || time.Cmp(compDuration) >= 0 {
		// Slice at boundaries is a no-op
		return nil
	}

	// Check for transitions at the slice point
	if config.RemoveTransitions {
		sliceRange := opentime.NewTimeRange(time, opentime.NewRationalTime(0, time.Rate()))
		transitions, _, _ := transitionsInRange(composition, sliceRange)
		if len(transitions) > 0 {
			removeTransitionsInRange(composition, sliceRange)
		}
	} else {
		// Check if there's a transition at this point
		sliceRange := opentime.NewTimeRange(time, opentime.NewRationalTime(0, time.Rate()))
		transitions, _, _ := transitionsInRange(composition, sliceRange)
		if len(transitions) > 0 {
			return newEditError("slice", "cannot slice through a transition")
		}
	}

	// Find the item at the slice time
	item, itemIndex, itemRange, err := itemAtTime(composition, time)
	if err != nil {
		return err
	}

	if item == nil {
		// No item at time
		return nil
	}

	// Check if slice is exactly at item boundary
	if time.Equal(itemRange.StartTime()) || time.Equal(itemRange.EndTimeExclusive()) {
		// Slice at item boundary is a no-op
		return nil
	}

	// Split the item
	firstPart, secondPart, err := splitItemAtTime(composition, item, itemIndex, itemRange, time)
	if err != nil {
		return err
	}

	// Remove the original item
	if err := composition.RemoveChild(itemIndex); err != nil {
		return err
	}

	// Insert the two parts
	if firstPart != nil {
		if err := composition.InsertChild(itemIndex, firstPart); err != nil {
			return err
		}
		itemIndex++
	}

	if secondPart != nil {
		if err := composition.InsertChild(itemIndex, secondPart); err != nil {
			return err
		}
	}

	return nil
}
