// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package algorithms

import (
	"github.com/Avalanche-io/gotio/opentime"
	"github.com/Avalanche-io/gotio"
)

// Fill places an item into a gap using 3/4-point edit logic.
// The composition is modified in place.
//
// Behavior depends on ReferencePoint:
//   - ReferencePointSource: Use clip's natural duration, overwrite from trackTime
//   - ReferencePointSequence: Trim clip to fit gap exactly
//   - ReferencePointFit: Add LinearTimeWarp effect to stretch/compress clip to gap
//
// Parameters:
//   - item: The item to place (will be cloned)
//   - composition: The composition to modify (usually a Track)
//   - trackTime: Time in the track where the gap exists
//   - referencePoint: How to fit the clip to the gap
func Fill(
	item gotio.Item,
	composition gotio.Composition,
	trackTime opentime.RationalTime,
	referencePoint ReferencePoint,
) error {
	// Find the item at trackTime
	gapItem, gapIndex, gapRange, err := itemAtTime(composition, trackTime)
	if err != nil {
		return err
	}

	if gapItem == nil {
		return newEditErrorAt("fill", "no item at time", trackTime)
	}

	// Verify it's a gap
	gap, isGap := gapItem.(*gotio.Gap)
	if !isGap {
		return newEditError("fill", "item at time is not a gap")
	}

	// Get gap duration
	gapDuration, err := gap.Duration()
	if err != nil {
		return err
	}

	// Clone the item
	clonedItem := item.Clone().(gotio.Item)

	// Get clip's source range
	var clipRange opentime.TimeRange
	if sr := clonedItem.SourceRange(); sr != nil {
		clipRange = *sr
	} else {
		ar, err := clonedItem.AvailableRange()
		if err != nil {
			return err
		}
		clipRange = ar
	}

	switch referencePoint {
	case ReferencePointSource:
		// Use clip's natural duration - perform an overwrite
		overwriteRange := opentime.NewTimeRange(gapRange.StartTime(), clipRange.Duration())
		return Overwrite(clonedItem, composition, overwriteRange)

	case ReferencePointSequence:
		// Trim clip to fit gap exactly
		return fillSequence(clonedItem, composition, gapIndex, gapRange, clipRange, gapDuration)

	case ReferencePointFit:
		// Add time warp to stretch/compress
		return fillFit(clonedItem, composition, gapIndex, gapRange, clipRange, gapDuration)

	default:
		return newEditError("fill", "unknown reference point")
	}
}

// fillSequence trims the clip to fit exactly in the gap.
func fillSequence(
	item gotio.Item,
	comp gotio.Composition,
	gapIndex int,
	gapRange opentime.TimeRange,
	clipRange opentime.TimeRange,
	gapDuration opentime.RationalTime,
) error {
	// Adjust clip to fit gap
	clipDuration := clipRange.Duration()

	if clipDuration.Cmp(gapDuration) > 0 {
		// Clip is longer than gap - trim it
		newRange := opentime.NewTimeRange(clipRange.StartTime(), gapDuration)
		item.SetSourceRange(&newRange)
	} else if clipDuration.Cmp(gapDuration) < 0 {
		// Clip is shorter than gap - will leave remaining gap
		// Set source range to clip's duration
		newRange := opentime.NewTimeRange(clipRange.StartTime(), clipDuration)
		item.SetSourceRange(&newRange)

		// Remove gap
		if err := comp.RemoveChild(gapIndex); err != nil {
			return err
		}

		// Insert clip
		if err := comp.InsertChild(gapIndex, item); err != nil {
			return err
		}

		// Create remaining gap
		remainingDuration := gapDuration.Sub(clipDuration)
		remainingGap := gotio.NewGapWithDuration(remainingDuration)
		return comp.InsertChild(gapIndex+1, remainingGap)
	} else {
		// Exact fit
		newRange := opentime.NewTimeRange(clipRange.StartTime(), gapDuration)
		item.SetSourceRange(&newRange)
	}

	// Remove gap and insert clip
	if err := comp.RemoveChild(gapIndex); err != nil {
		return err
	}
	return comp.InsertChild(gapIndex, item)
}

// fillFit adds a LinearTimeWarp to stretch/compress the clip to fit the gap.
func fillFit(
	item gotio.Item,
	comp gotio.Composition,
	gapIndex int,
	gapRange opentime.TimeRange,
	clipRange opentime.TimeRange,
	gapDuration opentime.RationalTime,
) error {
	// Calculate time scalar
	clipDuration := clipRange.Duration()
	if clipDuration.Value() == 0 {
		return newEditError("fill", "clip has zero duration")
	}

	// Time scalar = clip_duration / gap_duration
	// (higher scalar = faster playback to fit shorter gap)
	timeScalar := clipDuration.Value() / gapDuration.Value()

	// Create LinearTimeWarp effect
	timeWarp := gotio.NewLinearTimeWarp("time_fit", "LinearTimeWarp", timeScalar, nil)

	// Add effect to item
	effects := item.Effects()
	effects = append(effects, timeWarp)
	item.SetEffects(effects)

	// The source range stays the same (showing all media)
	// but the visible duration in the timeline matches the gap
	// Note: In OTIO, the effect affects playback, not the source_range
	// We need to adjust the source range to match what will be visible
	newRange := opentime.NewTimeRange(clipRange.StartTime(), clipRange.Duration())
	item.SetSourceRange(&newRange)

	// Remove gap and insert item
	if err := comp.RemoveChild(gapIndex); err != nil {
		return err
	}
	return comp.InsertChild(gapIndex, item)
}
