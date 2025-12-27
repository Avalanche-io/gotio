// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"encoding/json"

	"github.com/mrjoshuak/gotio/opentime"
)

// MissingReferenceSchema is the schema for MissingReference.
var MissingReferenceSchema = Schema{Name: "MissingReference", Version: 1}

// MissingReference represents media for which a concrete reference is missing.
type MissingReference struct {
	MediaReferenceBase
}

// NewMissingReference creates a new MissingReference.
func NewMissingReference(
	name string,
	availableRange *opentime.TimeRange,
	metadata AnyDictionary,
) *MissingReference {
	return &MissingReference{
		MediaReferenceBase: NewMediaReferenceBase(name, availableRange, metadata, nil),
	}
}

// IsMissingReference returns true for MissingReference.
func (m *MissingReference) IsMissingReference() bool {
	return true
}

// SchemaName returns the schema name.
func (m *MissingReference) SchemaName() string {
	return MissingReferenceSchema.Name
}

// SchemaVersion returns the schema version.
func (m *MissingReference) SchemaVersion() int {
	return MissingReferenceSchema.Version
}

// Clone creates a deep copy.
func (m *MissingReference) Clone() SerializableObject {
	return &MissingReference{
		MediaReferenceBase: MediaReferenceBase{
			SerializableObjectWithMetadataBase: SerializableObjectWithMetadataBase{
				name:     m.name,
				metadata: CloneAnyDictionary(m.metadata),
			},
			availableRange:       cloneAvailableRange(m.availableRange),
			availableImageBounds: cloneBox2d(m.availableImageBounds),
		},
	}
}

// IsEquivalentTo returns true if equivalent.
func (m *MissingReference) IsEquivalentTo(other SerializableObject) bool {
	otherM, ok := other.(*MissingReference)
	if !ok {
		return false
	}
	return m.name == otherM.name && areMetadataEqual(m.metadata, otherM.metadata)
}

// MarshalJSON implements json.Marshaler.
func (m *MissingReference) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.marshalMediaReferenceJSON(MissingReferenceSchema))
}

// UnmarshalJSON implements json.Unmarshaler.
func (m *MissingReference) UnmarshalJSON(data []byte) error {
	var j mediaReferenceJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	m.unmarshalMediaReferenceJSON(&j)
	return nil
}

func init() {
	RegisterSchema(MissingReferenceSchema, func() SerializableObject {
		return NewMissingReference("", nil, nil)
	})
}
