// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"github.com/Avalanche-io/gotio/opentime"
)

// Item is the interface for timeline items.
type Item interface {
	Composable

	// SourceRange returns the source range.
	SourceRange() *opentime.TimeRange

	// SetSourceRange sets the source range.
	SetSourceRange(sourceRange *opentime.TimeRange)

	// Effects returns the effects.
	Effects() []Effect

	// SetEffects sets the effects.
	SetEffects(effects []Effect)

	// Markers returns the markers.
	Markers() []*Marker

	// SetMarkers sets the markers.
	SetMarkers(markers []*Marker)

	// Enabled returns whether the item is enabled.
	Enabled() bool

	// SetEnabled sets whether the item is enabled.
	SetEnabled(enabled bool)

	// ItemColor returns the item color.
	ItemColor() *Color

	// SetItemColor sets the item color.
	SetItemColor(color *Color)

	// AvailableRange returns the available range.
	AvailableRange() (opentime.TimeRange, error)

	// TrimmedRange returns the trimmed range.
	TrimmedRange() (opentime.TimeRange, error)

	// VisibleRange returns the visible range.
	VisibleRange() (opentime.TimeRange, error)

	// TransformedTime transforms a time from this item to another.
	TransformedTime(t opentime.RationalTime, toItem Item) (opentime.RationalTime, error)

	// TransformedTimeRange transforms a time range from this item to another.
	TransformedTimeRange(tr opentime.TimeRange, toItem Item) (opentime.TimeRange, error)
}

// ItemBase is the base implementation of Item.
type ItemBase struct {
	ComposableBase
	sourceRange *opentime.TimeRange
	effects     []Effect
	markers     []*Marker
	enabled     bool
	color       *Color
}

// NewItemBase creates a new ItemBase.
func NewItemBase(
	name string,
	sourceRange *opentime.TimeRange,
	metadata AnyDictionary,
	effects []Effect,
	markers []*Marker,
	enabled bool,
	color *Color,
) ItemBase {
	if effects == nil {
		effects = make([]Effect, 0)
	}
	if markers == nil {
		markers = make([]*Marker, 0)
	}
	return ItemBase{
		ComposableBase: NewComposableBase(name, metadata),
		sourceRange:    sourceRange,
		effects:        effects,
		markers:        markers,
		enabled:        enabled,
		color:          color,
	}
}

// SourceRange returns the source range.
func (i *ItemBase) SourceRange() *opentime.TimeRange {
	return i.sourceRange
}

// SetSourceRange sets the source range.
func (i *ItemBase) SetSourceRange(sourceRange *opentime.TimeRange) {
	i.sourceRange = sourceRange
}

// Effects returns the effects.
func (i *ItemBase) Effects() []Effect {
	return i.effects
}

// SetEffects sets the effects.
func (i *ItemBase) SetEffects(effects []Effect) {
	if effects == nil {
		effects = make([]Effect, 0)
	}
	i.effects = effects
}

// Markers returns the markers.
func (i *ItemBase) Markers() []*Marker {
	return i.markers
}

// SetMarkers sets the markers.
func (i *ItemBase) SetMarkers(markers []*Marker) {
	if markers == nil {
		markers = make([]*Marker, 0)
	}
	i.markers = markers
}

// Enabled returns whether enabled.
func (i *ItemBase) Enabled() bool {
	return i.enabled
}

// SetEnabled sets whether enabled.
func (i *ItemBase) SetEnabled(enabled bool) {
	i.enabled = enabled
}

// ItemColor returns the color.
func (i *ItemBase) ItemColor() *Color {
	return i.color
}

// SetItemColor sets the color.
func (i *ItemBase) SetItemColor(color *Color) {
	i.color = color
}

// Note: Duration() is not implemented on ItemBase because all concrete types
// (Clip, Gap, Track, Stack, Transition, etc.) provide their own Duration()
// implementations with specific logic for each type.

// AvailableRange should be overridden by concrete types.
func (i *ItemBase) AvailableRange() (opentime.TimeRange, error) {
	return opentime.TimeRange{}, ErrCannotComputeAvailableRange
}

// TrimmedRange returns the trimmed range.
func (i *ItemBase) TrimmedRange() (opentime.TimeRange, error) {
	if i.sourceRange != nil {
		return *i.sourceRange, nil
	}
	// Use self reference for dynamic dispatch to concrete type's AvailableRange
	if selfItem, ok := i.Self().(Item); ok {
		return selfItem.AvailableRange()
	}
	return i.AvailableRange()
}

// VisibleRange returns the visible range (trimmed range with effects).
func (i *ItemBase) VisibleRange() (opentime.TimeRange, error) {
	if i.sourceRange != nil {
		return *i.sourceRange, nil
	}
	// Use self reference for dynamic dispatch to concrete type's AvailableRange
	if selfItem, ok := i.Self().(Item); ok {
		return selfItem.AvailableRange()
	}
	return i.AvailableRange()
}

// TransformedTime transforms a time from this item's coordinate space to another item's
// coordinate space. Both items must share a common ancestor in the composition hierarchy.
//
// The algorithm works by:
// 1. Walking UP from this item to the common ancestor, converting from internal time
//    to parent coordinates at each level
// 2. Walking DOWN from the common ancestor to the target item, converting from parent
//    coordinates to internal time at each level
func (i *ItemBase) TransformedTime(t opentime.RationalTime, toItem Item) (opentime.RationalTime, error) {
	if toItem == nil {
		return t, nil
	}

	// Get self as Composable for passing to parent methods
	selfComposable := i.Self()
	if selfComposable == nil {
		// Fallback: if self not set, return unchanged (shouldn't happen in proper usage)
		return t, nil
	}
	selfItem, ok := selfComposable.(Item)
	if !ok {
		return t, nil
	}

	// Get the highest ancestor (root) of this item
	root := i.highestAncestor()

	// Build path from this item to root, transforming coordinates along the way
	result := t
	item := selfItem

	// Walk UP from this item towards root or toItem
	for item != root && item != toItem {
		parent := item.Parent()
		if parent == nil {
			break
		}

		// Convert from item's internal time to parent's coordinate system
		// Step 1: Subtract the item's trimmed_range start time
		trimmedRange, err := item.TrimmedRange()
		if err != nil {
			return result, err
		}
		result = result.Sub(trimmedRange.StartTime())

		// Step 2: Add the parent's range_of_child start time
		// Need to pass the Composable interface to RangeOfChild
		itemComposable, ok := item.(Composable)
		if !ok {
			break
		}
		rangeInParent, err := parent.RangeOfChild(itemComposable)
		if err != nil {
			return result, err
		}
		result = result.Add(rangeInParent.StartTime())

		// Move up to parent
		if parentItem, ok := parent.(Item); ok {
			item = parentItem
		} else {
			break
		}
	}

	// If we found toItem while walking up, we're done
	if item == toItem {
		return result, nil
	}

	// Now walk UP from toItem towards the ancestor we reached
	ancestor := item
	item = toItem

	// Build a stack of transformations needed to go from ancestor down to toItem
	type transform struct {
		trimmedStart  opentime.RationalTime
		rangeInParent opentime.RationalTime
	}
	var transforms []transform

	for item != root && item != ancestor {
		parent := item.Parent()
		if parent == nil {
			break
		}

		trimmedRange, err := item.TrimmedRange()
		if err != nil {
			return result, err
		}

		itemComposable, ok := item.(Composable)
		if !ok {
			break
		}
		rangeInParent, err := parent.RangeOfChild(itemComposable)
		if err != nil {
			return result, err
		}

		transforms = append(transforms, transform{
			trimmedStart:  trimmedRange.StartTime(),
			rangeInParent: rangeInParent.StartTime(),
		})

		if parentItem, ok := parent.(Item); ok {
			item = parentItem
		} else {
			break
		}
	}

	// Apply transforms in reverse order (walking DOWN from ancestor to toItem)
	for j := len(transforms) - 1; j >= 0; j-- {
		tr := transforms[j]
		// Convert from parent's coordinate system to item's internal time
		// Step 1: Subtract the parent's range_of_child start time
		result = result.Sub(tr.rangeInParent)
		// Step 2: Add the item's trimmed_range start time
		result = result.Add(tr.trimmedStart)
	}

	return result, nil
}

// TransformedTimeRange transforms a time range from this item's coordinate space
// to another item's coordinate space. The duration is preserved; only the start
// time is transformed.
func (i *ItemBase) TransformedTimeRange(tr opentime.TimeRange, toItem Item) (opentime.TimeRange, error) {
	transformedStart, err := i.TransformedTime(tr.StartTime(), toItem)
	if err != nil {
		return opentime.TimeRange{}, err
	}
	return opentime.NewTimeRange(transformedStart, tr.Duration()), nil
}

// highestAncestor returns the root of the composition hierarchy containing this item.
// If the item has no parent, it returns itself.
func (i *ItemBase) highestAncestor() Item {
	selfComposable := i.Self()
	if selfComposable == nil {
		return nil
	}
	current, ok := selfComposable.(Item)
	if !ok {
		return nil
	}
	for {
		parent := current.Parent()
		if parent == nil {
			return current
		}
		if parentItem, ok := parent.(Item); ok {
			current = parentItem
		} else {
			return current
		}
	}
}

// cloneSourceRange creates a copy of a TimeRange pointer.
func cloneSourceRange(tr *opentime.TimeRange) *opentime.TimeRange {
	if tr == nil {
		return nil
	}
	clone := *tr
	return &clone
}

// cloneEffects creates a deep copy of effects.
func cloneEffects(effects []Effect) []Effect {
	if effects == nil {
		return nil
	}
	result := make([]Effect, len(effects))
	for i, e := range effects {
		result[i] = e.Clone().(Effect)
	}
	return result
}

// cloneMarkers creates a deep copy of markers.
func cloneMarkers(markers []*Marker) []*Marker {
	if markers == nil {
		return nil
	}
	result := make([]*Marker, len(markers))
	for i, m := range markers {
		result[i] = m.Clone().(*Marker)
	}
	return result
}
