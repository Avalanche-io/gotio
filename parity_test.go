// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package gotio

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoadSampleOTIOFiles tests loading sample OTIO files from the reference implementation.
func TestLoadSampleOTIOFiles(t *testing.T) {
	// Path to sample data
	sampleDir := "../otio-reference/tests/sample_data"

	// Check if sample directory exists
	if _, err := os.Stat(sampleDir); os.IsNotExist(err) {
		t.Skip("Sample data directory not found, skipping parity tests")
	}

	files, err := filepath.Glob(filepath.Join(sampleDir, "*.otio"))
	if err != nil {
		t.Fatalf("Failed to glob sample files: %v", err)
	}

	if len(files) == 0 {
		t.Skip("No sample OTIO files found")
	}

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			obj, err := FromJSONFile(file)
			if err != nil {
				t.Errorf("Failed to load %s: %v", file, err)
				return
			}

			if obj == nil {
				t.Errorf("Loaded nil object from %s", file)
				return
			}

			// Verify basic structure
			verifySerializableObject(t, obj, file)
		})
	}
}

func verifySerializableObject(t *testing.T, obj SerializableObject, filename string) {
	// Check type
	switch v := obj.(type) {
	case *Timeline:
		verifyTimeline(t, v, filename)
	case *SerializableCollection:
		verifyCollection(t, v, filename)
	case *Clip:
		verifyClip(t, v, filename)
	case *Track:
		verifyTrack(t, v, filename)
	case *Stack:
		verifyStack(t, v, filename)
	default:
		t.Logf("%s: Unknown root type: %T", filename, obj)
	}
}

func verifyTimeline(t *testing.T, timeline *Timeline, filename string) {
	t.Logf("%s: Timeline '%s' with %d tracks", filename, timeline.Name(), len(timeline.Tracks().Children()))

	// Verify tracks
	tracks := timeline.Tracks()
	if tracks == nil {
		t.Errorf("%s: Timeline has nil tracks", filename)
		return
	}

	// Count clips and other items
	clips := timeline.FindClips(nil, false)
	t.Logf("%s: Found %d clips", filename, len(clips))

	// Verify we can compute duration
	dur, err := timeline.Duration()
	if err != nil {
		t.Logf("%s: Could not compute duration: %v", filename, err)
	} else {
		t.Logf("%s: Duration: %v frames at %v fps", filename, dur.Value(), dur.Rate())
	}
}

func verifyCollection(t *testing.T, coll *SerializableCollection, filename string) {
	t.Logf("%s: Collection '%s' with %d children", filename, coll.Name(), len(coll.Children()))
}

func verifyClip(t *testing.T, clip *Clip, filename string) {
	t.Logf("%s: Clip '%s'", filename, clip.Name())

	// Verify source range if set
	if sr := clip.SourceRange(); sr != nil {
		t.Logf("%s: Source range: %v - %v", filename, sr.StartTime().Value(), sr.Duration().Value())
	}
}

func verifyTrack(t *testing.T, track *Track, filename string) {
	t.Logf("%s: Track '%s' (%s) with %d children", filename, track.Name(), track.Kind(), len(track.Children()))
}

func verifyStack(t *testing.T, stack *Stack, filename string) {
	t.Logf("%s: Stack '%s' with %d children", filename, stack.Name(), len(stack.Children()))
}

// TestRoundTripSampleFiles tests that we can load and re-serialize sample files.
func TestRoundTripSampleFiles(t *testing.T) {
	sampleDir := "../otio-reference/tests/sample_data"

	if _, err := os.Stat(sampleDir); os.IsNotExist(err) {
		t.Skip("Sample data directory not found")
	}

	files, err := filepath.Glob(filepath.Join(sampleDir, "*.otio"))
	if err != nil {
		t.Fatalf("Failed to glob: %v", err)
	}

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			// Load
			obj, err := FromJSONFile(file)
			if err != nil {
				t.Skipf("Failed to load: %v", err)
				return
			}

			// Serialize
			data, err := ToJSONBytes(obj)
			if err != nil {
				t.Errorf("Failed to serialize: %v", err)
				return
			}

			// Parse again
			obj2, err := FromJSONBytes(data)
			if err != nil {
				t.Errorf("Failed to re-parse: %v", err)
				return
			}

			// Basic equivalence check
			if obj2 == nil {
				t.Error("Re-parsed object is nil")
				return
			}

			t.Logf("Successfully round-tripped %s (%d bytes)", filepath.Base(file), len(data))
		})
	}
}

// TestClipExample tests loading the clip_example.otio file specifically.
func TestClipExample(t *testing.T) {
	file := "../otio-reference/tests/sample_data/clip_example.otio"

	if _, err := os.Stat(file); os.IsNotExist(err) {
		t.Skip("clip_example.otio not found")
	}

	obj, err := FromJSONFile(file)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	// This is a Timeline with clips inside
	timeline, ok := obj.(*Timeline)
	if !ok {
		t.Fatalf("Expected Timeline, got %T", obj)
	}

	t.Logf("Timeline name: %s", timeline.Name())

	// Find all clips
	clips := timeline.FindClips(nil, false)
	t.Logf("Clips found: %d", len(clips))

	for i, clip := range clips {
		t.Logf("Clip %d: '%s'", i, clip.Name())
	}
}

// TestSimpleCut tests loading the simple_cut.otio file specifically.
func TestSimpleCut(t *testing.T) {
	file := "../otio-reference/tests/sample_data/simple_cut.otio"

	if _, err := os.Stat(file); os.IsNotExist(err) {
		t.Skip("simple_cut.otio not found")
	}

	obj, err := FromJSONFile(file)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	timeline, ok := obj.(*Timeline)
	if !ok {
		t.Fatalf("Expected Timeline, got %T", obj)
	}

	t.Logf("Timeline: %s", timeline.Name())

	tracks := timeline.Tracks()
	if tracks == nil {
		t.Fatal("Tracks is nil")
	}

	t.Logf("Number of tracks: %d", len(tracks.Children()))

	// Find all clips
	clips := timeline.FindClips(nil, false)
	t.Logf("Number of clips: %d", len(clips))

	for i, clip := range clips {
		t.Logf("Clip %d: %s", i, clip.Name())
	}
}

// TestTransitionFile tests loading the transition.otio file.
func TestTransitionFile(t *testing.T) {
	file := "../otio-reference/tests/sample_data/transition.otio"

	if _, err := os.Stat(file); os.IsNotExist(err) {
		t.Skip("transition.otio not found")
	}

	obj, err := FromJSONFile(file)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	timeline, ok := obj.(*Timeline)
	if !ok {
		t.Fatalf("Expected Timeline, got %T", obj)
	}

	// Look for transitions
	tracks := timeline.Tracks()
	for _, child := range tracks.Children() {
		track, ok := child.(*Track)
		if !ok {
			continue
		}

		for _, trackChild := range track.Children() {
			if trans, ok := trackChild.(*Transition); ok {
				t.Logf("Found transition: %s (type: %s)", trans.Name(), trans.TransitionType())
				t.Logf("  In offset: %v", trans.InOffset().Value())
				t.Logf("  Out offset: %v", trans.OutOffset().Value())
			}
		}
	}
}

// TestEffectsFile tests loading the effects.otio file.
func TestEffectsFile(t *testing.T) {
	file := "../otio-reference/tests/sample_data/effects.otio"

	if _, err := os.Stat(file); os.IsNotExist(err) {
		t.Skip("effects.otio not found")
	}

	obj, err := FromJSONFile(file)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	timeline, ok := obj.(*Timeline)
	if !ok {
		t.Fatalf("Expected Timeline, got %T", obj)
	}

	// Look for effects
	tracks := timeline.Tracks()
	for _, child := range tracks.Children() {
		track, ok := child.(*Track)
		if !ok {
			continue
		}

		for _, trackChild := range track.Children() {
			if clip, ok := trackChild.(*Clip); ok {
				effects := clip.Effects()
				if len(effects) > 0 {
					t.Logf("Clip '%s' has %d effects", clip.Name(), len(effects))
					for _, eff := range effects {
						t.Logf("  Effect: %s (type: %s)", eff.Name(), eff.EffectName())
					}
				}
			}
		}
	}
}

// TestMultitrack tests loading the multitrack.otio file.
func TestMultitrackFile(t *testing.T) {
	file := "../otio-reference/tests/sample_data/multitrack.otio"

	if _, err := os.Stat(file); os.IsNotExist(err) {
		t.Skip("multitrack.otio not found")
	}

	obj, err := FromJSONFile(file)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	timeline, ok := obj.(*Timeline)
	if !ok {
		t.Fatalf("Expected Timeline, got %T", obj)
	}

	tracks := timeline.Tracks()
	t.Logf("Number of tracks: %d", len(tracks.Children()))

	for i, child := range tracks.Children() {
		if track, ok := child.(*Track); ok {
			t.Logf("Track %d: '%s' (%s) with %d children", i, track.Name(), track.Kind(), len(track.Children()))
		}
	}
}

// TestNestedExample tests loading the nested_example.otio file.
func TestNestedExampleFile(t *testing.T) {
	file := "../otio-reference/tests/sample_data/nested_example.otio"

	if _, err := os.Stat(file); os.IsNotExist(err) {
		t.Skip("nested_example.otio not found")
	}

	obj, err := FromJSONFile(file)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	timeline, ok := obj.(*Timeline)
	if !ok {
		t.Fatalf("Expected Timeline, got %T", obj)
	}

	// Walk through and look for nested structures
	tracks := timeline.Tracks()
	walkComposition(t, tracks, 0)
}

func walkComposition(t *testing.T, comp Composition, depth int) {
	indent := strings.Repeat("  ", depth)
	t.Logf("%sComposition: %T '%s'", indent, comp, comp.Name())

	for _, child := range comp.Children() {
		switch c := child.(type) {
		case Composition:
			walkComposition(t, c, depth+1)
		case *Clip:
			t.Logf("%s  Clip: '%s'", indent, c.Name())
		case *Gap:
			t.Logf("%s  Gap: '%s'", indent, c.Name())
		case *Transition:
			t.Logf("%s  Transition: '%s'", indent, c.Name())
		default:
			t.Logf("%s  Unknown: %T", indent, child)
		}
	}
}

// TestGeneratorReferenceFile tests loading the generator_reference_test.otio file.
func TestGeneratorReferenceFile(t *testing.T) {
	file := "../otio-reference/tests/sample_data/generator_reference_test.otio"

	if _, err := os.Stat(file); os.IsNotExist(err) {
		t.Skip("generator_reference_test.otio not found")
	}

	obj, err := FromJSONFile(file)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	timeline, ok := obj.(*Timeline)
	if !ok {
		t.Fatalf("Expected Timeline, got %T", obj)
	}

	// Look for generator references
	tracks := timeline.Tracks()
	for _, child := range tracks.Children() {
		track, ok := child.(*Track)
		if !ok {
			continue
		}

		for _, trackChild := range track.Children() {
			if clip, ok := trackChild.(*Clip); ok {
				ref := clip.MediaReference()
				if genRef, ok := ref.(*GeneratorReference); ok {
					t.Logf("Clip '%s' has generator reference:", clip.Name())
					t.Logf("  Generator kind: %s", genRef.GeneratorKind())
				}
			}
		}
	}
}
