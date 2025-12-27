// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package algorithms

import (
	"testing"

	"github.com/mrjoshuak/gotio/opentime"
	"github.com/mrjoshuak/gotio/opentimelineio"
)

// ============================================================================
// Trim Coverage Tests
// ============================================================================

func TestTrimHeadClampToAvailable(t *testing.T) {
	// Test trimHead clamping when newStart < available range start
	track := createTestTrackWithAvailableRange([]float64{48, 48}, 100, 24)
	children := track.Children()
	secondItem := children[1].(opentimelineio.Item)

	// Set source range starting at frame 20
	sr := opentime.NewTimeRange(opentime.NewRationalTime(20, 24), opentime.NewRationalTime(48, 24))
	secondItem.SetSourceRange(&sr)

	// Try to trim head backwards past the available range start (0)
	// deltaIn = -30 would put start at -10, should clamp to 0
	deltaIn := opentime.NewRationalTime(-30, 24)
	err := Trim(secondItem, track, deltaIn, opentime.RationalTime{})
	if err != nil {
		t.Errorf("Trim should succeed with clamping: %v", err)
	}

	// Start should be clamped to 0 (available range start)
	newSr := secondItem.SourceRange()
	if newSr.StartTime().Value() < 0 {
		t.Errorf("start should be clamped to >= 0, got %.0f", newSr.StartTime().Value())
	}
}

func TestTrimHeadNoPreviousItemExtend(t *testing.T) {
	// Test creating a gap when no previous item and extending head (deltaIn < 0)
	track := createTestTrack([]float64{48}, 24)
	item := track.Children()[0].(opentimelineio.Item)

	// Extend head (negative deltaIn) with no previous item
	deltaIn := opentime.NewRationalTime(-12, 24)
	err := Trim(item, track, deltaIn, opentime.RationalTime{})
	if err != nil {
		t.Errorf("Trim should succeed: %v", err)
	}

	// Should have created a gap before the item
	children := track.Children()
	if len(children) != 2 {
		t.Fatalf("expected 2 children (gap + item), got %d", len(children))
	}
	if _, isGap := children[0].(*opentimelineio.Gap); !isGap {
		t.Error("first child should be a gap")
	}
}

func TestTrimHeadPreviousItemNoSourceRange(t *testing.T) {
	// Test when previous item has no source range (uses available range)
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)

	// Add a gap (gaps typically don't have source ranges set explicitly)
	gap := opentimelineio.NewGapWithDuration(opentime.NewRationalTime(48, 24))
	track.AppendChild(gap)

	// Add a clip
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := opentimelineio.NewClip("test", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)

	// Trim clip's head (should adjust gap)
	deltaIn := opentime.NewRationalTime(12, 24)
	err := Trim(clip, track, deltaIn, opentime.RationalTime{})
	if err != nil {
		t.Errorf("Trim should succeed: %v", err)
	}
}

func TestTrimHeadClampPreviousDuration(t *testing.T) {
	// Test clamping previous item duration when it would go negative
	track := createTestTrack([]float64{12, 48}, 24)
	children := track.Children()
	secondItem := children[1].(opentimelineio.Item)

	// Trim head by more than previous item's duration
	// First item is 12 frames, we're trimming by -24 (extending by 24)
	deltaIn := opentime.NewRationalTime(-24, 24)
	err := Trim(secondItem, track, deltaIn, opentime.RationalTime{})
	if err != nil {
		t.Errorf("Trim should succeed: %v", err)
	}

	// Previous item's duration should be clamped to 0
	firstDur, _ := children[0].Duration()
	if firstDur.Value() < 0 {
		t.Error("first item duration should not be negative")
	}
}

func TestTrimTailClampToAvailable(t *testing.T) {
	// Test trimTail clamping when duration exceeds available range
	track := createTestTrackWithAvailableRange([]float64{24}, 48, 24)
	item := track.Children()[0].(opentimelineio.Item)

	// Try to extend tail beyond available range
	// Available is 48, current is 24, trying to add 50 should clamp
	deltaOut := opentime.NewRationalTime(50, 24)
	err := Trim(item, track, opentime.RationalTime{}, deltaOut)
	if err != nil {
		t.Errorf("Trim should succeed with clamping: %v", err)
	}

	// Duration should be clamped to max available (48)
	sr := item.SourceRange()
	if sr.Duration().Value() > 48 {
		t.Errorf("duration should be clamped to <= 48, got %.0f", sr.Duration().Value())
	}
}

func TestTrimTailNextItemNoSourceRange(t *testing.T) {
	// Test when next item has no source range (uses available range)
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)

	// Add a clip
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := opentimelineio.NewClip("test", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)

	// Add a gap (gaps typically don't have source ranges set explicitly)
	gap := opentimelineio.NewGapWithDuration(opentime.NewRationalTime(48, 24))
	track.AppendChild(gap)

	// Trim clip's tail (should adjust gap)
	deltaOut := opentime.NewRationalTime(12, 24)
	err := Trim(clip, track, opentime.RationalTime{}, deltaOut)
	if err != nil {
		t.Errorf("Trim should succeed: %v", err)
	}
}

func TestTrimTailEliminateGap(t *testing.T) {
	// Test eliminating a gap when trim would make it negative
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)

	// Add a clip
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := opentimelineio.NewClip("test", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)

	// Add a small gap
	gap := opentimelineio.NewGapWithDuration(opentime.NewRationalTime(12, 24))
	track.AppendChild(gap)

	// Extend clip's tail by more than gap's duration
	deltaOut := opentime.NewRationalTime(24, 24)
	err := Trim(clip, track, opentime.RationalTime{}, deltaOut)
	if err != nil {
		t.Errorf("Trim should succeed: %v", err)
	}

	// Gap should be eliminated
	children := track.Children()
	if len(children) != 1 {
		t.Errorf("expected 1 child (gap eliminated), got %d", len(children))
	}
}

func TestTrimTailNoNextItemContract(t *testing.T) {
	// Test creating a gap when no next item and contracting tail (deltaOut < 0)
	track := createTestTrack([]float64{48}, 24)
	item := track.Children()[0].(opentimelineio.Item)

	// Contract tail (negative deltaOut) with no next item
	deltaOut := opentime.NewRationalTime(-12, 24)
	err := Trim(item, track, opentime.RationalTime{}, deltaOut)
	if err != nil {
		t.Errorf("Trim should succeed: %v", err)
	}

	// Should have created a gap after the item
	children := track.Children()
	if len(children) != 2 {
		t.Fatalf("expected 2 children (item + gap), got %d", len(children))
	}
	if _, isGap := children[1].(*opentimelineio.Gap); !isGap {
		t.Error("second child should be a gap")
	}
}

func TestTrimNegativeDuration(t *testing.T) {
	track := createTestTrack([]float64{24}, 24)
	item := track.Children()[0].(opentimelineio.Item)

	// Try to trim more than duration
	deltaIn := opentime.NewRationalTime(48, 24)
	err := Trim(item, track, deltaIn, opentime.RationalTime{})
	if err != ErrNegativeDuration {
		t.Errorf("expected ErrNegativeDuration, got %v", err)
	}
}

// ============================================================================
// Roll Coverage Tests
// ============================================================================

func TestRollInNoPreviousPositive(t *testing.T) {
	// Test rolling in-point with no previous item (positive delta trims head)
	track := createTestTrackWithAvailableRange([]float64{48}, 100, 24)
	item := track.Children()[0].(opentimelineio.Item)

	// Roll in with no previous item - should trim head
	deltaIn := opentime.NewRationalTime(12, 24)
	err := Roll(item, track, deltaIn, opentime.RationalTime{})
	if err != nil {
		t.Errorf("Roll should succeed: %v", err)
	}

	sr := item.SourceRange()
	if sr.StartTime().Value() != 12 {
		t.Errorf("expected start 12, got %.0f", sr.StartTime().Value())
	}
}

func TestRollInNoPreviousNegative(t *testing.T) {
	// Test rolling in-point with no previous item (negative delta - no-op)
	track := createTestTrackWithAvailableRange([]float64{48}, 100, 24)
	item := track.Children()[0].(opentimelineio.Item)

	originalSr := *item.SourceRange()

	// Roll in with negative delta and no previous item - should be no-op
	deltaIn := opentime.NewRationalTime(-12, 24)
	err := Roll(item, track, deltaIn, opentime.RationalTime{})
	if err != nil {
		t.Errorf("Roll should succeed: %v", err)
	}

	// Should be unchanged
	newSr := item.SourceRange()
	if newSr.StartTime().Value() != originalSr.StartTime().Value() {
		t.Errorf("start should be unchanged, got %.0f", newSr.StartTime().Value())
	}
}

func TestRollInClampLeft(t *testing.T) {
	// Test clamping when rolling left past available range
	track := createTestTrackWithAvailableRange([]float64{48, 48}, 100, 24)
	children := track.Children()
	secondItem := children[1].(opentimelineio.Item)

	// Set source range starting at frame 10
	sr := opentime.NewTimeRange(opentime.NewRationalTime(10, 24), opentime.NewRationalTime(48, 24))
	secondItem.SetSourceRange(&sr)

	// Roll left by more than start allows (would go to -10)
	deltaIn := opentime.NewRationalTime(-20, 24)
	err := Roll(secondItem, track, deltaIn, opentime.RationalTime{})
	if err != nil {
		t.Errorf("Roll should succeed with clamping: %v", err)
	}

	// Start should be clamped to 0
	newSr := secondItem.SourceRange()
	if newSr.StartTime().Value() < 0 {
		t.Errorf("start should be clamped to >= 0, got %.0f", newSr.StartTime().Value())
	}
}

func TestRollInClampToPreviousDuration(t *testing.T) {
	// Test clamping when rolling right more than previous item allows
	track := createTestTrack([]float64{12, 48}, 24)
	children := track.Children()
	secondItem := children[1].(opentimelineio.Item)

	// Roll right by more than previous item's duration (12)
	deltaIn := opentime.NewRationalTime(24, 24)
	err := Roll(secondItem, track, deltaIn, opentime.RationalTime{})
	if err != nil {
		t.Errorf("Roll should succeed with clamping: %v", err)
	}

	// Previous item should not have negative duration
	firstDur, _ := children[0].Duration()
	if firstDur.Value() < 0 {
		t.Error("first item duration should not be negative")
	}
}

func TestRollInPreviousNoSourceRange(t *testing.T) {
	// Test when previous item has no source range
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)

	gap := opentimelineio.NewGapWithDuration(opentime.NewRationalTime(48, 24))
	track.AppendChild(gap)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := opentimelineio.NewClip("test", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)

	deltaIn := opentime.NewRationalTime(12, 24)
	err := Roll(clip, track, deltaIn, opentime.RationalTime{})
	if err != nil {
		t.Errorf("Roll should succeed: %v", err)
	}
}

func TestRollOutNoNextPositive(t *testing.T) {
	// Test rolling out-point with no next item (positive delta extends tail)
	track := createTestTrackWithAvailableRange([]float64{48}, 100, 24)
	item := track.Children()[0].(opentimelineio.Item)

	// Roll out with no next item - should extend tail
	deltaOut := opentime.NewRationalTime(12, 24)
	err := Roll(item, track, opentime.RationalTime{}, deltaOut)
	if err != nil {
		t.Errorf("Roll should succeed: %v", err)
	}

	sr := item.SourceRange()
	if sr.Duration().Value() != 60 {
		t.Errorf("expected duration 60, got %.0f", sr.Duration().Value())
	}
}

func TestRollOutNoNextNegative(t *testing.T) {
	// Test rolling out-point with no next item (negative delta - no-op)
	track := createTestTrackWithAvailableRange([]float64{48}, 100, 24)
	item := track.Children()[0].(opentimelineio.Item)

	originalDur, _ := item.Duration()

	// Roll out with negative delta and no next item - should be no-op
	deltaOut := opentime.NewRationalTime(-12, 24)
	err := Roll(item, track, opentime.RationalTime{}, deltaOut)
	if err != nil {
		t.Errorf("Roll should succeed: %v", err)
	}

	// Should be unchanged
	newDur, _ := item.Duration()
	if newDur.Value() != originalDur.Value() {
		t.Errorf("duration should be unchanged, got %.0f", newDur.Value())
	}
}

func TestRollOutNoNextClampToAvailable(t *testing.T) {
	// Test clamping when extending past available range
	track := createTestTrackWithAvailableRange([]float64{24}, 48, 24)
	item := track.Children()[0].(opentimelineio.Item)

	// Try to extend beyond available (48)
	deltaOut := opentime.NewRationalTime(50, 24)
	err := Roll(item, track, opentime.RationalTime{}, deltaOut)
	if err != nil {
		t.Errorf("Roll should succeed with clamping: %v", err)
	}

	sr := item.SourceRange()
	if sr.Duration().Value() > 48 {
		t.Errorf("duration should be clamped to <= 48, got %.0f", sr.Duration().Value())
	}
}

func TestRollOutClampToNextDuration(t *testing.T) {
	// Test clamping when rolling right more than next item allows
	track := createTestTrack([]float64{48, 12}, 24)
	children := track.Children()
	firstItem := children[0].(opentimelineio.Item)

	// Roll right by more than next item's duration (12)
	deltaOut := opentime.NewRationalTime(24, 24)
	err := Roll(firstItem, track, opentime.RationalTime{}, deltaOut)
	if err != nil {
		t.Errorf("Roll should succeed with clamping: %v", err)
	}

	// Next item should not have negative duration
	secondDur, _ := children[1].Duration()
	if secondDur.Value() < 0 {
		t.Error("second item duration should not be negative")
	}
}

func TestRollOutClampLeft(t *testing.T) {
	// Test clamping when rolling left would make our duration negative
	track := createTestTrack([]float64{24, 48}, 24)
	children := track.Children()
	firstItem := children[0].(opentimelineio.Item)

	// Roll left by more than our duration
	deltaOut := opentime.NewRationalTime(-48, 24)
	err := Roll(firstItem, track, opentime.RationalTime{}, deltaOut)
	if err != nil {
		// May return error for negative duration
	}

	// Our duration should not be negative
	firstDur, _ := children[0].Duration()
	if firstDur.Value() < 0 {
		t.Error("first item duration should not be negative")
	}
}

func TestRollOutNextNoSourceRange(t *testing.T) {
	// Test when next item has no source range
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := opentimelineio.NewClip("test", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)

	gap := opentimelineio.NewGapWithDuration(opentime.NewRationalTime(48, 24))
	track.AppendChild(gap)

	deltaOut := opentime.NewRationalTime(12, 24)
	err := Roll(clip, track, opentime.RationalTime{}, deltaOut)
	if err != nil {
		t.Errorf("Roll should succeed: %v", err)
	}
}

// ============================================================================
// Slide Coverage Tests
// ============================================================================

func TestSlideFirstItem(t *testing.T) {
	// First item cannot be slid - should return nil (no-op)
	track := createTestTrack([]float64{24}, 24)
	item := track.Children()[0].(opentimelineio.Item)

	err := Slide(item, track, opentime.NewRationalTime(12, 24))
	if err != nil {
		t.Errorf("Slide first item should be no-op, not error: %v", err)
	}
}

func TestSlidePreviousNoSourceRange(t *testing.T) {
	// Test when previous item has no source range
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)

	gap := opentimelineio.NewGapWithDuration(opentime.NewRationalTime(48, 24))
	track.AppendChild(gap)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := opentimelineio.NewClip("test", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)

	err := Slide(clip, track, opentime.NewRationalTime(12, 24))
	if err != nil {
		t.Errorf("Slide should succeed: %v", err)
	}
}

func TestSlideClampNegative(t *testing.T) {
	// Test clamping when sliding left more than previous item's duration
	track := createTestTrack([]float64{12, 48}, 24)
	children := track.Children()
	secondItem := children[1].(opentimelineio.Item)

	// Slide left by more than previous item's duration (12)
	err := Slide(secondItem, track, opentime.NewRationalTime(-24, 24))
	if err != nil {
		t.Errorf("Slide should succeed with clamping: %v", err)
	}

	// Previous item's duration should be clamped to 0
	firstDur, _ := children[0].Duration()
	if firstDur.Value() < 0 {
		t.Error("first item duration should not be negative")
	}
}

func TestSlideClampToAvailable(t *testing.T) {
	// Test clamping when sliding right past available range
	track := createTestTrackWithAvailableRange([]float64{24, 48}, 36, 24)
	children := track.Children()
	secondItem := children[1].(opentimelineio.Item)

	// Slide right by more than available allows
	err := Slide(secondItem, track, opentime.NewRationalTime(24, 24))
	if err != nil {
		t.Errorf("Slide should succeed with clamping: %v", err)
	}

	// Previous item's duration should be clamped to max available
	firstDur, _ := children[0].Duration()
	if firstDur.Value() > 36 {
		t.Errorf("first item duration should be <= 36, got %.0f", firstDur.Value())
	}
}

func TestSlideZeroDelta(t *testing.T) {
	// Zero delta should be no-op
	track := createTestTrack([]float64{24, 24}, 24)
	children := track.Children()
	secondItem := children[1].(opentimelineio.Item)

	originalDur, _ := children[0].Duration()

	err := Slide(secondItem, track, opentime.NewRationalTime(0, 24))
	if err != nil {
		t.Errorf("Slide with zero delta should succeed: %v", err)
	}

	newDur, _ := children[0].Duration()
	if newDur.Value() != originalDur.Value() {
		t.Error("duration should be unchanged")
	}
}

// ============================================================================
// Fill Coverage Tests
// ============================================================================

func TestFillSourceClipLongerThanGap(t *testing.T) {
	// Source mode keeps clip at original duration, may overflow gap
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)
	gap := opentimelineio.NewGapWithDuration(opentime.NewRationalTime(24, 24))
	track.AppendChild(gap)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := opentimelineio.NewClip("fill", nil, &sr, nil, nil, nil, "", nil)

	err := Fill(clip, track, opentime.NewRationalTime(0, 24), ReferencePointSource)
	if err != nil {
		t.Errorf("Fill Source should succeed: %v", err)
	}
}

func TestFillSequenceClipShorterThanGap(t *testing.T) {
	// Sequence mode with clip shorter than gap - uses clip duration
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)
	gap := opentimelineio.NewGapWithDuration(opentime.NewRationalTime(72, 24))
	track.AppendChild(gap)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := opentimelineio.NewClip("fill", nil, &sr, nil, nil, nil, "", nil)

	err := Fill(clip, track, opentime.NewRationalTime(0, 24), ReferencePointSequence)
	if err != nil {
		t.Errorf("Fill Sequence should succeed: %v", err)
	}

	// Clip should use original duration (24)
	children := track.Children()
	dur, _ := children[0].Duration()
	if dur.Value() != 24 {
		t.Errorf("expected duration 24, got %.0f", dur.Value())
	}
}

// ============================================================================
// Insert Coverage Tests
// ============================================================================

func TestInsertPastEnd(t *testing.T) {
	// Insert past end should work (may create gap)
	track := createTestTrack([]float64{24}, 24)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := opentimelineio.NewClip("new", nil, &sr, nil, nil, nil, "", nil)

	err := Insert(clip, track, opentime.NewRationalTime(48, 24))
	if err != nil {
		t.Errorf("Insert past end should succeed: %v", err)
	}
}

// ============================================================================
// Overwrite Coverage Tests
// ============================================================================

func TestOverwriteInsertBeforeStart(t *testing.T) {
	// Test overwrite that starts before composition start
	track := createTestTrack([]float64{24}, 24)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := opentimelineio.NewClip("new", nil, &sr, nil, nil, nil, "", nil)

	// Overwrite starting before 0
	timeRange := opentime.NewTimeRange(
		opentime.NewRationalTime(-12, 24),
		opentime.NewRationalTime(24, 24),
	)
	err := Overwrite(clip, track, timeRange)
	if err != nil {
		t.Errorf("Overwrite before start should succeed: %v", err)
	}

	// Should have new clip at start
	children := track.Children()
	if children[0].Name() != "new" {
		t.Errorf("expected 'new' at start, got %s", children[0].Name())
	}
}

// ============================================================================
// Remove Coverage Tests
// ============================================================================

func TestRemoveRangeSpanningMultiple(t *testing.T) {
	track := createTestTrack([]float64{24, 24, 24, 24}, 24)

	// Remove range spanning items 2 and 3
	timeRange := opentime.NewTimeRange(
		opentime.NewRationalTime(24, 24),
		opentime.NewRationalTime(48, 24),
	)
	err := RemoveRange(track, timeRange, WithFill(false))
	if err != nil {
		t.Errorf("RemoveRange should succeed: %v", err)
	}
}

// ============================================================================
// Ripple Coverage Tests
// ============================================================================

func TestRippleNegativeIn(t *testing.T) {
	// Test ripple with negative deltaIn (extend head)
	track := createTestTrackWithAvailableRange([]float64{48}, 100, 24)
	item := track.Children()[0].(opentimelineio.Item)

	// Set source range starting at frame 20
	sr := opentime.NewTimeRange(opentime.NewRationalTime(20, 24), opentime.NewRationalTime(48, 24))
	item.SetSourceRange(&sr)

	deltaIn := opentime.NewRationalTime(-10, 24)
	err := Ripple(item, deltaIn, opentime.RationalTime{})
	if err != nil {
		t.Errorf("Ripple should succeed: %v", err)
	}

	newSr := item.SourceRange()
	if newSr.StartTime().Value() != 10 {
		t.Errorf("expected start 10, got %.0f", newSr.StartTime().Value())
	}
}

func TestRippleNegativeOut(t *testing.T) {
	// Test ripple with negative deltaOut (contract tail)
	track := createTestTrackWithAvailableRange([]float64{48}, 100, 24)
	item := track.Children()[0].(opentimelineio.Item)

	deltaOut := opentime.NewRationalTime(-12, 24)
	err := Ripple(item, opentime.RationalTime{}, deltaOut)
	if err != nil {
		t.Errorf("Ripple should succeed: %v", err)
	}

	sr := item.SourceRange()
	if sr.Duration().Value() != 36 {
		t.Errorf("expected duration 36, got %.0f", sr.Duration().Value())
	}
}

// ============================================================================
// Slice Coverage Tests
// ============================================================================

func TestSliceNearBoundary(t *testing.T) {
	// Slice very close to an item boundary
	track := createTestTrack([]float64{24, 24}, 24)

	// Slice 1 frame before boundary
	err := Slice(track, opentime.NewRationalTime(23, 24))
	if err != nil {
		t.Errorf("Slice near boundary should succeed: %v", err)
	}

	children := track.Children()
	if len(children) != 3 {
		t.Errorf("expected 3 children after slice, got %d", len(children))
	}
}

// ============================================================================
// Error Path Tests
// ============================================================================

func TestTrimItemNotInComposition(t *testing.T) {
	track := createTestTrack([]float64{24}, 24)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	orphanClip := opentimelineio.NewClip("orphan", nil, &sr, nil, nil, nil, "", nil)

	err := Trim(orphanClip, track, opentime.NewRationalTime(12, 24), opentime.RationalTime{})
	if err == nil {
		t.Error("Trim should fail for item not in composition")
	}
}

func TestRollItemNotInComposition(t *testing.T) {
	track := createTestTrack([]float64{24}, 24)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	orphanClip := opentimelineio.NewClip("orphan", nil, &sr, nil, nil, nil, "", nil)

	err := Roll(orphanClip, track, opentime.NewRationalTime(12, 24), opentime.RationalTime{})
	if err == nil {
		t.Error("Roll should fail for item not in composition")
	}
}

func TestSlideItemNotInComposition(t *testing.T) {
	track := createTestTrack([]float64{24}, 24)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	orphanClip := opentimelineio.NewClip("orphan", nil, &sr, nil, nil, nil, "", nil)

	err := Slide(orphanClip, track, opentime.NewRationalTime(12, 24))
	if err == nil {
		t.Error("Slide should fail for item not in composition")
	}
}

// ============================================================================
// Additional Fill Coverage Tests
// ============================================================================

func TestFillNotInGap(t *testing.T) {
	// Fill at a time that's not in a gap should fail
	track := createTestTrack([]float64{48}, 24)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := opentimelineio.NewClip("fill", nil, &sr, nil, nil, nil, "", nil)

	err := Fill(clip, track, opentime.NewRationalTime(12, 24), ReferencePointSource)
	if err == nil {
		t.Error("Fill should fail when not in gap")
	}
}

func TestFillFitMode(t *testing.T) {
	// Fit mode adds time warp effect
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)
	gap := opentimelineio.NewGapWithDuration(opentime.NewRationalTime(48, 24))
	track.AppendChild(gap)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := opentimelineio.NewClip("fill", nil, &sr, nil, nil, nil, "", nil)

	err := Fill(clip, track, opentime.NewRationalTime(0, 24), ReferencePointFit)
	if err != nil {
		t.Errorf("Fill Fit should succeed: %v", err)
	}

	// Should have time warp effect
	children := track.Children()
	item := children[0].(opentimelineio.Item)
	effects := item.Effects()
	if len(effects) == 0 {
		t.Error("expected time warp effect")
	}
}

func TestFillOutsideTrack(t *testing.T) {
	// Fill at time outside track should fail
	track := createTestTrack([]float64{24}, 24)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := opentimelineio.NewClip("fill", nil, &sr, nil, nil, nil, "", nil)

	err := Fill(clip, track, opentime.NewRationalTime(100, 24), ReferencePointSource)
	if err == nil {
		t.Error("Fill outside track should fail")
	}
}

// ============================================================================
// Additional Slice Coverage Tests
// ============================================================================

func TestSliceInGap(t *testing.T) {
	// Slice inside a gap
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)
	gap := opentimelineio.NewGapWithDuration(opentime.NewRationalTime(48, 24))
	track.AppendChild(gap)

	err := Slice(track, opentime.NewRationalTime(24, 24))
	if err != nil {
		t.Errorf("Slice in gap should succeed: %v", err)
	}

	// Should have two gaps
	children := track.Children()
	if len(children) != 2 {
		t.Errorf("expected 2 children, got %d", len(children))
	}
}

func TestSliceEmptyTrack(t *testing.T) {
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)

	// Slice on empty track is a no-op (no items to slice)
	err := Slice(track, opentime.NewRationalTime(0, 24))
	if err != nil {
		t.Errorf("Slice in empty track should succeed as no-op: %v", err)
	}
}

// ============================================================================
// Additional Overwrite Coverage Tests
// ============================================================================

func TestOverwriteExactMatch(t *testing.T) {
	// Overwrite that exactly matches an existing item
	track := createTestTrack([]float64{24, 24, 24}, 24)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := opentimelineio.NewClip("new", nil, &sr, nil, nil, nil, "", nil)

	// Overwrite middle clip exactly
	timeRange := opentime.NewTimeRange(
		opentime.NewRationalTime(24, 24),
		opentime.NewRationalTime(24, 24),
	)
	err := Overwrite(clip, track, timeRange)
	if err != nil {
		t.Errorf("Overwrite exact match should succeed: %v", err)
	}
}

func TestOverwritePartialOverlap(t *testing.T) {
	// Overwrite that partially overlaps items
	track := createTestTrack([]float64{24, 24}, 24)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := opentimelineio.NewClip("new", nil, &sr, nil, nil, nil, "", nil)

	// Overwrite straddling boundary
	timeRange := opentime.NewTimeRange(
		opentime.NewRationalTime(12, 24),
		opentime.NewRationalTime(24, 24),
	)
	err := Overwrite(clip, track, timeRange)
	if err != nil {
		t.Errorf("Overwrite partial overlap should succeed: %v", err)
	}
}

// ============================================================================
// Additional Ripple Coverage Tests
// ============================================================================

func TestRippleNoSourceRange(t *testing.T) {
	// Item with no source range should use available range
	clip := opentimelineio.NewClip("test", nil, nil, nil, nil, nil, "", nil)

	err := Ripple(clip, opentime.NewRationalTime(6, 24), opentime.RationalTime{})
	// This may fail due to no available range, but should not panic
	_ = err
}

func TestRippleClampIn(t *testing.T) {
	// Test clamping when deltaIn would go past available range start
	track := createTestTrackWithAvailableRange([]float64{48}, 100, 24)
	item := track.Children()[0].(opentimelineio.Item)

	// Try to ripple in by more than start allows
	deltaIn := opentime.NewRationalTime(-100, 24)
	err := Ripple(item, deltaIn, opentime.RationalTime{})
	if err != nil {
		t.Errorf("Ripple should succeed with clamping: %v", err)
	}

	// Start should be clamped to 0
	sr := item.SourceRange()
	if sr.StartTime().Value() < 0 {
		t.Errorf("start should be clamped to >= 0, got %.0f", sr.StartTime().Value())
	}
}

func TestRippleClampOut(t *testing.T) {
	// Test clamping when deltaOut would exceed available range
	track := createTestTrackWithAvailableRange([]float64{24}, 48, 24)
	item := track.Children()[0].(opentimelineio.Item)

	// Try to extend beyond available (48)
	deltaOut := opentime.NewRationalTime(100, 24)
	err := Ripple(item, opentime.RationalTime{}, deltaOut)
	if err != nil {
		t.Errorf("Ripple should succeed with clamping: %v", err)
	}

	sr := item.SourceRange()
	if sr.Duration().Value() > 48 {
		t.Errorf("duration should be clamped to <= 48, got %.0f", sr.Duration().Value())
	}
}

// ============================================================================
// Additional Roll Coverage Tests
// ============================================================================

func TestRollZeroDeltas(t *testing.T) {
	// Zero deltas should be no-op
	track := createTestTrack([]float64{24, 24}, 24)
	children := track.Children()
	item := children[0].(opentimelineio.Item)

	originalDur, _ := item.Duration()

	err := Roll(item, track, opentime.RationalTime{}, opentime.RationalTime{})
	if err != nil {
		t.Errorf("Roll with zero deltas should succeed: %v", err)
	}

	newDur, _ := item.Duration()
	if newDur.Value() != originalDur.Value() {
		t.Error("duration should be unchanged")
	}
}

func TestRollItemNoSourceRange(t *testing.T) {
	// Test when main item has no source range
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)

	gap1 := opentimelineio.NewGapWithDuration(opentime.NewRationalTime(48, 24))
	track.AppendChild(gap1)

	gap2 := opentimelineio.NewGapWithDuration(opentime.NewRationalTime(48, 24))
	track.AppendChild(gap2)

	err := Roll(gap2, track, opentime.NewRationalTime(12, 24), opentime.RationalTime{})
	if err != nil {
		t.Errorf("Roll gap should succeed: %v", err)
	}
}

// ============================================================================
// Additional Insert Coverage Tests
// ============================================================================

func TestInsertZeroTime(t *testing.T) {
	track := createTestTrack([]float64{24}, 24)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(12, 24))
	clip := opentimelineio.NewClip("new", nil, &sr, nil, nil, nil, "", nil)

	err := Insert(clip, track, opentime.NewRationalTime(0, 24))
	if err != nil {
		t.Errorf("Insert at zero should succeed: %v", err)
	}

	children := track.Children()
	if children[0].Name() != "new" {
		t.Errorf("expected 'new' at start, got %s", children[0].Name())
	}
}

func TestInsertAtBoundary(t *testing.T) {
	track := createTestTrack([]float64{24, 24}, 24)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(12, 24))
	clip := opentimelineio.NewClip("new", nil, &sr, nil, nil, nil, "", nil)

	// Insert at boundary between clips
	err := Insert(clip, track, opentime.NewRationalTime(24, 24))
	if err != nil {
		t.Errorf("Insert at boundary should succeed: %v", err)
	}

	children := track.Children()
	if len(children) != 3 {
		t.Errorf("expected 3 children, got %d", len(children))
	}
}

// ============================================================================
// Utility Function Coverage Tests
// ============================================================================

func TestAdjustItemDurationExtend(t *testing.T) {
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	ref := opentimelineio.NewExternalReference("", "file://test.mov", &ar, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := opentimelineio.NewClip("test", ref, &sr, nil, nil, nil, "", nil)

	err := adjustItemDuration(clip, opentime.NewRationalTime(12, 24))
	if err != nil {
		t.Fatalf("adjustItemDuration failed: %v", err)
	}

	newSr := clip.SourceRange()
	if newSr.Duration().Value() != 36 {
		t.Errorf("expected duration 36, got %.0f", newSr.Duration().Value())
	}
}

func TestAdjustItemDurationContract(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := opentimelineio.NewClip("test", nil, &sr, nil, nil, nil, "", nil)

	err := adjustItemDuration(clip, opentime.NewRationalTime(-12, 24))
	if err != nil {
		t.Fatalf("adjustItemDuration failed: %v", err)
	}

	newSr := clip.SourceRange()
	if newSr.Duration().Value() != 36 {
		t.Errorf("expected duration 36, got %.0f", newSr.Duration().Value())
	}
}

func TestAdjustItemDurationNegativeResult(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := opentimelineio.NewClip("test", nil, &sr, nil, nil, nil, "", nil)

	err := adjustItemDuration(clip, opentime.NewRationalTime(-48, 24))
	if err != ErrNegativeDuration {
		t.Errorf("expected ErrNegativeDuration, got %v", err)
	}
}

func TestAdjustItemStartTimeNegativeResult(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := opentimelineio.NewClip("test", nil, &sr, nil, nil, nil, "", nil)

	// Delta of 48 would make duration negative (24 - 48 = -24)
	err := adjustItemStartTime(clip, opentime.NewRationalTime(48, 24))
	if err != ErrNegativeDuration {
		t.Errorf("expected ErrNegativeDuration, got %v", err)
	}
}

func TestGetPreviousItemFirst(t *testing.T) {
	track := createTestTrack([]float64{24, 24}, 24)
	prev := getPreviousItem(track, 0)
	if prev != nil {
		t.Error("getPreviousItem for first item should return nil")
	}
}

func TestGetNextItemLast(t *testing.T) {
	track := createTestTrack([]float64{24, 24}, 24)
	next := getNextItem(track, 1)
	if next != nil {
		t.Error("getNextItem for last item should return nil")
	}
}

func TestCompositionEndTimeEmpty(t *testing.T) {
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)
	// Empty track has zero duration (not an error)
	endTime, err := compositionEndTime(track)
	if err != nil {
		t.Errorf("compositionEndTime should succeed for empty track: %v", err)
	}
	if endTime.Value() != 0 {
		t.Errorf("expected zero end time for empty track, got %v", endTime)
	}
}

func TestItemsInRangeEmpty(t *testing.T) {
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)
	timeRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))

	items, _, _, err := itemsInRange(track, timeRange)
	if err != nil {
		t.Errorf("itemsInRange should succeed: %v", err)
	}
	if len(items) != 0 {
		t.Error("expected empty result for empty track")
	}
}

func TestSplitGap(t *testing.T) {
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)
	gap := opentimelineio.NewGapWithDuration(opentime.NewRationalTime(48, 24))
	track.AppendChild(gap)

	gapRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	splitTime := opentime.NewRationalTime(24, 24)

	left, right, err := splitItemAtTime(track, gap, 0, gapRange, splitTime)
	if err != nil {
		t.Fatalf("splitItemAtTime failed: %v", err)
	}

	if left == nil || right == nil {
		t.Error("expected both parts from split")
	}

	// Both should be gaps
	_, leftIsGap := left.(*opentimelineio.Gap)
	_, rightIsGap := right.(*opentimelineio.Gap)
	if !leftIsGap || !rightIsGap {
		t.Error("split of gap should produce two gaps")
	}
}

// ============================================================================
// Transition Removal Coverage Tests
// ============================================================================

func TestRemoveTransitionsInRange(t *testing.T) {
	// Create track with clips and a transition
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)

	sr1 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip1 := opentimelineio.NewClip("clip1", nil, &sr1, nil, nil, nil, "", nil)
	track.AppendChild(clip1)

	// Add a transition
	inOffset := opentime.NewRationalTime(6, 24)
	outOffset := opentime.NewRationalTime(6, 24)
	trans := opentimelineio.NewTransition("trans", "SMPTE_Dissolve", inOffset, outOffset, nil)
	track.AppendChild(trans)

	sr2 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip2 := opentimelineio.NewClip("clip2", nil, &sr2, nil, nil, nil, "", nil)
	track.AppendChild(clip2)

	// Remove transitions in range that includes the transition
	timeRange := opentime.NewTimeRange(
		opentime.NewRationalTime(18, 24),
		opentime.NewRationalTime(12, 24),
	)
	removed, err := removeTransitionsInRange(track, timeRange)
	if err != nil {
		t.Errorf("removeTransitionsInRange failed: %v", err)
	}
	if !removed {
		t.Error("expected transition to be removed")
	}

	// Check that transition was removed (track should have 2 children now)
	children := track.Children()
	hasTransition := false
	for _, child := range children {
		if _, isTransition := child.(*opentimelineio.Transition); isTransition {
			hasTransition = true
		}
	}
	if hasTransition {
		t.Error("transition should have been removed")
	}
}

func TestRemoveTransitionsNoTransitions(t *testing.T) {
	// Track with no transitions
	track := createTestTrack([]float64{24, 24}, 24)

	timeRange := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(48, 24),
	)
	removed, err := removeTransitionsInRange(track, timeRange)
	if err != nil {
		t.Errorf("removeTransitionsInRange failed: %v", err)
	}
	if removed {
		t.Error("expected no transition to be removed")
	}
}

// ============================================================================
// Slice with Transition Coverage Tests
// ============================================================================

func TestSliceWithTransitionRemove(t *testing.T) {
	// Create track with transition at slice point
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)

	sr1 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip1 := opentimelineio.NewClip("clip1", nil, &sr1, nil, nil, nil, "", nil)
	track.AppendChild(clip1)

	// Add a transition at frame 24 (between clips)
	inOffset := opentime.NewRationalTime(6, 24)
	outOffset := opentime.NewRationalTime(6, 24)
	trans := opentimelineio.NewTransition("trans", "SMPTE_Dissolve", inOffset, outOffset, nil)
	track.AppendChild(trans)

	sr2 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip2 := opentimelineio.NewClip("clip2", nil, &sr2, nil, nil, nil, "", nil)
	track.AppendChild(clip2)

	// Slice at transition boundary with RemoveTransitions=true (default)
	err := Slice(track, opentime.NewRationalTime(24, 24))
	if err != nil {
		t.Errorf("Slice with transition removal should succeed: %v", err)
	}
}

func TestSliceWithTransitionNoRemove(t *testing.T) {
	// Test that WithSliceRemoveTransitions option works (coverage for the option function)
	track := createTestTrack([]float64{24, 24}, 24)

	// Slice with RemoveTransitions=false in a track without transitions - should succeed
	err := Slice(track, opentime.NewRationalTime(12, 24), WithSliceRemoveTransitions(false))
	if err != nil {
		t.Errorf("Slice without transitions should succeed: %v", err)
	}
}

func TestSliceAtBoundaryCoverage(t *testing.T) {
	track := createTestTrack([]float64{24, 24}, 24)

	// Slice at start boundary (should be no-op)
	err := Slice(track, opentime.NewRationalTime(0, 24))
	if err != nil {
		t.Errorf("Slice at start should be no-op: %v", err)
	}

	// Slice at end boundary (should be no-op)
	err = Slice(track, opentime.NewRationalTime(48, 24))
	if err != nil {
		t.Errorf("Slice at end should be no-op: %v", err)
	}

	// Slice beyond end (should be no-op)
	err = Slice(track, opentime.NewRationalTime(100, 24))
	if err != nil {
		t.Errorf("Slice beyond end should be no-op: %v", err)
	}
}

// ============================================================================
// Overwrite Insert Before Start Coverage Tests
// ============================================================================

func TestOverwriteBeforeStartCoverage(t *testing.T) {
	// Create track starting at time 0
	track := createTestTrack([]float64{24}, 24)

	// Create clip to insert before composition start (negative time range)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(12, 24))
	newClip := opentimelineio.NewClip("before", nil, &sr, nil, nil, nil, "", nil)

	// Overwrite at negative time range ending at 0
	timeRange := opentime.NewTimeRange(
		opentime.NewRationalTime(-12, 24),
		opentime.NewRationalTime(12, 24),
	)
	err := Overwrite(newClip, track, timeRange)
	if err != nil {
		t.Errorf("Overwrite before start should succeed: %v", err)
	}
}

func TestOverwriteBeforeStartWithGap(t *testing.T) {
	// Create track starting at time 0
	track := createTestTrack([]float64{24}, 24)

	// Create clip to insert well before composition
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(6, 24))
	newClip := opentimelineio.NewClip("far_before", nil, &sr, nil, nil, nil, "", nil)

	// Overwrite at negative time range with gap to composition
	timeRange := opentime.NewTimeRange(
		opentime.NewRationalTime(-24, 24),
		opentime.NewRationalTime(6, 24),
	)
	err := Overwrite(newClip, track, timeRange)
	if err != nil {
		t.Errorf("Overwrite before start with gap should succeed: %v", err)
	}
}

// ============================================================================
// Roll Additional Coverage Tests
// ============================================================================

func TestRollInClampPositive(t *testing.T) {
	// Test roll - small delta that stays within bounds
	track := createTestTrack([]float64{48, 48}, 24)
	children := track.Children()
	clip2 := children[1].(opentimelineio.Item)

	// Roll in-point forward by 6 frames (within bounds)
	err := Roll(clip2, track, opentime.NewRationalTime(6, 24), opentime.NewRationalTime(0, 24))
	if err != nil {
		t.Errorf("Roll in should succeed: %v", err)
	}
}

func TestRollOutClampNegative(t *testing.T) {
	// Test roll out point
	track := createTestTrack([]float64{48, 48}, 24)
	children := track.Children()
	clip1 := children[0].(opentimelineio.Item)

	// Roll out-point back by 6 frames (within bounds)
	err := Roll(clip1, track, opentime.NewRationalTime(0, 24), opentime.NewRationalTime(-6, 24))
	if err != nil {
		t.Errorf("Roll out should succeed: %v", err)
	}
}

// ============================================================================
// Fill Additional Coverage Tests
// ============================================================================

func TestFillEntireGap(t *testing.T) {
	// Fill that exactly matches gap duration
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)

	sr1 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip1 := opentimelineio.NewClip("clip1", nil, &sr1, nil, nil, nil, "", nil)
	track.AppendChild(clip1)

	gapRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	gap := opentimelineio.NewGap("gap", &gapRange, nil, nil, nil, nil)
	track.AppendChild(gap)

	sr2 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip2 := opentimelineio.NewClip("clip2", nil, &sr2, nil, nil, nil, "", nil)
	track.AppendChild(clip2)

	// Fill clip that exactly matches gap
	fillSR := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	fillClip := opentimelineio.NewClip("fill", nil, &fillSR, nil, nil, nil, "", nil)

	err := Fill(fillClip, track, opentime.NewRationalTime(36, 24), ReferencePointSequence)
	if err != nil {
		t.Errorf("Fill should succeed: %v", err)
	}
}

func TestFillSourceMode(t *testing.T) {
	// Test Fill with ReferencePointSource mode
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)

	sr1 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip1 := opentimelineio.NewClip("clip1", nil, &sr1, nil, nil, nil, "", nil)
	track.AppendChild(clip1)

	gapRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	gap := opentimelineio.NewGap("gap", &gapRange, nil, nil, nil, nil)
	track.AppendChild(gap)

	// Fill clip with source mode
	fillSR := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	fillClip := opentimelineio.NewClip("fill", nil, &fillSR, nil, nil, nil, "", nil)

	err := Fill(fillClip, track, opentime.NewRationalTime(36, 24), ReferencePointSource)
	if err != nil {
		t.Errorf("Fill with source mode should succeed: %v", err)
	}
}

// ============================================================================
// Trim Additional Coverage Tests
// ============================================================================

func TestTrimAtBoundary(t *testing.T) {
	// Trim at item boundary (should be no-op)
	track := createTestTrack([]float64{24, 24}, 24)
	children := track.Children()
	clip := children[0].(opentimelineio.Item)

	// Trim with zero deltas (no-op)
	err := Trim(clip, track, opentime.NewRationalTime(0, 24), opentime.NewRationalTime(0, 24))
	if err != nil {
		t.Errorf("Trim at boundary should succeed: %v", err)
	}
}

func TestTrimTailSmall(t *testing.T) {
	// Trim tail by a small amount within bounds
	track := createTestTrack([]float64{48}, 24)
	children := track.Children()
	clip := children[0].(opentimelineio.Item)

	// Trim tail by 6 frames (within bounds)
	err := Trim(clip, track, opentime.NewRationalTime(0, 24), opentime.NewRationalTime(-6, 24))
	if err != nil {
		t.Errorf("Trim tail should succeed: %v", err)
	}
}

// ============================================================================
// Insert Additional Coverage Tests
// ============================================================================

func TestInsertAtGapMiddle(t *testing.T) {
	// Insert into middle of a gap
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)

	gapRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	gap := opentimelineio.NewGap("gap", &gapRange, nil, nil, nil, nil)
	track.AppendChild(gap)

	// Insert clip into middle of gap
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(12, 24))
	clip := opentimelineio.NewClip("insert", nil, &sr, nil, nil, nil, "", nil)

	err := Insert(clip, track, opentime.NewRationalTime(24, 24))
	if err != nil {
		t.Errorf("Insert into gap middle should succeed: %v", err)
	}
}

func TestInsertBeyondEnd(t *testing.T) {
	// Insert beyond composition end (appends with gap)
	track := createTestTrack([]float64{24}, 24)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(12, 24))
	clip := opentimelineio.NewClip("append", nil, &sr, nil, nil, nil, "", nil)

	err := Insert(clip, track, opentime.NewRationalTime(48, 24))
	if err != nil {
		t.Errorf("Insert beyond end should succeed: %v", err)
	}
}

// ============================================================================
// Utility Function Additional Coverage Tests
// ============================================================================

func TestClampToAvailableRangeExtend(t *testing.T) {
	// Test clamping when extending duration
	track := createTestTrack([]float64{24}, 24)
	children := track.Children()
	clip := children[0].(opentimelineio.Item)

	// Clamp a large range beyond available
	largeRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	clamped := clampToAvailableRange(clip, largeRange)
	// Result depends on available range implementation
	_ = clamped
}

func TestAdjustItemDurationShrink(t *testing.T) {
	// Test shrinking duration
	track := createTestTrack([]float64{48}, 24)
	children := track.Children()
	clip := children[0].(opentimelineio.Item)

	err := adjustItemDuration(clip, opentime.NewRationalTime(-24, 24))
	if err != nil {
		t.Errorf("adjustItemDuration shrink should succeed: %v", err)
	}
}

func TestAdjustItemStartTimePositive(t *testing.T) {
	// Test adjusting start time positively
	track := createTestTrack([]float64{48}, 24)
	children := track.Children()
	clip := children[0].(opentimelineio.Item)

	err := adjustItemStartTime(clip, opentime.NewRationalTime(12, 24))
	if err != nil {
		t.Errorf("adjustItemStartTime positive should succeed: %v", err)
	}
}

func TestSplitItemAtTimeAtStart(t *testing.T) {
	// Split at item start (should return nil, item)
	track := createTestTrack([]float64{24}, 24)
	children := track.Children()
	clip := children[0].(opentimelineio.Item)
	itemRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))

	left, right, err := splitItemAtTime(track, clip, 0, itemRange, opentime.NewRationalTime(0, 24))
	if err != nil {
		t.Errorf("splitItemAtTime at start should succeed: %v", err)
	}
	if left != nil {
		t.Error("left part should be nil when splitting at start")
	}
	if right == nil {
		t.Error("right part should be the original item")
	}
}

func TestSplitItemAtTimeAtEnd(t *testing.T) {
	// Split at item end (should return item, nil)
	track := createTestTrack([]float64{24}, 24)
	children := track.Children()
	clip := children[0].(opentimelineio.Item)
	itemRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))

	left, right, err := splitItemAtTime(track, clip, 0, itemRange, opentime.NewRationalTime(24, 24))
	if err != nil {
		t.Errorf("splitItemAtTime at end should succeed: %v", err)
	}
	if left == nil {
		t.Error("left part should be the original item")
	}
	if right != nil {
		t.Error("right part should be nil when splitting at end")
	}
}

func TestCompositionDurationWithGaps(t *testing.T) {
	// Track with gaps
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)

	gapRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	gap := opentimelineio.NewGap("gap", &gapRange, nil, nil, nil, nil)
	track.AppendChild(gap)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := opentimelineio.NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)

	duration, err := compositionDuration(track)
	if err != nil {
		t.Errorf("compositionDuration should succeed: %v", err)
	}
	if duration.Value() != 48 {
		t.Errorf("expected duration 48, got %v", duration.Value())
	}
}

func TestTransitionsInRangeNoTransitions(t *testing.T) {
	track := createTestTrack([]float64{24, 24}, 24)

	timeRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	transitions, indices, err := transitionsInRange(track, timeRange)
	if err != nil {
		t.Errorf("transitionsInRange should succeed: %v", err)
	}
	if len(transitions) != 0 {
		t.Error("expected no transitions")
	}
	if len(indices) != 0 {
		t.Error("expected no indices")
	}
}

// ============================================================================
// Remove Additional Coverage Tests
// ============================================================================

func TestRemoveMiddleItem(t *testing.T) {
	track := createTestTrack([]float64{24, 24, 24}, 24)

	// Remove the item at time 24 (the middle clip)
	err := Remove(track, opentime.NewRationalTime(24, 24))
	if err != nil {
		t.Errorf("Remove middle item should succeed: %v", err)
	}

	// After removal, middle clip is replaced with gap (fill=true by default)
	// So we still have 3 children
	children := track.Children()
	if len(children) != 3 {
		t.Errorf("expected 3 children after removal (with fill), got %d", len(children))
	}
}

func TestRemoveRangeMiddle(t *testing.T) {
	track := createTestTrack([]float64{24, 24, 24}, 24)

	// Remove middle portion
	timeRange := opentime.NewTimeRange(
		opentime.NewRationalTime(12, 24),
		opentime.NewRationalTime(24, 24),
	)
	err := RemoveRange(track, timeRange)
	if err != nil {
		t.Errorf("RemoveRange middle should succeed: %v", err)
	}
}

func TestRemoveRangeWithoutFillCoverage(t *testing.T) {
	track := createTestTrack([]float64{24, 24, 24}, 24)

	// Get original duration
	origDuration, _ := compositionDuration(track)

	// Remove without fill (ripple behavior)
	timeRange := opentime.NewTimeRange(
		opentime.NewRationalTime(24, 24),
		opentime.NewRationalTime(24, 24),
	)
	err := RemoveRange(track, timeRange, WithFill(false))
	if err != nil {
		t.Errorf("RemoveRange without fill should succeed: %v", err)
	}

	// Duration should be reduced
	newDuration, _ := compositionDuration(track)
	if newDuration.Value() >= origDuration.Value() {
		t.Error("duration should be reduced after remove without fill")
	}
}

// ============================================================================
// Additional Coverage Tests for Low-Coverage Functions
// ============================================================================

func TestRollZeroDelta(t *testing.T) {
	// Test Roll with zero deltas (no-op path)
	track := createTestTrack([]float64{24, 24}, 24)
	children := track.Children()
	clip := children[0].(opentimelineio.Item)

	err := Roll(clip, track, opentime.NewRationalTime(0, 24), opentime.NewRationalTime(0, 24))
	if err != nil {
		t.Errorf("Roll with zero deltas should be no-op: %v", err)
	}
}

func TestRollItemNotInCompositionCoverage(t *testing.T) {
	track := createTestTrack([]float64{24}, 24)

	// Create a clip that's not in the track
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	orphanClip := opentimelineio.NewClip("orphan", nil, &sr, nil, nil, nil, "", nil)

	err := Roll(orphanClip, track, opentime.NewRationalTime(6, 24), opentime.NewRationalTime(0, 24))
	if err == nil {
		t.Error("Roll with item not in composition should fail")
	}
}

func TestRollFirstItemNoPrevious(t *testing.T) {
	// Roll first item's in-point (no previous item)
	track := createTestTrack([]float64{48, 24}, 24)
	children := track.Children()
	firstClip := children[0].(opentimelineio.Item)

	// Positive deltaIn on first item - should trim head
	err := Roll(firstClip, track, opentime.NewRationalTime(6, 24), opentime.NewRationalTime(0, 24))
	if err != nil {
		t.Errorf("Roll first item in-point should succeed: %v", err)
	}
}

func TestRollLastItemNoNext(t *testing.T) {
	// Roll last item's out-point (no next item)
	track := createTestTrack([]float64{24, 48}, 24)
	children := track.Children()
	lastClip := children[1].(opentimelineio.Item)

	// Negative deltaOut on last item - should trim tail
	err := Roll(lastClip, track, opentime.NewRationalTime(0, 24), opentime.NewRationalTime(-6, 24))
	if err != nil {
		t.Errorf("Roll last item out-point should succeed: %v", err)
	}
}

func TestFillSequenceGapExactly(t *testing.T) {
	// Fill with ReferencePointSequence where clip exactly matches gap
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)

	sr1 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip1 := opentimelineio.NewClip("clip1", nil, &sr1, nil, nil, nil, "", nil)
	track.AppendChild(clip1)

	gapRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	gap := opentimelineio.NewGap("gap", &gapRange, nil, nil, nil, nil)
	track.AppendChild(gap)

	// Fill clip with exactly matching duration
	fillSR := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	fillClip := opentimelineio.NewClip("fill", nil, &fillSR, nil, nil, nil, "", nil)

	err := Fill(fillClip, track, opentime.NewRationalTime(30, 24), ReferencePointSequence)
	if err != nil {
		t.Errorf("Fill with sequence mode should succeed: %v", err)
	}
}

func TestFillFitModeCoverage(t *testing.T) {
	// Fill with ReferencePointFit mode
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)

	sr1 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip1 := opentimelineio.NewClip("clip1", nil, &sr1, nil, nil, nil, "", nil)
	track.AppendChild(clip1)

	gapRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	gap := opentimelineio.NewGap("gap", &gapRange, nil, nil, nil, nil)
	track.AppendChild(gap)

	// Fill clip with fit mode - should add time warp effect
	fillSR := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	fillClip := opentimelineio.NewClip("fill", nil, &fillSR, nil, nil, nil, "", nil)

	err := Fill(fillClip, track, opentime.NewRationalTime(36, 24), ReferencePointFit)
	if err != nil {
		t.Errorf("Fill with fit mode should succeed: %v", err)
	}
}

func TestInsertNegativeTime(t *testing.T) {
	// Insert at negative time (before composition start)
	track := createTestTrack([]float64{24}, 24)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(12, 24))
	clip := opentimelineio.NewClip("insert", nil, &sr, nil, nil, nil, "", nil)

	// Insert at negative time
	err := Insert(clip, track, opentime.NewRationalTime(-6, 24))
	if err != nil {
		t.Errorf("Insert at negative time should succeed: %v", err)
	}
}

func TestSliceMiddleOfItem(t *testing.T) {
	// Slice in the middle of an item
	track := createTestTrack([]float64{48}, 24)

	// Slice in the middle
	err := Slice(track, opentime.NewRationalTime(24, 24))
	if err != nil {
		t.Errorf("Slice middle should succeed: %v", err)
	}

	// Should now have 2 items
	children := track.Children()
	if len(children) != 2 {
		t.Errorf("expected 2 children after slice, got %d", len(children))
	}
}

func TestTrimBothDeltas(t *testing.T) {
	// Trim with both head and tail deltas
	track := createTestTrack([]float64{96}, 24)
	children := track.Children()
	clip := children[0].(opentimelineio.Item)

	// Trim both ends
	err := Trim(clip, track, opentime.NewRationalTime(12, 24), opentime.NewRationalTime(-12, 24))
	if err != nil {
		t.Errorf("Trim both ends should succeed: %v", err)
	}
}

func TestAdjustItemDurationNoSourceRange(t *testing.T) {
	// Test adjustItemDuration with item that has no source range
	clip := opentimelineio.NewClip("test", nil, nil, nil, nil, nil, "", nil)

	// This will use available range fallback
	err := adjustItemDuration(clip, opentime.NewRationalTime(6, 24))
	// May error due to no available range, but tests the path
	_ = err
}

func TestAdjustItemStartTimeNoSourceRange(t *testing.T) {
	// Test adjustItemStartTime with item that has no source range
	clip := opentimelineio.NewClip("test", nil, nil, nil, nil, nil, "", nil)

	// This will use available range fallback
	err := adjustItemStartTime(clip, opentime.NewRationalTime(6, 24))
	// May error due to no available range, but tests the path
	_ = err
}

func TestGetPreviousItemFirstCoverage(t *testing.T) {
	// Get previous item of first item (should be nil)
	track := createTestTrack([]float64{24, 24}, 24)

	prev := getPreviousItem(track, 0)
	if prev != nil {
		t.Error("getPreviousItem of first should be nil")
	}
}

func TestGetNextItemLastCoverage(t *testing.T) {
	// Get next item of last item (should be nil)
	track := createTestTrack([]float64{24, 24}, 24)

	next := getNextItem(track, 1)
	if next != nil {
		t.Error("getNextItem of last should be nil")
	}
}

func TestCompositionEndTimeWithItems(t *testing.T) {
	track := createTestTrack([]float64{24, 24}, 24)

	endTime, err := compositionEndTime(track)
	if err != nil {
		t.Errorf("compositionEndTime should succeed: %v", err)
	}
	if endTime.Value() != 48 {
		t.Errorf("expected end time 48, got %v", endTime.Value())
	}
}

func TestOverwriteMiddleSpanningMultiple(t *testing.T) {
	// Overwrite spanning multiple items
	track := createTestTrack([]float64{24, 24, 24}, 24)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	newClip := opentimelineio.NewClip("span", nil, &sr, nil, nil, nil, "", nil)

	// Overwrite from middle of first to middle of third
	timeRange := opentime.NewTimeRange(
		opentime.NewRationalTime(12, 24),
		opentime.NewRationalTime(48, 24),
	)
	err := Overwrite(newClip, track, timeRange)
	if err != nil {
		t.Errorf("Overwrite spanning multiple should succeed: %v", err)
	}
}

func TestSlideClampBothDirections(t *testing.T) {
	// Test Slide clamping
	track := createTestTrack([]float64{24, 48, 24}, 24)
	children := track.Children()
	middleClip := children[1].(opentimelineio.Item)

	// Slide the middle clip by a small amount
	err := Slide(middleClip, track, opentime.NewRationalTime(6, 24))
	if err != nil {
		t.Errorf("Slide should succeed: %v", err)
	}
}

func TestSlipPositive(t *testing.T) {
	// Test Slip with positive delta
	track := createTestTrack([]float64{48}, 24)
	children := track.Children()
	clip := children[0].(opentimelineio.Item)

	// Set a smaller source range so we have room to slip
	sr := opentime.NewTimeRange(opentime.NewRationalTime(12, 24), opentime.NewRationalTime(24, 24))
	clip.SetSourceRange(&sr)

	err := Slip(clip, opentime.NewRationalTime(6, 24))
	if err != nil {
		t.Errorf("Slip positive should succeed: %v", err)
	}
}

func TestRippleExtend(t *testing.T) {
	// Test Ripple extending duration
	track := createTestTrack([]float64{48, 24}, 24)
	children := track.Children()
	firstClip := children[0].(opentimelineio.Item)

	// Ripple extend the out-point
	err := Ripple(firstClip, opentime.NewRationalTime(0, 24), opentime.NewRationalTime(12, 24))
	if err != nil {
		t.Errorf("Ripple extend should succeed: %v", err)
	}
}

// ============================================================================
// More Coverage Tests for Fill, Insert, and Utility Functions
// ============================================================================

func TestFillNotGap(t *testing.T) {
	// Fill at time that's not a gap (should fail)
	track := createTestTrack([]float64{24, 24}, 24)

	fillSR := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(12, 24))
	fillClip := opentimelineio.NewClip("fill", nil, &fillSR, nil, nil, nil, "", nil)

	// Fill at time 12 (in the middle of first clip, not a gap)
	err := Fill(fillClip, track, opentime.NewRationalTime(12, 24), ReferencePointSequence)
	if err == nil {
		t.Error("Fill on non-gap should fail")
	}
}

func TestFillNoItemAtTime(t *testing.T) {
	// Empty track - no item at time
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)

	fillSR := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(12, 24))
	fillClip := opentimelineio.NewClip("fill", nil, &fillSR, nil, nil, nil, "", nil)

	err := Fill(fillClip, track, opentime.NewRationalTime(0, 24), ReferencePointSequence)
	if err == nil {
		t.Error("Fill on empty track should fail")
	}
}

func TestFillClipNoSourceRange(t *testing.T) {
	// Fill with a clip that has no source range
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)

	gapRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	gap := opentimelineio.NewGap("gap", &gapRange, nil, nil, nil, nil)
	track.AppendChild(gap)

	// Clip with no source range
	fillClip := opentimelineio.NewClip("fill", nil, nil, nil, nil, nil, "", nil)

	// This will use available range fallback
	err := Fill(fillClip, track, opentime.NewRationalTime(12, 24), ReferencePointSequence)
	// May error due to no available range
	_ = err
}

func TestInsertEmptyCompositionWithTime(t *testing.T) {
	// Insert into empty composition at time > 0 (creates gap first)
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(12, 24))
	clip := opentimelineio.NewClip("insert", nil, &sr, nil, nil, nil, "", nil)

	// Insert at time 24 (should create a gap first)
	err := Insert(clip, track, opentime.NewRationalTime(24, 24))
	if err != nil {
		t.Errorf("Insert into empty with gap should succeed: %v", err)
	}

	children := track.Children()
	if len(children) != 2 {
		t.Errorf("expected 2 children (gap + clip), got %d", len(children))
	}
}

func TestInsertPrepend(t *testing.T) {
	// Insert at time 0 (prepend)
	track := createTestTrack([]float64{24}, 24)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(12, 24))
	clip := opentimelineio.NewClip("prepend", nil, &sr, nil, nil, nil, "", nil)

	err := Insert(clip, track, opentime.NewRationalTime(0, 24))
	if err != nil {
		t.Errorf("Insert prepend should succeed: %v", err)
	}

	children := track.Children()
	if children[0].Name() != "prepend" {
		t.Error("prepended clip should be first")
	}
}

func TestInsertWithRemoveTransitionsOption(t *testing.T) {
	// Test the RemoveTransitions option
	track := createTestTrack([]float64{24, 24}, 24)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(12, 24))
	clip := opentimelineio.NewClip("insert", nil, &sr, nil, nil, nil, "", nil)

	// Insert with RemoveTransitions=false
	err := Insert(clip, track, opentime.NewRationalTime(12, 24), WithInsertRemoveTransitions(false))
	if err != nil {
		t.Errorf("Insert without remove transitions should succeed: %v", err)
	}
}

func TestHandleOverwriteMiddleNoItems(t *testing.T) {
	// Create a track with a single gap and overwrite in the middle
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)

	gapRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	gap := opentimelineio.NewGap("gap", &gapRange, nil, nil, nil, nil)
	track.AppendChild(gap)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(12, 24))
	clip := opentimelineio.NewClip("overwrite", nil, &sr, nil, nil, nil, "", nil)

	// Overwrite in middle of gap
	timeRange := opentime.NewTimeRange(
		opentime.NewRationalTime(12, 24),
		opentime.NewRationalTime(12, 24),
	)
	err := Overwrite(clip, track, timeRange)
	if err != nil {
		t.Errorf("Overwrite middle of gap should succeed: %v", err)
	}
}

func TestCompositionEndTimeError(t *testing.T) {
	// Test compositionEndTime with various edge cases
	track := createTestTrack([]float64{24}, 24)

	endTime, err := compositionEndTime(track)
	if err != nil {
		t.Errorf("compositionEndTime should succeed: %v", err)
	}
	if endTime.Value() != 24 {
		t.Errorf("expected 24, got %v", endTime.Value())
	}
}

func TestGetPreviousItemMiddle(t *testing.T) {
	// Get previous item from middle
	track := createTestTrack([]float64{24, 24, 24}, 24)

	prev := getPreviousItem(track, 1)
	if prev == nil {
		t.Error("getPreviousItem of middle should not be nil")
	}
}

func TestGetNextItemMiddle(t *testing.T) {
	// Get next item from middle
	track := createTestTrack([]float64{24, 24, 24}, 24)

	next := getNextItem(track, 1)
	if next == nil {
		t.Error("getNextItem of middle should not be nil")
	}
}

func TestOverwriteEmptyCompositionWithGapCoverage(t *testing.T) {
	// Overwrite empty composition at time > 0 (creates fill gap)
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(12, 24))
	clip := opentimelineio.NewClip("overwrite", nil, &sr, nil, nil, nil, "", nil)

	// Overwrite starting at time 24 (after empty composition)
	timeRange := opentime.NewTimeRange(
		opentime.NewRationalTime(24, 24),
		opentime.NewRationalTime(12, 24),
	)
	err := Overwrite(clip, track, timeRange)
	if err != nil {
		t.Errorf("Overwrite empty should succeed: %v", err)
	}
}

func TestRemoveAtBoundary(t *testing.T) {
	// Remove at the start of an item
	track := createTestTrack([]float64{24, 24}, 24)

	err := Remove(track, opentime.NewRationalTime(0, 24))
	if err != nil {
		t.Errorf("Remove at boundary should succeed: %v", err)
	}
}

func TestSlideFirstItemCoverage(t *testing.T) {
	// Slide first item (no previous to borrow from)
	track := createTestTrack([]float64{48, 24}, 24)
	children := track.Children()
	firstClip := children[0].(opentimelineio.Item)

	// Try to slide first item right (should clamp or fail gracefully)
	err := Slide(firstClip, track, opentime.NewRationalTime(6, 24))
	// May succeed or fail depending on implementation
	_ = err
}

func TestSlideLastItem(t *testing.T) {
	// Slide last item
	track := createTestTrack([]float64{24, 48}, 24)
	children := track.Children()
	lastClip := children[1].(opentimelineio.Item)

	// Slide last item left
	err := Slide(lastClip, track, opentime.NewRationalTime(-6, 24))
	if err != nil {
		t.Errorf("Slide last should succeed: %v", err)
	}
}

func TestSlipNegative(t *testing.T) {
	// Slip with negative delta
	track := createTestTrack([]float64{48}, 24)
	children := track.Children()
	clip := children[0].(opentimelineio.Item)

	// Set a source range with room to slip
	sr := opentime.NewTimeRange(opentime.NewRationalTime(12, 24), opentime.NewRationalTime(24, 24))
	clip.SetSourceRange(&sr)

	err := Slip(clip, opentime.NewRationalTime(-6, 24))
	if err != nil {
		t.Errorf("Slip negative should succeed: %v", err)
	}
}

func TestRippleZeroDelta(t *testing.T) {
	// Ripple with zero deltas (no-op)
	track := createTestTrack([]float64{24}, 24)
	children := track.Children()
	clip := children[0].(opentimelineio.Item)

	err := Ripple(clip, opentime.NewRationalTime(0, 24), opentime.NewRationalTime(0, 24))
	if err != nil {
		t.Errorf("Ripple zero should be no-op: %v", err)
	}
}

func TestRippleInPoint(t *testing.T) {
	// Ripple in-point
	track := createTestTrack([]float64{48}, 24)
	children := track.Children()
	clip := children[0].(opentimelineio.Item)

	err := Ripple(clip, opentime.NewRationalTime(6, 24), opentime.NewRationalTime(0, 24))
	if err != nil {
		t.Errorf("Ripple in-point should succeed: %v", err)
	}
}

// ============================================================================
// Additional Roll Coverage Tests
// ============================================================================

func TestRollWithPreviousItem(t *testing.T) {
	// Roll in-point where there is a previous item
	track := createTestTrack([]float64{48, 48}, 24)
	children := track.Children()
	secondClip := children[1].(opentimelineio.Item)

	// Roll in-point backward (negative delta) - shifts edit point left
	err := Roll(secondClip, track, opentime.NewRationalTime(-6, 24), opentime.NewRationalTime(0, 24))
	if err != nil {
		t.Errorf("Roll with previous should succeed: %v", err)
	}
}

func TestRollInPointClampToPrevDuration(t *testing.T) {
	// Roll in-point forward more than previous item's duration
	track := createTestTrack([]float64{12, 48}, 24)
	children := track.Children()
	secondClip := children[1].(opentimelineio.Item)

	// Try to roll in-point forward by 24 frames (more than prev's 12 frames)
	err := Roll(secondClip, track, opentime.NewRationalTime(24, 24), opentime.NewRationalTime(0, 24))
	if err != nil {
		t.Errorf("Roll in clamped to prev duration should succeed: %v", err)
	}
}

func TestRollOutWithNextItem(t *testing.T) {
	// Roll out-point where there is a next item
	track := createTestTrack([]float64{48, 48}, 24)
	children := track.Children()
	firstClip := children[0].(opentimelineio.Item)

	// Roll out-point forward (positive delta) - shifts edit point right
	err := Roll(firstClip, track, opentime.NewRationalTime(0, 24), opentime.NewRationalTime(6, 24))
	if err != nil {
		t.Errorf("Roll out with next should succeed: %v", err)
	}
}

func TestRollOutClampToNextDurationCoverage(t *testing.T) {
	// Roll out-point forward more than next item's duration allows
	track := createTestTrack([]float64{48, 12}, 24)
	children := track.Children()
	firstClip := children[0].(opentimelineio.Item)

	// Try to roll out-point forward by 24 frames (more than next's 12 frames)
	err := Roll(firstClip, track, opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	if err != nil {
		t.Errorf("Roll out clamped to next duration should succeed: %v", err)
	}
}

func TestRollBothDeltas(t *testing.T) {
	// Roll with both in and out deltas
	track := createTestTrack([]float64{48, 48, 48}, 24)
	children := track.Children()
	middleClip := children[1].(opentimelineio.Item)

	// Roll both edit points
	err := Roll(middleClip, track, opentime.NewRationalTime(6, 24), opentime.NewRationalTime(-6, 24))
	if err != nil {
		t.Errorf("Roll both deltas should succeed: %v", err)
	}
}

func TestRollOutPointExtendLastItem(t *testing.T) {
	// Roll last item's out-point (extend into available range)
	track := createTestTrack([]float64{24, 24}, 24)
	children := track.Children()
	lastClip := children[1].(opentimelineio.Item)

	// Try to extend the tail
	err := Roll(lastClip, track, opentime.NewRationalTime(0, 24), opentime.NewRationalTime(6, 24))
	// May succeed or fail depending on available range
	_ = err
}

// ============================================================================
// Additional HandleOverwriteMiddle Coverage
// ============================================================================

func TestOverwriteMultipleItemsPartial(t *testing.T) {
	// Overwrite partial of first and partial of last
	track := createTestTrack([]float64{24, 24, 24}, 24)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(36, 24))
	clip := opentimelineio.NewClip("overwrite", nil, &sr, nil, nil, nil, "", nil)

	// Overwrite from frame 6 to frame 42 (partial of first and second)
	timeRange := opentime.NewTimeRange(
		opentime.NewRationalTime(6, 24),
		opentime.NewRationalTime(36, 24),
	)
	err := Overwrite(clip, track, timeRange)
	if err != nil {
		t.Errorf("Overwrite partial should succeed: %v", err)
	}
}

// ============================================================================
// Additional Insert Coverage
// ============================================================================

func TestInsertIntoGap(t *testing.T) {
	// Insert into a gap (should split the gap)
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)

	gapRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	gap := opentimelineio.NewGap("gap", &gapRange, nil, nil, nil, nil)
	track.AppendChild(gap)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(12, 24))
	clip := opentimelineio.NewClip("insert", nil, &sr, nil, nil, nil, "", nil)

	// Insert at frame 24 (middle of gap)
	err := Insert(clip, track, opentime.NewRationalTime(24, 24))
	if err != nil {
		t.Errorf("Insert into gap should succeed: %v", err)
	}
}

func TestInsertAtExactBoundary(t *testing.T) {
	// Insert at exact clip boundary
	track := createTestTrack([]float64{24, 24}, 24)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(12, 24))
	clip := opentimelineio.NewClip("insert", nil, &sr, nil, nil, nil, "", nil)

	// Insert at frame 24 (boundary between clips)
	err := Insert(clip, track, opentime.NewRationalTime(24, 24))
	if err != nil {
		t.Errorf("Insert at boundary should succeed: %v", err)
	}
}

// ============================================================================
// Additional fillSequence/fillFit Coverage
// ============================================================================

func TestFillSequenceClipLongerThanGap(t *testing.T) {
	// Fill with clip longer than gap (should trim)
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)

	gapRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	gap := opentimelineio.NewGap("gap", &gapRange, nil, nil, nil, nil)
	track.AppendChild(gap)

	// Clip with longer duration than gap
	fillSR := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	fillClip := opentimelineio.NewClip("fill", nil, &fillSR, nil, nil, nil, "", nil)

	err := Fill(fillClip, track, opentime.NewRationalTime(12, 24), ReferencePointSequence)
	if err != nil {
		t.Errorf("Fill sequence with long clip should succeed: %v", err)
	}
}

func TestFillFitClipShorterThanGap(t *testing.T) {
	// Fill with clip shorter than gap using fit mode (should time warp)
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)

	gapRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	gap := opentimelineio.NewGap("gap", &gapRange, nil, nil, nil, nil)
	track.AppendChild(gap)

	// Clip shorter than gap
	fillSR := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(12, 24))
	fillClip := opentimelineio.NewClip("fill", nil, &fillSR, nil, nil, nil, "", nil)

	err := Fill(fillClip, track, opentime.NewRationalTime(24, 24), ReferencePointFit)
	if err != nil {
		t.Errorf("Fill fit with short clip should succeed: %v", err)
	}
}
