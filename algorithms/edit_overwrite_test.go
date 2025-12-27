// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package algorithms

import (
	"testing"

	"github.com/mrjoshuak/gotio/opentime"
	"github.com/mrjoshuak/gotio/opentimelineio"
)

func TestOverwriteAppendAfterEnd(t *testing.T) {
	// Track: [A:24] (duration 24)
	// Overwrite at 48-72 with X:24
	// Result: [A:24][Gap:24][X:24]
	track := createTestTrack([]float64{24}, 24)

	sr := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(24, 24),
	)
	newClip := opentimelineio.NewClip("X", nil, &sr, nil, nil, nil, "", nil)

	overwriteRange := opentime.NewTimeRange(
		opentime.NewRationalTime(48, 24),
		opentime.NewRationalTime(24, 24),
	)

	err := Overwrite(newClip, track, overwriteRange)
	if err != nil {
		t.Fatalf("Overwrite failed: %v", err)
	}

	children := track.Children()
	if len(children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(children))
	}

	// Verify structure: [A][Gap][X]
	if children[0].Name() != "clip_A" {
		t.Errorf("child 0: expected clip_A, got %s", children[0].Name())
	}

	if _, ok := children[1].(*opentimelineio.Gap); !ok {
		t.Errorf("child 1: expected Gap, got %T", children[1])
	}
	gapDur, _ := children[1].Duration()
	if gapDur.Value() != 24 {
		t.Errorf("gap duration: expected 24, got %.0f", gapDur.Value())
	}

	if children[2].Name() != "X" {
		t.Errorf("child 2: expected X, got %s", children[2].Name())
	}
}

func TestOverwriteAtEnd(t *testing.T) {
	// Track: [A:24] (duration 24)
	// Overwrite at 24-48 with X:24
	// Result: [A:24][X:24]
	track := createTestTrack([]float64{24}, 24)

	sr := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(24, 24),
	)
	newClip := opentimelineio.NewClip("X", nil, &sr, nil, nil, nil, "", nil)

	overwriteRange := opentime.NewTimeRange(
		opentime.NewRationalTime(24, 24),
		opentime.NewRationalTime(24, 24),
	)

	err := Overwrite(newClip, track, overwriteRange)
	if err != nil {
		t.Fatalf("Overwrite failed: %v", err)
	}

	children := track.Children()
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}

	if children[0].Name() != "clip_A" {
		t.Errorf("child 0: expected clip_A, got %s", children[0].Name())
	}
	if children[1].Name() != "X" {
		t.Errorf("child 1: expected X, got %s", children[1].Name())
	}
}

func TestOverwriteMiddle(t *testing.T) {
	// Track: [A:48] (duration 48)
	// Overwrite at 12-36 with X:24
	// Result: [A:12][X:24][A':12] where A' is the remaining part
	track := createTestTrack([]float64{48}, 24)

	sr := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(24, 24),
	)
	newClip := opentimelineio.NewClip("X", nil, &sr, nil, nil, nil, "", nil)

	overwriteRange := opentime.NewTimeRange(
		opentime.NewRationalTime(12, 24),
		opentime.NewRationalTime(24, 24),
	)

	err := Overwrite(newClip, track, overwriteRange)
	if err != nil {
		t.Fatalf("Overwrite failed: %v", err)
	}

	children := track.Children()
	if len(children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(children))
	}

	// First part of A
	dur0, _ := children[0].Duration()
	if dur0.Value() != 12 {
		t.Errorf("child 0 duration: expected 12, got %.0f", dur0.Value())
	}

	// X
	if children[1].Name() != "X" {
		t.Errorf("child 1: expected X, got %s", children[1].Name())
	}
	dur1, _ := children[1].Duration()
	if dur1.Value() != 24 {
		t.Errorf("child 1 duration: expected 24, got %.0f", dur1.Value())
	}

	// Second part of A
	dur2, _ := children[2].Duration()
	if dur2.Value() != 12 {
		t.Errorf("child 2 duration: expected 12, got %.0f", dur2.Value())
	}
}

func TestOverwriteSpanningMultiple(t *testing.T) {
	// Track: [A:24][B:24][C:24] (duration 72)
	// Overwrite at 12-60 with X:48
	// Result: [A:12][X:48][C':12]
	track := createTestTrack([]float64{24, 24, 24}, 24)

	sr := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(48, 24),
	)
	newClip := opentimelineio.NewClip("X", nil, &sr, nil, nil, nil, "", nil)

	overwriteRange := opentime.NewTimeRange(
		opentime.NewRationalTime(12, 24),
		opentime.NewRationalTime(48, 24),
	)

	err := Overwrite(newClip, track, overwriteRange)
	if err != nil {
		t.Fatalf("Overwrite failed: %v", err)
	}

	children := track.Children()
	if len(children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(children))
	}

	// First part of A
	dur0, _ := children[0].Duration()
	if dur0.Value() != 12 {
		t.Errorf("child 0 duration: expected 12, got %.0f", dur0.Value())
	}

	// X
	if children[1].Name() != "X" {
		t.Errorf("child 1: expected X, got %s", children[1].Name())
	}
	dur1, _ := children[1].Duration()
	if dur1.Value() != 48 {
		t.Errorf("child 1 duration: expected 48, got %.0f", dur1.Value())
	}

	// Remaining part of C
	dur2, _ := children[2].Duration()
	if dur2.Value() != 12 {
		t.Errorf("child 2 duration: expected 12, got %.0f", dur2.Value())
	}
}

func TestOverwriteReplaceEntireClip(t *testing.T) {
	// Track: [A:24][B:24][C:24] (duration 72)
	// Overwrite at 24-48 with X:24 (exactly replacing B)
	// Result: [A:24][X:24][C:24]
	track := createTestTrack([]float64{24, 24, 24}, 24)

	sr := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(24, 24),
	)
	newClip := opentimelineio.NewClip("X", nil, &sr, nil, nil, nil, "", nil)

	overwriteRange := opentime.NewTimeRange(
		opentime.NewRationalTime(24, 24),
		opentime.NewRationalTime(24, 24),
	)

	err := Overwrite(newClip, track, overwriteRange)
	if err != nil {
		t.Fatalf("Overwrite failed: %v", err)
	}

	children := track.Children()
	if len(children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(children))
	}

	if children[0].Name() != "clip_A" {
		t.Errorf("child 0: expected clip_A, got %s", children[0].Name())
	}
	if children[1].Name() != "X" {
		t.Errorf("child 1: expected X, got %s", children[1].Name())
	}
	if children[2].Name() != "clip_C" {
		t.Errorf("child 2: expected clip_C, got %s", children[2].Name())
	}
}

func TestOverwriteEmptyComposition(t *testing.T) {
	track := opentimelineio.NewTrack("empty", nil, opentimelineio.TrackKindVideo, nil, nil)

	sr := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(24, 24),
	)
	newClip := opentimelineio.NewClip("X", nil, &sr, nil, nil, nil, "", nil)

	overwriteRange := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(24, 24),
	)

	err := Overwrite(newClip, track, overwriteRange)
	if err != nil {
		t.Fatalf("Overwrite failed: %v", err)
	}

	children := track.Children()
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}

	if children[0].Name() != "X" {
		t.Errorf("child 0: expected X, got %s", children[0].Name())
	}
}

func TestOverwriteEmptyCompositionWithGap(t *testing.T) {
	track := opentimelineio.NewTrack("empty", nil, opentimelineio.TrackKindVideo, nil, nil)

	sr := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(24, 24),
	)
	newClip := opentimelineio.NewClip("X", nil, &sr, nil, nil, nil, "", nil)

	// Start at frame 24, not 0
	overwriteRange := opentime.NewTimeRange(
		opentime.NewRationalTime(24, 24),
		opentime.NewRationalTime(24, 24),
	)

	err := Overwrite(newClip, track, overwriteRange)
	if err != nil {
		t.Fatalf("Overwrite failed: %v", err)
	}

	children := track.Children()
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}

	if _, ok := children[0].(*opentimelineio.Gap); !ok {
		t.Errorf("child 0: expected Gap, got %T", children[0])
	}
	gapDur, _ := children[0].Duration()
	if gapDur.Value() != 24 {
		t.Errorf("gap duration: expected 24, got %.0f", gapDur.Value())
	}

	if children[1].Name() != "X" {
		t.Errorf("child 1: expected X, got %s", children[1].Name())
	}
}

func TestOverwritePreservesTrackDuration(t *testing.T) {
	// Track: [A:24][B:24][C:24] (duration 72)
	// Overwrite at 24-48 with X:24
	// Track duration should remain 72
	track := createTestTrack([]float64{24, 24, 24}, 24)

	sr := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(24, 24),
	)
	newClip := opentimelineio.NewClip("X", nil, &sr, nil, nil, nil, "", nil)

	overwriteRange := opentime.NewTimeRange(
		opentime.NewRationalTime(24, 24),
		opentime.NewRationalTime(24, 24),
	)

	beforeDur, _ := compositionDuration(track)

	err := Overwrite(newClip, track, overwriteRange)
	if err != nil {
		t.Fatalf("Overwrite failed: %v", err)
	}

	afterDur, _ := compositionDuration(track)

	if !beforeDur.Equal(afterDur) {
		t.Errorf("duration changed: before=%.0f, after=%.0f", beforeDur.Value(), afterDur.Value())
	}
}

func TestOverwriteAtStart(t *testing.T) {
	// Track: [A:48] (duration 48)
	// Overwrite at 0-24 with X:24
	// Result: [X:24][A':24]
	track := createTestTrack([]float64{48}, 24)

	sr := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(24, 24),
	)
	newClip := opentimelineio.NewClip("X", nil, &sr, nil, nil, nil, "", nil)

	overwriteRange := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(24, 24),
	)

	err := Overwrite(newClip, track, overwriteRange)
	if err != nil {
		t.Fatalf("Overwrite failed: %v", err)
	}

	children := track.Children()
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}

	if children[0].Name() != "X" {
		t.Errorf("child 0: expected X, got %s", children[0].Name())
	}
	dur0, _ := children[0].Duration()
	if dur0.Value() != 24 {
		t.Errorf("child 0 duration: expected 24, got %.0f", dur0.Value())
	}

	// Remaining part of A
	dur1, _ := children[1].Duration()
	if dur1.Value() != 24 {
		t.Errorf("child 1 duration: expected 24, got %.0f", dur1.Value())
	}
}
