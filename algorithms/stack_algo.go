// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package algorithms

import (
	"github.com/mrjoshuak/gotio/opentime"
	"github.com/mrjoshuak/gotio/opentimelineio"
)

// FlattenStack flattens a stack (multitrack composition) down to a single track.
// Tracks are composited in order (later tracks on top of earlier tracks).
// Overlapping segments are handled by trimming away overlaps from lower tracks.
func FlattenStack(stack *opentimelineio.Stack) (*opentimelineio.Track, error) {
	children := stack.Children()
	if len(children) == 0 {
		return opentimelineio.NewTrack("Flattened", nil, opentimelineio.TrackKindVideo, nil, nil), nil
	}

	// Get tracks from stack
	var tracks []*opentimelineio.Track
	for _, child := range children {
		if track, ok := child.(*opentimelineio.Track); ok {
			tracks = append(tracks, track)
		}
	}

	return FlattenTracks(tracks)
}

// FlattenTracks flattens multiple tracks down to a single track.
// Later tracks take priority over earlier tracks (later tracks are "on top").
func FlattenTracks(tracks []*opentimelineio.Track) (*opentimelineio.Track, error) {
	if len(tracks) == 0 {
		return opentimelineio.NewTrack("Flattened", nil, opentimelineio.TrackKindVideo, nil, nil), nil
	}

	if len(tracks) == 1 {
		return tracks[0].Clone().(*opentimelineio.Track), nil
	}

	// Start with the first track
	result := tracks[0].Clone().(*opentimelineio.Track)

	// For each subsequent track, composite it on top
	for i := 1; i < len(tracks); i++ {
		composited, err := compositeTrackOnTop(result, tracks[i])
		if err != nil {
			return nil, err
		}
		result = composited
	}

	result.SetName("Flattened")
	return result, nil
}

// compositeTrackOnTop composites the top track onto the base track.
// Items from the top track take priority over items from the base track.
func compositeTrackOnTop(base, top *opentimelineio.Track) (*opentimelineio.Track, error) {
	// Get time ranges for all items in the top track
	topRanges := make([]opentime.TimeRange, 0)
	topItems := make([]opentimelineio.Composable, 0)

	for i, child := range top.Children() {
		// Skip non-visible items
		if item, ok := child.(opentimelineio.Item); ok {
			if !item.Enabled() {
				continue
			}
		}

		childRange, err := top.RangeOfChildAtIndex(i)
		if err != nil {
			continue
		}

		// Skip transitions and gaps for flattening purposes
		if _, isTransition := child.(*opentimelineio.Transition); isTransition {
			continue
		}
		if _, isGap := child.(*opentimelineio.Gap); isGap {
			continue
		}

		topRanges = append(topRanges, childRange)
		topItems = append(topItems, child)
	}

	// Build result track with base items trimmed around top items
	result := opentimelineio.NewTrack(
		base.Name(),
		base.SourceRange(),
		base.Kind(),
		opentimelineio.CloneAnyDictionary(base.Metadata()),
		nil,
	)

	// Process base track items
	for i, child := range base.Children() {
		childRange, err := base.RangeOfChildAtIndex(i)
		if err != nil {
			continue
		}

		// Check for overlaps with top items
		remainingRanges := []opentime.TimeRange{childRange}
		for _, topRange := range topRanges {
			var newRemainingRanges []opentime.TimeRange
			for _, r := range remainingRanges {
				split := subtractRange(r, topRange)
				newRemainingRanges = append(newRemainingRanges, split...)
			}
			remainingRanges = newRemainingRanges
		}

		// Add portions of this item that aren't covered by top items
		for _, r := range remainingRanges {
			if r.Duration().Value() <= 0 {
				continue
			}

			// Clone and trim the item
			cloned := child.Clone().(opentimelineio.Composable)
			if item, ok := cloned.(opentimelineio.Item); ok {
				trimItemToRange(item, childRange, r)
			}
			result.AppendChild(cloned)
		}
	}

	// Add all top items
	for _, item := range topItems {
		result.AppendChild(item.Clone().(opentimelineio.Composable))
	}

	return result, nil
}

// subtractRange subtracts b from a, returning the remaining portions of a.
func subtractRange(a, b opentime.TimeRange) []opentime.TimeRange {
	// If no intersection, return a unchanged
	if !a.Intersects(b, opentime.DefaultEpsilon) {
		return []opentime.TimeRange{a}
	}

	var result []opentime.TimeRange

	// Portion before b
	if a.StartTime().Cmp(b.StartTime()) < 0 {
		beforeEnd := b.StartTime()
		if beforeEnd.Cmp(a.EndTimeExclusive()) > 0 {
			beforeEnd = a.EndTimeExclusive()
		}
		beforeRange := opentime.RangeFromStartEndTime(a.StartTime(), beforeEnd)
		if beforeRange.Duration().Value() > 0 {
			result = append(result, beforeRange)
		}
	}

	// Portion after b
	if a.EndTimeExclusive().Cmp(b.EndTimeExclusive()) > 0 {
		afterStart := b.EndTimeExclusive()
		if afterStart.Cmp(a.StartTime()) < 0 {
			afterStart = a.StartTime()
		}
		afterRange := opentime.RangeFromStartEndTime(afterStart, a.EndTimeExclusive())
		if afterRange.Duration().Value() > 0 {
			result = append(result, afterRange)
		}
	}

	return result
}

// trimItemToRange trims an item to a sub-range within its original range.
func trimItemToRange(item opentimelineio.Item, originalRange, newRange opentime.TimeRange) {
	var itemSourceRange opentime.TimeRange
	if sr := item.SourceRange(); sr != nil {
		itemSourceRange = *sr
	} else {
		ar, err := item.AvailableRange()
		if err != nil {
			return
		}
		itemSourceRange = ar
	}

	// Calculate offset from original range start
	offsetFromStart := newRange.StartTime().Sub(originalRange.StartTime())

	// Calculate new source range
	newSourceStart := itemSourceRange.StartTime().Add(offsetFromStart.RescaledTo(itemSourceRange.StartTime().Rate()))
	newSourceDuration := newRange.Duration().RescaledTo(itemSourceRange.Duration().Rate())
	newSourceRange := opentime.NewTimeRange(newSourceStart, newSourceDuration)

	item.SetSourceRange(&newSourceRange)
}

// TopClipAtTime returns the topmost visible clip that overlaps with a given time.
// It walks through tracks from top to bottom (last to first) to find
// the first visible clip at the specified time.
func TopClipAtTime(stack *opentimelineio.Stack, t opentime.RationalTime) *opentimelineio.Clip {
	children := stack.Children()

	// Walk tracks in reverse order (top to bottom)
	for i := len(children) - 1; i >= 0; i-- {
		track, ok := children[i].(*opentimelineio.Track)
		if !ok {
			continue
		}

		// Find clip at this time in the track
		clip := clipAtTimeInTrack(track, t)
		if clip != nil && clip.Enabled() {
			return clip
		}
	}

	return nil
}

// clipAtTimeInTrack finds the clip at the given time in a track.
func clipAtTimeInTrack(track *opentimelineio.Track, t opentime.RationalTime) *opentimelineio.Clip {
	for i, child := range track.Children() {
		childRange, err := track.RangeOfChildAtIndex(i)
		if err != nil {
			continue
		}

		// Check if the time is within this child's range
		if !childRange.Contains(t) {
			continue
		}

		// If it's a clip, return it
		if clip, ok := child.(*opentimelineio.Clip); ok {
			return clip
		}

		// If it's a nested composition, recurse
		if nestedStack, ok := child.(*opentimelineio.Stack); ok {
			// Adjust time for nested composition
			nestedTime := t.Sub(childRange.StartTime())
			return TopClipAtTime(nestedStack, nestedTime)
		}
	}

	return nil
}
