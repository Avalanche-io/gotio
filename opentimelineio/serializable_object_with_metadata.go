// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

// SerializableObjectWithMetadata adds name and metadata to SerializableObject.
type SerializableObjectWithMetadata interface {
	SerializableObject

	// Name returns the name.
	Name() string

	// SetName sets the name.
	SetName(name string)

	// Metadata returns the metadata dictionary.
	Metadata() AnyDictionary

	// SetMetadata sets the metadata dictionary.
	SetMetadata(metadata AnyDictionary)
}

// SerializableObjectWithMetadataBase is the base implementation.
type SerializableObjectWithMetadataBase struct {
	name     string
	metadata AnyDictionary
}

// NewSerializableObjectWithMetadataBase creates a new base.
func NewSerializableObjectWithMetadataBase(name string, metadata AnyDictionary) SerializableObjectWithMetadataBase {
	if metadata == nil {
		metadata = make(AnyDictionary)
	}
	return SerializableObjectWithMetadataBase{
		name:     name,
		metadata: metadata,
	}
}

// Name returns the name.
func (s *SerializableObjectWithMetadataBase) Name() string {
	return s.name
}

// SetName sets the name.
func (s *SerializableObjectWithMetadataBase) SetName(name string) {
	s.name = name
}

// Metadata returns the metadata.
func (s *SerializableObjectWithMetadataBase) Metadata() AnyDictionary {
	return s.metadata
}

// SetMetadata sets the metadata.
func (s *SerializableObjectWithMetadataBase) SetMetadata(metadata AnyDictionary) {
	if metadata == nil {
		metadata = make(AnyDictionary)
	}
	s.metadata = metadata
}
