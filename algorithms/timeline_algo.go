// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package algorithms

import (
	"github.com/Avalanche-io/gotio/opentime"
	"github.com/Avalanche-io/gotio/opentimelineio"
)

// TimelineTrimmedToRange returns a new timeline trimmed to the given time range.
// All tracks in the timeline are trimmed to the specified range.
func TimelineTrimmedToRange(timeline *opentimelineio.Timeline, trimRange opentime.TimeRange) (*opentimelineio.Timeline, error) {
	// Clone the timeline
	cloned := timeline.Clone().(*opentimelineio.Timeline)

	// Get the tracks stack
	tracksStack := cloned.Tracks()
	if tracksStack == nil {
		return cloned, nil
	}

	// Create a new stack for the trimmed tracks
	newTracks := opentimelineio.NewStack(
		tracksStack.Name(),
		tracksStack.SourceRange(),
		opentimelineio.CloneAnyDictionary(tracksStack.Metadata()),
		nil,
		nil,
		nil,
	)

	// Trim each track
	for _, child := range tracksStack.Children() {
		track, ok := child.(*opentimelineio.Track)
		if !ok {
			// Keep non-track children as-is
			newTracks.AppendChild(child.Clone().(opentimelineio.Composable))
			continue
		}

		// Trim the track
		trimmedTrack, err := TrackTrimmedToRange(track, trimRange)
		if err != nil {
			return nil, err
		}

		newTracks.AppendChild(trimmedTrack)
	}

	// Create the result timeline with the new tracks
	result := opentimelineio.NewTimeline(
		cloned.Name(),
		cloned.GlobalStartTime(),
		opentimelineio.CloneAnyDictionary(cloned.Metadata()),
	)
	result.SetTracks(newTracks)

	return result, nil
}

// TimelineAudioTracks returns all audio tracks from a timeline.
func TimelineAudioTracks(timeline *opentimelineio.Timeline) []*opentimelineio.Track {
	tracks := timeline.Tracks()
	if tracks == nil {
		return nil
	}

	var audioTracks []*opentimelineio.Track
	for _, child := range tracks.Children() {
		track, ok := child.(*opentimelineio.Track)
		if !ok {
			continue
		}
		if track.Kind() == opentimelineio.TrackKindAudio {
			audioTracks = append(audioTracks, track)
		}
	}

	return audioTracks
}

// TimelineVideoTracks returns all video tracks from a timeline.
func TimelineVideoTracks(timeline *opentimelineio.Timeline) []*opentimelineio.Track {
	tracks := timeline.Tracks()
	if tracks == nil {
		return nil
	}

	var videoTracks []*opentimelineio.Track
	for _, child := range tracks.Children() {
		track, ok := child.(*opentimelineio.Track)
		if !ok {
			continue
		}
		if track.Kind() == opentimelineio.TrackKindVideo {
			videoTracks = append(videoTracks, track)
		}
	}

	return videoTracks
}

// FlattenTimelineVideoTracks flattens all video tracks in a timeline to a single track.
// Audio tracks are preserved unchanged.
func FlattenTimelineVideoTracks(timeline *opentimelineio.Timeline) (*opentimelineio.Timeline, error) {
	// Clone the timeline
	cloned := timeline.Clone().(*opentimelineio.Timeline)

	tracks := cloned.Tracks()
	if tracks == nil {
		return cloned, nil
	}

	// Separate video and audio tracks
	var videoTracks []*opentimelineio.Track
	var audioTracks []*opentimelineio.Track
	var otherChildren []opentimelineio.Composable

	for _, child := range tracks.Children() {
		track, ok := child.(*opentimelineio.Track)
		if !ok {
			otherChildren = append(otherChildren, child)
			continue
		}

		switch track.Kind() {
		case opentimelineio.TrackKindVideo:
			videoTracks = append(videoTracks, track)
		case opentimelineio.TrackKindAudio:
			audioTracks = append(audioTracks, track)
		default:
			otherChildren = append(otherChildren, track)
		}
	}

	// Flatten video tracks
	var flattenedVideo *opentimelineio.Track
	if len(videoTracks) > 0 {
		var err error
		flattenedVideo, err = FlattenTracks(videoTracks)
		if err != nil {
			return nil, err
		}
		flattenedVideo.SetKind(opentimelineio.TrackKindVideo)
	}

	// Create new tracks stack
	newTracks := opentimelineio.NewStack(
		tracks.Name(),
		tracks.SourceRange(),
		opentimelineio.CloneAnyDictionary(tracks.Metadata()),
		nil,
		nil,
		nil,
	)

	// Add flattened video track
	if flattenedVideo != nil {
		newTracks.AppendChild(flattenedVideo)
	}

	// Add audio tracks
	for _, track := range audioTracks {
		newTracks.AppendChild(track.Clone().(opentimelineio.Composable))
	}

	// Add other children
	for _, child := range otherChildren {
		newTracks.AppendChild(child.Clone().(opentimelineio.Composable))
	}

	// Create result timeline
	result := opentimelineio.NewTimeline(
		cloned.Name(),
		cloned.GlobalStartTime(),
		opentimelineio.CloneAnyDictionary(cloned.Metadata()),
	)
	result.SetTracks(newTracks)

	return result, nil
}
