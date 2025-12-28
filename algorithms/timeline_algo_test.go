// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package algorithms

import (
	"testing"

	"github.com/Avalanche-io/gotio/opentime"
	"github.com/Avalanche-io/gotio"
)

func TestTimelineTrimmedToRange(t *testing.T) {
	timeline := gotio.NewTimeline("test", nil, nil)

	// Add a video track with clips
	videoTrack := gotio.NewTrack("video", nil, gotio.TrackKindVideo, nil, nil)
	sr1 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip1 := gotio.NewClip("clip1", nil, &sr1, nil, nil, nil, "", nil)
	sr2 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip2 := gotio.NewClip("clip2", nil, &sr2, nil, nil, nil, "", nil)
	videoTrack.AppendChild(clip1)
	videoTrack.AppendChild(clip2)
	timeline.Tracks().AppendChild(videoTrack)

	// Trim to first 48 frames
	trimRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))

	result, err := TimelineTrimmedToRange(timeline, trimRange)
	if err != nil {
		t.Fatalf("TimelineTrimmedToRange error: %v", err)
	}

	if result.Name() != "test" {
		t.Errorf("Name = %s, want test", result.Name())
	}

	t.Logf("Trimmed timeline tracks: %d", len(result.Tracks().Children()))
}

func TestTimelineTrimmedToRangeEmpty(t *testing.T) {
	timeline := gotio.NewTimeline("empty", nil, nil)

	trimRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))

	result, err := TimelineTrimmedToRange(timeline, trimRange)
	if err != nil {
		t.Fatalf("TimelineTrimmedToRange error: %v", err)
	}

	if result.Name() != "empty" {
		t.Errorf("Name = %s, want empty", result.Name())
	}
}

func TestTimelineAudioTracks(t *testing.T) {
	timeline := gotio.NewTimeline("test", nil, nil)

	videoTrack := gotio.NewTrack("video", nil, gotio.TrackKindVideo, nil, nil)
	audioTrack1 := gotio.NewTrack("audio1", nil, gotio.TrackKindAudio, nil, nil)
	audioTrack2 := gotio.NewTrack("audio2", nil, gotio.TrackKindAudio, nil, nil)

	timeline.Tracks().AppendChild(videoTrack)
	timeline.Tracks().AppendChild(audioTrack1)
	timeline.Tracks().AppendChild(audioTrack2)

	audioTracks := TimelineAudioTracks(timeline)
	if len(audioTracks) != 2 {
		t.Errorf("Expected 2 audio tracks, got %d", len(audioTracks))
	}
}

func TestTimelineAudioTracksEmpty(t *testing.T) {
	timeline := gotio.NewTimeline("empty", nil, nil)

	audioTracks := TimelineAudioTracks(timeline)
	if len(audioTracks) != 0 {
		t.Errorf("Expected 0 audio tracks, got %d", len(audioTracks))
	}
}

func TestTimelineVideoTracks(t *testing.T) {
	timeline := gotio.NewTimeline("test", nil, nil)

	videoTrack1 := gotio.NewTrack("video1", nil, gotio.TrackKindVideo, nil, nil)
	videoTrack2 := gotio.NewTrack("video2", nil, gotio.TrackKindVideo, nil, nil)
	audioTrack := gotio.NewTrack("audio", nil, gotio.TrackKindAudio, nil, nil)

	timeline.Tracks().AppendChild(videoTrack1)
	timeline.Tracks().AppendChild(videoTrack2)
	timeline.Tracks().AppendChild(audioTrack)

	videoTracks := TimelineVideoTracks(timeline)
	if len(videoTracks) != 2 {
		t.Errorf("Expected 2 video tracks, got %d", len(videoTracks))
	}
}

func TestTimelineVideoTracksEmpty(t *testing.T) {
	timeline := gotio.NewTimeline("empty", nil, nil)

	videoTracks := TimelineVideoTracks(timeline)
	if len(videoTracks) != 0 {
		t.Errorf("Expected 0 video tracks, got %d", len(videoTracks))
	}
}

func TestFlattenTimelineVideoTracks(t *testing.T) {
	timeline := gotio.NewTimeline("test", nil, nil)

	// Add two video tracks
	videoTrack1 := gotio.NewTrack("video1", nil, gotio.TrackKindVideo, nil, nil)
	sr1 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip1 := gotio.NewClip("clip1", nil, &sr1, nil, nil, nil, "", nil)
	videoTrack1.AppendChild(clip1)

	videoTrack2 := gotio.NewTrack("video2", nil, gotio.TrackKindVideo, nil, nil)
	sr2 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip2 := gotio.NewClip("clip2", nil, &sr2, nil, nil, nil, "", nil)
	videoTrack2.AppendChild(clip2)

	// Add an audio track
	audioTrack := gotio.NewTrack("audio", nil, gotio.TrackKindAudio, nil, nil)
	sr3 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	audioClip := gotio.NewClip("audio_clip", nil, &sr3, nil, nil, nil, "", nil)
	audioTrack.AppendChild(audioClip)

	timeline.Tracks().AppendChild(videoTrack1)
	timeline.Tracks().AppendChild(videoTrack2)
	timeline.Tracks().AppendChild(audioTrack)

	result, err := FlattenTimelineVideoTracks(timeline)
	if err != nil {
		t.Fatalf("FlattenTimelineVideoTracks error: %v", err)
	}

	// Should have 1 video track (flattened) + 1 audio track
	videoTracks := TimelineVideoTracks(result)
	audioTracks := TimelineAudioTracks(result)

	if len(videoTracks) != 1 {
		t.Errorf("Expected 1 video track after flattening, got %d", len(videoTracks))
	}
	if len(audioTracks) != 1 {
		t.Errorf("Expected 1 audio track after flattening, got %d", len(audioTracks))
	}
}

func TestFlattenTimelineVideoTracksEmpty(t *testing.T) {
	timeline := gotio.NewTimeline("empty", nil, nil)

	result, err := FlattenTimelineVideoTracks(timeline)
	if err != nil {
		t.Fatalf("FlattenTimelineVideoTracks error: %v", err)
	}

	if result.Name() != "empty" {
		t.Errorf("Name = %s, want empty", result.Name())
	}
}

func TestFlattenTimelineVideoTracksNoVideo(t *testing.T) {
	timeline := gotio.NewTimeline("test", nil, nil)

	// Only audio track
	audioTrack := gotio.NewTrack("audio", nil, gotio.TrackKindAudio, nil, nil)
	timeline.Tracks().AppendChild(audioTrack)

	result, err := FlattenTimelineVideoTracks(timeline)
	if err != nil {
		t.Fatalf("FlattenTimelineVideoTracks error: %v", err)
	}

	audioTracks := TimelineAudioTracks(result)
	if len(audioTracks) != 1 {
		t.Errorf("Expected 1 audio track, got %d", len(audioTracks))
	}
}
