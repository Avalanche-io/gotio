// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package algorithms

import (
	"testing"

	"github.com/mrjoshuak/gotio/opentime"
	"github.com/mrjoshuak/gotio/opentimelineio"
)

func TestTrackTrimmedToRange(t *testing.T) {
	// Create a track with multiple clips
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)

	sr1 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip1 := opentimelineio.NewClip("clip1", nil, &sr1, nil, nil, nil, "", nil)

	sr2 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip2 := opentimelineio.NewClip("clip2", nil, &sr2, nil, nil, nil, "", nil)

	sr3 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip3 := opentimelineio.NewClip("clip3", nil, &sr3, nil, nil, nil, "", nil)

	track.AppendChild(clip1)
	track.AppendChild(clip2)
	track.AppendChild(clip3)

	// Trim to a range that includes the middle clip
	trimRange := opentime.NewTimeRange(opentime.NewRationalTime(24, 24), opentime.NewRationalTime(48, 24))

	result, err := TrackTrimmedToRange(track, trimRange)
	if err != nil {
		t.Fatalf("TrackTrimmedToRange error: %v", err)
	}

	// The result should contain trimmed portions of the original clips
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	t.Logf("Original track children: %d", len(track.Children()))
	t.Logf("Trimmed track children: %d", len(result.Children()))
}

func TestTrackTrimmedToRangeEmpty(t *testing.T) {
	track := opentimelineio.NewTrack("empty", nil, opentimelineio.TrackKindVideo, nil, nil)

	trimRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))

	result, err := TrackTrimmedToRange(track, trimRange)
	if err != nil {
		t.Fatalf("TrackTrimmedToRange error: %v", err)
	}

	if len(result.Children()) != 0 {
		t.Errorf("Expected 0 children, got %d", len(result.Children()))
	}
}

func TestIntersectRanges(t *testing.T) {
	a := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	b := opentime.NewTimeRange(opentime.NewRationalTime(50, 24), opentime.NewRationalTime(100, 24))

	result := intersectRanges(a, b)

	// Intersection should be 50-100
	if result.StartTime().Value() != 50 {
		t.Errorf("Start = %v, want 50", result.StartTime().Value())
	}
	if result.Duration().Value() != 50 {
		t.Errorf("Duration = %v, want 50", result.Duration().Value())
	}
}

func TestTrackWithExpandedTransitions(t *testing.T) {
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)

	sr1 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip1 := opentimelineio.NewClip("clip1", nil, &sr1, nil, nil, nil, "", nil)

	sr2 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip2 := opentimelineio.NewClip("clip2", nil, &sr2, nil, nil, nil, "", nil)

	inOffset := opentime.NewRationalTime(12, 24)
	outOffset := opentime.NewRationalTime(12, 24)
	transition := opentimelineio.NewTransition("dissolve", "SMPTE_Dissolve", inOffset, outOffset, nil)

	track.AppendChild(clip1)
	track.AppendChild(transition)
	track.AppendChild(clip2)

	result, err := TrackWithExpandedTransitions(track)
	if err != nil {
		t.Fatalf("TrackWithExpandedTransitions error: %v", err)
	}

	t.Logf("Original children: %d", len(track.Children()))
	t.Logf("Expanded children: %d", len(result.Children()))
}

func TestTrackWithExpandedTransitionsNoTransitions(t *testing.T) {
	track := opentimelineio.NewTrack("test", nil, opentimelineio.TrackKindVideo, nil, nil)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := opentimelineio.NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)

	result, err := TrackWithExpandedTransitions(track)
	if err != nil {
		t.Fatalf("TrackWithExpandedTransitions error: %v", err)
	}

	if len(result.Children()) != 1 {
		t.Errorf("Expected 1 child, got %d", len(result.Children()))
	}
}

func TestTrackWithExpandedTransitionsEmpty(t *testing.T) {
	track := opentimelineio.NewTrack("empty", nil, opentimelineio.TrackKindVideo, nil, nil)

	result, err := TrackWithExpandedTransitions(track)
	if err != nil {
		t.Fatalf("TrackWithExpandedTransitions error: %v", err)
	}

	if len(result.Children()) != 0 {
		t.Errorf("Expected 0 children, got %d", len(result.Children()))
	}
}
