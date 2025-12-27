// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"encoding/json"

	"github.com/mrjoshuak/gotio/opentime"
)

// MarkerSchema is the schema for Marker.
var MarkerSchema = Schema{Name: "Marker", Version: 2}

// Marker represents a marked time or range on an item.
type Marker struct {
	SerializableObjectWithMetadataBase
	markedRange opentime.TimeRange
	color       MarkerColor
	comment     string
}

// NewMarker creates a new Marker.
func NewMarker(
	name string,
	markedRange opentime.TimeRange,
	color MarkerColor,
	comment string,
	metadata AnyDictionary,
) *Marker {
	if color == "" {
		color = MarkerColorGreen
	}
	return &Marker{
		SerializableObjectWithMetadataBase: NewSerializableObjectWithMetadataBase(name, metadata),
		markedRange:                        markedRange,
		color:                              color,
		comment:                            comment,
	}
}

// MarkedRange returns the marked range.
func (m *Marker) MarkedRange() opentime.TimeRange {
	return m.markedRange
}

// SetMarkedRange sets the marked range.
func (m *Marker) SetMarkedRange(markedRange opentime.TimeRange) {
	m.markedRange = markedRange
}

// Color returns the marker color.
func (m *Marker) Color() MarkerColor {
	return m.color
}

// SetColor sets the marker color.
func (m *Marker) SetColor(color MarkerColor) {
	m.color = color
}

// Comment returns the comment.
func (m *Marker) Comment() string {
	return m.comment
}

// SetComment sets the comment.
func (m *Marker) SetComment(comment string) {
	m.comment = comment
}

// SchemaName returns the schema name.
func (m *Marker) SchemaName() string {
	return MarkerSchema.Name
}

// SchemaVersion returns the schema version.
func (m *Marker) SchemaVersion() int {
	return MarkerSchema.Version
}

// Clone creates a deep copy.
func (m *Marker) Clone() SerializableObject {
	return &Marker{
		SerializableObjectWithMetadataBase: SerializableObjectWithMetadataBase{
			name:     m.name,
			metadata: CloneAnyDictionary(m.metadata),
		},
		markedRange: m.markedRange,
		color:       m.color,
		comment:     m.comment,
	}
}

// IsEquivalentTo returns true if equivalent.
func (m *Marker) IsEquivalentTo(other SerializableObject) bool {
	otherM, ok := other.(*Marker)
	if !ok {
		return false
	}
	return m.name == otherM.name &&
		m.markedRange.Equal(otherM.markedRange) &&
		m.color == otherM.color &&
		m.comment == otherM.comment
}

// markerJSON is the JSON representation.
type markerJSON struct {
	Schema      string             `json:"OTIO_SCHEMA"`
	Name        string             `json:"name"`
	Metadata    AnyDictionary      `json:"metadata"`
	MarkedRange opentime.TimeRange `json:"marked_range"`
	Color       MarkerColor        `json:"color"`
	Comment     string             `json:"comment"`
}

// MarshalJSON implements json.Marshaler.
func (m *Marker) MarshalJSON() ([]byte, error) {
	return json.Marshal(&markerJSON{
		Schema:      MarkerSchema.String(),
		Name:        m.name,
		Metadata:    m.metadata,
		MarkedRange: m.markedRange,
		Color:       m.color,
		Comment:     m.comment,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (m *Marker) UnmarshalJSON(data []byte) error {
	var j markerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	m.name = j.Name
	m.metadata = j.Metadata
	if m.metadata == nil {
		m.metadata = make(AnyDictionary)
	}
	m.markedRange = j.MarkedRange
	m.color = j.Color
	m.comment = j.Comment
	return nil
}

func init() {
	RegisterSchema(MarkerSchema, func() SerializableObject {
		return NewMarker("", opentime.TimeRange{}, "", "", nil)
	})
}
