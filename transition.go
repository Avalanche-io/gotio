// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package gotio

import (
	"encoding/json"

	"github.com/Avalanche-io/gotio/opentime"
)

// TransitionType defines types of transitions.
type TransitionType string

const (
	// TransitionTypeSMPTEDissolve is a SMPTE dissolve transition.
	TransitionTypeSMPTEDissolve TransitionType = "SMPTE_Dissolve"
	// TransitionTypeCustom is a custom transition.
	TransitionTypeCustom TransitionType = "Custom_Transition"
)

// TransitionSchema is the schema for Transition.
var TransitionSchema = Schema{Name: "Transition", Version: 1}

// Transition represents a transition between two adjacent items in a track.
type Transition struct {
	ComposableBase
	transitionType TransitionType
	inOffset       opentime.RationalTime
	outOffset      opentime.RationalTime
}

// NewTransition creates a new Transition.
func NewTransition(
	name string,
	transitionType TransitionType,
	inOffset opentime.RationalTime,
	outOffset opentime.RationalTime,
	metadata AnyDictionary,
) *Transition {
	tr := &Transition{
		ComposableBase: NewComposableBase(name, metadata),
		transitionType: transitionType,
		inOffset:       inOffset,
		outOffset:      outOffset,
	}
	tr.SetSelf(tr)
	return tr
}

// TransitionType returns the transition type.
func (t *Transition) TransitionType() TransitionType {
	return t.transitionType
}

// SetTransitionType sets the transition type.
func (t *Transition) SetTransitionType(transitionType TransitionType) {
	t.transitionType = transitionType
}

// InOffset returns the in offset.
func (t *Transition) InOffset() opentime.RationalTime {
	return t.inOffset
}

// SetInOffset sets the in offset.
func (t *Transition) SetInOffset(inOffset opentime.RationalTime) {
	t.inOffset = inOffset
}

// OutOffset returns the out offset.
func (t *Transition) OutOffset() opentime.RationalTime {
	return t.outOffset
}

// SetOutOffset sets the out offset.
func (t *Transition) SetOutOffset(outOffset opentime.RationalTime) {
	t.outOffset = outOffset
}

// Duration returns the duration of the transition.
func (t *Transition) Duration() (opentime.RationalTime, error) {
	return t.inOffset.Add(t.outOffset), nil
}

// Visible returns false for transitions (they don't take up space).
func (t *Transition) Visible() bool {
	return false
}

// Overlapping returns true for transitions.
func (t *Transition) Overlapping() bool {
	return true
}

// SchemaName returns the schema name.
func (t *Transition) SchemaName() string {
	return TransitionSchema.Name
}

// SchemaVersion returns the schema version.
func (t *Transition) SchemaVersion() int {
	return TransitionSchema.Version
}

// Clone creates a deep copy.
func (t *Transition) Clone() SerializableObject {
	clone := &Transition{
		ComposableBase: ComposableBase{
			SerializableObjectWithMetadataBase: SerializableObjectWithMetadataBase{
				name:     t.name,
				metadata: CloneAnyDictionary(t.metadata),
			},
		},
		transitionType: t.transitionType,
		inOffset:       t.inOffset,
		outOffset:      t.outOffset,
	}
	clone.SetSelf(clone)
	return clone
}

// IsEquivalentTo returns true if equivalent.
func (t *Transition) IsEquivalentTo(other SerializableObject) bool {
	otherT, ok := other.(*Transition)
	if !ok {
		return false
	}
	return t.name == otherT.name &&
		t.transitionType == otherT.transitionType &&
		t.inOffset.Equal(otherT.inOffset) &&
		t.outOffset.Equal(otherT.outOffset)
}

// transitionJSON is the JSON representation.
type transitionJSON struct {
	Schema         string                `json:"OTIO_SCHEMA"`
	Name           string                `json:"name"`
	Metadata       AnyDictionary         `json:"metadata"`
	TransitionType TransitionType        `json:"transition_type"`
	InOffset       opentime.RationalTime `json:"in_offset"`
	OutOffset      opentime.RationalTime `json:"out_offset"`
}

// MarshalJSON implements json.Marshaler.
func (t *Transition) MarshalJSON() ([]byte, error) {
	return json.Marshal(&transitionJSON{
		Schema:         TransitionSchema.String(),
		Name:           t.name,
		Metadata:       t.metadata,
		TransitionType: t.transitionType,
		InOffset:       t.inOffset,
		OutOffset:      t.outOffset,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *Transition) UnmarshalJSON(data []byte) error {
	var j transitionJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	t.name = j.Name
	t.metadata = j.Metadata
	if t.metadata == nil {
		t.metadata = make(AnyDictionary)
	}
	t.transitionType = j.TransitionType
	t.inOffset = j.InOffset
	t.outOffset = j.OutOffset
	t.SetSelf(t)
	return nil
}

func init() {
	RegisterSchema(TransitionSchema, func() SerializableObject {
		return NewTransition("", TransitionTypeSMPTEDissolve, opentime.RationalTime{}, opentime.RationalTime{}, nil)
	})
}
