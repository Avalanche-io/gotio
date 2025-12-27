// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"encoding/json"

	"github.com/Avalanche-io/gotio/opentime"
)

// StackSchema is the schema for Stack.
var StackSchema = Schema{Name: "Stack", Version: 1}

// Stack is a composition of items arranged in layers (overlapping in time).
type Stack struct {
	CompositionBase
}

// NewStack creates a new Stack.
func NewStack(
	name string,
	sourceRange *opentime.TimeRange,
	metadata AnyDictionary,
	effects []Effect,
	markers []*Marker,
	color *Color,
) *Stack {
	stack := &Stack{
		CompositionBase: NewCompositionBase(name, sourceRange, metadata, effects, markers, color),
	}
	stack.SetSelf(stack)
	return stack
}

// CompositionKind returns "Stack".
func (s *Stack) CompositionKind() string {
	return "Stack"
}

// InsertChild inserts a child at the given index.
func (s *Stack) InsertChild(index int, child Composable) error {
	if index < 0 || index > len(s.children) {
		return &IndexError{Index: index, Size: len(s.children)}
	}
	child.SetParent(s)
	s.children = append(s.children[:index], append([]Composable{child}, s.children[index:]...)...)
	return nil
}

// AppendChild appends a child.
func (s *Stack) AppendChild(child Composable) error {
	return s.InsertChild(len(s.children), child)
}

// SetChild sets the child at the given index.
func (s *Stack) SetChild(index int, child Composable) error {
	if index < 0 || index >= len(s.children) {
		return &IndexError{Index: index, Size: len(s.children)}
	}
	s.children[index].SetParent(nil)
	child.SetParent(s)
	s.children[index] = child
	return nil
}

// RemoveChild removes the child at the given index.
func (s *Stack) RemoveChild(index int) error {
	if index < 0 || index >= len(s.children) {
		return &IndexError{Index: index, Size: len(s.children)}
	}
	s.children[index].SetParent(nil)
	s.children = append(s.children[:index], s.children[index+1:]...)
	return nil
}

// RangeOfChildAtIndex returns the range of the child at the given index.
// For a Stack, all children start at time 0 and have their own duration.
func (s *Stack) RangeOfChildAtIndex(index int) (opentime.TimeRange, error) {
	if index < 0 || index >= len(s.children) {
		return opentime.TimeRange{}, &IndexError{Index: index, Size: len(s.children)}
	}

	// In a stack, all children start at time 0
	dur, err := s.children[index].Duration()
	if err != nil {
		return opentime.TimeRange{}, err
	}

	// Start time is zero at the same rate as duration
	startTime := opentime.NewRationalTime(0, dur.Rate())
	return opentime.NewTimeRange(startTime, dur), nil
}

// AvailableRange returns the available range of the stack.
// The duration is the maximum duration of all children.
func (s *Stack) AvailableRange() (opentime.TimeRange, error) {
	if len(s.children) == 0 {
		return opentime.TimeRange{}, nil
	}

	// Initialize max duration with first child
	maxDuration, err := s.children[0].Duration()
	if err != nil {
		return opentime.TimeRange{}, err
	}

	for i := 1; i < len(s.children); i++ {
		dur, err := s.children[i].Duration()
		if err != nil {
			return opentime.TimeRange{}, err
		}
		if dur.ToSeconds() > maxDuration.ToSeconds() {
			maxDuration = dur
		}
	}

	startTime := opentime.NewRationalTime(0, maxDuration.Rate())
	return opentime.NewTimeRange(startTime, maxDuration), nil
}

// Duration returns the duration of the stack.
func (s *Stack) Duration() (opentime.RationalTime, error) {
	if s.sourceRange != nil {
		return s.sourceRange.Duration(), nil
	}
	ar, err := s.AvailableRange()
	if err != nil {
		return opentime.RationalTime{}, err
	}
	return ar.Duration(), nil
}

// ChildAtTime returns the child at the given time.
// For a Stack, this returns the topmost (last) child that contains the time.
func (s *Stack) ChildAtTime(searchTime opentime.RationalTime, shallowSearch bool) (Composable, error) {
	// Search from top to bottom (reverse order)
	for i := len(s.children) - 1; i >= 0; i-- {
		child := s.children[i]
		childRange, err := s.RangeOfChildAtIndex(i)
		if err != nil {
			return nil, err
		}
		if childRange.Contains(searchTime) {
			if !shallowSearch {
				if comp, ok := child.(Composition); ok {
					return comp.ChildAtTime(searchTime, false)
				}
			}
			return child, nil
		}
	}
	return nil, nil
}

// ChildrenInRange returns all children within the given range.
func (s *Stack) ChildrenInRange(searchRange opentime.TimeRange) ([]Composable, error) {
	var result []Composable
	for i, child := range s.children {
		childRange, err := s.RangeOfChildAtIndex(i)
		if err != nil {
			return nil, err
		}
		if searchRange.Intersects(childRange, opentime.DefaultEpsilon) {
			result = append(result, child)
		}
	}
	return result, nil
}

// RangeOfAllChildren returns a map of child to range.
func (s *Stack) RangeOfAllChildren() (map[Composable]opentime.TimeRange, error) {
	result := make(map[Composable]opentime.TimeRange)

	for i, child := range s.children {
		dur, err := s.children[i].Duration()
		if err != nil {
			return nil, err
		}
		result[child] = opentime.NewTimeRange(opentime.RationalTime{}, dur)
	}

	return result, nil
}

// AvailableImageBounds returns the union of all clips' image bounds.
func (s *Stack) AvailableImageBounds() (*Box2d, error) {
	var result *Box2d

	for _, child := range s.children {
		var bounds *Box2d
		var err error

		if clip, ok := child.(*Clip); ok {
			bounds, err = clip.AvailableImageBounds()
		} else if track, ok := child.(*Track); ok {
			bounds, err = track.AvailableImageBounds()
		} else if stack, ok := child.(*Stack); ok {
			bounds, err = stack.AvailableImageBounds()
		}

		if err != nil || bounds == nil {
			continue
		}

		if result == nil {
			result = &Box2d{
				Min: bounds.Min,
				Max: bounds.Max,
			}
		} else {
			if bounds.Min.X < result.Min.X {
				result.Min.X = bounds.Min.X
			}
			if bounds.Min.Y < result.Min.Y {
				result.Min.Y = bounds.Min.Y
			}
			if bounds.Max.X > result.Max.X {
				result.Max.X = bounds.Max.X
			}
			if bounds.Max.Y > result.Max.Y {
				result.Max.Y = bounds.Max.Y
			}
		}
	}

	return result, nil
}

// SchemaName returns the schema name.
func (s *Stack) SchemaName() string {
	return StackSchema.Name
}

// SchemaVersion returns the schema version.
func (s *Stack) SchemaVersion() int {
	return StackSchema.Version
}

// Clone creates a deep copy.
func (s *Stack) Clone() SerializableObject {
	clone := &Stack{
		CompositionBase: CompositionBase{
			ItemBase: ItemBase{
				ComposableBase: ComposableBase{
					SerializableObjectWithMetadataBase: SerializableObjectWithMetadataBase{
						name:     s.name,
						metadata: CloneAnyDictionary(s.metadata),
					},
				},
				sourceRange: cloneSourceRange(s.sourceRange),
				effects:     cloneEffects(s.effects),
				markers:     cloneMarkers(s.markers),
				enabled:     s.enabled,
				color:       cloneColor(s.color),
			},
			children: cloneChildren(s.children),
		},
	}
	clone.SetSelf(clone)
	for _, child := range clone.children {
		child.SetParent(clone)
	}
	return clone
}

// IsEquivalentTo returns true if equivalent.
func (s *Stack) IsEquivalentTo(other SerializableObject) bool {
	otherS, ok := other.(*Stack)
	if !ok {
		return false
	}
	if s.name != otherS.name {
		return false
	}
	if len(s.children) != len(otherS.children) {
		return false
	}
	for i := range s.children {
		if !s.children[i].IsEquivalentTo(otherS.children[i]) {
			return false
		}
	}
	return true
}

// MarshalJSON implements json.Marshaler.
func (s *Stack) MarshalJSON() ([]byte, error) {
	j, err := s.marshalCompositionJSON(StackSchema)
	if err != nil {
		return nil, err
	}
	return json.Marshal(j)
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *Stack) UnmarshalJSON(data []byte) error {
	var j compositionJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := s.unmarshalCompositionJSON(&j); err != nil {
		return err
	}
	s.SetSelf(s)
	return nil
}

func init() {
	RegisterSchema(StackSchema, func() SerializableObject {
		return NewStack("", nil, nil, nil, nil, nil)
	})
}
