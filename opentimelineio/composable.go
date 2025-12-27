// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"github.com/mrjoshuak/gotio/opentime"
)

// Composable is the interface for items that can be composed in a composition.
type Composable interface {
	SerializableObjectWithMetadata

	// Parent returns the parent composition.
	Parent() Composition

	// SetParent sets the parent composition.
	SetParent(parent Composition)

	// Duration returns the duration.
	Duration() (opentime.RationalTime, error)

	// Visible returns whether this item is visible (takes up time in parent).
	Visible() bool

	// Overlapping returns whether this item overlaps with neighbors.
	Overlapping() bool
}

// ComposableBase is the base implementation of Composable.
type ComposableBase struct {
	SerializableObjectWithMetadataBase
	parent any        // stores Composition, but uses any to avoid type constraints on embedded methods
	self   Composable // reference to the containing struct (set by concrete types)
}

// NewComposableBase creates a new ComposableBase.
func NewComposableBase(name string, metadata AnyDictionary) ComposableBase {
	return ComposableBase{
		SerializableObjectWithMetadataBase: NewSerializableObjectWithMetadataBase(name, metadata),
	}
}

// Parent returns the parent.
func (c *ComposableBase) Parent() Composition {
	if c.parent == nil {
		return nil
	}
	if comp, ok := c.parent.(Composition); ok {
		return comp
	}
	return nil
}

// SetParent sets the parent.
func (c *ComposableBase) SetParent(parent Composition) {
	c.parent = parent
}

// setParentRaw sets the parent without type checking (used internally by CompositionBase).
func (c *ComposableBase) setParentRaw(parent any) {
	c.parent = parent
}

// Visible returns true by default.
func (c *ComposableBase) Visible() bool {
	return true
}

// Overlapping returns false by default.
func (c *ComposableBase) Overlapping() bool {
	return false
}

// Self returns the Composable interface value for this object.
// This is set by concrete types to enable hierarchy traversal.
func (c *ComposableBase) Self() Composable {
	return c.self
}

// SetSelf sets the Composable interface value for this object.
// Concrete types must call this after construction.
func (c *ComposableBase) SetSelf(self Composable) {
	c.self = self
}
