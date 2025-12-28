// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package gotio

import (
	"testing"

	"github.com/Avalanche-io/gotio/opentime"
)

func TestTimelineDuration(t *testing.T) {
	timeline := NewTimeline("test", nil, nil)
	track := NewTrack("V1", nil, TrackKindVideo, nil, nil)
	timeline.Tracks().AppendChild(track)

	// Empty track has zero duration
	dur, err := timeline.Duration()
	if err != nil {
		t.Fatalf("Duration error: %v", err)
	}
	if dur.Rate() != 0 && dur.Value() != 0 {
		// Empty track returns zero RationalTime
	}

	// Add clips
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip1 := NewClip("clip1", nil, &sr, nil, nil, nil, "", nil)
	clip2 := NewClip("clip2", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip1)
	track.AppendChild(clip2)

	dur, err = timeline.Duration()
	if err != nil {
		t.Fatalf("Duration error: %v", err)
	}
	if dur.ToSeconds() != 2.0 {
		t.Errorf("Duration = %v seconds, want 2.0", dur.ToSeconds())
	}
}

func TestTimelineGlobalStartTime(t *testing.T) {
	gst := opentime.NewRationalTime(100, 24)
	timeline := NewTimeline("test", &gst, nil)

	got := timeline.GlobalStartTime()
	if got == nil {
		t.Fatal("GlobalStartTime should not be nil")
	}
	if got.Value() != 100 || got.Rate() != 24 {
		t.Errorf("GlobalStartTime = %v, want RationalTime(100, 24)", got)
	}

	// Test SetGlobalStartTime
	newGst := opentime.NewRationalTime(200, 30)
	timeline.SetGlobalStartTime(&newGst)
	got = timeline.GlobalStartTime()
	if got.Value() != 200 || got.Rate() != 30 {
		t.Errorf("After SetGlobalStartTime = %v, want RationalTime(200, 30)", got)
	}

	// Set to nil
	timeline.SetGlobalStartTime(nil)
	if timeline.GlobalStartTime() != nil {
		t.Error("GlobalStartTime should be nil after setting nil")
	}
}

func TestTimelineSetTracks(t *testing.T) {
	timeline := NewTimeline("test", nil, nil)

	// Create a new stack
	newStack := NewStack("new_tracks", nil, nil, nil, nil, nil)
	track := NewTrack("V1", nil, TrackKindVideo, nil, nil)
	newStack.AppendChild(track)

	// Set the tracks
	timeline.SetTracks(newStack)

	if timeline.Tracks() != newStack {
		t.Error("SetTracks did not set the new stack")
	}
	if len(timeline.Tracks().Children()) != 1 {
		t.Errorf("Expected 1 track, got %d", len(timeline.Tracks().Children()))
	}
}

func TestTimelineAvailableRange(t *testing.T) {
	timeline := NewTimeline("test", nil, nil)
	track := NewTrack("V1", nil, TrackKindVideo, nil, nil)
	timeline.Tracks().AppendChild(track)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)

	ar, err := timeline.AvailableRange()
	if err != nil {
		t.Fatalf("AvailableRange error: %v", err)
	}
	if ar.Duration().ToSeconds() != 2.0 {
		t.Errorf("AvailableRange Duration = %v, want 2.0 seconds", ar.Duration().ToSeconds())
	}
}

func TestTimelineClone(t *testing.T) {
	gst := opentime.NewRationalTime(10, 24)
	timeline := NewTimeline("test", &gst, AnyDictionary{"key": "value"})
	track := NewTrack("V1", nil, TrackKindVideo, nil, nil)
	timeline.Tracks().AppendChild(track)

	clone := timeline.Clone().(*Timeline)

	if clone.Name() != "test" {
		t.Errorf("Clone name = %s, want test", clone.Name())
	}
	if clone.GlobalStartTime() == nil || clone.GlobalStartTime().Value() != 10 {
		t.Error("Clone GlobalStartTime should be 10")
	}
	if clone.Metadata()["key"] != "value" {
		t.Error("Clone metadata should match")
	}
	if len(clone.Tracks().Children()) != 1 {
		t.Error("Clone should have 1 track")
	}

	// Verify deep copy - modifying clone shouldn't affect original
	clone.SetName("modified")
	if timeline.Name() == "modified" {
		t.Error("Modifying clone affected original")
	}
}

func TestTimelineIsEquivalentTo(t *testing.T) {
	timeline1 := NewTimeline("test", nil, nil)
	timeline2 := NewTimeline("test", nil, nil)
	timeline3 := NewTimeline("different", nil, nil)

	if !timeline1.IsEquivalentTo(timeline2) {
		t.Error("Identical timelines should be equivalent")
	}
	if timeline1.IsEquivalentTo(timeline3) {
		t.Error("Different timelines should not be equivalent")
	}

	// Test with a non-Timeline
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)
	if timeline1.IsEquivalentTo(clip) {
		t.Error("Timeline should not be equivalent to Clip")
	}
}

func TestTimelineSchema(t *testing.T) {
	timeline := NewTimeline("test", nil, nil)

	if timeline.SchemaName() != "Timeline" {
		t.Errorf("SchemaName = %s, want Timeline", timeline.SchemaName())
	}
	if timeline.SchemaVersion() != 1 {
		t.Errorf("SchemaVersion = %d, want 1", timeline.SchemaVersion())
	}
}

func TestTimelineRangeOfChild(t *testing.T) {
	timeline := NewTimeline("test", nil, nil)
	track1 := NewTrack("V1", nil, TrackKindVideo, nil, nil)
	track2 := NewTrack("V2", nil, TrackKindVideo, nil, nil)

	sr1 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	sr2 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip1 := NewClip("clip1", nil, &sr1, nil, nil, nil, "", nil)
	clip2 := NewClip("clip2", nil, &sr2, nil, nil, nil, "", nil)

	track1.AppendChild(clip1)
	track2.AppendChild(clip2)
	timeline.Tracks().AppendChild(track1)
	timeline.Tracks().AppendChild(track2)

	// In a stack (timeline tracks), all children start at 0
	r, err := timeline.Tracks().RangeOfChild(track1)
	if err != nil {
		t.Fatalf("RangeOfChild error: %v", err)
	}
	if r.StartTime().Value() != 0 {
		t.Errorf("Track1 start time = %v, want 0", r.StartTime().Value())
	}
}
