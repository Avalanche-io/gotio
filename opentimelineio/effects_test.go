// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"encoding/json"
	"testing"
)

func TestEffectSetters(t *testing.T) {
	effect := NewEffect("blur", "BlurEffect", nil)

	// Test SetEffectName
	effect.SetEffectName("NewBlur")
	if effect.EffectName() != "NewBlur" {
		t.Errorf("SetEffectName: got %s, want NewBlur", effect.EffectName())
	}
}

func TestTimeEffectBasics(t *testing.T) {
	te := NewTimeEffect("speed", "SpeedEffect", nil)

	// Test schema
	if te.SchemaName() != "TimeEffect" {
		t.Errorf("SchemaName = %s, want TimeEffect", te.SchemaName())
	}
	if te.SchemaVersion() != 1 {
		t.Errorf("SchemaVersion = %d, want 1", te.SchemaVersion())
	}
}

func TestTimeEffectClone(t *testing.T) {
	te := NewTimeEffect("speed", "SpeedEffect", AnyDictionary{"factor": 2.0})

	clone := te.Clone().(*TimeEffectImpl)

	if clone.Name() != "speed" {
		t.Errorf("Clone name = %s, want speed", clone.Name())
	}
	if clone.EffectName() != "SpeedEffect" {
		t.Errorf("Clone effect name = %s, want SpeedEffect", clone.EffectName())
	}
}

func TestTimeEffectIsEquivalentTo(t *testing.T) {
	te1 := NewTimeEffect("speed", "SpeedEffect", nil)
	te2 := NewTimeEffect("speed", "SpeedEffect", nil)
	te3 := NewTimeEffect("slow", "SlowEffect", nil)

	if !te1.IsEquivalentTo(te2) {
		t.Error("Identical time effects should be equivalent")
	}
	if te1.IsEquivalentTo(te3) {
		t.Error("Different time effects should not be equivalent")
	}

	// Test with different type
	effect := NewEffect("effect", "Effect", nil)
	if te1.IsEquivalentTo(effect) {
		t.Error("TimeEffect should not be equivalent to Effect")
	}
}

func TestTimeEffectJSON(t *testing.T) {
	te := NewTimeEffect("speed", "SpeedEffect", nil)

	data, err := json.Marshal(te)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	te2 := &TimeEffectImpl{}
	if err := json.Unmarshal(data, te2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if te2.EffectName() != "SpeedEffect" {
		t.Errorf("EffectName mismatch: got %s", te2.EffectName())
	}
}

func TestLinearTimeWarpSetters(t *testing.T) {
	ltw := NewLinearTimeWarp("", "", 1.0, nil)

	ltw.SetTimeScalar(2.0)
	if ltw.TimeScalar() != 2.0 {
		t.Errorf("SetTimeScalar: got %f, want 2.0", ltw.TimeScalar())
	}
}

func TestLinearTimeWarpSchema(t *testing.T) {
	ltw := NewLinearTimeWarp("", "", 1.0, nil)

	if ltw.SchemaName() != "LinearTimeWarp" {
		t.Errorf("SchemaName = %s, want LinearTimeWarp", ltw.SchemaName())
	}
	if ltw.SchemaVersion() != 1 {
		t.Errorf("SchemaVersion = %d, want 1", ltw.SchemaVersion())
	}
}

func TestLinearTimeWarpClone(t *testing.T) {
	ltw := NewLinearTimeWarp("warp", "TimeWarp", 2.5, AnyDictionary{"key": "value"})

	clone := ltw.Clone().(*LinearTimeWarp)

	if clone.Name() != "warp" {
		t.Errorf("Clone name = %s, want warp", clone.Name())
	}
	if clone.TimeScalar() != 2.5 {
		t.Errorf("Clone time scalar = %f, want 2.5", clone.TimeScalar())
	}
}

func TestLinearTimeWarpJSON(t *testing.T) {
	ltw := NewLinearTimeWarp("warp", "TimeWarp", 0.5, nil)

	data, err := json.Marshal(ltw)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	ltw2 := &LinearTimeWarp{}
	if err := json.Unmarshal(data, ltw2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if ltw2.TimeScalar() != 0.5 {
		t.Errorf("TimeScalar mismatch: got %f", ltw2.TimeScalar())
	}
}

func TestFreezeFrameSchema(t *testing.T) {
	ff := NewFreezeFrame("freeze", nil)

	if ff.SchemaName() != "FreezeFrame" {
		t.Errorf("SchemaName = %s, want FreezeFrame", ff.SchemaName())
	}
	if ff.SchemaVersion() != 1 {
		t.Errorf("SchemaVersion = %d, want 1", ff.SchemaVersion())
	}
}

func TestFreezeFrameClone(t *testing.T) {
	ff := NewFreezeFrame("freeze", AnyDictionary{"frame": 10})

	clone := ff.Clone().(*FreezeFrame)

	if clone.Name() != "freeze" {
		t.Errorf("Clone name = %s, want freeze", clone.Name())
	}
	if clone.TimeScalar() != 0 {
		t.Errorf("Clone time scalar = %f, want 0", clone.TimeScalar())
	}
}

func TestFreezeFrameIsEquivalentTo(t *testing.T) {
	ff1 := NewFreezeFrame("freeze", nil)
	ff2 := NewFreezeFrame("freeze", nil)
	ff3 := NewFreezeFrame("different", nil)

	if !ff1.IsEquivalentTo(ff2) {
		t.Error("Identical freeze frames should be equivalent")
	}
	if ff1.IsEquivalentTo(ff3) {
		t.Error("Different freeze frames should not be equivalent")
	}
}

func TestFreezeFrameJSON(t *testing.T) {
	ff := NewFreezeFrame("freeze", nil)

	data, err := json.Marshal(ff)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	ff2 := &FreezeFrame{}
	if err := json.Unmarshal(data, ff2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if ff2.Name() != "freeze" {
		t.Errorf("Name mismatch: got %s", ff2.Name())
	}
	if ff2.TimeScalar() != 0 {
		t.Errorf("TimeScalar should be 0, got %f", ff2.TimeScalar())
	}
}

func TestEffectClone(t *testing.T) {
	effect := NewEffect("blur", "BlurEffect", AnyDictionary{"strength": 10})

	clone := effect.Clone().(*EffectImpl)

	if clone.Name() != "blur" {
		t.Errorf("Clone name = %s, want blur", clone.Name())
	}
	if clone.EffectName() != "BlurEffect" {
		t.Errorf("Clone effect name = %s, want BlurEffect", clone.EffectName())
	}
}

func TestEffectIsEquivalentTo(t *testing.T) {
	e1 := NewEffect("blur", "BlurEffect", nil)
	e2 := NewEffect("blur", "BlurEffect", nil)
	e3 := NewEffect("sharpen", "SharpenEffect", nil)

	if !e1.IsEquivalentTo(e2) {
		t.Error("Identical effects should be equivalent")
	}
	if e1.IsEquivalentTo(e3) {
		t.Error("Different effects should not be equivalent")
	}
}

func TestEffectJSON(t *testing.T) {
	effect := NewEffect("blur", "BlurEffect", nil)

	data, err := json.Marshal(effect)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	effect2 := &EffectImpl{}
	if err := json.Unmarshal(data, effect2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if effect2.EffectName() != "BlurEffect" {
		t.Errorf("EffectName mismatch: got %s", effect2.EffectName())
	}
}
