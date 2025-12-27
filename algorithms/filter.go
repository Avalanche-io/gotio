// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package algorithms

import (
	"reflect"

	"github.com/Avalanche-io/gotio/opentimelineio"
)

// FilterFunc is a function that filters/transforms a composable.
// Return values:
// - The modified composable (or original to keep unchanged)
// - A slice of composables to expand the item into multiple items
// - nil to remove the item
type FilterFunc func(composable opentimelineio.Composable) []opentimelineio.Composable

// ContextFilterFunc is like FilterFunc but also receives previous and next items
// for context-aware filtering.
type ContextFilterFunc func(prev, current, next opentimelineio.Composable) []opentimelineio.Composable

// FilteredComposition returns a deep-copy filtered composition tree.
// The filter function is called for each composable and can:
// - Return a single-element slice to keep (possibly modified)
// - Return a multi-element slice to expand into multiple items
// - Return nil or empty slice to prune the item
// typesToPrune optionally specifies types to automatically prune.
func FilteredComposition(
	root opentimelineio.SerializableObject,
	filter FilterFunc,
	typesToPrune []reflect.Type,
) opentimelineio.SerializableObject {
	return filteredCompositionRecursive(root, filter, typesToPrune)
}

func filteredCompositionRecursive(
	obj opentimelineio.SerializableObject,
	filter FilterFunc,
	typesToPrune []reflect.Type,
) opentimelineio.SerializableObject {
	if obj == nil {
		return nil
	}

	// Check if this type should be pruned
	for _, prunedType := range typesToPrune {
		if reflect.TypeOf(obj) == prunedType {
			return nil
		}
	}

	// Clone the object to avoid modifying the original
	cloned := obj.Clone()

	// If it's a composable, apply the filter
	if composable, ok := cloned.(opentimelineio.Composable); ok {
		result := filter(composable)
		if len(result) == 0 {
			return nil
		}
		if len(result) == 1 {
			cloned = result[0]
		}
		// For expansion (multiple items), the caller needs to handle this
		// at the composition level, so we just return the first for now
	}

	// Process children if this is a composition
	switch comp := cloned.(type) {
	case *opentimelineio.Timeline:
		// Process the tracks
		if tracks := comp.Tracks(); tracks != nil {
			filtered := filteredCompositionRecursive(tracks, filter, typesToPrune)
			if stack, ok := filtered.(*opentimelineio.Stack); ok {
				comp.SetTracks(stack)
			}
		}

	case *opentimelineio.Stack:
		newChildren := filterChildren(comp.Children(), filter, typesToPrune)
		comp.SetChildren(nil)
		for _, child := range newChildren {
			comp.AppendChild(child)
		}

	case *opentimelineio.Track:
		newChildren := filterChildren(comp.Children(), filter, typesToPrune)
		comp.SetChildren(nil)
		for _, child := range newChildren {
			comp.AppendChild(child)
		}

	case *opentimelineio.SerializableCollection:
		var newChildren []opentimelineio.SerializableObject
		for _, child := range comp.Children() {
			filtered := filteredCompositionRecursive(child, filter, typesToPrune)
			if filtered != nil {
				newChildren = append(newChildren, filtered)
			}
		}
		comp.SetChildren(newChildren)
	}

	return cloned
}

// filterChildren filters a slice of composable children.
func filterChildren(
	children []opentimelineio.Composable,
	filter FilterFunc,
	typesToPrune []reflect.Type,
) []opentimelineio.Composable {
	var result []opentimelineio.Composable

	for _, child := range children {
		// Check if this type should be pruned
		shouldPrune := false
		for _, prunedType := range typesToPrune {
			if reflect.TypeOf(child) == prunedType {
				shouldPrune = true
				break
			}
		}
		if shouldPrune {
			continue
		}

		// Apply filter
		filtered := filter(child)
		if len(filtered) == 0 {
			continue
		}

		// Process each resulting item
		for _, item := range filtered {
			// Recursively filter if it's a composition
			if so, ok := item.(opentimelineio.SerializableObject); ok {
				recursed := filteredCompositionRecursive(so, filter, typesToPrune)
				if composable, ok := recursed.(opentimelineio.Composable); ok && composable != nil {
					result = append(result, composable)
				}
			} else {
				result = append(result, item)
			}
		}
	}

	return result
}

// FilteredWithSequenceContext returns a filtered composition where the filter
// function receives previous and next items for context-aware filtering.
func FilteredWithSequenceContext(
	root opentimelineio.SerializableObject,
	filter ContextFilterFunc,
	typesToPrune []reflect.Type,
) opentimelineio.SerializableObject {
	return filteredWithContextRecursive(root, filter, typesToPrune)
}

func filteredWithContextRecursive(
	obj opentimelineio.SerializableObject,
	filter ContextFilterFunc,
	typesToPrune []reflect.Type,
) opentimelineio.SerializableObject {
	if obj == nil {
		return nil
	}

	// Check if this type should be pruned
	for _, prunedType := range typesToPrune {
		if reflect.TypeOf(obj) == prunedType {
			return nil
		}
	}

	// Clone the object
	cloned := obj.Clone()

	// Process children if this is a composition
	switch comp := cloned.(type) {
	case *opentimelineio.Timeline:
		if tracks := comp.Tracks(); tracks != nil {
			filtered := filteredWithContextRecursive(tracks, filter, typesToPrune)
			if stack, ok := filtered.(*opentimelineio.Stack); ok {
				comp.SetTracks(stack)
			}
		}

	case *opentimelineio.Stack:
		newChildren := filterChildrenWithContext(comp.Children(), filter, typesToPrune)
		comp.SetChildren(nil)
		for _, child := range newChildren {
			comp.AppendChild(child)
		}

	case *opentimelineio.Track:
		newChildren := filterChildrenWithContext(comp.Children(), filter, typesToPrune)
		comp.SetChildren(nil)
		for _, child := range newChildren {
			comp.AppendChild(child)
		}

	case *opentimelineio.SerializableCollection:
		var newChildren []opentimelineio.SerializableObject
		children := comp.Children()
		for i, child := range children {
			var prev, next opentimelineio.SerializableObject
			if i > 0 {
				prev = children[i-1]
			}
			if i < len(children)-1 {
				next = children[i+1]
			}

			// Apply context filter (wrapping in adapter)
			if composable, ok := child.(opentimelineio.Composable); ok {
				var prevComp, nextComp opentimelineio.Composable
				if prev != nil {
					prevComp, _ = prev.(opentimelineio.Composable)
				}
				if next != nil {
					nextComp, _ = next.(opentimelineio.Composable)
				}

				result := filter(prevComp, composable, nextComp)
				for _, item := range result {
					if so, ok := item.(opentimelineio.SerializableObject); ok {
						filtered := filteredWithContextRecursive(so, filter, typesToPrune)
						if filtered != nil {
							newChildren = append(newChildren, filtered)
						}
					}
				}
			} else {
				// Non-composable, just recurse
				filtered := filteredWithContextRecursive(child, filter, typesToPrune)
				if filtered != nil {
					newChildren = append(newChildren, filtered)
				}
			}
		}
		comp.SetChildren(newChildren)
	}

	return cloned
}

// filterChildrenWithContext filters children with context.
func filterChildrenWithContext(
	children []opentimelineio.Composable,
	filter ContextFilterFunc,
	typesToPrune []reflect.Type,
) []opentimelineio.Composable {
	var result []opentimelineio.Composable

	for i, child := range children {
		// Check if this type should be pruned
		shouldPrune := false
		for _, prunedType := range typesToPrune {
			if reflect.TypeOf(child) == prunedType {
				shouldPrune = true
				break
			}
		}
		if shouldPrune {
			continue
		}

		// Get context
		var prev, next opentimelineio.Composable
		if i > 0 {
			prev = children[i-1]
		}
		if i < len(children)-1 {
			next = children[i+1]
		}

		// Apply filter
		filtered := filter(prev, child, next)
		if len(filtered) == 0 {
			continue
		}

		// Process each resulting item
		for _, item := range filtered {
			if so, ok := item.(opentimelineio.SerializableObject); ok {
				recursed := filteredWithContextRecursive(so, filter, typesToPrune)
				if composable, ok := recursed.(opentimelineio.Composable); ok && composable != nil {
					result = append(result, composable)
				}
			} else {
				result = append(result, item)
			}
		}
	}

	return result
}

// KeepFilter is a helper filter that keeps all items unchanged.
func KeepFilter(composable opentimelineio.Composable) []opentimelineio.Composable {
	return []opentimelineio.Composable{composable}
}

// PruneFilter is a helper filter that removes all items.
func PruneFilter(composable opentimelineio.Composable) []opentimelineio.Composable {
	return nil
}

// TypeFilter returns a filter that keeps only items of the specified types.
func TypeFilter(keepTypes ...reflect.Type) FilterFunc {
	return func(composable opentimelineio.Composable) []opentimelineio.Composable {
		itemType := reflect.TypeOf(composable)
		for _, keepType := range keepTypes {
			if itemType == keepType {
				return []opentimelineio.Composable{composable}
			}
		}
		return nil
	}
}

// NameFilter returns a filter that keeps only items with names matching the predicate.
func NameFilter(predicate func(string) bool) FilterFunc {
	return func(composable opentimelineio.Composable) []opentimelineio.Composable {
		if so, ok := composable.(opentimelineio.SerializableObjectWithMetadata); ok {
			if predicate(so.Name()) {
				return []opentimelineio.Composable{composable}
			}
		}
		return nil
	}
}
