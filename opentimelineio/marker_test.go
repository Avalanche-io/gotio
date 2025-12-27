// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"encoding/json"
	"testing"

	"github.com/Avalanche-io/gotio/opentime"
)

func TestMarkerSetters(t *testing.T) {
	mr := opentime.NewTimeRange(opentime.NewRationalTime(10, 24), opentime.NewRationalTime(5, 24))
	marker := NewMarker("marker", mr, MarkerColorRed, "", nil)

	// Test SetMarkedRange
	newRange := opentime.NewTimeRange(opentime.NewRationalTime(20, 24), opentime.NewRationalTime(10, 24))
	marker.SetMarkedRange(newRange)
	if marker.MarkedRange().StartTime().Value() != 20 {
		t.Errorf("SetMarkedRange: got start time %v, want 20", marker.MarkedRange().StartTime().Value())
	}

	// Test SetColor
	marker.SetColor(MarkerColorGreen)
	if marker.Color() != MarkerColorGreen {
		t.Errorf("SetColor: got %s, want %s", marker.Color(), MarkerColorGreen)
	}
}

func TestMarkerSchema(t *testing.T) {
	mr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(1, 24))
	marker := NewMarker("marker", mr, MarkerColorRed, "", nil)

	if marker.SchemaName() != "Marker" {
		t.Errorf("SchemaName = %s, want Marker", marker.SchemaName())
	}
	if marker.SchemaVersion() != 2 {
		t.Errorf("SchemaVersion = %d, want 2", marker.SchemaVersion())
	}
}

func TestMarkerClone(t *testing.T) {
	mr := opentime.NewTimeRange(opentime.NewRationalTime(10, 24), opentime.NewRationalTime(5, 24))
	marker := NewMarker("test_marker", mr, MarkerColorBlue, "", AnyDictionary{"note": "important"})

	clone := marker.Clone().(*Marker)

	if clone.Name() != "test_marker" {
		t.Errorf("Clone name = %s, want test_marker", clone.Name())
	}
	if clone.Color() != MarkerColorBlue {
		t.Errorf("Clone color = %s, want %s", clone.Color(), MarkerColorBlue)
	}
	if clone.MarkedRange().StartTime().Value() != 10 {
		t.Error("Clone marked range mismatch")
	}

	// Verify deep copy
	clone.SetName("modified")
	if marker.Name() == "modified" {
		t.Error("Modifying clone affected original")
	}
}

func TestMarkerIsEquivalentTo(t *testing.T) {
	mr := opentime.NewTimeRange(opentime.NewRationalTime(10, 24), opentime.NewRationalTime(5, 24))
	m1 := NewMarker("marker", mr, MarkerColorRed, "", nil)
	m2 := NewMarker("marker", mr, MarkerColorRed, "", nil)
	m3 := NewMarker("different", mr, MarkerColorGreen, "", nil)

	if !m1.IsEquivalentTo(m2) {
		t.Error("Identical markers should be equivalent")
	}
	if m1.IsEquivalentTo(m3) {
		t.Error("Different markers should not be equivalent")
	}

	// Test with non-Marker
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)
	if m1.IsEquivalentTo(clip) {
		t.Error("Marker should not be equivalent to Clip")
	}
}

func TestMarkerJSON(t *testing.T) {
	mr := opentime.NewTimeRange(opentime.NewRationalTime(10, 24), opentime.NewRationalTime(5, 24))
	marker := NewMarker("test", mr, MarkerColorYellow, "hello", nil)

	data, err := json.Marshal(marker)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	marker2 := &Marker{}
	if err := json.Unmarshal(data, marker2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if marker2.Name() != "test" {
		t.Errorf("Name mismatch: got %s", marker2.Name())
	}
	if marker2.Color() != MarkerColorYellow {
		t.Errorf("Color mismatch: got %s", marker2.Color())
	}
	if marker2.MarkedRange().StartTime().Value() != 10 {
		t.Errorf("MarkedRange start mismatch: got %v", marker2.MarkedRange().StartTime().Value())
	}
}

func TestMarkerColors(t *testing.T) {
	// Test all defined marker colors
	colors := []MarkerColor{
		MarkerColorPink,
		MarkerColorRed,
		MarkerColorOrange,
		MarkerColorYellow,
		MarkerColorGreen,
		MarkerColorCyan,
		MarkerColorBlue,
		MarkerColorPurple,
		MarkerColorMagenta,
		MarkerColorBlack,
		MarkerColorWhite,
	}

	mr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(1, 24))
	for _, color := range colors {
		marker := NewMarker("test", mr, color, "", nil)
		if marker.Color() != color {
			t.Errorf("Color mismatch for %s: got %s", color, marker.Color())
		}
	}
}
