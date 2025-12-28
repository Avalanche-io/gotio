// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package algorithms

import (
	"testing"

	"github.com/Avalanche-io/gotio/opentime"
	"github.com/Avalanche-io/gotio"
)

// Helper function to create a test track with clips
func createTestTrack(clipDurations []float64, rate float64) *gotio.Track {
	track := gotio.NewTrack("test_track", nil, gotio.TrackKindVideo, nil, nil)
	for i, dur := range clipDurations {
		sr := opentime.NewTimeRange(
			opentime.NewRationalTime(0, rate),
			opentime.NewRationalTime(dur, rate),
		)
		clip := gotio.NewClip(
			"clip_"+string(rune('A'+i)),
			nil,
			&sr,
			nil, nil, nil, "", nil,
		)
		track.AppendChild(clip)
	}
	return track
}

// Helper function to create a test track with clips that have available ranges
func createTestTrackWithAvailableRange(clipDurations []float64, availableRange float64, rate float64) *gotio.Track {
	track := gotio.NewTrack("test_track", nil, gotio.TrackKindVideo, nil, nil)
	for i, dur := range clipDurations {
		ar := opentime.NewTimeRange(
			opentime.NewRationalTime(0, rate),
			opentime.NewRationalTime(availableRange, rate),
		)
		ref := gotio.NewExternalReference("", "file://test.mov", &ar, nil)

		sr := opentime.NewTimeRange(
			opentime.NewRationalTime(0, rate),
			opentime.NewRationalTime(dur, rate),
		)
		clip := gotio.NewClip(
			"clip_"+string(rune('A'+i)),
			ref,
			&sr,
			nil, nil, nil, "", nil,
		)
		track.AppendChild(clip)
	}
	return track
}

func TestItemAtTime(t *testing.T) {
	track := createTestTrack([]float64{24, 48, 24}, 24)

	tests := []struct {
		name      string
		time      float64
		wantName  string
		wantIndex int
		wantFound bool
	}{
		{"at start", 0, "clip_A", 0, true},
		{"middle of first", 12, "clip_A", 0, true},
		{"start of second", 24, "clip_B", 1, true},
		{"middle of second", 48, "clip_B", 1, true},
		{"start of third", 72, "clip_C", 2, true},
		{"past end", 100, "", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			time := opentime.NewRationalTime(tt.time, 24)
			item, index, _, err := itemAtTime(track, time)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantFound {
				if item == nil {
					t.Fatalf("expected to find item, got nil")
				}
				if item.Name() != tt.wantName {
					t.Errorf("expected item name %s, got %s", tt.wantName, item.Name())
				}
				if index != tt.wantIndex {
					t.Errorf("expected index %d, got %d", tt.wantIndex, index)
				}
			} else {
				if item != nil {
					t.Errorf("expected no item, got %s", item.Name())
				}
				if index != -1 {
					t.Errorf("expected index -1, got %d", index)
				}
			}
		})
	}
}

func TestItemsInRange(t *testing.T) {
	track := createTestTrack([]float64{24, 24, 24, 24}, 24)

	tests := []struct {
		name       string
		rangeStart float64
		rangeDur   float64
		wantCount  int
		wantNames  []string
	}{
		{"single item", 0, 24, 1, []string{"clip_A"}},
		{"spanning two", 12, 24, 2, []string{"clip_A", "clip_B"}},
		{"spanning all", 0, 96, 4, []string{"clip_A", "clip_B", "clip_C", "clip_D"}},
		{"middle items", 24, 48, 2, []string{"clip_B", "clip_C"}},
		{"past end", 100, 24, 0, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeRange := opentime.NewTimeRange(
				opentime.NewRationalTime(tt.rangeStart, 24),
				opentime.NewRationalTime(tt.rangeDur, 24),
			)
			items, _, _, err := itemsInRange(track, timeRange)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(items) != tt.wantCount {
				t.Fatalf("expected %d items, got %d", tt.wantCount, len(items))
			}

			for i, item := range items {
				if item.Name() != tt.wantNames[i] {
					t.Errorf("item %d: expected name %s, got %s", i, tt.wantNames[i], item.Name())
				}
			}
		})
	}
}

func TestSplitItemAtTime(t *testing.T) {
	rate := 24.0

	tests := []struct {
		name         string
		itemDuration float64
		splitOffset  float64
		wantFirst    float64
		wantSecond   float64
		wantNilFirst bool
		wantNilSec   bool
	}{
		{"split in middle", 48, 24, 24, 24, false, false},
		{"split at 1/4", 48, 12, 12, 36, false, false},
		{"split at 3/4", 48, 36, 36, 12, false, false},
		{"split at start", 48, 0, 0, 48, true, false},
		{"split at end", 48, 48, 48, 0, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a track with a single clip
			track := createTestTrack([]float64{tt.itemDuration}, rate)
			children := track.Children()
			item := children[0].(gotio.Item)
			itemRange, _ := track.RangeOfChildAtIndex(0)

			splitTime := opentime.NewRationalTime(tt.splitOffset, rate)
			first, second, err := splitItemAtTime(track, item, 0, itemRange, splitTime)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantNilFirst {
				if first != nil {
					t.Error("expected nil first part")
				}
			} else {
				if first == nil {
					t.Fatal("expected non-nil first part")
				}
				firstDur, _ := first.Duration()
				if firstDur.Value() != tt.wantFirst {
					t.Errorf("first part duration: expected %.0f, got %.0f", tt.wantFirst, firstDur.Value())
				}
			}

			if tt.wantNilSec {
				if second != nil {
					t.Error("expected nil second part")
				}
			} else {
				if second == nil {
					t.Fatal("expected non-nil second part")
				}
				secondDur, _ := second.Duration()
				if secondDur.Value() != tt.wantSecond {
					t.Errorf("second part duration: expected %.0f, got %.0f", tt.wantSecond, secondDur.Value())
				}
			}
		})
	}
}

func TestClampToAvailableRange(t *testing.T) {
	rate := 24.0

	// Create a clip with available range of 0-100 frames
	ar := opentime.NewTimeRange(
		opentime.NewRationalTime(0, rate),
		opentime.NewRationalTime(100, rate),
	)
	ref := gotio.NewExternalReference("", "file://test.mov", &ar, nil)
	clip := gotio.NewClip("test", ref, nil, nil, nil, nil, "", nil)

	tests := []struct {
		name       string
		start      float64
		dur        float64
		wantStart  float64
		wantDur    float64
	}{
		{"within bounds", 10, 50, 10, 50},
		{"start before", -10, 50, 0, 40},
		{"end after", 80, 50, 80, 20},
		{"both out of bounds", -10, 200, 0, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceRange := opentime.NewTimeRange(
				opentime.NewRationalTime(tt.start, rate),
				opentime.NewRationalTime(tt.dur, rate),
			)
			clamped := clampToAvailableRange(clip, sourceRange)

			if clamped.StartTime().Value() != tt.wantStart {
				t.Errorf("start: expected %.0f, got %.0f", tt.wantStart, clamped.StartTime().Value())
			}
			if clamped.Duration().Value() != tt.wantDur {
				t.Errorf("duration: expected %.0f, got %.0f", tt.wantDur, clamped.Duration().Value())
			}
		})
	}
}

func TestCompositionDuration(t *testing.T) {
	track := createTestTrack([]float64{24, 48, 24}, 24)
	dur, err := compositionDuration(track)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dur.Value() != 96 {
		t.Errorf("expected duration 96, got %.0f", dur.Value())
	}
}

func TestCreateFillGap(t *testing.T) {
	duration := opentime.NewRationalTime(24, 24)

	// Without template
	gap := createFillGap(duration, nil)
	if gap == nil {
		t.Fatal("expected non-nil gap")
	}
	gapDur, _ := gap.Duration()
	if gapDur.Value() != 24 {
		t.Errorf("expected duration 24, got %.0f", gapDur.Value())
	}

	// With gap template
	templateRange := opentime.NewTimeRange(opentime.RationalTime{}, opentime.NewRationalTime(48, 24))
	template := gotio.NewGap("template", &templateRange, nil, nil, nil, nil)
	gap = createFillGap(duration, template)
	if gap == nil {
		t.Fatal("expected non-nil gap")
	}
	gapDur, _ = gap.Duration()
	if gapDur.Value() != 24 {
		t.Errorf("expected duration 24, got %.0f", gapDur.Value())
	}
}

func TestGetPreviousNextItem(t *testing.T) {
	track := createTestTrack([]float64{24, 24, 24}, 24)

	// Test getPreviousItem
	prev := getPreviousItem(track, 0)
	if prev != nil {
		t.Error("expected nil previous for index 0")
	}

	prev = getPreviousItem(track, 1)
	if prev == nil || prev.Name() != "clip_A" {
		t.Error("expected clip_A as previous for index 1")
	}

	// Test getNextItem
	next := getNextItem(track, 2)
	if next != nil {
		t.Error("expected nil next for last index")
	}

	next = getNextItem(track, 1)
	if next == nil || next.Name() != "clip_C" {
		t.Error("expected clip_C as next for index 1")
	}
}

func TestAdjustItemDuration(t *testing.T) {
	rate := 24.0
	track := createTestTrack([]float64{48}, rate)
	item := track.Children()[0].(gotio.Item)

	// Extend duration
	delta := opentime.NewRationalTime(12, rate)
	err := adjustItemDuration(item, delta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	dur, _ := item.Duration()
	if dur.Value() != 60 {
		t.Errorf("expected duration 60, got %.0f", dur.Value())
	}

	// Reduce duration
	delta = opentime.NewRationalTime(-24, rate)
	err = adjustItemDuration(item, delta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	dur, _ = item.Duration()
	if dur.Value() != 36 {
		t.Errorf("expected duration 36, got %.0f", dur.Value())
	}

	// Negative result should error
	delta = opentime.NewRationalTime(-100, rate)
	err = adjustItemDuration(item, delta)
	if err == nil {
		t.Error("expected error for negative duration")
	}
}

func TestMaxMinRationalTime(t *testing.T) {
	a := opentime.NewRationalTime(24, 24)
	b := opentime.NewRationalTime(48, 24)

	max := maxRationalTime(a, b)
	if max.Value() != 48 {
		t.Errorf("max: expected 48, got %.0f", max.Value())
	}

	min := minRationalTime(a, b)
	if min.Value() != 24 {
		t.Errorf("min: expected 24, got %.0f", min.Value())
	}
}
