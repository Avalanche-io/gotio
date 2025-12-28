// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package gotio

import (
	"encoding/json"
)

// LinearTimeWarpSchema is the schema for LinearTimeWarp.
var LinearTimeWarpSchema = Schema{Name: "LinearTimeWarp", Version: 1}

// LinearTimeWarp is a time effect that applies a linear speed change.
type LinearTimeWarp struct {
	EffectBase
	timeScalar float64
}

// NewLinearTimeWarp creates a new LinearTimeWarp.
func NewLinearTimeWarp(name string, effectName string, timeScalar float64, metadata AnyDictionary) *LinearTimeWarp {
	if timeScalar == 0 {
		timeScalar = 1.0
	}
	return &LinearTimeWarp{
		EffectBase: NewEffectBase(name, effectName, metadata),
		timeScalar: timeScalar,
	}
}

// TimeScalar returns the time scalar (speed multiplier).
func (l *LinearTimeWarp) TimeScalar() float64 {
	return l.timeScalar
}

// SetTimeScalar sets the time scalar.
func (l *LinearTimeWarp) SetTimeScalar(scalar float64) {
	l.timeScalar = scalar
}

// SchemaName returns the schema name.
func (l *LinearTimeWarp) SchemaName() string {
	return LinearTimeWarpSchema.Name
}

// SchemaVersion returns the schema version.
func (l *LinearTimeWarp) SchemaVersion() int {
	return LinearTimeWarpSchema.Version
}

// Clone creates a deep copy.
func (l *LinearTimeWarp) Clone() SerializableObject {
	return &LinearTimeWarp{
		EffectBase: EffectBase{
			SerializableObjectWithMetadataBase: SerializableObjectWithMetadataBase{
				name:     l.name,
				metadata: CloneAnyDictionary(l.metadata),
			},
			effectName: l.effectName,
		},
		timeScalar: l.timeScalar,
	}
}

// IsEquivalentTo returns true if equivalent.
func (l *LinearTimeWarp) IsEquivalentTo(other SerializableObject) bool {
	otherL, ok := other.(*LinearTimeWarp)
	if !ok {
		return false
	}
	return l.name == otherL.name && l.effectName == otherL.effectName && l.timeScalar == otherL.timeScalar
}

// linearTimeWarpJSON is the JSON representation.
type linearTimeWarpJSON struct {
	Schema     string        `json:"OTIO_SCHEMA"`
	Name       string        `json:"name"`
	Metadata   AnyDictionary `json:"metadata"`
	EffectName string        `json:"effect_name"`
	TimeScalar float64       `json:"time_scalar"`
}

// MarshalJSON implements json.Marshaler.
func (l *LinearTimeWarp) MarshalJSON() ([]byte, error) {
	return json.Marshal(&linearTimeWarpJSON{
		Schema:     LinearTimeWarpSchema.String(),
		Name:       l.name,
		Metadata:   l.metadata,
		EffectName: l.effectName,
		TimeScalar: l.timeScalar,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (l *LinearTimeWarp) UnmarshalJSON(data []byte) error {
	var j linearTimeWarpJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	l.name = j.Name
	l.metadata = j.Metadata
	if l.metadata == nil {
		l.metadata = make(AnyDictionary)
	}
	l.effectName = j.EffectName
	l.timeScalar = j.TimeScalar
	if l.timeScalar == 0 {
		l.timeScalar = 1.0
	}
	return nil
}

func init() {
	RegisterSchema(LinearTimeWarpSchema, func() SerializableObject {
		return NewLinearTimeWarp("", "", 1.0, nil)
	})
}
