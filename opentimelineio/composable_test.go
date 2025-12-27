// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"testing"

	"github.com/mrjoshuak/gotio/opentime"
)

func TestComposableParent(t *testing.T) {
	track := NewTrack("V1", nil, TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	// Initially, clip has no parent
	if clip.Parent() != nil {
		t.Error("Clip should have no parent initially")
	}

	// Add to track
	track.AppendChild(clip)

	// Now clip should have track as parent
	if clip.Parent() != track {
		t.Error("Clip parent should be track")
	}

	// Remove from track
	track.RemoveChild(0)

	// Clip should have no parent again
	if clip.Parent() != nil {
		t.Error("Clip should have no parent after removal")
	}
}

func TestComposableSetParentNil(t *testing.T) {
	track := NewTrack("V1", nil, TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip)
	clip.SetParent(nil)

	if clip.Parent() != nil {
		t.Error("Parent should be nil after SetParent(nil)")
	}
}

func TestSerializableObjectWithMetadataSetMetadata(t *testing.T) {
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)

	// Set metadata
	newMeta := AnyDictionary{"new_key": "new_value", "number": 42}
	clip.SetMetadata(newMeta)

	if clip.Metadata()["new_key"] != "new_value" {
		t.Error("SetMetadata should update metadata")
	}
	if clip.Metadata()["number"] != 42 {
		t.Error("SetMetadata should preserve all values")
	}
}

func TestItemSourceRange(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(10, 24), opentime.NewRationalTime(30, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	// Test SourceRange
	if clip.SourceRange() == nil {
		t.Error("SourceRange should not be nil")
	}
	if clip.SourceRange().StartTime().Value() != 10 {
		t.Errorf("SourceRange start = %v, want 10", clip.SourceRange().StartTime().Value())
	}

	// Test SetSourceRange
	newSr := opentime.NewTimeRange(opentime.NewRationalTime(5, 24), opentime.NewRationalTime(20, 24))
	clip.SetSourceRange(&newSr)
	if clip.SourceRange().StartTime().Value() != 5 {
		t.Errorf("After SetSourceRange, start = %v, want 5", clip.SourceRange().StartTime().Value())
	}

	// Set to nil
	clip.SetSourceRange(nil)
	if clip.SourceRange() != nil {
		t.Error("SourceRange should be nil after setting nil")
	}
}

func TestItemEnabled(t *testing.T) {
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)

	// Default is enabled
	if !clip.Enabled() {
		t.Error("Clip should be enabled by default")
	}

	// Disable
	clip.SetEnabled(false)
	if clip.Enabled() {
		t.Error("Clip should be disabled after SetEnabled(false)")
	}
}

func TestItemColor(t *testing.T) {
	color := NewColorRGB(255, 0, 0)
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", color)

	if clip.ItemColor() == nil {
		t.Error("ItemColor should not be nil")
	}
	if clip.ItemColor().R != 255 {
		t.Errorf("ItemColor R = %v, want 255", clip.ItemColor().R)
	}

	// Set new color
	newColor := NewColorRGB(0, 255, 0)
	clip.SetItemColor(newColor)
	if clip.ItemColor().G != 255 {
		t.Errorf("After SetItemColor, G = %v, want 255", clip.ItemColor().G)
	}

	// Set to nil
	clip.SetItemColor(nil)
	if clip.ItemColor() != nil {
		t.Error("ItemColor should be nil after setting nil")
	}
}

func TestItemEffects(t *testing.T) {
	effect1 := NewEffect("blur", "BlurEffect", nil)
	effect2 := NewEffect("sharpen", "SharpenEffect", nil)
	effects := []Effect{effect1, effect2}

	clip := NewClip("clip", nil, nil, nil, effects, nil, "", nil) // name, ref, sr, meta, effects, markers, key, color

	if len(clip.Effects()) != 2 {
		t.Errorf("Effects count = %d, want 2", len(clip.Effects()))
	}

	// Set new effects
	newEffects := []Effect{NewEffect("glow", "GlowEffect", nil)}
	clip.SetEffects(newEffects)
	if len(clip.Effects()) != 1 {
		t.Errorf("After SetEffects, count = %d, want 1", len(clip.Effects()))
	}
}

func TestItemMarkers(t *testing.T) {
	mr := opentime.NewTimeRange(opentime.NewRationalTime(10, 24), opentime.NewRationalTime(1, 24))
	marker1 := NewMarker("m1", mr, MarkerColorRed, "", nil)
	marker2 := NewMarker("m2", mr, MarkerColorBlue, "", nil)
	markers := []*Marker{marker1, marker2}

	clip := NewClip("clip", nil, nil, nil, nil, markers, "", nil)

	if len(clip.Markers()) != 2 {
		t.Errorf("Markers count = %d, want 2", len(clip.Markers()))
	}

	// Set new markers
	newMarkers := []*Marker{NewMarker("m3", mr, MarkerColorGreen, "", nil)}
	clip.SetMarkers(newMarkers)
	if len(clip.Markers()) != 1 {
		t.Errorf("After SetMarkers, count = %d, want 1", len(clip.Markers()))
	}
}

func TestClipDuration(t *testing.T) {
	// Test with source range
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	dur, err := clip.Duration()
	if err != nil {
		t.Fatalf("Duration error: %v", err)
	}
	if dur.Value() != 48 {
		t.Errorf("Duration = %v, want 48", dur.Value())
	}

	// Test with media reference
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	ref := NewExternalReference("", "/path/file.mov", &ar, nil)
	clip2 := NewClip("clip2", ref, nil, nil, nil, nil, "", nil)

	dur, err = clip2.Duration()
	if err != nil {
		t.Fatalf("Duration error: %v", err)
	}
	if dur.Value() != 100 {
		t.Errorf("Duration from media ref = %v, want 100", dur.Value())
	}
}

func TestClipAvailableRange(t *testing.T) {
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	ref := NewExternalReference("", "/path/file.mov", &ar, nil)
	clip := NewClip("clip", ref, nil, nil, nil, nil, "", nil)

	availRange, err := clip.AvailableRange()
	if err != nil {
		t.Fatalf("AvailableRange error: %v", err)
	}
	// AvailableRange returns a TimeRange value (not pointer), check duration
	if availRange.Duration().Value() != 100 {
		t.Errorf("AvailableRange duration = %v, want 100", availRange.Duration().Value())
	}
}

func TestClipTrimmedRange(t *testing.T) {
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	ref := NewExternalReference("", "/path/file.mov", &ar, nil)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(10, 24), opentime.NewRationalTime(50, 24))
	clip := NewClip("clip", ref, &sr, nil, nil, nil, "", nil)

	trimmed, err := clip.TrimmedRange()
	if err != nil {
		t.Fatalf("TrimmedRange error: %v", err)
	}
	if trimmed.StartTime().Value() != 10 {
		t.Errorf("TrimmedRange start = %v, want 10", trimmed.StartTime().Value())
	}
	if trimmed.Duration().Value() != 50 {
		t.Errorf("TrimmedRange duration = %v, want 50", trimmed.Duration().Value())
	}
}
