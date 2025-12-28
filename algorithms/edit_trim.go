// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package algorithms

import (
	"github.com/Avalanche-io/gotio/opentime"
	"github.com/Avalanche-io/gotio"
)

// TrimConfig holds configuration for the Trim operation.
type TrimConfig struct {
	FillTemplate gotio.Item
}

// TrimOption is a functional option for Trim.
type TrimOption func(*TrimConfig)

// WithTrimFillTemplate sets the template item to use for filling gaps.
func WithTrimFillTemplate(template gotio.Item) TrimOption {
	return func(c *TrimConfig) {
		c.FillTemplate = template
	}
}

// Trim adjusts an item's in/out points without affecting composition duration.
// Adjacent items are adjusted to compensate.
// The item and adjacent items are modified in place.
//
// Behavior:
//   - deltaIn > 0: moves source start forward, previous item expands
//   - deltaIn < 0: moves source start backward, previous item contracts
//   - deltaOut > 0: extends duration, next item contracts
//   - deltaOut < 0: reduces duration, next item expands
//
// Parameters:
//   - item: The item to trim
//   - composition: The composition containing the item
//   - deltaIn: Adjustment to source_range start (positive = trim head)
//   - deltaOut: Adjustment to source_range end (positive = extend tail)
//   - opts: Optional configuration
func Trim(
	item gotio.Item,
	composition gotio.Composition,
	deltaIn opentime.RationalTime,
	deltaOut opentime.RationalTime,
	opts ...TrimOption,
) error {
	// Apply options
	config := &TrimConfig{}
	for _, opt := range opts {
		opt(config)
	}

	if deltaIn.Value() == 0 && deltaOut.Value() == 0 {
		return nil
	}

	// Find item's index
	itemIndex, err := composition.IndexOfChild(item)
	if err != nil {
		return newEditErrorForItem("trim", "item not in composition", item)
	}

	// Get current source range
	sourceRange, err := itemSourceRange(item)
	if err != nil {
		return err
	}

	// Handle deltaIn (head trim)
	if deltaIn.Value() != 0 {
		if err := trimHead(item, composition, itemIndex, sourceRange, deltaIn, config); err != nil {
			return err
		}
		// Update source range for deltaOut processing
		if sr := item.SourceRange(); sr != nil {
			sourceRange = *sr
		}
	}

	// Handle deltaOut (tail trim)
	if deltaOut.Value() != 0 {
		if err := trimTail(item, composition, itemIndex, sourceRange, deltaOut, config); err != nil {
			return err
		}
	}

	return nil
}

// trimHead handles the head (in-point) trim.
func trimHead(
	item gotio.Item,
	composition gotio.Composition,
	itemIndex int,
	sourceRange opentime.TimeRange,
	deltaIn opentime.RationalTime,
	config *TrimConfig,
) error {
	// Calculate new source start and duration
	newStart := sourceRange.StartTime().Add(deltaIn)
	newDuration := sourceRange.Duration().Sub(deltaIn)

	// Ensure duration doesn't go negative
	if newDuration.Value() <= 0 {
		return ErrNegativeDuration
	}

	// Clamp to available range
	availRange, err := item.AvailableRange()
	if err == nil {
		if newStart.Cmp(availRange.StartTime()) < 0 {
			diff := availRange.StartTime().Sub(newStart)
			newStart = availRange.StartTime()
			newDuration = newDuration.Add(diff)
		}
	}

	// Update item's source range
	newRange := opentime.NewTimeRange(newStart, newDuration)
	item.SetSourceRange(&newRange)

	// Adjust previous item to compensate
	prevItem := getPreviousItem(composition, itemIndex)
	if prevItem != nil {
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

		// Previous item's duration changes by deltaIn
		newPrevDuration := prevRange.Duration().Add(deltaIn)
		if newPrevDuration.Value() < 0 {
			newPrevDuration = opentime.NewRationalTime(0, prevRange.Duration().Rate())
		}

		newPrevRange := opentime.NewTimeRange(prevRange.StartTime(), newPrevDuration)
		prevItem.SetSourceRange(&newPrevRange)
	} else if deltaIn.Value() < 0 {
		// No previous item and we're extending head - need to create a gap
		gapDuration := deltaIn.Neg()
		gap := createFillGap(gapDuration, config.FillTemplate)
		if err := composition.InsertChild(itemIndex, gap); err != nil {
			return err
		}
	}

	return nil
}

// trimTail handles the tail (out-point) trim.
func trimTail(
	item gotio.Item,
	composition gotio.Composition,
	itemIndex int,
	sourceRange opentime.TimeRange,
	deltaOut opentime.RationalTime,
	config *TrimConfig,
) error {
	// Calculate new duration
	newDuration := sourceRange.Duration().Add(deltaOut)

	// Ensure duration doesn't go negative
	if newDuration.Value() <= 0 {
		return ErrNegativeDuration
	}

	// Clamp to available range
	availRange, err := item.AvailableRange()
	if err == nil {
		maxDuration := availRange.EndTimeExclusive().Sub(sourceRange.StartTime())
		if newDuration.Cmp(maxDuration) > 0 {
			newDuration = maxDuration
		}
	}

	// Update item's source range
	newRange := opentime.NewTimeRange(sourceRange.StartTime(), newDuration)
	item.SetSourceRange(&newRange)

	// Adjust next item to compensate
	nextItem := getNextItem(composition, itemIndex)
	if nextItem != nil {
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

		// For positive deltaOut (extending), we need to trim next item's head
		// For negative deltaOut (contracting), we extend next item's head
		newNextStart := nextRange.StartTime().Add(deltaOut)
		newNextDuration := nextRange.Duration().Sub(deltaOut)

		// Handle next item becoming a gap
		if _, isGap := nextItem.(*gotio.Gap); isGap {
			if newNextDuration.Value() <= 0 {
				// Gap is eliminated - remove it
				// Note: Recalculate itemIndex since our item might have shifted
				composition.RemoveChild(itemIndex + 1)
				return nil
			}
		} else {
			if newNextDuration.Value() < 0 {
				newNextDuration = opentime.NewRationalTime(0, nextRange.Duration().Rate())
			}
		}

		newNextRange := opentime.NewTimeRange(newNextStart, newNextDuration)
		nextItem.SetSourceRange(&newNextRange)
	} else if deltaOut.Value() < 0 {
		// No next item and we're contracting - need to create a gap
		gapDuration := deltaOut.Neg()
		gap := createFillGap(gapDuration, config.FillTemplate)
		if err := composition.AppendChild(gap); err != nil {
			return err
		}
	}

	return nil
}
