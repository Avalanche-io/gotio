// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"encoding/json"

	"github.com/mrjoshuak/gotio/opentime"
)

// Composition is the interface for compositions that contain children.
type Composition interface {
	Item

	// CompositionKind returns the kind of composition.
	CompositionKind() string

	// Children returns the list of children.
	Children() []Composable

	// ClearChildren removes all children.
	ClearChildren()

	// SetChildren sets the list of children.
	SetChildren(children []Composable) error

	// InsertChild inserts a child at the given index.
	InsertChild(index int, child Composable) error

	// SetChild sets the child at the given index.
	SetChild(index int, child Composable) error

	// RemoveChild removes the child at the given index.
	RemoveChild(index int) error

	// AppendChild appends a child.
	AppendChild(child Composable) error

	// IndexOfChild returns the index of the given child.
	IndexOfChild(child Composable) (int, error)

	// IsParentOf returns whether this is the parent of the given child.
	IsParentOf(child Composable) bool

	// RangeOfChild returns the range of the given child.
	RangeOfChild(child Composable) (opentime.TimeRange, error)

	// RangeOfChildAtIndex returns the range of the child at the given index.
	RangeOfChildAtIndex(index int) (opentime.TimeRange, error)

	// TrimmedRangeOfChild returns the trimmed range of the given child.
	TrimmedRangeOfChild(child Composable) (*opentime.TimeRange, error)

	// TrimmedRangeOfChildAtIndex returns the trimmed range of the child at the given index.
	TrimmedRangeOfChildAtIndex(index int) (opentime.TimeRange, error)

	// HasChild returns whether this contains the given child.
	HasChild(child Composable) bool

	// HasClips returns whether this contains any clips.
	HasClips() bool

	// RangeOfAllChildren returns a map of child to range.
	RangeOfAllChildren() (map[Composable]opentime.TimeRange, error)

	// ChildAtTime returns the child at the given time.
	ChildAtTime(searchTime opentime.RationalTime, shallowSearch bool) (Composable, error)

	// ChildrenInRange returns all children within the given range.
	ChildrenInRange(searchRange opentime.TimeRange) ([]Composable, error)

	// FindChildren finds children matching the given type.
	FindChildren(searchRange *opentime.TimeRange, shallowSearch bool, filter func(Composable) bool) []Composable

	// FindClips finds all clips.
	FindClips(searchRange *opentime.TimeRange, shallowSearch bool) []*Clip
}

// CompositionSchema is the schema for Composition.
var CompositionSchema = Schema{Name: "Composition", Version: 1}

// CompositionBase is the base implementation of Composition.
type CompositionBase struct {
	ItemBase
	children []Composable
}

// NewCompositionBase creates a new CompositionBase.
func NewCompositionBase(
	name string,
	sourceRange *opentime.TimeRange,
	metadata AnyDictionary,
	effects []Effect,
	markers []*Marker,
	color *Color,
) CompositionBase {
	return CompositionBase{
		ItemBase: NewItemBase(name, sourceRange, metadata, effects, markers, true, color),
		children: make([]Composable, 0),
	}
}

// CompositionKind returns the kind of composition.
func (c *CompositionBase) CompositionKind() string {
	return "Composition"
}

// Children returns the children.
func (c *CompositionBase) Children() []Composable {
	return c.children
}

// ClearChildren removes all children.
func (c *CompositionBase) ClearChildren() {
	for _, child := range c.children {
		child.SetParent(nil)
	}
	c.children = make([]Composable, 0)
}

// SetChildren sets the children.
func (c *CompositionBase) SetChildren(children []Composable) error {
	c.ClearChildren()
	if children == nil {
		return nil
	}
	for _, child := range children {
		if err := c.AppendChild(child); err != nil {
			return err
		}
	}
	return nil
}

// InsertChild inserts a child at the given index.
// Note: The concrete composition type should call this and then set itself as parent.
func (c *CompositionBase) InsertChild(index int, child Composable) error {
	if index < 0 || index > len(c.children) {
		return &IndexError{Index: index, Size: len(c.children)}
	}
	// Use setParentRaw to set 'c' as parent - concrete types will override as needed
	if cb, ok := child.(interface{ setParentRaw(any) }); ok {
		cb.setParentRaw(c)
	}
	c.children = append(c.children[:index], append([]Composable{child}, c.children[index:]...)...)
	return nil
}

// SetChild sets the child at the given index.
func (c *CompositionBase) SetChild(index int, child Composable) error {
	if index < 0 || index >= len(c.children) {
		return &IndexError{Index: index, Size: len(c.children)}
	}
	c.children[index].SetParent(nil)
	if cb, ok := child.(interface{ setParentRaw(any) }); ok {
		cb.setParentRaw(c)
	}
	c.children[index] = child
	return nil
}

// RemoveChild removes the child at the given index.
func (c *CompositionBase) RemoveChild(index int) error {
	if index < 0 || index >= len(c.children) {
		return &IndexError{Index: index, Size: len(c.children)}
	}
	c.children[index].SetParent(nil)
	c.children = append(c.children[:index], c.children[index+1:]...)
	return nil
}

// AppendChild appends a child.
func (c *CompositionBase) AppendChild(child Composable) error {
	return c.InsertChild(len(c.children), child)
}

// IndexOfChild returns the index of the given child.
func (c *CompositionBase) IndexOfChild(child Composable) (int, error) {
	for i, ch := range c.children {
		if ch == child {
			return i, nil
		}
	}
	return -1, ErrNotFound
}

// IsParentOf returns whether this is the parent of the given child.
func (c *CompositionBase) IsParentOf(child Composable) bool {
	return c.HasChild(child)
}

// HasChild returns whether this contains the given child.
func (c *CompositionBase) HasChild(child Composable) bool {
	for _, ch := range c.children {
		if ch == child {
			return true
		}
	}
	return false
}

// HasClips returns whether this contains any clips.
func (c *CompositionBase) HasClips() bool {
	for _, child := range c.children {
		if _, ok := child.(*Clip); ok {
			return true
		}
		if comp, ok := child.(Composition); ok {
			if comp.HasClips() {
				return true
			}
		}
	}
	return false
}

// RangeOfChild returns the range of the given child.
func (c *CompositionBase) RangeOfChild(child Composable) (opentime.TimeRange, error) {
	index, err := c.IndexOfChild(child)
	if err != nil {
		return opentime.TimeRange{}, err
	}
	return c.RangeOfChildAtIndex(index)
}

// RangeOfChildAtIndex returns the range of the child at the given index.
// This default implementation assumes sequential children (like a Track).
func (c *CompositionBase) RangeOfChildAtIndex(index int) (opentime.TimeRange, error) {
	if index < 0 || index >= len(c.children) {
		return opentime.TimeRange{}, &IndexError{Index: index, Size: len(c.children)}
	}

	// Calculate start time by summing durations of previous children
	var startTime opentime.RationalTime
	for i := 0; i < index; i++ {
		dur, err := c.children[i].Duration()
		if err != nil {
			return opentime.TimeRange{}, err
		}
		startTime = startTime.Add(dur)
	}

	// Get duration of this child
	dur, err := c.children[index].Duration()
	if err != nil {
		return opentime.TimeRange{}, err
	}

	return opentime.NewTimeRange(startTime, dur), nil
}

// TrimmedRangeOfChild returns the trimmed range of the given child.
func (c *CompositionBase) TrimmedRangeOfChild(child Composable) (*opentime.TimeRange, error) {
	childRange, err := c.RangeOfChild(child)
	if err != nil {
		return nil, err
	}
	trimmed := c.trimChildRange(childRange)
	return trimmed, nil
}

// TrimmedRangeOfChildAtIndex returns the trimmed range of the child at the given index.
func (c *CompositionBase) TrimmedRangeOfChildAtIndex(index int) (opentime.TimeRange, error) {
	childRange, err := c.RangeOfChildAtIndex(index)
	if err != nil {
		return opentime.TimeRange{}, err
	}
	trimmed := c.trimChildRange(childRange)
	if trimmed == nil {
		return opentime.TimeRange{}, nil
	}
	return *trimmed, nil
}

// trimChildRange trims a child range to the source range.
func (c *CompositionBase) trimChildRange(childRange opentime.TimeRange) *opentime.TimeRange {
	if c.sourceRange == nil {
		return &childRange
	}
	if !c.sourceRange.Intersects(childRange, opentime.DefaultEpsilon) {
		return nil
	}
	result := c.sourceRange.ClampedRange(childRange)
	return &result
}

// RangeOfAllChildren returns a map of child to range.
func (c *CompositionBase) RangeOfAllChildren() (map[Composable]opentime.TimeRange, error) {
	result := make(map[Composable]opentime.TimeRange)
	for i, child := range c.children {
		r, err := c.RangeOfChildAtIndex(i)
		if err != nil {
			return nil, err
		}
		result[child] = r
	}
	return result, nil
}

// ChildAtTime returns the child at the given time.
func (c *CompositionBase) ChildAtTime(searchTime opentime.RationalTime, shallowSearch bool) (Composable, error) {
	children, err := c.childrenAtTime(searchTime)
	if err != nil {
		return nil, err
	}
	if len(children) == 0 {
		return nil, nil
	}

	child := children[0]

	if !shallowSearch {
		if comp, ok := child.(Composition); ok {
			// Transform time to child's coordinate system
			childRange, err := c.RangeOfChild(child)
			if err != nil {
				return nil, err
			}
			childTime := searchTime.Sub(childRange.StartTime())
			return comp.ChildAtTime(childTime, false)
		}
	}

	return child, nil
}

// childrenAtTime returns all children at the given time.
func (c *CompositionBase) childrenAtTime(searchTime opentime.RationalTime) ([]Composable, error) {
	var result []Composable
	for i, child := range c.children {
		childRange, err := c.RangeOfChildAtIndex(i)
		if err != nil {
			return nil, err
		}
		if childRange.Contains(searchTime) {
			result = append(result, child)
		}
	}
	return result, nil
}

// ChildrenInRange returns all children within the given range.
func (c *CompositionBase) ChildrenInRange(searchRange opentime.TimeRange) ([]Composable, error) {
	var result []Composable
	for i, child := range c.children {
		childRange, err := c.RangeOfChildAtIndex(i)
		if err != nil {
			return nil, err
		}
		if searchRange.Intersects(childRange, opentime.DefaultEpsilon) {
			result = append(result, child)
		}
	}
	return result, nil
}

// FindChildren finds children matching the given filter.
func (c *CompositionBase) FindChildren(searchRange *opentime.TimeRange, shallowSearch bool, filter func(Composable) bool) []Composable {
	var result []Composable
	var children []Composable

	if searchRange != nil {
		children, _ = c.ChildrenInRange(*searchRange)
	} else {
		children = c.children
	}

	for _, child := range children {
		if filter == nil || filter(child) {
			result = append(result, child)
		}

		if !shallowSearch {
			if comp, ok := child.(Composition); ok {
				var childRange *opentime.TimeRange
				if searchRange != nil {
					// Transform search range to child's coordinate system
					r, err := c.RangeOfChild(child)
					if err == nil {
						transformed := opentime.NewTimeRange(
							searchRange.StartTime().Sub(r.StartTime()),
							searchRange.Duration(),
						)
						childRange = &transformed
					}
				}
				childResults := comp.FindChildren(childRange, false, filter)
				result = append(result, childResults...)
			}
		}
	}

	return result
}

// FindClips finds all clips.
func (c *CompositionBase) FindClips(searchRange *opentime.TimeRange, shallowSearch bool) []*Clip {
	children := c.FindChildren(searchRange, shallowSearch, func(child Composable) bool {
		_, ok := child.(*Clip)
		return ok
	})
	result := make([]*Clip, len(children))
	for i, child := range children {
		result[i] = child.(*Clip)
	}
	return result
}

// Duration returns the duration of the composition.
func (c *CompositionBase) Duration() (opentime.RationalTime, error) {
	if c.sourceRange != nil {
		return c.sourceRange.Duration(), nil
	}
	return c.computedDuration()
}

// computedDuration computes the duration from children.
func (c *CompositionBase) computedDuration() (opentime.RationalTime, error) {
	var total opentime.RationalTime
	for _, child := range c.children {
		dur, err := child.Duration()
		if err != nil {
			return opentime.RationalTime{}, err
		}
		total = total.Add(dur)
	}
	return total, nil
}

// AvailableRange returns the available range.
func (c *CompositionBase) AvailableRange() (opentime.TimeRange, error) {
	dur, err := c.computedDuration()
	if err != nil {
		return opentime.TimeRange{}, err
	}
	return opentime.NewTimeRange(opentime.RationalTime{}, dur), nil
}

// cloneChildren creates a deep copy of children.
func cloneChildren(children []Composable) []Composable {
	if children == nil {
		return nil
	}
	result := make([]Composable, len(children))
	for i, child := range children {
		result[i] = child.Clone().(Composable)
	}
	return result
}

// compositionJSON is the JSON representation.
type compositionJSON struct {
	Schema      string              `json:"OTIO_SCHEMA"`
	Name        string              `json:"name"`
	Metadata    AnyDictionary       `json:"metadata"`
	SourceRange *opentime.TimeRange `json:"source_range"`
	Effects     []RawMessage   `json:"effects"`
	Markers     []*Marker           `json:"markers"`
	Enabled     bool                `json:"enabled"`
	Color       *Color              `json:"color"`
	Children    []RawMessage   `json:"children"`
}

// marshalCompositionJSON creates a compositionJSON.
func (c *CompositionBase) marshalCompositionJSON(schema Schema) (compositionJSON, error) {
	effects := make([]RawMessage, len(c.effects))
	for i, e := range c.effects {
		data, err := json.Marshal(e)
		if err != nil {
			return compositionJSON{}, err
		}
		effects[i] = data
	}

	children := make([]RawMessage, len(c.children))
	for i, child := range c.children {
		data, err := json.Marshal(child)
		if err != nil {
			return compositionJSON{}, err
		}
		children[i] = data
	}

	return compositionJSON{
		Schema:      schema.String(),
		Name:        c.name,
		Metadata:    c.metadata,
		SourceRange: c.sourceRange,
		Effects:     effects,
		Markers:     c.markers,
		Enabled:     c.enabled,
		Color:       c.color,
		Children:    children,
	}, nil
}

// unmarshalCompositionJSON unmarshals JSON.
func (c *CompositionBase) unmarshalCompositionJSON(j *compositionJSON) error {
	c.name = j.Name
	c.metadata = j.Metadata
	if c.metadata == nil {
		c.metadata = make(AnyDictionary)
	}
	c.sourceRange = j.SourceRange
	c.enabled = j.Enabled
	c.color = j.Color

	// Unmarshal effects
	c.effects = make([]Effect, len(j.Effects))
	for i, data := range j.Effects {
		obj, err := FromJSONString(string(data))
		if err != nil {
			return err
		}
		effect, ok := obj.(Effect)
		if !ok {
			return &TypeMismatchError{Expected: "Effect", Got: obj.SchemaName()}
		}
		c.effects[i] = effect
	}

	// Copy markers
	c.markers = j.Markers
	if c.markers == nil {
		c.markers = make([]*Marker, 0)
	}

	// Unmarshal children
	c.children = make([]Composable, len(j.Children))
	for i, data := range j.Children {
		obj, err := FromJSONString(string(data))
		if err != nil {
			return err
		}
		composable, ok := obj.(Composable)
		if !ok {
			return &TypeMismatchError{Expected: "Composable", Got: obj.SchemaName()}
		}
		// Use setParentRaw since we're in CompositionBase context
		if cb, ok := composable.(interface{ setParentRaw(any) }); ok {
			cb.setParentRaw(c)
		}
		c.children[i] = composable
	}

	return nil
}

// CompositionImpl is a standalone implementation of Composition.
type CompositionImpl struct {
	CompositionBase
}

// NewComposition creates a new Composition.
func NewComposition(
	name string,
	sourceRange *opentime.TimeRange,
	metadata AnyDictionary,
	effects []Effect,
	markers []*Marker,
	color *Color,
) *CompositionImpl {
	comp := &CompositionImpl{
		CompositionBase: NewCompositionBase(name, sourceRange, metadata, effects, markers, color),
	}
	comp.SetSelf(comp)
	return comp
}

// SchemaName returns the schema name.
func (c *CompositionImpl) SchemaName() string {
	return CompositionSchema.Name
}

// SchemaVersion returns the schema version.
func (c *CompositionImpl) SchemaVersion() int {
	return CompositionSchema.Version
}

// Clone creates a deep copy.
func (c *CompositionImpl) Clone() SerializableObject {
	clone := &CompositionImpl{
		CompositionBase: CompositionBase{
			ItemBase: ItemBase{
				ComposableBase: ComposableBase{
					SerializableObjectWithMetadataBase: SerializableObjectWithMetadataBase{
						name:     c.name,
						metadata: CloneAnyDictionary(c.metadata),
					},
				},
				sourceRange: cloneSourceRange(c.sourceRange),
				effects:     cloneEffects(c.effects),
				markers:     cloneMarkers(c.markers),
				enabled:     c.enabled,
				color:       cloneColor(c.color),
			},
			children: cloneChildren(c.children),
		},
	}
	clone.SetSelf(clone)
	// Set parent references
	for _, child := range clone.children {
		child.SetParent(clone)
	}
	return clone
}

// IsEquivalentTo returns true if equivalent to another.
func (c *CompositionImpl) IsEquivalentTo(other SerializableObject) bool {
	otherC, ok := other.(*CompositionImpl)
	if !ok {
		return false
	}
	if c.name != otherC.name || c.enabled != otherC.enabled {
		return false
	}
	if len(c.children) != len(otherC.children) {
		return false
	}
	for i := range c.children {
		if !c.children[i].IsEquivalentTo(otherC.children[i]) {
			return false
		}
	}
	return true
}

// MarshalJSON implements json.Marshaler.
func (c *CompositionImpl) MarshalJSON() ([]byte, error) {
	j, err := c.marshalCompositionJSON(CompositionSchema)
	if err != nil {
		return nil, err
	}
	return json.Marshal(j)
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *CompositionImpl) UnmarshalJSON(data []byte) error {
	var j compositionJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := c.unmarshalCompositionJSON(&j); err != nil {
		return err
	}
	c.SetSelf(c)
	return nil
}

// SetParent implements Composable.SetParent.
func (c *CompositionImpl) SetParent(parent Composition) {
	c.parent = parent
}

func init() {
	RegisterSchema(CompositionSchema, func() SerializableObject {
		return NewComposition("", nil, nil, nil, nil, nil)
	})
}
