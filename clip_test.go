// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package gotio

import (
	"encoding/json"
	"testing"

	"github.com/Avalanche-io/gotio/opentime"
)

func TestClipSetMediaReference(t *testing.T) {
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)

	// Initially has missing reference (default behavior)
	if clip.MediaReference() == nil {
		t.Error("Clip should have a default missing reference")
	}

	// Set a media reference
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	ref := NewExternalReference("", "/path/to/file.mov", &ar, nil)
	clip.SetMediaReference(ref)

	if clip.MediaReference() == nil {
		t.Error("MediaReference should not be nil after setting")
	}
	extRef, ok := clip.MediaReference().(*ExternalReference)
	if !ok {
		t.Error("MediaReference should be ExternalReference")
	}
	if extRef.TargetURL() != "/path/to/file.mov" {
		t.Error("MediaReference URL mismatch")
	}

	// Set to nil (will set MissingReference)
	clip.SetMediaReference(nil)
	if clip.MediaReference() == nil {
		t.Error("MediaReference should not be nil (should be MissingReference)")
	}
}

func TestClipActiveMediaReferenceKey(t *testing.T) {
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	mainRef := NewExternalReference("main", "/path/main.mov", &ar, nil)
	proxyRef := NewExternalReference("proxy", "/path/proxy.mov", &ar, nil)

	clip := NewClip("clip", mainRef, nil, nil, nil, nil, "main", nil)

	// Set up multiple media references
	clip.SetMediaReferences(map[string]MediaReference{
		"main":  mainRef,
		"proxy": proxyRef,
	}, "main")

	if clip.ActiveMediaReferenceKey() != "main" {
		t.Errorf("ActiveMediaReferenceKey = %s, want main", clip.ActiveMediaReferenceKey())
	}

	if err := clip.SetActiveMediaReferenceKey("proxy"); err != nil {
		t.Errorf("SetActiveMediaReferenceKey error: %v", err)
	}
	if clip.ActiveMediaReferenceKey() != "proxy" {
		t.Errorf("After SetActiveMediaReferenceKey = %s, want proxy", clip.ActiveMediaReferenceKey())
	}
}

func TestClipMediaReferences(t *testing.T) {
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	mainRef := NewExternalReference("main", "/path/to/main.mov", &ar, nil)
	proxyRef := NewExternalReference("proxy", "/path/to/proxy.mov", &ar, nil)

	clip := NewClip("clip", mainRef, nil, nil, nil, nil, "main", nil)

	// Set multiple references
	mediaRefs := map[string]MediaReference{
		"main":  mainRef,
		"proxy": proxyRef,
	}
	clip.SetMediaReferences(mediaRefs, "main")

	refs := clip.MediaReferences()
	if len(refs) != 2 {
		t.Errorf("MediaReferences count = %d, want 2", len(refs))
	}

	// The active reference should be main
	if clip.MediaReference() != mainRef {
		t.Error("MediaReference should be main ref")
	}

	// Switch active
	clip.SetActiveMediaReferenceKey("proxy")
	if clip.MediaReference() != proxyRef {
		t.Error("MediaReference should be proxy ref after switching")
	}

	// Set new references
	newRefs := map[string]MediaReference{
		"new": NewExternalReference("new", "/path/to/new.mov", &ar, nil),
	}
	clip.SetMediaReferences(newRefs, "new")
	if len(clip.MediaReferences()) != 1 {
		t.Errorf("After SetMediaReferences, count = %d, want 1", len(clip.MediaReferences()))
	}
}

func TestClipSchema(t *testing.T) {
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)

	if clip.SchemaName() != "Clip" {
		t.Errorf("SchemaName = %s, want Clip", clip.SchemaName())
	}
	if clip.SchemaVersion() != 2 {
		t.Errorf("SchemaVersion = %d, want 2", clip.SchemaVersion())
	}
}

func TestClipClone(t *testing.T) {
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	ref := NewExternalReference("", "/path/to/file.mov", &ar, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(10, 24), opentime.NewRationalTime(50, 24))
	clip := NewClip("test_clip", ref, &sr, AnyDictionary{"key": "value"}, nil, nil, "", nil)

	clone := clip.Clone().(*Clip)

	if clone.Name() != "test_clip" {
		t.Errorf("Clone name = %s, want test_clip", clone.Name())
	}
	if clone.MediaReference() == nil {
		t.Error("Clone should have media reference")
	}
	if clone.SourceRange() == nil || clone.SourceRange().StartTime().Value() != 10 {
		t.Error("Clone source range mismatch")
	}

	// Verify deep copy
	clone.SetName("modified")
	if clip.Name() == "modified" {
		t.Error("Modifying clone affected original")
	}
}

func TestClipIsEquivalentTo(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	c1 := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	c2 := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	c3 := NewClip("different", nil, &sr, nil, nil, nil, "", nil)

	if !c1.IsEquivalentTo(c2) {
		t.Error("Identical clips should be equivalent")
	}
	if c1.IsEquivalentTo(c3) {
		t.Error("Different clips should not be equivalent")
	}

	// Test with non-Clip
	gap := NewGapWithDuration(opentime.NewRationalTime(24, 24))
	if c1.IsEquivalentTo(gap) {
		t.Error("Clip should not be equivalent to Gap")
	}
}

func TestClipVisibility(t *testing.T) {
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)

	if !clip.Visible() {
		t.Error("Clip.Visible() should be true")
	}
	if clip.Overlapping() {
		t.Error("Clip.Overlapping() should be false")
	}
}

func TestClipAvailableImageBounds(t *testing.T) {
	// Clip without media reference
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)

	bounds, err := clip.AvailableImageBounds()
	if err != nil {
		t.Fatalf("AvailableImageBounds error: %v", err)
	}
	if bounds != nil {
		t.Error("Clip without external ref should have nil bounds")
	}
}

func TestClipJSONComplete(t *testing.T) {
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	ref := NewExternalReference("", "/path/to/file.mov", &ar, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(10, 24), opentime.NewRationalTime(50, 24))

	mr := opentime.NewTimeRange(opentime.NewRationalTime(5, 24), opentime.NewRationalTime(1, 24))
	marker := NewMarker("marker", mr, MarkerColorRed, "", nil)
	effect := NewEffect("blur", "BlurEffect", nil)

	clip := NewClip("test_clip", ref, &sr, AnyDictionary{"note": "test"}, []Effect{effect}, []*Marker{marker}, "", nil)

	data, err := json.Marshal(clip)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	clip2 := &Clip{}
	if err := json.Unmarshal(data, clip2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if clip2.Name() != "test_clip" {
		t.Errorf("Name mismatch: got %s", clip2.Name())
	}
	if clip2.MediaReference() == nil {
		t.Error("MediaReference should not be nil")
	}
	if clip2.SourceRange() == nil {
		t.Error("SourceRange should not be nil")
	}
	if len(clip2.Markers()) != 1 {
		t.Errorf("Markers count = %d, want 1", len(clip2.Markers()))
	}
	if len(clip2.Effects()) != 1 {
		t.Errorf("Effects count = %d, want 1", len(clip2.Effects()))
	}
}

func TestClipWithMediaReferencesJSON(t *testing.T) {
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	mainRef := NewExternalReference("main", "/path/main.mov", &ar, nil)
	proxyRef := NewExternalReference("proxy", "/path/proxy.mov", &ar, nil)

	mediaRefs := map[string]MediaReference{
		"main":  mainRef,
		"proxy": proxyRef,
	}

	clip := NewClip("clip", mainRef, nil, nil, nil, nil, "main", nil)
	clip.SetMediaReferences(mediaRefs, "main")

	data, err := json.Marshal(clip)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	clip2 := &Clip{}
	if err := json.Unmarshal(data, clip2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if len(clip2.MediaReferences()) != 2 {
		t.Errorf("MediaReferences count = %d, want 2", len(clip2.MediaReferences()))
	}
	if clip2.ActiveMediaReferenceKey() != "main" {
		t.Errorf("ActiveMediaReferenceKey = %s, want main", clip2.ActiveMediaReferenceKey())
	}
}
