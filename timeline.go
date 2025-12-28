// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package gotio

import (
	"encoding/json"

	"github.com/Avalanche-io/gotio/opentime"
)

// TimelineSchema is the schema for Timeline.
var TimelineSchema = Schema{Name: "Timeline", Version: 1}

// Timeline represents a complete timeline with tracks.
type Timeline struct {
	SerializableObjectWithMetadataBase
	globalStartTime *opentime.RationalTime
	tracks          *Stack
}

// NewTimeline creates a new Timeline.
func NewTimeline(
	name string,
	globalStartTime *opentime.RationalTime,
	metadata AnyDictionary,
) *Timeline {
	return &Timeline{
		SerializableObjectWithMetadataBase: NewSerializableObjectWithMetadataBase(name, metadata),
		globalStartTime:                    globalStartTime,
		tracks:                             NewStack("tracks", nil, nil, nil, nil, nil),
	}
}

// GlobalStartTime returns the global start time.
func (t *Timeline) GlobalStartTime() *opentime.RationalTime {
	return t.globalStartTime
}

// SetGlobalStartTime sets the global start time.
func (t *Timeline) SetGlobalStartTime(globalStartTime *opentime.RationalTime) {
	t.globalStartTime = globalStartTime
}

// Tracks returns the tracks stack.
func (t *Timeline) Tracks() *Stack {
	return t.tracks
}

// SetTracks sets the tracks stack.
func (t *Timeline) SetTracks(tracks *Stack) {
	t.tracks = tracks
}

// Duration returns the duration of the timeline.
func (t *Timeline) Duration() (opentime.RationalTime, error) {
	if t.tracks == nil {
		return opentime.RationalTime{}, nil
	}
	return t.tracks.Duration()
}

// AvailableRange returns the available range of the timeline.
func (t *Timeline) AvailableRange() (opentime.TimeRange, error) {
	if t.tracks == nil {
		return opentime.TimeRange{}, nil
	}
	return t.tracks.AvailableRange()
}

// VideoTracks returns all video tracks.
func (t *Timeline) VideoTracks() []*Track {
	return t.tracksByKind(TrackKindVideo)
}

// AudioTracks returns all audio tracks.
func (t *Timeline) AudioTracks() []*Track {
	return t.tracksByKind(TrackKindAudio)
}

// tracksByKind returns tracks of the given kind.
func (t *Timeline) tracksByKind(kind string) []*Track {
	var result []*Track
	if t.tracks == nil {
		return result
	}
	for _, child := range t.tracks.Children() {
		if track, ok := child.(*Track); ok {
			if track.Kind() == kind {
				result = append(result, track)
			}
		}
	}
	return result
}

// FindClips finds all clips in the timeline.
func (t *Timeline) FindClips(searchRange *opentime.TimeRange, shallowSearch bool) []*Clip {
	if t.tracks == nil {
		return nil
	}
	return t.tracks.FindClips(searchRange, shallowSearch)
}

// FindChildren finds children matching the given filter.
func (t *Timeline) FindChildren(searchRange *opentime.TimeRange, shallowSearch bool, filter func(Composable) bool) []Composable {
	if t.tracks == nil {
		return nil
	}
	return t.tracks.FindChildren(searchRange, shallowSearch, filter)
}

// AvailableImageBounds returns the union of all clips' image bounds.
func (t *Timeline) AvailableImageBounds() (*Box2d, error) {
	if t.tracks == nil {
		return nil, nil
	}
	return t.tracks.AvailableImageBounds()
}

// SchemaName returns the schema name.
func (t *Timeline) SchemaName() string {
	return TimelineSchema.Name
}

// SchemaVersion returns the schema version.
func (t *Timeline) SchemaVersion() int {
	return TimelineSchema.Version
}

// Clone creates a deep copy.
func (t *Timeline) Clone() SerializableObject {
	var gst *opentime.RationalTime
	if t.globalStartTime != nil {
		clone := *t.globalStartTime
		gst = &clone
	}

	var tracks *Stack
	if t.tracks != nil {
		tracks = t.tracks.Clone().(*Stack)
	}

	return &Timeline{
		SerializableObjectWithMetadataBase: SerializableObjectWithMetadataBase{
			name:     t.name,
			metadata: CloneAnyDictionary(t.metadata),
		},
		globalStartTime: gst,
		tracks:          tracks,
	}
}

// IsEquivalentTo returns true if equivalent.
func (t *Timeline) IsEquivalentTo(other SerializableObject) bool {
	otherT, ok := other.(*Timeline)
	if !ok {
		return false
	}
	if t.name != otherT.name {
		return false
	}
	if t.tracks == nil && otherT.tracks == nil {
		return true
	}
	if t.tracks == nil || otherT.tracks == nil {
		return false
	}
	return t.tracks.IsEquivalentTo(otherT.tracks)
}

// timelineJSON is the JSON representation.
type timelineJSON struct {
	Schema          string                 `json:"OTIO_SCHEMA"`
	Name            string                 `json:"name"`
	Metadata        AnyDictionary          `json:"metadata"`
	GlobalStartTime *opentime.RationalTime `json:"global_start_time"`
	Tracks          RawMessage        `json:"tracks"`
}

// MarshalJSON implements json.Marshaler.
func (t *Timeline) MarshalJSON() ([]byte, error) {
	var tracksData RawMessage
	if t.tracks != nil {
		data, err := json.Marshal(t.tracks)
		if err != nil {
			return nil, err
		}
		tracksData = data
	}

	return json.Marshal(&timelineJSON{
		Schema:          TimelineSchema.String(),
		Name:            t.name,
		Metadata:        t.metadata,
		GlobalStartTime: t.globalStartTime,
		Tracks:          tracksData,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *Timeline) UnmarshalJSON(data []byte) error {
	var j timelineJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}

	t.name = j.Name
	t.metadata = j.Metadata
	if t.metadata == nil {
		t.metadata = make(AnyDictionary)
	}
	t.globalStartTime = j.GlobalStartTime

	if j.Tracks != nil {
		obj, err := FromJSONString(string(j.Tracks))
		if err != nil {
			return err
		}
		stack, ok := obj.(*Stack)
		if !ok {
			return &TypeMismatchError{Expected: "Stack", Got: obj.SchemaName()}
		}
		t.tracks = stack
	}

	return nil
}

func init() {
	RegisterSchema(TimelineSchema, func() SerializableObject {
		return NewTimeline("", nil, nil)
	})
}
