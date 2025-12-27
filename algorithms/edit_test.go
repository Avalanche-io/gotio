// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package algorithms

import (
	"testing"

	"github.com/mrjoshuak/gotio/opentime"
	"github.com/mrjoshuak/gotio/opentimelineio"
)

// ============================================================================
// Insert Tests
// ============================================================================

func TestInsertAtEnd(t *testing.T) {
	// Track: [A:24] -> Insert X at 24 -> [A:24][X:24]
	track := createTestTrack([]float64{24}, 24)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	newClip := opentimelineio.NewClip("X", nil, &sr, nil, nil, nil, "", nil)

	time := opentime.NewRationalTime(24, 24)
	err := Insert(newClip, track, time)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
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

func TestInsertInMiddle(t *testing.T) {
	// Track: [A:48] -> Insert X at 24 -> [A:24][X:24][A':24]
	track := createTestTrack([]float64{48}, 24)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	newClip := opentimelineio.NewClip("X", nil, &sr, nil, nil, nil, "", nil)

	time := opentime.NewRationalTime(24, 24)
	err := Insert(newClip, track, time)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	children := track.Children()
	if len(children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(children))
	}

	// Duration should grow by 24
	dur, _ := compositionDuration(track)
	if dur.Value() != 72 {
		t.Errorf("expected duration 72, got %.0f", dur.Value())
	}
}

func TestInsertAtStart(t *testing.T) {
	// Track: [A:24] -> Insert X at 0 -> [X:24][A:24]
	track := createTestTrack([]float64{24}, 24)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	newClip := opentimelineio.NewClip("X", nil, &sr, nil, nil, nil, "", nil)

	time := opentime.NewRationalTime(0, 24)
	err := Insert(newClip, track, time)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	children := track.Children()
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}

	if children[0].Name() != "X" {
		t.Errorf("child 0: expected X, got %s", children[0].Name())
	}
}

// ============================================================================
// Slice Tests
// ============================================================================

func TestSliceMiddle(t *testing.T) {
	// Track: [A:48] -> Slice at 24 -> [A:24][A':24]
	track := createTestTrack([]float64{48}, 24)

	time := opentime.NewRationalTime(24, 24)
	err := Slice(track, time)
	if err != nil {
		t.Fatalf("Slice failed: %v", err)
	}

	children := track.Children()
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}

	dur0, _ := children[0].Duration()
	dur1, _ := children[1].Duration()
	if dur0.Value() != 24 || dur1.Value() != 24 {
		t.Errorf("expected 24/24, got %.0f/%.0f", dur0.Value(), dur1.Value())
	}

	// Total duration unchanged
	totalDur, _ := compositionDuration(track)
	if totalDur.Value() != 48 {
		t.Errorf("expected total 48, got %.0f", totalDur.Value())
	}
}

func TestSliceAtBoundary(t *testing.T) {
	// Slice at boundary should be no-op
	track := createTestTrack([]float64{24, 24}, 24)

	time := opentime.NewRationalTime(24, 24) // Between clips
	err := Slice(track, time)
	if err != nil {
		t.Fatalf("Slice failed: %v", err)
	}

	children := track.Children()
	if len(children) != 2 {
		t.Fatalf("expected 2 children (no-op), got %d", len(children))
	}
}

// ============================================================================
// Slip Tests
// ============================================================================

func TestSlipForward(t *testing.T) {
	track := createTestTrackWithAvailableRange([]float64{24}, 100, 24)
	item := track.Children()[0].(opentimelineio.Item)

	delta := opentime.NewRationalTime(10, 24)
	err := Slip(item, delta)
	if err != nil {
		t.Fatalf("Slip failed: %v", err)
	}

	sr := item.SourceRange()
	if sr.StartTime().Value() != 10 {
		t.Errorf("expected start 10, got %.0f", sr.StartTime().Value())
	}
	if sr.Duration().Value() != 24 {
		t.Errorf("expected duration 24, got %.0f", sr.Duration().Value())
	}
}

func TestSlipBackward(t *testing.T) {
	// Create clip starting at frame 20
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	ref := opentimelineio.NewExternalReference("", "file://test.mov", &ar, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(20, 24), opentime.NewRationalTime(24, 24))
	clip := opentimelineio.NewClip("test", ref, &sr, nil, nil, nil, "", nil)

	delta := opentime.NewRationalTime(-10, 24)
	err := Slip(clip, delta)
	if err != nil {
		t.Fatalf("Slip failed: %v", err)
	}

	newSr := clip.SourceRange()
	if newSr.StartTime().Value() != 10 {
		t.Errorf("expected start 10, got %.0f", newSr.StartTime().Value())
	}
}

func TestSlipClampToAvailable(t *testing.T) {
	track := createTestTrackWithAvailableRange([]float64{24}, 50, 24)
	item := track.Children()[0].(opentimelineio.Item)

	// Try to slip past available range
	delta := opentime.NewRationalTime(40, 24) // Would put end at 64, past 50
	err := Slip(item, delta)
	if err != nil {
		t.Fatalf("Slip failed: %v", err)
	}

	sr := item.SourceRange()
	// Should be clamped so end = 50, start = 50 - 24 = 26
	if sr.StartTime().Value() != 26 {
		t.Errorf("expected start 26, got %.0f", sr.StartTime().Value())
	}
}

// ============================================================================
// Slide Tests
// ============================================================================

func TestSlideRight(t *testing.T) {
	track := createTestTrackWithAvailableRange([]float64{48, 24}, 100, 24)
	children := track.Children()
	secondItem := children[1].(opentimelineio.Item)
	firstItem := children[0].(opentimelineio.Item)

	delta := opentime.NewRationalTime(12, 24)
	err := Slide(secondItem, track, delta)
	if err != nil {
		t.Fatalf("Slide failed: %v", err)
	}

	// Previous item should be longer
	firstDur, _ := firstItem.Duration()
	if firstDur.Value() != 60 {
		t.Errorf("expected first duration 60, got %.0f", firstDur.Value())
	}
}

func TestSlideLeft(t *testing.T) {
	track := createTestTrack([]float64{48, 24}, 24)
	children := track.Children()
	secondItem := children[1].(opentimelineio.Item)
	firstItem := children[0].(opentimelineio.Item)

	delta := opentime.NewRationalTime(-12, 24)
	err := Slide(secondItem, track, delta)
	if err != nil {
		t.Fatalf("Slide failed: %v", err)
	}

	// Previous item should be shorter
	firstDur, _ := firstItem.Duration()
	if firstDur.Value() != 36 {
		t.Errorf("expected first duration 36, got %.0f", firstDur.Value())
	}
}

// ============================================================================
// Trim Tests
// ============================================================================

func TestTrimHead(t *testing.T) {
	track := createTestTrack([]float64{24, 48}, 24)
	children := track.Children()
	secondItem := children[1].(opentimelineio.Item)

	deltaIn := opentime.NewRationalTime(12, 24)
	err := Trim(secondItem, track, deltaIn, opentime.RationalTime{})
	if err != nil {
		t.Fatalf("Trim failed: %v", err)
	}

	// Second item should be shorter
	dur, _ := secondItem.Duration()
	if dur.Value() != 36 {
		t.Errorf("expected duration 36, got %.0f", dur.Value())
	}

	// First item should be longer (compensating)
	firstDur, _ := children[0].Duration()
	if firstDur.Value() != 36 {
		t.Errorf("expected first duration 36, got %.0f", firstDur.Value())
	}
}

func TestTrimTail(t *testing.T) {
	track := createTestTrack([]float64{48, 24}, 24)
	children := track.Children()
	firstItem := children[0].(opentimelineio.Item)

	deltaOut := opentime.NewRationalTime(-12, 24)
	err := Trim(firstItem, track, opentime.RationalTime{}, deltaOut)
	if err != nil {
		t.Fatalf("Trim failed: %v", err)
	}

	// First item should be shorter
	dur, _ := firstItem.Duration()
	if dur.Value() != 36 {
		t.Errorf("expected duration 36, got %.0f", dur.Value())
	}
}

// ============================================================================
// Ripple Tests
// ============================================================================

func TestRippleIn(t *testing.T) {
	track := createTestTrackWithAvailableRange([]float64{48}, 100, 24)
	item := track.Children()[0].(opentimelineio.Item)

	deltaIn := opentime.NewRationalTime(12, 24)
	err := Ripple(item, deltaIn, opentime.RationalTime{})
	if err != nil {
		t.Fatalf("Ripple failed: %v", err)
	}

	sr := item.SourceRange()
	if sr.StartTime().Value() != 12 {
		t.Errorf("expected start 12, got %.0f", sr.StartTime().Value())
	}
}

func TestRippleOut(t *testing.T) {
	track := createTestTrackWithAvailableRange([]float64{48}, 100, 24)
	item := track.Children()[0].(opentimelineio.Item)

	deltaOut := opentime.NewRationalTime(12, 24)
	err := Ripple(item, opentime.RationalTime{}, deltaOut)
	if err != nil {
		t.Fatalf("Ripple failed: %v", err)
	}

	sr := item.SourceRange()
	if sr.Duration().Value() != 60 {
		t.Errorf("expected duration 60, got %.0f", sr.Duration().Value())
	}
}

// ============================================================================
// Roll Tests
// ============================================================================

func TestRollIn(t *testing.T) {
	track := createTestTrack([]float64{48, 48}, 24)
	children := track.Children()
	secondItem := children[1].(opentimelineio.Item)

	deltaIn := opentime.NewRationalTime(12, 24)
	err := Roll(secondItem, track, deltaIn, opentime.RationalTime{})
	if err != nil {
		t.Fatalf("Roll failed: %v", err)
	}

	// Second item's head trimmed, first item extended
	firstDur, _ := children[0].Duration()
	secondDur, _ := children[1].Duration()

	if firstDur.Value() != 60 {
		t.Errorf("expected first duration 60, got %.0f", firstDur.Value())
	}
	if secondDur.Value() != 36 {
		t.Errorf("expected second duration 36, got %.0f", secondDur.Value())
	}

	// Total should be unchanged
	totalDur, _ := compositionDuration(track)
	if totalDur.Value() != 96 {
		t.Errorf("expected total 96, got %.0f", totalDur.Value())
	}
}

func TestRollOut(t *testing.T) {
	track := createTestTrack([]float64{48, 48}, 24)
	children := track.Children()
	firstItem := children[0].(opentimelineio.Item)

	deltaOut := opentime.NewRationalTime(12, 24)
	err := Roll(firstItem, track, opentime.RationalTime{}, deltaOut)
	if err != nil {
		t.Fatalf("Roll failed: %v", err)
	}

	// First item extended, second item's head trimmed
	firstDur, _ := children[0].Duration()
	secondDur, _ := children[1].Duration()

	if firstDur.Value() != 60 {
		t.Errorf("expected first duration 60, got %.0f", firstDur.Value())
	}
	if secondDur.Value() != 36 {
		t.Errorf("expected second duration 36, got %.0f", secondDur.Value())
	}
}

// ============================================================================
// Fill Tests
// ============================================================================

func TestFillSource(t *testing.T) {
	// Create track with gap
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)
	gap := opentimelineio.NewGapWithDuration(opentime.NewRationalTime(48, 24))
	track.AppendChild(gap)

	// Create clip
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := opentimelineio.NewClip("fill", nil, &sr, nil, nil, nil, "", nil)

	// Fill at start with Source mode
	time := opentime.NewRationalTime(0, 24)
	err := Fill(clip, track, time, ReferencePointSource)
	if err != nil {
		t.Fatalf("Fill failed: %v", err)
	}

	children := track.Children()
	// Should have clip and remaining gap
	if len(children) < 1 {
		t.Fatal("expected at least 1 child")
	}

	if children[0].Name() != "fill" {
		t.Errorf("expected fill clip, got %s", children[0].Name())
	}
}

func TestFillSequence(t *testing.T) {
	// Create track with gap
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)
	gap := opentimelineio.NewGapWithDuration(opentime.NewRationalTime(48, 24))
	track.AppendChild(gap)

	// Create clip longer than gap
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(72, 24))
	clip := opentimelineio.NewClip("fill", nil, &sr, nil, nil, nil, "", nil)

	// Fill with Sequence mode (should trim to gap)
	time := opentime.NewRationalTime(0, 24)
	err := Fill(clip, track, time, ReferencePointSequence)
	if err != nil {
		t.Fatalf("Fill failed: %v", err)
	}

	children := track.Children()
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}

	// Should be trimmed to 48
	dur, _ := children[0].Duration()
	if dur.Value() != 48 {
		t.Errorf("expected duration 48, got %.0f", dur.Value())
	}
}

func TestFillFit(t *testing.T) {
	// Create track with gap
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)
	gap := opentimelineio.NewGapWithDuration(opentime.NewRationalTime(48, 24))
	track.AppendChild(gap)

	// Create clip
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := opentimelineio.NewClip("fill", nil, &sr, nil, nil, nil, "", nil)

	// Fill with Fit mode (should add time warp)
	time := opentime.NewRationalTime(0, 24)
	err := Fill(clip, track, time, ReferencePointFit)
	if err != nil {
		t.Fatalf("Fill failed: %v", err)
	}

	children := track.Children()
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}

	// Should have time warp effect
	item := children[0].(opentimelineio.Item)
	effects := item.Effects()
	if len(effects) == 0 {
		t.Error("expected time warp effect")
	}
}

// ============================================================================
// Remove Tests
// ============================================================================

func TestRemoveWithFill(t *testing.T) {
	track := createTestTrack([]float64{24, 24, 24}, 24)

	// Remove middle clip
	time := opentime.NewRationalTime(36, 24) // Middle of second clip
	err := Remove(track, time)
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	children := track.Children()
	if len(children) != 3 {
		t.Fatalf("expected 3 children (with gap), got %d", len(children))
	}

	// Middle should be a gap
	if _, ok := children[1].(*opentimelineio.Gap); !ok {
		t.Errorf("expected gap, got %T", children[1])
	}

	// Duration preserved
	dur, _ := compositionDuration(track)
	if dur.Value() != 72 {
		t.Errorf("expected duration 72, got %.0f", dur.Value())
	}
}

func TestRemoveWithoutFill(t *testing.T) {
	track := createTestTrack([]float64{24, 24, 24}, 24)

	// Remove middle clip without fill
	time := opentime.NewRationalTime(36, 24)
	err := Remove(track, time, WithFill(false))
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	children := track.Children()
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}

	// Duration should shrink
	dur, _ := compositionDuration(track)
	if dur.Value() != 48 {
		t.Errorf("expected duration 48, got %.0f", dur.Value())
	}
}

func TestRemoveRange(t *testing.T) {
	track := createTestTrack([]float64{24, 24, 24, 24}, 24)

	// Remove middle two clips
	timeRange := opentime.NewTimeRange(
		opentime.NewRationalTime(24, 24),
		opentime.NewRationalTime(48, 24),
	)
	err := RemoveRange(track, timeRange)
	if err != nil {
		t.Fatalf("RemoveRange failed: %v", err)
	}

	children := track.Children()
	// Should have: first clip, gap, last clip
	if len(children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(children))
	}

	// Middle should be gap
	if _, ok := children[1].(*opentimelineio.Gap); !ok {
		t.Errorf("expected gap, got %T", children[1])
	}
}

// ============================================================================
// Error Type Tests
// ============================================================================

func TestReferencePointString(t *testing.T) {
	tests := []struct {
		rp       ReferencePoint
		expected string
	}{
		{ReferencePointSource, "Source"},
		{ReferencePointSequence, "Sequence"},
		{ReferencePointFit, "Fit"},
		{ReferencePoint(99), "ReferencePoint(99)"},
	}

	for _, tt := range tests {
		if got := tt.rp.String(); got != tt.expected {
			t.Errorf("ReferencePoint(%d).String() = %s, want %s", tt.rp, got, tt.expected)
		}
	}
}

func TestEditErrorString(t *testing.T) {
	t.Run("basic error", func(t *testing.T) {
		err := &EditError{
			Operation: "overwrite",
			Message:   "failed",
		}
		s := err.Error()
		if s == "" {
			t.Error("expected non-empty error")
		}
	})

	t.Run("with time", func(t *testing.T) {
		time := opentime.NewRationalTime(24, 24)
		err := &EditError{
			Operation: "insert",
			Message:   "invalid time",
			Time:      &time,
		}
		s := err.Error()
		if s == "" {
			t.Error("expected non-empty error")
		}
	})

	t.Run("with item", func(t *testing.T) {
		clip := opentimelineio.NewClip("test", nil, nil, nil, nil, nil, "", nil)
		err := &EditError{
			Operation: "slice",
			Message:   "cannot slice",
			Item:      clip,
		}
		s := err.Error()
		if s == "" {
			t.Error("expected non-empty error")
		}
	})

	t.Run("newEditError", func(t *testing.T) {
		err := newEditError("test_op", "test message")
		if err.Operation != "test_op" {
			t.Errorf("expected operation test_op, got %s", err.Operation)
		}
	})

	t.Run("newEditErrorAt", func(t *testing.T) {
		time := opentime.NewRationalTime(10, 24)
		err := newEditErrorAt("test_op", "test message", time)
		if err.Time == nil {
			t.Error("expected time to be set")
		}
	})

	t.Run("newEditErrorForItem", func(t *testing.T) {
		clip := opentimelineio.NewClip("test", nil, nil, nil, nil, nil, "", nil)
		err := newEditErrorForItem("test_op", "test message", clip)
		if err.Item == nil {
			t.Error("expected item to be set")
		}
	})
}

// ============================================================================
// Option Tests
// ============================================================================

func TestOverwriteWithOptions(t *testing.T) {
	t.Run("WithRemoveTransitions", func(t *testing.T) {
		track := createTestTrack([]float64{48}, 24)
		sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
		clip := opentimelineio.NewClip("new", nil, &sr, nil, nil, nil, "", nil)
		timeRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))

		err := Overwrite(clip, track, timeRange, WithRemoveTransitions(true))
		if err != nil {
			t.Errorf("Overwrite with option failed: %v", err)
		}
	})

	t.Run("WithFillTemplate", func(t *testing.T) {
		track := createTestTrack([]float64{48}, 24)
		sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
		clip := opentimelineio.NewClip("new", nil, &sr, nil, nil, nil, "", nil)
		template := opentimelineio.NewGapWithDuration(opentime.NewRationalTime(1, 24))
		timeRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))

		err := Overwrite(clip, track, timeRange, WithFillTemplate(template))
		if err != nil {
			t.Errorf("Overwrite with option failed: %v", err)
		}
	})
}

func TestInsertWithOptions(t *testing.T) {
	t.Run("WithInsertRemoveTransitions", func(t *testing.T) {
		track := createTestTrack([]float64{24}, 24)
		sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
		clip := opentimelineio.NewClip("new", nil, &sr, nil, nil, nil, "", nil)

		err := Insert(clip, track, opentime.NewRationalTime(12, 24), WithInsertRemoveTransitions(true))
		if err != nil {
			t.Errorf("Insert with option failed: %v", err)
		}
	})

	t.Run("WithInsertFillTemplate", func(t *testing.T) {
		track := createTestTrack([]float64{24}, 24)
		sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
		clip := opentimelineio.NewClip("new", nil, &sr, nil, nil, nil, "", nil)
		template := opentimelineio.NewGapWithDuration(opentime.NewRationalTime(1, 24))

		err := Insert(clip, track, opentime.NewRationalTime(12, 24), WithInsertFillTemplate(template))
		if err != nil {
			t.Errorf("Insert with option failed: %v", err)
		}
	})
}

func TestSliceWithOptions(t *testing.T) {
	track := createTestTrack([]float64{48}, 24)
	err := Slice(track, opentime.NewRationalTime(24, 24), WithSliceRemoveTransitions(true))
	if err != nil {
		t.Errorf("Slice with option failed: %v", err)
	}
}

func TestTrimWithOptions(t *testing.T) {
	track := createTestTrack([]float64{24, 48}, 24)
	item := track.Children()[1].(opentimelineio.Item)
	template := opentimelineio.NewGapWithDuration(opentime.NewRationalTime(1, 24))

	err := Trim(item, track, opentime.NewRationalTime(12, 24), opentime.RationalTime{}, WithTrimFillTemplate(template))
	if err != nil {
		t.Errorf("Trim with option failed: %v", err)
	}
}

func TestRemoveWithOptions(t *testing.T) {
	track := createTestTrack([]float64{24, 24, 24}, 24)
	template := opentimelineio.NewGapWithDuration(opentime.NewRationalTime(1, 24))

	err := Remove(track, opentime.NewRationalTime(36, 24), WithRemoveFillTemplate(template))
	if err != nil {
		t.Errorf("Remove with option failed: %v", err)
	}
}

// ============================================================================
// Edge Case Tests
// ============================================================================

func TestOverwriteBeforeStart(t *testing.T) {
	track := createTestTrack([]float64{24}, 24)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := opentimelineio.NewClip("before", nil, &sr, nil, nil, nil, "", nil)

	// Insert before start with gap
	timeRange := opentime.NewTimeRange(opentime.NewRationalTime(-24, 24), opentime.NewRationalTime(24, 24))
	err := Overwrite(clip, track, timeRange)
	if err != nil {
		t.Fatalf("Overwrite before start failed: %v", err)
	}

	// Should have new clip at start
	children := track.Children()
	if len(children) < 1 {
		t.Fatal("expected at least 1 child")
	}
	if children[0].Name() != "before" {
		t.Errorf("expected 'before', got %s", children[0].Name())
	}
}

func TestSliceAtEdges(t *testing.T) {
	t.Run("slice at start", func(t *testing.T) {
		track := createTestTrack([]float64{24}, 24)
		err := Slice(track, opentime.NewRationalTime(0, 24))
		// Should be no-op at exact start
		if err != nil {
			t.Errorf("Slice at start failed: %v", err)
		}
	})

	t.Run("slice at end", func(t *testing.T) {
		track := createTestTrack([]float64{24}, 24)
		err := Slice(track, opentime.NewRationalTime(24, 24))
		// Should be no-op at exact end (slice at boundary)
		if err != nil {
			t.Errorf("Slice at end failed: %v", err)
		}
	})
}

func TestSlideWithTwoItems(t *testing.T) {
	// Slide requires a previous item to adjust
	track := createTestTrackWithAvailableRange([]float64{48, 24}, 100, 24)
	item := track.Children()[1].(opentimelineio.Item)

	// Sliding left should shorten previous item
	err := Slide(item, track, opentime.NewRationalTime(-12, 24))
	if err != nil {
		t.Errorf("Slide failed: %v", err)
	}

	firstDur, _ := track.Children()[0].Duration()
	if firstDur.Value() != 36 {
		t.Errorf("expected first duration 36, got %.0f", firstDur.Value())
	}
}

func TestRollEdgeCases(t *testing.T) {
	t.Run("roll first item in", func(t *testing.T) {
		track := createTestTrack([]float64{48}, 24)
		item := track.Children()[0].(opentimelineio.Item)

		// Rolling in on first item should work
		err := Roll(item, track, opentime.NewRationalTime(12, 24), opentime.RationalTime{})
		if err != nil {
			t.Errorf("Roll first item failed: %v", err)
		}
	})

	t.Run("roll last item out", func(t *testing.T) {
		track := createTestTrack([]float64{48}, 24)
		item := track.Children()[0].(opentimelineio.Item)

		// Rolling out on last item should work
		err := Roll(item, track, opentime.RationalTime{}, opentime.NewRationalTime(12, 24))
		if err != nil {
			t.Errorf("Roll last item failed: %v", err)
		}
	})
}

func TestFillNoGap(t *testing.T) {
	track := createTestTrack([]float64{24}, 24)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := opentimelineio.NewClip("fill", nil, &sr, nil, nil, nil, "", nil)

	// Try to fill where there's no gap
	err := Fill(clip, track, opentime.NewRationalTime(12, 24), ReferencePointSource)
	if err == nil {
		t.Error("expected error when filling non-gap")
	}
}

func TestRemoveRangeWithoutFill(t *testing.T) {
	track := createTestTrack([]float64{24, 24, 24}, 24)
	timeRange := opentime.NewTimeRange(opentime.NewRationalTime(24, 24), opentime.NewRationalTime(24, 24))

	err := RemoveRange(track, timeRange, WithFill(false))
	if err != nil {
		t.Errorf("RemoveRange without fill failed: %v", err)
	}

	children := track.Children()
	if len(children) != 2 {
		t.Errorf("expected 2 children, got %d", len(children))
	}
}

func TestRemoveOutOfBounds(t *testing.T) {
	track := createTestTrack([]float64{24}, 24)

	err := Remove(track, opentime.NewRationalTime(48, 24))
	if err == nil {
		t.Error("expected error when removing out of bounds")
	}
}

func TestSlipNoSourceRange(t *testing.T) {
	clip := opentimelineio.NewClip("test", nil, nil, nil, nil, nil, "", nil)

	err := Slip(clip, opentime.NewRationalTime(10, 24))
	if err == nil {
		t.Error("expected error when slipping item without source range")
	}
}

func TestRippleBothDirections(t *testing.T) {
	t.Run("ripple in and out together", func(t *testing.T) {
		track := createTestTrackWithAvailableRange([]float64{48}, 100, 24)
		item := track.Children()[0].(opentimelineio.Item)

		// Ripple both in and out
		deltaIn := opentime.NewRationalTime(6, 24)
		deltaOut := opentime.NewRationalTime(6, 24)
		err := Ripple(item, deltaIn, deltaOut)
		if err != nil {
			t.Fatalf("Ripple failed: %v", err)
		}

		sr := item.SourceRange()
		if sr.StartTime().Value() != 6 {
			t.Errorf("expected start 6, got %.0f", sr.StartTime().Value())
		}
		if sr.Duration().Value() != 48 {
			t.Errorf("expected duration 48, got %.0f", sr.Duration().Value())
		}
	})
}

func TestTrimExpandWithNeighbor(t *testing.T) {
	track := createTestTrack([]float64{48, 48}, 24)
	children := track.Children()
	firstItem := children[0].(opentimelineio.Item)

	// Trim tail to extend into second item's space
	deltaOut := opentime.NewRationalTime(12, 24)
	err := Trim(firstItem, track, opentime.RationalTime{}, deltaOut)
	if err != nil {
		t.Errorf("Trim expand failed: %v", err)
	}

	// First item extended, second item shortened
	firstDur, _ := children[0].Duration()
	secondDur, _ := children[1].Duration()
	if firstDur.Value() != 60 {
		t.Errorf("expected first duration 60, got %.0f", firstDur.Value())
	}
	if secondDur.Value() != 36 {
		t.Errorf("expected second duration 36, got %.0f", secondDur.Value())
	}
}

// ============================================================================
// Utility Function Tests
// ============================================================================

func TestCompositionEndTime(t *testing.T) {
	track := createTestTrack([]float64{24, 24}, 24)
	end, err := compositionEndTime(track)
	if err != nil {
		t.Fatalf("compositionEndTime failed: %v", err)
	}
	if end.Value() != 48 {
		t.Errorf("expected end time 48, got %.0f", end.Value())
	}
}

func TestIsPositive(t *testing.T) {
	pos := opentime.NewRationalTime(10, 24)
	neg := opentime.NewRationalTime(-10, 24)
	zero := opentime.NewRationalTime(0, 24)

	if !isPositive(pos) {
		t.Error("expected 10/24 to be positive")
	}
	if isPositive(neg) {
		t.Error("expected -10/24 to not be positive")
	}
	if isPositive(zero) {
		t.Error("expected 0/24 to not be positive")
	}
}

func TestAdjustItemStartTime(t *testing.T) {
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	ref := opentimelineio.NewExternalReference("", "file://test.mov", &ar, nil)
	// Start at frame 0 with duration 48
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := opentimelineio.NewClip("test", ref, &sr, nil, nil, nil, "", nil)

	// Add delta of 10 frames to start: start = 0+10 = 10, duration = 48-10 = 38
	err := adjustItemStartTime(clip, opentime.NewRationalTime(10, 24))
	if err != nil {
		t.Fatalf("adjustItemStartTime failed: %v", err)
	}

	newSr := clip.SourceRange()
	if newSr.StartTime().Value() != 10 {
		t.Errorf("expected start 10, got %.0f", newSr.StartTime().Value())
	}
	if newSr.Duration().Value() != 38 {
		t.Errorf("expected duration 38, got %.0f", newSr.Duration().Value())
	}
}

func TestMinMaxRationalTime(t *testing.T) {
	a := opentime.NewRationalTime(10, 24)
	b := opentime.NewRationalTime(20, 24)

	min := minRationalTime(a, b)
	max := maxRationalTime(a, b)

	if min.Value() != 10 {
		t.Errorf("expected min 10, got %.0f", min.Value())
	}
	if max.Value() != 20 {
		t.Errorf("expected max 20, got %.0f", max.Value())
	}

	// Reverse order
	min2 := minRationalTime(b, a)
	max2 := maxRationalTime(b, a)

	if min2.Value() != 10 {
		t.Errorf("expected min 10, got %.0f", min2.Value())
	}
	if max2.Value() != 20 {
		t.Errorf("expected max 20, got %.0f", max2.Value())
	}
}

// ============================================================================
// Fill Sequence Tests
// ============================================================================

func TestFillSequenceTrimToGap(t *testing.T) {
	// Create track with gap
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)
	gap := opentimelineio.NewGapWithDuration(opentime.NewRationalTime(24, 24))
	track.AppendChild(gap)

	// Create clip longer than gap
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := opentimelineio.NewClip("fill", nil, &sr, nil, nil, nil, "", nil)

	// Fill with Sequence mode (should trim to gap size)
	time := opentime.NewRationalTime(0, 24)
	err := Fill(clip, track, time, ReferencePointSequence)
	if err != nil {
		t.Fatalf("Fill failed: %v", err)
	}

	children := track.Children()
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}

	// Clip should be trimmed to gap size
	dur, _ := children[0].Duration()
	if dur.Value() != 24 {
		t.Errorf("expected duration 24, got %.0f", dur.Value())
	}
}

func TestFillSequenceInMiddleOfGap(t *testing.T) {
	// Create track with gap
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)
	gap := opentimelineio.NewGapWithDuration(opentime.NewRationalTime(72, 24))
	track.AppendChild(gap)

	// Create clip shorter than gap
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := opentimelineio.NewClip("fill", nil, &sr, nil, nil, nil, "", nil)

	// Fill in middle of gap with Sequence mode
	time := opentime.NewRationalTime(24, 24)
	err := Fill(clip, track, time, ReferencePointSequence)
	if err != nil {
		t.Fatalf("Fill failed: %v", err)
	}

	children := track.Children()
	// Verify at least the clip was added
	if len(children) < 2 {
		t.Fatalf("expected at least 2 children, got %d", len(children))
	}
}

// ============================================================================
// Transition Tests
// ============================================================================

func createTestTrackWithTransitions() *opentimelineio.Track {
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)

	sr1 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip1 := opentimelineio.NewClip("clip1", nil, &sr1, nil, nil, nil, "", nil)
	track.AppendChild(clip1)

	// Add a transition
	inOffset := opentime.NewRationalTime(6, 24)
	outOffset := opentime.NewRationalTime(6, 24)
	transition := opentimelineio.NewTransition("cross_dissolve", "SMPTE_Dissolve", inOffset, outOffset, nil)
	track.AppendChild(transition)

	sr2 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip2 := opentimelineio.NewClip("clip2", nil, &sr2, nil, nil, nil, "", nil)
	track.AppendChild(clip2)

	return track
}

func TestSliceWithTransitions(t *testing.T) {
	track := createTestTrackWithTransitions()

	// Slice in first clip with transition removal enabled
	time := opentime.NewRationalTime(24, 24)
	err := Slice(track, time, WithSliceRemoveTransitions(true))
	if err != nil {
		t.Errorf("Slice with transitions failed: %v", err)
	}
}

func TestInsertRemovesTransitions(t *testing.T) {
	track := createTestTrackWithTransitions()
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	newClip := opentimelineio.NewClip("new", nil, &sr, nil, nil, nil, "", nil)

	// Insert with transition removal
	time := opentime.NewRationalTime(48, 24)
	err := Insert(newClip, track, time, WithInsertRemoveTransitions(true))
	if err != nil {
		t.Errorf("Insert with transition removal failed: %v", err)
	}
}

// ============================================================================
// Remove Range Tests
// ============================================================================

func TestRemoveRangeEntireTrack(t *testing.T) {
	track := createTestTrack([]float64{24, 24}, 24)

	// Remove entire track contents
	timeRange := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(48, 24),
	)
	err := RemoveRange(track, timeRange)
	if err != nil {
		t.Errorf("RemoveRange entire track failed: %v", err)
	}

	// Should have one gap
	children := track.Children()
	if len(children) != 1 {
		t.Fatalf("expected 1 child (gap), got %d", len(children))
	}
	_, isGap := children[0].(*opentimelineio.Gap)
	if !isGap {
		t.Error("expected Gap")
	}
}

func TestRemoveRangePartialClip(t *testing.T) {
	track := createTestTrack([]float64{48}, 24)

	// Remove middle portion
	timeRange := opentime.NewTimeRange(
		opentime.NewRationalTime(12, 24),
		opentime.NewRationalTime(24, 24),
	)
	err := RemoveRange(track, timeRange)
	if err != nil {
		t.Errorf("RemoveRange partial clip failed: %v", err)
	}

	// Should have: clip_part1, gap, clip_part2
	children := track.Children()
	if len(children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(children))
	}
}

// ============================================================================
// Trim Edge Cases
// ============================================================================

func TestTrimHeadWithPreviousItem(t *testing.T) {
	track := createTestTrack([]float64{48, 48}, 24)
	children := track.Children()
	secondItem := children[1].(opentimelineio.Item)

	// Trim head (positive delta) - makes clip shorter, extends previous
	deltaIn := opentime.NewRationalTime(12, 24)
	err := Trim(secondItem, track, deltaIn, opentime.RationalTime{})
	if err != nil {
		t.Errorf("Trim head failed: %v", err)
	}

	// Second item shorter, first item longer
	firstDur, _ := children[0].Duration()
	secondDur, _ := children[1].Duration()
	if firstDur.Value() != 60 {
		t.Errorf("expected first duration 60, got %.0f", firstDur.Value())
	}
	if secondDur.Value() != 36 {
		t.Errorf("expected second duration 36, got %.0f", secondDur.Value())
	}
}

func TestTrimTailWithNextItem(t *testing.T) {
	track := createTestTrack([]float64{48, 48}, 24)
	children := track.Children()
	firstItem := children[0].(opentimelineio.Item)

	// Trim tail (negative delta) - makes clip shorter, extends next
	deltaOut := opentime.NewRationalTime(-12, 24)
	err := Trim(firstItem, track, opentime.RationalTime{}, deltaOut)
	if err != nil {
		t.Errorf("Trim tail failed: %v", err)
	}

	// First item shorter, second item longer
	firstDur, _ := children[0].Duration()
	secondDur, _ := children[1].Duration()
	if firstDur.Value() != 36 {
		t.Errorf("expected first duration 36, got %.0f", firstDur.Value())
	}
	if secondDur.Value() != 60 {
		t.Errorf("expected second duration 60, got %.0f", secondDur.Value())
	}
}

func TestTrimBothEnds(t *testing.T) {
	track := createTestTrack([]float64{24, 72, 24}, 24)
	children := track.Children()
	middleItem := children[1].(opentimelineio.Item)

	// Trim both head and tail
	deltaIn := opentime.NewRationalTime(12, 24)
	deltaOut := opentime.NewRationalTime(-12, 24)
	err := Trim(middleItem, track, deltaIn, deltaOut)
	if err != nil {
		t.Errorf("Trim both ends failed: %v", err)
	}

	// Middle item should be 48 frames (72 - 12 - 12)
	dur, _ := children[1].Duration()
	if dur.Value() != 48 {
		t.Errorf("expected duration 48, got %.0f", dur.Value())
	}
}

// ============================================================================
// Roll Edge Cases
// ============================================================================

func TestRollInWithPreviousNeighbor(t *testing.T) {
	track := createTestTrackWithAvailableRange([]float64{48, 48}, 100, 24)
	children := track.Children()
	secondItem := children[1].(opentimelineio.Item)

	// Roll in point (trim head, extend previous)
	deltaIn := opentime.NewRationalTime(12, 24)
	err := Roll(secondItem, track, deltaIn, opentime.RationalTime{})
	if err != nil {
		t.Errorf("Roll in failed: %v", err)
	}

	// First extends, second shortens
	firstDur, _ := children[0].Duration()
	secondDur, _ := children[1].Duration()
	if firstDur.Value() != 60 {
		t.Errorf("expected first 60, got %.0f", firstDur.Value())
	}
	if secondDur.Value() != 36 {
		t.Errorf("expected second 36, got %.0f", secondDur.Value())
	}
}

func TestRollOutWithNextNeighbor(t *testing.T) {
	track := createTestTrackWithAvailableRange([]float64{48, 48}, 100, 24)
	children := track.Children()
	firstItem := children[0].(opentimelineio.Item)

	// Roll out point (extend first, trim next head)
	deltaOut := opentime.NewRationalTime(12, 24)
	err := Roll(firstItem, track, opentime.RationalTime{}, deltaOut)
	if err != nil {
		t.Errorf("Roll out failed: %v", err)
	}

	// First extends, second shortens
	firstDur, _ := children[0].Duration()
	secondDur, _ := children[1].Duration()
	if firstDur.Value() != 60 {
		t.Errorf("expected first 60, got %.0f", firstDur.Value())
	}
	if secondDur.Value() != 36 {
		t.Errorf("expected second 36, got %.0f", secondDur.Value())
	}
}

// ============================================================================
// Slide Edge Cases
// ============================================================================

func TestSlideNegative(t *testing.T) {
	track := createTestTrackWithAvailableRange([]float64{48, 48}, 100, 24)
	children := track.Children()
	secondItem := children[1].(opentimelineio.Item)

	// Slide left (shorten previous)
	delta := opentime.NewRationalTime(-12, 24)
	err := Slide(secondItem, track, delta)
	if err != nil {
		t.Errorf("Slide left failed: %v", err)
	}

	// First item shorter
	firstDur, _ := children[0].Duration()
	if firstDur.Value() != 36 {
		t.Errorf("expected first 36, got %.0f", firstDur.Value())
	}
}

// ============================================================================
// Insert Edge Cases
// ============================================================================

func TestInsertAtStartCoverage(t *testing.T) {
	track := createTestTrack([]float64{24}, 24)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	newClip := opentimelineio.NewClip("new", nil, &sr, nil, nil, nil, "", nil)

	err := Insert(newClip, track, opentime.NewRationalTime(0, 24))
	if err != nil {
		t.Errorf("Insert at start failed: %v", err)
	}

	children := track.Children()
	if children[0].Name() != "new" {
		t.Errorf("expected 'new' at start, got %s", children[0].Name())
	}
}

func TestInsertIntoEmptyTrack(t *testing.T) {
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	newClip := opentimelineio.NewClip("new", nil, &sr, nil, nil, nil, "", nil)

	err := Insert(newClip, track, opentime.NewRationalTime(0, 24))
	if err != nil {
		t.Errorf("Insert into empty track failed: %v", err)
	}

	children := track.Children()
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}
}

// ============================================================================
// Overwrite Edge Cases
// ============================================================================

func TestOverwriteEmptyCompositionCoverage(t *testing.T) {
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	newClip := opentimelineio.NewClip("new", nil, &sr, nil, nil, nil, "", nil)

	timeRange := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(24, 24),
	)
	err := Overwrite(newClip, track, timeRange)
	if err != nil {
		t.Errorf("Overwrite empty composition failed: %v", err)
	}

	children := track.Children()
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}
}

func TestOverwriteAppendAfterEndCoverage(t *testing.T) {
	track := createTestTrack([]float64{24}, 24)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	newClip := opentimelineio.NewClip("new", nil, &sr, nil, nil, nil, "", nil)

	// Overwrite starting after current end (with gap)
	timeRange := opentime.NewTimeRange(
		opentime.NewRationalTime(48, 24), // after end of track (which is 24)
		opentime.NewRationalTime(24, 24),
	)
	err := Overwrite(newClip, track, timeRange)
	if err != nil {
		t.Errorf("Overwrite append after end failed: %v", err)
	}

	// Should have: original clip, gap, new clip
	children := track.Children()
	if len(children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(children))
	}
}

// ============================================================================
// Ripple Edge Cases
// ============================================================================

func TestRippleClampToAvailable(t *testing.T) {
	track := createTestTrackWithAvailableRange([]float64{24}, 48, 24)
	item := track.Children()[0].(opentimelineio.Item)

	// Try to extend beyond available range
	deltaOut := opentime.NewRationalTime(100, 24)
	err := Ripple(item, opentime.RationalTime{}, deltaOut)
	if err != nil {
		t.Errorf("Ripple clamp failed: %v", err)
	}

	// Should be clamped to max available (48)
	sr := item.SourceRange()
	if sr.Duration().Value() > 48 {
		t.Errorf("expected duration <= 48, got %.0f", sr.Duration().Value())
	}
}
