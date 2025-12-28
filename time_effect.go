// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package gotio

import (
	"encoding/json"
)

// TimeEffect is an effect that modifies time.
type TimeEffect interface {
	Effect
}

// TimeEffectSchema is the schema for TimeEffect.
var TimeEffectSchema = Schema{Name: "TimeEffect", Version: 1}

// TimeEffectImpl is the base implementation of TimeEffect.
type TimeEffectImpl struct {
	EffectBase
}

// NewTimeEffect creates a new TimeEffect.
func NewTimeEffect(name string, effectName string, metadata AnyDictionary) *TimeEffectImpl {
	return &TimeEffectImpl{
		EffectBase: NewEffectBase(name, effectName, metadata),
	}
}

// SchemaName returns the schema name.
func (t *TimeEffectImpl) SchemaName() string {
	return TimeEffectSchema.Name
}

// SchemaVersion returns the schema version.
func (t *TimeEffectImpl) SchemaVersion() int {
	return TimeEffectSchema.Version
}

// Clone creates a deep copy.
func (t *TimeEffectImpl) Clone() SerializableObject {
	return &TimeEffectImpl{
		EffectBase: EffectBase{
			SerializableObjectWithMetadataBase: SerializableObjectWithMetadataBase{
				name:     t.name,
				metadata: CloneAnyDictionary(t.metadata),
			},
			effectName: t.effectName,
		},
	}
}

// IsEquivalentTo returns true if equivalent.
func (t *TimeEffectImpl) IsEquivalentTo(other SerializableObject) bool {
	otherT, ok := other.(*TimeEffectImpl)
	if !ok {
		return false
	}
	return t.name == otherT.name && t.effectName == otherT.effectName
}

// MarshalJSON implements json.Marshaler.
func (t *TimeEffectImpl) MarshalJSON() ([]byte, error) {
	return json.Marshal(&effectJSON{
		Schema:     TimeEffectSchema.String(),
		Name:       t.name,
		Metadata:   t.metadata,
		EffectName: t.effectName,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *TimeEffectImpl) UnmarshalJSON(data []byte) error {
	var j effectJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	t.name = j.Name
	t.metadata = j.Metadata
	if t.metadata == nil {
		t.metadata = make(AnyDictionary)
	}
	t.effectName = j.EffectName
	return nil
}

func init() {
	RegisterSchema(TimeEffectSchema, func() SerializableObject {
		return NewTimeEffect("", "", nil)
	})
}
