// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package gotio

import (
	"encoding/json"
)

// UnknownSchema represents an object with an unregistered schema.
// It preserves all JSON data for round-trip serialization.
type UnknownSchema struct {
	schema string
	data   map[string]any
}

// NewUnknownSchema creates a new UnknownSchema.
func NewUnknownSchema(schema string, data AnyDictionary) *UnknownSchema {
	d := make(map[string]any)
	for k, v := range data {
		d[k] = v
	}
	return &UnknownSchema{
		schema: schema,
		data:   d,
	}
}

// SchemaName returns the original schema name.
func (u *UnknownSchema) SchemaName() string {
	name, _, _ := ParseSchema(u.schema)
	return name
}

// SchemaVersion returns the original schema version.
func (u *UnknownSchema) SchemaVersion() int {
	_, version, _ := ParseSchema(u.schema)
	return version
}

// OriginalSchema returns the original schema string.
func (u *UnknownSchema) OriginalSchema() string {
	return u.schema
}

// Data returns the preserved JSON data.
func (u *UnknownSchema) Data() map[string]any {
	return u.data
}

// Clone creates a deep copy.
func (u *UnknownSchema) Clone() SerializableObject {
	clone := &UnknownSchema{
		schema: u.schema,
		data:   make(map[string]any),
	}
	for k, v := range u.data {
		clone.data[k] = v
	}
	return clone
}

// IsEquivalentTo returns true if equivalent.
func (u *UnknownSchema) IsEquivalentTo(other SerializableObject) bool {
	otherU, ok := other.(*UnknownSchema)
	if !ok {
		return false
	}
	if u.schema != otherU.schema {
		return false
	}
	if len(u.data) != len(otherU.data) {
		return false
	}
	// Simple comparison - not deep
	for k, v := range u.data {
		if otherV, ok := otherU.data[k]; !ok || v != otherV {
			return false
		}
	}
	return true
}

// MarshalJSON implements json.Marshaler.
func (u *UnknownSchema) MarshalJSON() ([]byte, error) {
	result := make(map[string]any)
	for k, v := range u.data {
		result[k] = v
	}
	result["OTIO_SCHEMA"] = u.schema
	return json.Marshal(result)
}

// UnmarshalJSON implements json.Unmarshaler.
func (u *UnknownSchema) UnmarshalJSON(data []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if schema, ok := raw["OTIO_SCHEMA"].(string); ok {
		u.schema = schema
	}

	u.data = raw
	return nil
}
