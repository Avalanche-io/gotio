// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"encoding/json"
)

// FreezeFrameSchema is the schema for FreezeFrame.
var FreezeFrameSchema = Schema{Name: "FreezeFrame", Version: 1}

// FreezeFrame is a time effect that holds a single frame.
type FreezeFrame struct {
	EffectBase
}

// NewFreezeFrame creates a new FreezeFrame.
func NewFreezeFrame(name string, metadata AnyDictionary) *FreezeFrame {
	return &FreezeFrame{
		EffectBase: NewEffectBase(name, "FreezeFrame", metadata),
	}
}

// TimeScalar returns 0 for freeze frame (time is frozen).
func (f *FreezeFrame) TimeScalar() float64 {
	return 0
}

// SchemaName returns the schema name.
func (f *FreezeFrame) SchemaName() string {
	return FreezeFrameSchema.Name
}

// SchemaVersion returns the schema version.
func (f *FreezeFrame) SchemaVersion() int {
	return FreezeFrameSchema.Version
}

// Clone creates a deep copy.
func (f *FreezeFrame) Clone() SerializableObject {
	return &FreezeFrame{
		EffectBase: EffectBase{
			SerializableObjectWithMetadataBase: SerializableObjectWithMetadataBase{
				name:     f.name,
				metadata: CloneAnyDictionary(f.metadata),
			},
			effectName: f.effectName,
		},
	}
}

// IsEquivalentTo returns true if equivalent.
func (f *FreezeFrame) IsEquivalentTo(other SerializableObject) bool {
	otherF, ok := other.(*FreezeFrame)
	if !ok {
		return false
	}
	return f.name == otherF.name
}

// MarshalJSON implements json.Marshaler.
func (f *FreezeFrame) MarshalJSON() ([]byte, error) {
	return json.Marshal(&effectJSON{
		Schema:     FreezeFrameSchema.String(),
		Name:       f.name,
		Metadata:   f.metadata,
		EffectName: f.effectName,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (f *FreezeFrame) UnmarshalJSON(data []byte) error {
	var j effectJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	f.name = j.Name
	f.metadata = j.Metadata
	if f.metadata == nil {
		f.metadata = make(AnyDictionary)
	}
	f.effectName = j.EffectName
	if f.effectName == "" {
		f.effectName = "FreezeFrame"
	}
	return nil
}

func init() {
	RegisterSchema(FreezeFrameSchema, func() SerializableObject {
		return NewFreezeFrame("", nil)
	})
}
