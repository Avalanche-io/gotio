// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"encoding/json"

	"github.com/mrjoshuak/gotio/opentime"
)

// Track kinds.
const (
	TrackKindVideo = "Video"
	TrackKindAudio = "Audio"
)

// NeighborGapPolicy defines policies for inserting gaps.
type NeighborGapPolicy int

const (
	// NeighborGapPolicyNever never inserts gaps.
	NeighborGapPolicyNever NeighborGapPolicy = 0
	// NeighborGapPolicyAroundTransitions inserts gaps around transitions.
	NeighborGapPolicyAroundTransitions NeighborGapPolicy = 1
)

// TrackSchema is the schema for Track.
var TrackSchema = Schema{Name: "Track", Version: 1}

// Track is a composition of items arranged sequentially in time.
type Track struct {
	CompositionBase
	kind string
}

// NewTrack creates a new Track.
func NewTrack(
	name string,
	sourceRange *opentime.TimeRange,
	kind string,
	metadata AnyDictionary,
	color *Color,
) *Track {
	if kind == "" {
		kind = TrackKindVideo
	}
	track := &Track{
		CompositionBase: NewCompositionBase(name, sourceRange, metadata, nil, nil, color),
		kind:            kind,
	}
	track.SetSelf(track)
	return track
}

// Kind returns the kind of track.
func (t *Track) Kind() string {
	return t.kind
}

// SetKind sets the kind of track.
func (t *Track) SetKind(kind string) {
	t.kind = kind
}

// CompositionKind returns "Track".
func (t *Track) CompositionKind() string {
	return "Track"
}

// InsertChild inserts a child at the given index.
func (t *Track) InsertChild(index int, child Composable) error {
	if index < 0 || index > len(t.children) {
		return &IndexError{Index: index, Size: len(t.children)}
	}
	child.SetParent(t)
	t.children = append(t.children[:index], append([]Composable{child}, t.children[index:]...)...)
	return nil
}

// AppendChild appends a child.
func (t *Track) AppendChild(child Composable) error {
	return t.InsertChild(len(t.children), child)
}

// SetChild sets the child at the given index.
func (t *Track) SetChild(index int, child Composable) error {
	if index < 0 || index >= len(t.children) {
		return &IndexError{Index: index, Size: len(t.children)}
	}
	t.children[index].SetParent(nil)
	child.SetParent(t)
	t.children[index] = child
	return nil
}

// RemoveChild removes the child at the given index.
func (t *Track) RemoveChild(index int) error {
	if index < 0 || index >= len(t.children) {
		return &IndexError{Index: index, Size: len(t.children)}
	}
	t.children[index].SetParent(nil)
	t.children = append(t.children[:index], t.children[index+1:]...)
	return nil
}

// RangeOfChildAtIndex returns the range of the child at the given index.
// For a Track, children are arranged sequentially.
func (t *Track) RangeOfChildAtIndex(index int) (opentime.TimeRange, error) {
	if index < 0 || index >= len(t.children) {
		return opentime.TimeRange{}, &IndexError{Index: index, Size: len(t.children)}
	}

	// Get duration of this child first so we know the rate
	dur, err := t.children[index].Duration()
	if err != nil {
		return opentime.TimeRange{}, err
	}

	// Calculate start time by summing durations of visible children before this index
	// Start with zero at the same rate as the duration
	startTime := opentime.NewRationalTime(0, dur.Rate())
	for i := 0; i < index; i++ {
		if t.children[i].Visible() {
			childDur, err := t.children[i].Duration()
			if err != nil {
				return opentime.TimeRange{}, err
			}
			startTime = startTime.Add(childDur)
		}
	}

	return opentime.NewTimeRange(startTime, dur), nil
}

// TrimmedRangeOfChildAtIndex returns the trimmed range of the child at the given index.
func (t *Track) TrimmedRangeOfChildAtIndex(index int) (opentime.TimeRange, error) {
	childRange, err := t.RangeOfChildAtIndex(index)
	if err != nil {
		return opentime.TimeRange{}, err
	}
	trimmed := t.trimChildRange(childRange)
	if trimmed == nil {
		return opentime.TimeRange{}, nil
	}
	return *trimmed, nil
}

// AvailableRange returns the available range of the track.
func (t *Track) AvailableRange() (opentime.TimeRange, error) {
	if len(t.children) == 0 {
		return opentime.TimeRange{}, nil
	}

	// Get first visible child's duration to establish rate
	var total opentime.RationalTime
	for _, child := range t.children {
		if child.Visible() {
			dur, err := child.Duration()
			if err != nil {
				return opentime.TimeRange{}, err
			}
			if total.Rate() <= 0 {
				total = dur
			} else {
				total = total.Add(dur)
			}
		}
	}

	startTime := opentime.NewRationalTime(0, total.Rate())
	return opentime.NewTimeRange(startTime, total), nil
}

// Duration returns the duration of the track.
// For a Track, this is the sum of visible children's durations.
func (t *Track) Duration() (opentime.RationalTime, error) {
	if t.sourceRange != nil {
		return t.sourceRange.Duration(), nil
	}
	ar, err := t.AvailableRange()
	if err != nil {
		return opentime.RationalTime{}, err
	}
	return ar.Duration(), nil
}

// HandlesOfChild returns the in and out handles of the given child.
func (t *Track) HandlesOfChild(child Composable) (*opentime.RationalTime, *opentime.RationalTime, error) {
	index, err := t.IndexOfChild(child)
	if err != nil {
		return nil, nil, err
	}

	var inHandle, outHandle *opentime.RationalTime

	// Check previous neighbor
	if index > 0 {
		prev := t.children[index-1]
		if tr, ok := prev.(*Transition); ok {
			in := tr.InOffset()
			inHandle = &in
		}
	}

	// Check next neighbor
	if index < len(t.children)-1 {
		next := t.children[index+1]
		if tr, ok := next.(*Transition); ok {
			out := tr.OutOffset()
			outHandle = &out
		}
	}

	return inHandle, outHandle, nil
}

// NeighborsOf returns the neighbors of the given item.
func (t *Track) NeighborsOf(item Composable, insertGap NeighborGapPolicy) (Composable, Composable, error) {
	index, err := t.IndexOfChild(item)
	if err != nil {
		return nil, nil, err
	}

	var prev, next Composable

	if index > 0 {
		prev = t.children[index-1]
	} else if insertGap == NeighborGapPolicyAroundTransitions {
		if _, ok := item.(*Transition); ok {
			prev = NewGapWithDuration(opentime.RationalTime{})
		}
	}

	if index < len(t.children)-1 {
		next = t.children[index+1]
	} else if insertGap == NeighborGapPolicyAroundTransitions {
		if _, ok := item.(*Transition); ok {
			next = NewGapWithDuration(opentime.RationalTime{})
		}
	}

	return prev, next, nil
}

// RangeOfAllChildren returns a map of child to range.
func (t *Track) RangeOfAllChildren() (map[Composable]opentime.TimeRange, error) {
	result := make(map[Composable]opentime.TimeRange)
	var startTime opentime.RationalTime

	for i, child := range t.children {
		dur, err := t.children[i].Duration()
		if err != nil {
			return nil, err
		}

		result[child] = opentime.NewTimeRange(startTime, dur)

		if child.Visible() {
			startTime = startTime.Add(dur)
		}
	}

	return result, nil
}

// ChildAtTime returns the child at the given time.
// For a Track, this finds the first child whose range contains the search time.
func (t *Track) ChildAtTime(searchTime opentime.RationalTime, shallowSearch bool) (Composable, error) {
	for i, child := range t.children {
		childRange, err := t.RangeOfChildAtIndex(i)
		if err != nil {
			return nil, err
		}
		if childRange.Contains(searchTime) {
			if !shallowSearch {
				if comp, ok := child.(Composition); ok {
					childTime := searchTime.Sub(childRange.StartTime())
					return comp.ChildAtTime(childTime, false)
				}
			}
			return child, nil
		}
	}
	return nil, nil
}

// ChildrenInRange returns all children within the given range.
func (t *Track) ChildrenInRange(searchRange opentime.TimeRange) ([]Composable, error) {
	var result []Composable
	for i, child := range t.children {
		childRange, err := t.RangeOfChildAtIndex(i)
		if err != nil {
			return nil, err
		}
		if searchRange.Intersects(childRange, opentime.DefaultEpsilon) {
			result = append(result, child)
		}
	}
	return result, nil
}

// AvailableImageBounds returns the union of all clips' image bounds.
func (t *Track) AvailableImageBounds() (*Box2d, error) {
	var result *Box2d

	for _, child := range t.children {
		if clip, ok := child.(*Clip); ok {
			bounds, err := clip.AvailableImageBounds()
			if err != nil {
				continue
			}
			if bounds != nil {
				if result == nil {
					result = &Box2d{
						Min: bounds.Min,
						Max: bounds.Max,
					}
				} else {
					if bounds.Min.X < result.Min.X {
						result.Min.X = bounds.Min.X
					}
					if bounds.Min.Y < result.Min.Y {
						result.Min.Y = bounds.Min.Y
					}
					if bounds.Max.X > result.Max.X {
						result.Max.X = bounds.Max.X
					}
					if bounds.Max.Y > result.Max.Y {
						result.Max.Y = bounds.Max.Y
					}
				}
			}
		}
	}

	return result, nil
}

// SchemaName returns the schema name.
func (t *Track) SchemaName() string {
	return TrackSchema.Name
}

// SchemaVersion returns the schema version.
func (t *Track) SchemaVersion() int {
	return TrackSchema.Version
}

// Clone creates a deep copy.
func (t *Track) Clone() SerializableObject {
	clone := &Track{
		CompositionBase: CompositionBase{
			ItemBase: ItemBase{
				ComposableBase: ComposableBase{
					SerializableObjectWithMetadataBase: SerializableObjectWithMetadataBase{
						name:     t.name,
						metadata: CloneAnyDictionary(t.metadata),
					},
				},
				sourceRange: cloneSourceRange(t.sourceRange),
				effects:     cloneEffects(t.effects),
				markers:     cloneMarkers(t.markers),
				enabled:     t.enabled,
				color:       cloneColor(t.color),
			},
			children: cloneChildren(t.children),
		},
		kind: t.kind,
	}
	clone.SetSelf(clone)
	for _, child := range clone.children {
		child.SetParent(clone)
	}
	return clone
}

// IsEquivalentTo returns true if equivalent.
func (t *Track) IsEquivalentTo(other SerializableObject) bool {
	otherT, ok := other.(*Track)
	if !ok {
		return false
	}
	if t.name != otherT.name || t.kind != otherT.kind {
		return false
	}
	if len(t.children) != len(otherT.children) {
		return false
	}
	for i := range t.children {
		if !t.children[i].IsEquivalentTo(otherT.children[i]) {
			return false
		}
	}
	return true
}

// trackJSON is the JSON representation.
type trackJSON struct {
	Schema      string              `json:"OTIO_SCHEMA"`
	Name        string              `json:"name"`
	Metadata    AnyDictionary       `json:"metadata"`
	SourceRange *opentime.TimeRange `json:"source_range"`
	Effects     []RawMessage   `json:"effects"`
	Markers     []*Marker           `json:"markers"`
	Enabled     bool                `json:"enabled"`
	Color       *Color              `json:"color"`
	Children    []RawMessage   `json:"children"`
	Kind        string              `json:"kind"`
}

// MarshalJSON implements json.Marshaler.
func (t *Track) MarshalJSON() ([]byte, error) {
	effects := make([]RawMessage, len(t.effects))
	for i, e := range t.effects {
		data, err := json.Marshal(e)
		if err != nil {
			return nil, err
		}
		effects[i] = data
	}

	children := make([]RawMessage, len(t.children))
	for i, child := range t.children {
		data, err := json.Marshal(child)
		if err != nil {
			return nil, err
		}
		children[i] = data
	}

	return json.Marshal(&trackJSON{
		Schema:      TrackSchema.String(),
		Name:        t.name,
		Metadata:    t.metadata,
		SourceRange: t.sourceRange,
		Effects:     effects,
		Markers:     t.markers,
		Enabled:     t.enabled,
		Color:       t.color,
		Children:    children,
		Kind:        t.kind,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *Track) UnmarshalJSON(data []byte) error {
	var j trackJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}

	t.name = j.Name
	t.metadata = j.Metadata
	if t.metadata == nil {
		t.metadata = make(AnyDictionary)
	}
	t.sourceRange = j.SourceRange
	t.enabled = j.Enabled
	t.color = j.Color
	t.kind = j.Kind
	if t.kind == "" {
		t.kind = TrackKindVideo
	}

	// Unmarshal effects
	t.effects = make([]Effect, len(j.Effects))
	for i, data := range j.Effects {
		obj, err := FromJSONString(string(data))
		if err != nil {
			return err
		}
		effect, ok := obj.(Effect)
		if !ok {
			return &TypeMismatchError{Expected: "Effect", Got: obj.SchemaName()}
		}
		t.effects[i] = effect
	}

	// Copy markers
	t.markers = j.Markers
	if t.markers == nil {
		t.markers = make([]*Marker, 0)
	}

	// Unmarshal children
	t.children = make([]Composable, len(j.Children))
	for i, data := range j.Children {
		obj, err := FromJSONString(string(data))
		if err != nil {
			return err
		}
		composable, ok := obj.(Composable)
		if !ok {
			return &TypeMismatchError{Expected: "Composable", Got: obj.SchemaName()}
		}
		composable.SetParent(t)
		t.children[i] = composable
	}

	t.SetSelf(t)
	return nil
}

func init() {
	RegisterSchema(TrackSchema, func() SerializableObject {
		return NewTrack("", nil, "", nil, nil)
	})
	// Register legacy alias: "Sequence" was the old name for "Track"
	RegisterSchemaAlias("Sequence", "Track")
}
