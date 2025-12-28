// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package gotio

import (
	"encoding/json"
)

// SerializableCollectionSchema is the schema for SerializableCollection.
var SerializableCollectionSchema = Schema{Name: "SerializableCollection", Version: 1}

// SerializableCollection is a collection of serializable objects.
type SerializableCollection struct {
	SerializableObjectWithMetadataBase
	children []SerializableObject
}

// NewSerializableCollection creates a new SerializableCollection.
func NewSerializableCollection(
	name string,
	children []SerializableObject,
	metadata AnyDictionary,
) *SerializableCollection {
	if children == nil {
		children = make([]SerializableObject, 0)
	}
	return &SerializableCollection{
		SerializableObjectWithMetadataBase: NewSerializableObjectWithMetadataBase(name, metadata),
		children:                           children,
	}
}

// Children returns the children.
func (s *SerializableCollection) Children() []SerializableObject {
	return s.children
}

// SetChildren sets the children.
func (s *SerializableCollection) SetChildren(children []SerializableObject) {
	if children == nil {
		children = make([]SerializableObject, 0)
	}
	s.children = children
}

// AppendChild appends a child.
func (s *SerializableCollection) AppendChild(child SerializableObject) {
	s.children = append(s.children, child)
}

// InsertChild inserts a child at the given index.
func (s *SerializableCollection) InsertChild(index int, child SerializableObject) error {
	if index < 0 || index > len(s.children) {
		return &IndexError{Index: index, Size: len(s.children)}
	}
	s.children = append(s.children[:index], append([]SerializableObject{child}, s.children[index:]...)...)
	return nil
}

// RemoveChild removes the child at the given index.
func (s *SerializableCollection) RemoveChild(index int) error {
	if index < 0 || index >= len(s.children) {
		return &IndexError{Index: index, Size: len(s.children)}
	}
	s.children = append(s.children[:index], s.children[index+1:]...)
	return nil
}

// ClearChildren removes all children.
func (s *SerializableCollection) ClearChildren() {
	s.children = make([]SerializableObject, 0)
}

// FindChildren finds children matching the filter.
func (s *SerializableCollection) FindChildren(filter func(SerializableObject) bool) []SerializableObject {
	var result []SerializableObject
	for _, child := range s.children {
		if filter == nil || filter(child) {
			result = append(result, child)
		}
	}
	return result
}

// SchemaName returns the schema name.
func (s *SerializableCollection) SchemaName() string {
	return SerializableCollectionSchema.Name
}

// SchemaVersion returns the schema version.
func (s *SerializableCollection) SchemaVersion() int {
	return SerializableCollectionSchema.Version
}

// Clone creates a deep copy.
func (s *SerializableCollection) Clone() SerializableObject {
	children := make([]SerializableObject, len(s.children))
	for i, child := range s.children {
		children[i] = child.Clone()
	}
	return &SerializableCollection{
		SerializableObjectWithMetadataBase: SerializableObjectWithMetadataBase{
			name:     s.name,
			metadata: CloneAnyDictionary(s.metadata),
		},
		children: children,
	}
}

// IsEquivalentTo returns true if equivalent.
func (s *SerializableCollection) IsEquivalentTo(other SerializableObject) bool {
	otherS, ok := other.(*SerializableCollection)
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

// serializableCollectionJSON is the JSON representation.
type serializableCollectionJSON struct {
	Schema   string            `json:"OTIO_SCHEMA"`
	Name     string            `json:"name"`
	Metadata AnyDictionary     `json:"metadata"`
	Children []RawMessage `json:"children"`
}

// MarshalJSON implements json.Marshaler.
func (s *SerializableCollection) MarshalJSON() ([]byte, error) {
	children := make([]RawMessage, len(s.children))
	for i, child := range s.children {
		data, err := json.Marshal(child)
		if err != nil {
			return nil, err
		}
		children[i] = data
	}

	return json.Marshal(&serializableCollectionJSON{
		Schema:   SerializableCollectionSchema.String(),
		Name:     s.name,
		Metadata: s.metadata,
		Children: children,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *SerializableCollection) UnmarshalJSON(data []byte) error {
	var j serializableCollectionJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}

	s.name = j.Name
	s.metadata = j.Metadata
	if s.metadata == nil {
		s.metadata = make(AnyDictionary)
	}

	s.children = make([]SerializableObject, len(j.Children))
	for i, data := range j.Children {
		obj, err := FromJSONString(string(data))
		if err != nil {
			return err
		}
		s.children[i] = obj
	}

	return nil
}

func init() {
	RegisterSchema(SerializableCollectionSchema, func() SerializableObject {
		return NewSerializableCollection("", nil, nil)
	})
}
