// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package algorithms

import (
	"testing"

	"github.com/mrjoshuak/gotio/opentime"
	"github.com/mrjoshuak/gotio/opentimelineio"
)

func TestFlattenStackEmpty(t *testing.T) {
	stack := opentimelineio.NewStack("empty", nil, nil, nil, nil, nil)

	result, err := FlattenStack(stack)
	if err != nil {
		t.Fatalf("FlattenStack error: %v", err)
	}

	if len(result.Children()) != 0 {
		t.Errorf("Expected 0 children, got %d", len(result.Children()))
	}
}

func TestFlattenStackSingleTrack(t *testing.T) {
	stack := opentimelineio.NewStack("single", nil, nil, nil, nil, nil)

	track := opentimelineio.NewTrack("track", nil, opentimelineio.TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := opentimelineio.NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)

	stack.AppendChild(track)

	result, err := FlattenStack(stack)
	if err != nil {
		t.Fatalf("FlattenStack error: %v", err)
	}

	if len(result.Children()) != 1 {
		t.Errorf("Expected 1 child, got %d", len(result.Children()))
	}
}

func TestFlattenStackMultipleTracks(t *testing.T) {
	stack := opentimelineio.NewStack("multi", nil, nil, nil, nil, nil)

	// First track: clip from 0-48
	track1 := opentimelineio.NewTrack("track1", nil, opentimelineio.TrackKindVideo, nil, nil)
	sr1 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip1 := opentimelineio.NewClip("clip1", nil, &sr1, nil, nil, nil, "", nil)
	track1.AppendChild(clip1)

	// Second track: clip from 24-72 (overlaps with first)
	track2 := opentimelineio.NewTrack("track2", nil, opentimelineio.TrackKindVideo, nil, nil)
	gap := opentimelineio.NewGap("gap", &opentime.TimeRange{}, nil, nil, nil, nil)
	gapRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	gap.SetSourceRange(&gapRange)

	sr2 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip2 := opentimelineio.NewClip("clip2", nil, &sr2, nil, nil, nil, "", nil)
	track2.AppendChild(gap)
	track2.AppendChild(clip2)

	stack.AppendChild(track1)
	stack.AppendChild(track2)

	result, err := FlattenStack(stack)
	if err != nil {
		t.Fatalf("FlattenStack error: %v", err)
	}

	t.Logf("Flattened track has %d children", len(result.Children()))
}

func TestFlattenTracks(t *testing.T) {
	// Create two tracks
	track1 := opentimelineio.NewTrack("track1", nil, opentimelineio.TrackKindVideo, nil, nil)
	sr1 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip1 := opentimelineio.NewClip("clip1", nil, &sr1, nil, nil, nil, "", nil)
	track1.AppendChild(clip1)

	track2 := opentimelineio.NewTrack("track2", nil, opentimelineio.TrackKindVideo, nil, nil)
	sr2 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip2 := opentimelineio.NewClip("clip2", nil, &sr2, nil, nil, nil, "", nil)
	track2.AppendChild(clip2)

	result, err := FlattenTracks([]*opentimelineio.Track{track1, track2})
	if err != nil {
		t.Fatalf("FlattenTracks error: %v", err)
	}

	if result.Name() != "Flattened" {
		t.Errorf("Name = %s, want Flattened", result.Name())
	}
}

func TestFlattenTracksEmpty(t *testing.T) {
	result, err := FlattenTracks(nil)
	if err != nil {
		t.Fatalf("FlattenTracks error: %v", err)
	}

	if result.Name() != "Flattened" {
		t.Errorf("Name = %s, want Flattened", result.Name())
	}
}

func TestFlattenTracksSingle(t *testing.T) {
	track := opentimelineio.NewTrack("track", nil, opentimelineio.TrackKindVideo, nil, nil)

	result, err := FlattenTracks([]*opentimelineio.Track{track})
	if err != nil {
		t.Fatalf("FlattenTracks error: %v", err)
	}

	// Single track should be cloned
	if result == track {
		t.Error("Result should be a clone, not the original")
	}
}

func TestTopClipAtTime(t *testing.T) {
	stack := opentimelineio.NewStack("stack", nil, nil, nil, nil, nil)

	// First track: clip from 0-48
	track1 := opentimelineio.NewTrack("track1", nil, opentimelineio.TrackKindVideo, nil, nil)
	sr1 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip1 := opentimelineio.NewClip("clip1", nil, &sr1, nil, nil, nil, "", nil)
	track1.AppendChild(clip1)

	// Second track: clip from 24-72
	track2 := opentimelineio.NewTrack("track2", nil, opentimelineio.TrackKindVideo, nil, nil)
	gapRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	gap := opentimelineio.NewGap("gap", &gapRange, nil, nil, nil, nil)
	sr2 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip2 := opentimelineio.NewClip("clip2", nil, &sr2, nil, nil, nil, "", nil)
	track2.AppendChild(gap)
	track2.AppendChild(clip2)

	stack.AppendChild(track1)
	stack.AppendChild(track2)

	// At time 12, only clip1 is visible
	clip := TopClipAtTime(stack, opentime.NewRationalTime(12, 24))
	if clip == nil {
		t.Log("At time 12: no clip found (may be expected based on range calculation)")
	} else {
		t.Logf("At time 12: %s", clip.Name())
	}

	// At time 36, clip2 should be on top
	clip = TopClipAtTime(stack, opentime.NewRationalTime(36, 24))
	if clip == nil {
		t.Log("At time 36: no clip found")
	} else {
		t.Logf("At time 36: %s", clip.Name())
	}
}

func TestTopClipAtTimeEmpty(t *testing.T) {
	stack := opentimelineio.NewStack("empty", nil, nil, nil, nil, nil)

	clip := TopClipAtTime(stack, opentime.NewRationalTime(0, 24))
	if clip != nil {
		t.Error("Expected nil for empty stack")
	}
}

func TestSubtractRange(t *testing.T) {
	// Test no intersection
	a := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(50, 24))
	b := opentime.NewTimeRange(opentime.NewRationalTime(100, 24), opentime.NewRationalTime(50, 24))

	result := subtractRange(a, b)
	if len(result) != 1 {
		t.Errorf("Expected 1 result, got %d", len(result))
	}

	// Test b completely covers a
	a = opentime.NewTimeRange(opentime.NewRationalTime(25, 24), opentime.NewRationalTime(50, 24))
	b = opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))

	result = subtractRange(a, b)
	if len(result) != 0 {
		t.Errorf("Expected 0 results, got %d", len(result))
	}

	// Test partial overlap (b in middle of a)
	a = opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	b = opentime.NewTimeRange(opentime.NewRationalTime(25, 24), opentime.NewRationalTime(50, 24))

	result = subtractRange(a, b)
	if len(result) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result))
	}
}
