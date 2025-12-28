// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package algorithms

import (
	"reflect"

	"github.com/Avalanche-io/gotio"
)

// FilterFunc is a function that filters/transforms a composable.
// Return values:
// - The modified composable (or original to keep unchanged)
// - A slice of composables to expand the item into multiple items
// - nil to remove the item
type FilterFunc func(composable gotio.Composable) []gotio.Composable

// ContextFilterFunc is like FilterFunc but also receives previous and next items
// for context-aware filtering.
type ContextFilterFunc func(prev, current, next gotio.Composable) []gotio.Composable

// FilteredComposition returns a deep-copy filtered composition tree.
// The filter function is called for each composable and can:
// - Return a single-element slice to keep (possibly modified)
// - Return a multi-element slice to expand into multiple items
// - Return nil or empty slice to prune the item
// typesToPrune optionally specifies types to automatically prune.
func FilteredComposition(
	root gotio.SerializableObject,
	filter FilterFunc,
	typesToPrune []reflect.Type,
) gotio.SerializableObject {
	return filteredCompositionRecursive(root, filter, typesToPrune)
}

func filteredCompositionRecursive(
	obj gotio.SerializableObject,
	filter FilterFunc,
	typesToPrune []reflect.Type,
) gotio.SerializableObject {
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
	if composable, ok := cloned.(gotio.Composable); ok {
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
	case *gotio.Timeline:
		// Process the tracks
		if tracks := comp.Tracks(); tracks != nil {
			filtered := filteredCompositionRecursive(tracks, filter, typesToPrune)
			if stack, ok := filtered.(*gotio.Stack); ok {
				comp.SetTracks(stack)
			}
		}

	case *gotio.Stack:
		newChildren := filterChildren(comp.Children(), filter, typesToPrune)
		comp.SetChildren(nil)
		for _, child := range newChildren {
			comp.AppendChild(child)
		}

	case *gotio.Track:
		newChildren := filterChildren(comp.Children(), filter, typesToPrune)
		comp.SetChildren(nil)
		for _, child := range newChildren {
			comp.AppendChild(child)
		}

	case *gotio.SerializableCollection:
		var newChildren []gotio.SerializableObject
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
	children []gotio.Composable,
	filter FilterFunc,
	typesToPrune []reflect.Type,
) []gotio.Composable {
	var result []gotio.Composable

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
			if so, ok := item.(gotio.SerializableObject); ok {
				recursed := filteredCompositionRecursive(so, filter, typesToPrune)
				if composable, ok := recursed.(gotio.Composable); ok && composable != nil {
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
	root gotio.SerializableObject,
	filter ContextFilterFunc,
	typesToPrune []reflect.Type,
) gotio.SerializableObject {
	return filteredWithContextRecursive(root, filter, typesToPrune)
}

func filteredWithContextRecursive(
	obj gotio.SerializableObject,
	filter ContextFilterFunc,
	typesToPrune []reflect.Type,
) gotio.SerializableObject {
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
	case *gotio.Timeline:
		if tracks := comp.Tracks(); tracks != nil {
			filtered := filteredWithContextRecursive(tracks, filter, typesToPrune)
			if stack, ok := filtered.(*gotio.Stack); ok {
				comp.SetTracks(stack)
			}
		}

	case *gotio.Stack:
		newChildren := filterChildrenWithContext(comp.Children(), filter, typesToPrune)
		comp.SetChildren(nil)
		for _, child := range newChildren {
			comp.AppendChild(child)
		}

	case *gotio.Track:
		newChildren := filterChildrenWithContext(comp.Children(), filter, typesToPrune)
		comp.SetChildren(nil)
		for _, child := range newChildren {
			comp.AppendChild(child)
		}

	case *gotio.SerializableCollection:
		var newChildren []gotio.SerializableObject
		children := comp.Children()
		for i, child := range children {
			var prev, next gotio.SerializableObject
			if i > 0 {
				prev = children[i-1]
			}
			if i < len(children)-1 {
				next = children[i+1]
			}

			// Apply context filter (wrapping in adapter)
			if composable, ok := child.(gotio.Composable); ok {
				var prevComp, nextComp gotio.Composable
				if prev != nil {
					prevComp, _ = prev.(gotio.Composable)
				}
				if next != nil {
					nextComp, _ = next.(gotio.Composable)
				}

				result := filter(prevComp, composable, nextComp)
				for _, item := range result {
					if so, ok := item.(gotio.SerializableObject); ok {
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
	children []gotio.Composable,
	filter ContextFilterFunc,
	typesToPrune []reflect.Type,
) []gotio.Composable {
	var result []gotio.Composable

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
		var prev, next gotio.Composable
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
			if so, ok := item.(gotio.SerializableObject); ok {
				recursed := filteredWithContextRecursive(so, filter, typesToPrune)
				if composable, ok := recursed.(gotio.Composable); ok && composable != nil {
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
func KeepFilter(composable gotio.Composable) []gotio.Composable {
	return []gotio.Composable{composable}
}

// PruneFilter is a helper filter that removes all items.
func PruneFilter(composable gotio.Composable) []gotio.Composable {
	return nil
}

// TypeFilter returns a filter that keeps only items of the specified types.
func TypeFilter(keepTypes ...reflect.Type) FilterFunc {
	return func(composable gotio.Composable) []gotio.Composable {
		itemType := reflect.TypeOf(composable)
		for _, keepType := range keepTypes {
			if itemType == keepType {
				return []gotio.Composable{composable}
			}
		}
		return nil
	}
}

// NameFilter returns a filter that keeps only items with names matching the predicate.
func NameFilter(predicate func(string) bool) FilterFunc {
	return func(composable gotio.Composable) []gotio.Composable {
		if so, ok := composable.(gotio.SerializableObjectWithMetadata); ok {
			if predicate(so.Name()) {
				return []gotio.Composable{composable}
			}
		}
		return nil
	}
}
