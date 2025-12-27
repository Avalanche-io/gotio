// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"encoding/json"
	"testing"

	"github.com/mrjoshuak/gotio/opentime"
)

func TestTransitionSetters(t *testing.T) {
	transition := NewTransition("dissolve", TransitionTypeSMPTEDissolve,
		opentime.NewRationalTime(12, 24),
		opentime.NewRationalTime(12, 24),
		nil)

	// Test SetTransitionType
	transition.SetTransitionType(TransitionTypeCustom)
	if transition.TransitionType() != TransitionTypeCustom {
		t.Errorf("SetTransitionType: got %s, want %s", transition.TransitionType(), TransitionTypeCustom)
	}

	// Test SetInOffset
	transition.SetInOffset(opentime.NewRationalTime(6, 24))
	if transition.InOffset().Value() != 6 {
		t.Errorf("SetInOffset: got %v, want 6", transition.InOffset().Value())
	}

	// Test SetOutOffset
	transition.SetOutOffset(opentime.NewRationalTime(8, 24))
	if transition.OutOffset().Value() != 8 {
		t.Errorf("SetOutOffset: got %v, want 8", transition.OutOffset().Value())
	}
}

func TestTransitionDuration(t *testing.T) {
	transition := NewTransition("dissolve", TransitionTypeSMPTEDissolve,
		opentime.NewRationalTime(12, 24),
		opentime.NewRationalTime(12, 24),
		nil)

	dur, err := transition.Duration()
	if err != nil {
		t.Fatalf("Duration error: %v", err)
	}
	if dur.Value() != 24 {
		t.Errorf("Duration = %v, want 24", dur.Value())
	}
}

func TestTransitionVisibility(t *testing.T) {
	transition := NewTransition("dissolve", TransitionTypeSMPTEDissolve,
		opentime.NewRationalTime(12, 24),
		opentime.NewRationalTime(12, 24),
		nil)

	if transition.Visible() {
		t.Error("Transition.Visible() should be false")
	}
	if !transition.Overlapping() {
		t.Error("Transition.Overlapping() should be true")
	}
}

func TestTransitionSchema(t *testing.T) {
	transition := NewTransition("", TransitionTypeSMPTEDissolve,
		opentime.RationalTime{}, opentime.RationalTime{}, nil)

	if transition.SchemaName() != "Transition" {
		t.Errorf("SchemaName = %s, want Transition", transition.SchemaName())
	}
	if transition.SchemaVersion() != 1 {
		t.Errorf("SchemaVersion = %d, want 1", transition.SchemaVersion())
	}
}

func TestTransitionClone(t *testing.T) {
	transition := NewTransition("dissolve", TransitionTypeSMPTEDissolve,
		opentime.NewRationalTime(10, 24),
		opentime.NewRationalTime(15, 24),
		AnyDictionary{"key": "value"})

	clone := transition.Clone().(*Transition)

	if clone.Name() != "dissolve" {
		t.Errorf("Clone name = %s, want dissolve", clone.Name())
	}
	if clone.TransitionType() != TransitionTypeSMPTEDissolve {
		t.Errorf("Clone type = %s, want %s", clone.TransitionType(), TransitionTypeSMPTEDissolve)
	}
	if clone.InOffset().Value() != 10 {
		t.Errorf("Clone InOffset = %v, want 10", clone.InOffset().Value())
	}
	if clone.OutOffset().Value() != 15 {
		t.Errorf("Clone OutOffset = %v, want 15", clone.OutOffset().Value())
	}
}

func TestTransitionIsEquivalentTo(t *testing.T) {
	t1 := NewTransition("dissolve", TransitionTypeSMPTEDissolve,
		opentime.NewRationalTime(12, 24),
		opentime.NewRationalTime(12, 24),
		nil)
	t2 := NewTransition("dissolve", TransitionTypeSMPTEDissolve,
		opentime.NewRationalTime(12, 24),
		opentime.NewRationalTime(12, 24),
		nil)
	t3 := NewTransition("wipe", TransitionTypeCustom,
		opentime.NewRationalTime(6, 24),
		opentime.NewRationalTime(6, 24),
		nil)

	if !t1.IsEquivalentTo(t2) {
		t.Error("Identical transitions should be equivalent")
	}
	if t1.IsEquivalentTo(t3) {
		t.Error("Different transitions should not be equivalent")
	}

	// Test with non-Transition
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)
	if t1.IsEquivalentTo(clip) {
		t.Error("Transition should not be equivalent to Clip")
	}
}

func TestTransitionJSON(t *testing.T) {
	transition := NewTransition("dissolve", TransitionTypeSMPTEDissolve,
		opentime.NewRationalTime(12, 24),
		opentime.NewRationalTime(18, 24),
		nil)

	data, err := json.Marshal(transition)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	transition2 := &Transition{}
	if err := json.Unmarshal(data, transition2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if transition2.Name() != "dissolve" {
		t.Errorf("Name mismatch: got %s", transition2.Name())
	}
	if transition2.TransitionType() != TransitionTypeSMPTEDissolve {
		t.Errorf("Type mismatch: got %s", transition2.TransitionType())
	}
	if transition2.InOffset().Value() != 12 {
		t.Errorf("InOffset mismatch: got %v", transition2.InOffset().Value())
	}
	if transition2.OutOffset().Value() != 18 {
		t.Errorf("OutOffset mismatch: got %v", transition2.OutOffset().Value())
	}
}
