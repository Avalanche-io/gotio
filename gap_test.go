// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package gotio

import (
	"encoding/json"
	"testing"

	"github.com/Avalanche-io/gotio/opentime"
)

func TestGapAvailableRange(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	gap := NewGap("gap", &sr, nil, nil, nil, nil)

	ar, err := gap.AvailableRange()
	if err != nil {
		t.Fatalf("AvailableRange error: %v", err)
	}
	if ar.Duration().Value() != 48 {
		t.Errorf("AvailableRange Duration = %v, want 48", ar.Duration().Value())
	}
}

func TestGapSchema(t *testing.T) {
	gap := NewGapWithDuration(opentime.NewRationalTime(24, 24))

	if gap.SchemaName() != "Gap" {
		t.Errorf("SchemaName = %s, want Gap", gap.SchemaName())
	}
	if gap.SchemaVersion() != 1 {
		t.Errorf("SchemaVersion = %d, want 1", gap.SchemaVersion())
	}
}

func TestGapClone(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(36, 24))
	gap := NewGap("test_gap", &sr, AnyDictionary{"key": "value"}, nil, nil, nil)

	clone := gap.Clone().(*Gap)

	if clone.Name() != "test_gap" {
		t.Errorf("Clone name = %s, want test_gap", clone.Name())
	}
	dur, _ := clone.Duration()
	if dur.Value() != 36 {
		t.Errorf("Clone duration = %v, want 36", dur.Value())
	}
}

func TestGapIsEquivalentTo(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	g1 := NewGap("gap", &sr, nil, nil, nil, nil)
	g2 := NewGap("gap", &sr, nil, nil, nil, nil)

	sr2 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	g3 := NewGap("different", &sr2, nil, nil, nil, nil)

	if !g1.IsEquivalentTo(g2) {
		t.Error("Identical gaps should be equivalent")
	}
	if g1.IsEquivalentTo(g3) {
		t.Error("Different gaps should not be equivalent")
	}

	// Test with non-Gap
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)
	if g1.IsEquivalentTo(clip) {
		t.Error("Gap should not be equivalent to Clip")
	}
}

func TestGapVisibility(t *testing.T) {
	gap := NewGapWithDuration(opentime.NewRationalTime(24, 24))

	if !gap.Visible() {
		t.Error("Gap.Visible() should be true")
	}
	if gap.Overlapping() {
		t.Error("Gap.Overlapping() should be false")
	}
}

func TestGapJSON(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	gap := NewGap("test_gap", &sr, AnyDictionary{"note": "empty space"}, nil, nil, nil)

	data, err := json.Marshal(gap)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	gap2 := &Gap{}
	if err := json.Unmarshal(data, gap2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if gap2.Name() != "test_gap" {
		t.Errorf("Name mismatch: got %s", gap2.Name())
	}
	dur, _ := gap2.Duration()
	if dur.Value() != 48 {
		t.Errorf("Duration mismatch: got %v", dur.Value())
	}
}

func TestGapDuration(t *testing.T) {
	// Test with source range
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 30), opentime.NewRationalTime(60, 30))
	gap := NewGap("", &sr, nil, nil, nil, nil)

	dur, err := gap.Duration()
	if err != nil {
		t.Fatalf("Duration error: %v", err)
	}
	if dur.Value() != 60 || dur.Rate() != 30 {
		t.Errorf("Duration = %v, want 60@30fps", dur)
	}
}
