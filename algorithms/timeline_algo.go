// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package algorithms

import (
	"github.com/Avalanche-io/gotio/opentime"
	"github.com/Avalanche-io/gotio"
)

// TimelineTrimmedToRange returns a new timeline trimmed to the given time range.
// All tracks in the timeline are trimmed to the specified range.
func TimelineTrimmedToRange(timeline *gotio.Timeline, trimRange opentime.TimeRange) (*gotio.Timeline, error) {
	// Clone the timeline
	cloned := timeline.Clone().(*gotio.Timeline)

	// Get the tracks stack
	tracksStack := cloned.Tracks()
	if tracksStack == nil {
		return cloned, nil
	}

	// Create a new stack for the trimmed tracks
	newTracks := gotio.NewStack(
		tracksStack.Name(),
		tracksStack.SourceRange(),
		gotio.CloneAnyDictionary(tracksStack.Metadata()),
		nil,
		nil,
		nil,
	)

	// Trim each track
	for _, child := range tracksStack.Children() {
		track, ok := child.(*gotio.Track)
		if !ok {
			// Keep non-track children as-is
			newTracks.AppendChild(child.Clone().(gotio.Composable))
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
	result := gotio.NewTimeline(
		cloned.Name(),
		cloned.GlobalStartTime(),
		gotio.CloneAnyDictionary(cloned.Metadata()),
	)
	result.SetTracks(newTracks)

	return result, nil
}

// TimelineAudioTracks returns all audio tracks from a timeline.
func TimelineAudioTracks(timeline *gotio.Timeline) []*gotio.Track {
	tracks := timeline.Tracks()
	if tracks == nil {
		return nil
	}

	var audioTracks []*gotio.Track
	for _, child := range tracks.Children() {
		track, ok := child.(*gotio.Track)
		if !ok {
			continue
		}
		if track.Kind() == gotio.TrackKindAudio {
			audioTracks = append(audioTracks, track)
		}
	}

	return audioTracks
}

// TimelineVideoTracks returns all video tracks from a timeline.
func TimelineVideoTracks(timeline *gotio.Timeline) []*gotio.Track {
	tracks := timeline.Tracks()
	if tracks == nil {
		return nil
	}

	var videoTracks []*gotio.Track
	for _, child := range tracks.Children() {
		track, ok := child.(*gotio.Track)
		if !ok {
			continue
		}
		if track.Kind() == gotio.TrackKindVideo {
			videoTracks = append(videoTracks, track)
		}
	}

	return videoTracks
}

// FlattenTimelineVideoTracks flattens all video tracks in a timeline to a single track.
// Audio tracks are preserved unchanged.
func FlattenTimelineVideoTracks(timeline *gotio.Timeline) (*gotio.Timeline, error) {
	// Clone the timeline
	cloned := timeline.Clone().(*gotio.Timeline)

	tracks := cloned.Tracks()
	if tracks == nil {
		return cloned, nil
	}

	// Separate video and audio tracks
	var videoTracks []*gotio.Track
	var audioTracks []*gotio.Track
	var otherChildren []gotio.Composable

	for _, child := range tracks.Children() {
		track, ok := child.(*gotio.Track)
		if !ok {
			otherChildren = append(otherChildren, child)
			continue
		}

		switch track.Kind() {
		case gotio.TrackKindVideo:
			videoTracks = append(videoTracks, track)
		case gotio.TrackKindAudio:
			audioTracks = append(audioTracks, track)
		default:
			otherChildren = append(otherChildren, track)
		}
	}

	// Flatten video tracks
	var flattenedVideo *gotio.Track
	if len(videoTracks) > 0 {
		var err error
		flattenedVideo, err = FlattenTracks(videoTracks)
		if err != nil {
			return nil, err
		}
		flattenedVideo.SetKind(gotio.TrackKindVideo)
	}

	// Create new tracks stack
	newTracks := gotio.NewStack(
		tracks.Name(),
		tracks.SourceRange(),
		gotio.CloneAnyDictionary(tracks.Metadata()),
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
		newTracks.AppendChild(track.Clone().(gotio.Composable))
	}

	// Add other children
	for _, child := range otherChildren {
		newTracks.AppendChild(child.Clone().(gotio.Composable))
	}

	// Create result timeline
	result := gotio.NewTimeline(
		cloned.Name(),
		cloned.GlobalStartTime(),
		gotio.CloneAnyDictionary(cloned.Metadata()),
	)
	result.SetTracks(newTracks)

	return result, nil
}
