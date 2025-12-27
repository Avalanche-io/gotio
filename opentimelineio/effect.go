// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"encoding/json"
)

// Effect is the interface for effects.
type Effect interface {
	SerializableObjectWithMetadata

	// EffectName returns the effect name.
	EffectName() string

	// SetEffectName sets the effect name.
	SetEffectName(name string)
}

// EffectSchema is the schema for Effect.
var EffectSchema = Schema{Name: "Effect", Version: 1}

// EffectBase is the base implementation of Effect.
type EffectBase struct {
	SerializableObjectWithMetadataBase
	effectName string
}

// NewEffectBase creates a new EffectBase.
func NewEffectBase(name string, effectName string, metadata AnyDictionary) EffectBase {
	return EffectBase{
		SerializableObjectWithMetadataBase: NewSerializableObjectWithMetadataBase(name, metadata),
		effectName:                         effectName,
	}
}

// EffectName returns the effect name.
func (e *EffectBase) EffectName() string {
	return e.effectName
}

// SetEffectName sets the effect name.
func (e *EffectBase) SetEffectName(name string) {
	e.effectName = name
}

// EffectImpl is the standard Effect implementation.
type EffectImpl struct {
	EffectBase
}

// NewEffect creates a new Effect.
func NewEffect(name string, effectName string, metadata AnyDictionary) *EffectImpl {
	return &EffectImpl{
		EffectBase: NewEffectBase(name, effectName, metadata),
	}
}

// SchemaName returns the schema name.
func (e *EffectImpl) SchemaName() string {
	return EffectSchema.Name
}

// SchemaVersion returns the schema version.
func (e *EffectImpl) SchemaVersion() int {
	return EffectSchema.Version
}

// Clone creates a deep copy.
func (e *EffectImpl) Clone() SerializableObject {
	return &EffectImpl{
		EffectBase: EffectBase{
			SerializableObjectWithMetadataBase: SerializableObjectWithMetadataBase{
				name:     e.name,
				metadata: CloneAnyDictionary(e.metadata),
			},
			effectName: e.effectName,
		},
	}
}

// IsEquivalentTo returns true if equivalent.
func (e *EffectImpl) IsEquivalentTo(other SerializableObject) bool {
	otherE, ok := other.(*EffectImpl)
	if !ok {
		return false
	}
	return e.name == otherE.name && e.effectName == otherE.effectName
}

// effectJSON is the JSON representation.
type effectJSON struct {
	Schema     string        `json:"OTIO_SCHEMA"`
	Name       string        `json:"name"`
	Metadata   AnyDictionary `json:"metadata"`
	EffectName string        `json:"effect_name"`
}

// MarshalJSON implements json.Marshaler.
func (e *EffectImpl) MarshalJSON() ([]byte, error) {
	return json.Marshal(&effectJSON{
		Schema:     EffectSchema.String(),
		Name:       e.name,
		Metadata:   e.metadata,
		EffectName: e.effectName,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (e *EffectImpl) UnmarshalJSON(data []byte) error {
	var j effectJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	e.name = j.Name
	e.metadata = j.Metadata
	if e.metadata == nil {
		e.metadata = make(AnyDictionary)
	}
	e.effectName = j.EffectName
	return nil
}

func init() {
	RegisterSchema(EffectSchema, func() SerializableObject {
		return NewEffect("", "", nil)
	})
}
