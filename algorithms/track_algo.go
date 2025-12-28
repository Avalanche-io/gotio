// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

// Package algorithms provides algorithm implementations for OpenTimelineIO.
// This includes track trimming, stack flattening, timeline operations, and
// composition filtering.
package algorithms

import (
	"github.com/Avalanche-io/gotio/opentime"
	"github.com/Avalanche-io/gotio"
)

// TrackTrimmedToRange returns a new track trimmed to the given time range.
// Items outside the range are removed, items on the ends are trimmed.
// This never expands the track, only shortens it.
func TrackTrimmedToRange(track *gotio.Track, trimRange opentime.TimeRange) (*gotio.Track, error) {
	// Clone the track to not modify the original
	cloned := track.Clone().(*gotio.Track)

	// Calculate what to keep
	var newChildren []gotio.Composable

	runningOffset := opentime.NewRationalTime(0, trimRange.StartTime().Rate())

	for i, child := range cloned.Children() {
		childRange, err := cloned.RangeOfChildAtIndex(i)
		if err != nil {
			continue
		}

		// Check if this child overlaps with the trim range
		if !childRange.Intersects(trimRange, opentime.DefaultEpsilon) {
			continue
		}

		// Calculate the intersection
		intersection := intersectRanges(childRange, trimRange)

		// Get the child's trimmed range
		item, isItem := child.(gotio.Item)
		if !isItem {
			continue
		}

		// Calculate offset from original child start
		offsetFromChildStart := intersection.StartTime().Sub(childRange.StartTime())

		// Get existing source range or create one
		var itemSourceRange opentime.TimeRange
		if sr := item.SourceRange(); sr != nil {
			itemSourceRange = *sr
		} else {
			ar, err := item.AvailableRange()
			if err != nil {
				continue
			}
			itemSourceRange = ar
		}

		// Calculate new source range
		newSourceStart := itemSourceRange.StartTime().Add(offsetFromChildStart.RescaledTo(itemSourceRange.StartTime().Rate()))
		newSourceDuration := intersection.Duration().RescaledTo(itemSourceRange.Duration().Rate())
		newSourceRange := opentime.NewTimeRange(newSourceStart, newSourceDuration)

		// Clone the child and set the new source range
		clonedChild := child.Clone().(gotio.Composable)
		if clonedItem, ok := clonedChild.(gotio.Item); ok {
			clonedItem.SetSourceRange(&newSourceRange)
		}

		newChildren = append(newChildren, clonedChild)
		runningOffset = runningOffset.Add(intersection.Duration())
	}

	// Create a new track with the trimmed children
	result := gotio.NewTrack(
		cloned.Name(),
		cloned.SourceRange(),
		cloned.Kind(),
		gotio.CloneAnyDictionary(cloned.Metadata()),
		nil,
	)

	for _, child := range newChildren {
		result.AppendChild(child)
	}

	return result, nil
}

// intersectRanges returns the intersection of two time ranges.
func intersectRanges(a, b opentime.TimeRange) opentime.TimeRange {
	// Find the later start time
	startA := a.StartTime()
	startB := b.StartTime()
	var start opentime.RationalTime
	if startA.Cmp(startB) > 0 {
		start = startA
	} else {
		start = startB
	}

	// Find the earlier end time
	endA := a.EndTimeExclusive()
	endB := b.EndTimeExclusive()
	var end opentime.RationalTime
	if endA.Cmp(endB) < 0 {
		end = endA
	} else {
		end = endB
	}

	return opentime.RangeFromStartEndTime(start, end)
}

// TrackWithExpandedTransitions returns a new track where transitions are
// expanded to show the overlapping portions of neighboring clips.
// For example, [Clip1, T, Clip2] becomes [Clip1', (Clip1_t, T, Clip2_t), Clip2']
// where the clips adjacent to the transition are trimmed and the overlapping
// portions are placed alongside the transition.
func TrackWithExpandedTransitions(track *gotio.Track) (*gotio.Track, error) {
	// Clone the track
	cloned := track.Clone().(*gotio.Track)

	children := cloned.Children()
	if len(children) == 0 {
		return cloned, nil
	}

	// Find transitions and expand them
	var newChildren []gotio.Composable

	for i, child := range children {
		transition, isTransition := child.(*gotio.Transition)
		if !isTransition {
			newChildren = append(newChildren, child.Clone().(gotio.Composable))
			continue
		}

		// Get adjacent clips
		var prevItem, nextItem gotio.Item
		if i > 0 {
			if item, ok := children[i-1].(gotio.Item); ok {
				prevItem = item
			}
		}
		if i < len(children)-1 {
			if item, ok := children[i+1].(gotio.Item); ok {
				nextItem = item
			}
		}

		// Expand transition
		expanded := expandTransition(transition, prevItem, nextItem)
		newChildren = append(newChildren, expanded...)
	}

	// Create result track
	result := gotio.NewTrack(
		cloned.Name(),
		cloned.SourceRange(),
		cloned.Kind(),
		gotio.CloneAnyDictionary(cloned.Metadata()),
		nil,
	)

	for _, child := range newChildren {
		result.AppendChild(child)
	}

	return result, nil
}

// expandTransition expands a transition into its overlapping clip portions.
func expandTransition(transition *gotio.Transition, prevItem, nextItem gotio.Item) []gotio.Composable {
	var result []gotio.Composable

	// Get transition durations
	inOffset := transition.InOffset()
	outOffset := transition.OutOffset()

	// Create the overlapping portion from previous item (if any)
	if prevItem != nil {
		// Clone and trim the previous item to show only the overlap portion
		clonedPrev := prevItem.Clone().(gotio.Item)
		if sr := clonedPrev.SourceRange(); sr != nil {
			// Trim to just the out-going portion
			trimStart := sr.EndTimeExclusive().Sub(inOffset)
			trimRange := opentime.NewTimeRange(trimStart, inOffset)
			clonedPrev.SetSourceRange(&trimRange)
		}
		result = append(result, clonedPrev.(gotio.Composable))
	}

	// Add the transition itself
	result = append(result, transition.Clone().(gotio.Composable))

	// Create the overlapping portion from next item (if any)
	if nextItem != nil {
		// Clone and trim the next item to show only the overlap portion
		clonedNext := nextItem.Clone().(gotio.Item)
		if sr := clonedNext.SourceRange(); sr != nil {
			// Trim to just the in-coming portion
			trimRange := opentime.NewTimeRange(sr.StartTime(), outOffset)
			clonedNext.SetSourceRange(&trimRange)
		}
		result = append(result, clonedNext.(gotio.Composable))
	}

	return result
}
