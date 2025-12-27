// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"encoding/json"

	"github.com/Avalanche-io/gotio/opentime"
)

// GapSchema is the schema for Gap.
var GapSchema = Schema{Name: "Gap", Version: 1}

// Gap represents an empty space in a track.
type Gap struct {
	ItemBase
}

// NewGap creates a new Gap.
func NewGap(
	name string,
	sourceRange *opentime.TimeRange,
	metadata AnyDictionary,
	effects []Effect,
	markers []*Marker,
	color *Color,
) *Gap {
	gap := &Gap{
		ItemBase: NewItemBase(name, sourceRange, metadata, effects, markers, true, color),
	}
	gap.SetSelf(gap)
	return gap
}

// NewGapWithDuration creates a new Gap with the given duration.
func NewGapWithDuration(duration opentime.RationalTime) *Gap {
	sourceRange := opentime.NewTimeRange(opentime.RationalTime{}, duration)
	return NewGap("", &sourceRange, nil, nil, nil, nil)
}

// AvailableRange returns the available range (same as source range for Gap).
func (g *Gap) AvailableRange() (opentime.TimeRange, error) {
	if g.sourceRange != nil {
		return *g.sourceRange, nil
	}
	return opentime.TimeRange{}, ErrCannotComputeAvailableRange
}

// Duration returns the duration from source range or available range.
func (g *Gap) Duration() (opentime.RationalTime, error) {
	if g.sourceRange != nil {
		return g.sourceRange.Duration(), nil
	}
	ar, err := g.AvailableRange()
	if err != nil {
		return opentime.RationalTime{}, err
	}
	return ar.Duration(), nil
}

// SchemaName returns the schema name.
func (g *Gap) SchemaName() string {
	return GapSchema.Name
}

// SchemaVersion returns the schema version.
func (g *Gap) SchemaVersion() int {
	return GapSchema.Version
}

// Clone creates a deep copy.
func (g *Gap) Clone() SerializableObject {
	clone := &Gap{
		ItemBase: ItemBase{
			ComposableBase: ComposableBase{
				SerializableObjectWithMetadataBase: SerializableObjectWithMetadataBase{
					name:     g.name,
					metadata: CloneAnyDictionary(g.metadata),
				},
			},
			sourceRange: cloneSourceRange(g.sourceRange),
			effects:     cloneEffects(g.effects),
			markers:     cloneMarkers(g.markers),
			enabled:     g.enabled,
			color:       cloneColor(g.color),
		},
	}
	clone.SetSelf(clone)
	return clone
}

// IsEquivalentTo returns true if equivalent.
func (g *Gap) IsEquivalentTo(other SerializableObject) bool {
	otherG, ok := other.(*Gap)
	if !ok {
		return false
	}
	return g.name == otherG.name && g.enabled == otherG.enabled
}

// gapJSON is the JSON representation.
type gapJSON struct {
	Schema      string              `json:"OTIO_SCHEMA"`
	Name        string              `json:"name"`
	Metadata    AnyDictionary       `json:"metadata"`
	SourceRange *opentime.TimeRange `json:"source_range"`
	Effects     []RawMessage   `json:"effects"`
	Markers     []*Marker           `json:"markers"`
	Enabled     bool                `json:"enabled"`
	Color       *Color              `json:"color"`
}

// MarshalJSON implements json.Marshaler.
func (g *Gap) MarshalJSON() ([]byte, error) {
	effects := make([]RawMessage, len(g.effects))
	for i, e := range g.effects {
		data, err := json.Marshal(e)
		if err != nil {
			return nil, err
		}
		effects[i] = data
	}

	return json.Marshal(&gapJSON{
		Schema:      GapSchema.String(),
		Name:        g.name,
		Metadata:    g.metadata,
		SourceRange: g.sourceRange,
		Effects:     effects,
		Markers:     g.markers,
		Enabled:     g.enabled,
		Color:       g.color,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *Gap) UnmarshalJSON(data []byte) error {
	var j gapJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}

	g.name = j.Name
	g.metadata = j.Metadata
	if g.metadata == nil {
		g.metadata = make(AnyDictionary)
	}
	g.sourceRange = j.SourceRange
	g.enabled = j.Enabled
	g.color = j.Color

	// Unmarshal effects
	g.effects = make([]Effect, len(j.Effects))
	for i, data := range j.Effects {
		obj, err := FromJSONString(string(data))
		if err != nil {
			return err
		}
		effect, ok := obj.(Effect)
		if !ok {
			return &TypeMismatchError{Expected: "Effect", Got: obj.SchemaName()}
		}
		g.effects[i] = effect
	}

	// Copy markers
	g.markers = j.Markers
	if g.markers == nil {
		g.markers = make([]*Marker, 0)
	}

	g.SetSelf(g)
	return nil
}

func init() {
	RegisterSchema(GapSchema, func() SerializableObject {
		return NewGap("", nil, nil, nil, nil, nil)
	})
}
