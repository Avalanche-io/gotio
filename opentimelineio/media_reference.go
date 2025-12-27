// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"github.com/Avalanche-io/gotio/opentime"
)

// Box2d represents a 2D bounding box.
type Box2d struct {
	Min Vec2d `json:"min"`
	Max Vec2d `json:"max"`
}

// Vec2d represents a 2D vector.
type Vec2d struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// MediaReference is the interface for media references.
type MediaReference interface {
	SerializableObjectWithMetadata

	// AvailableRange returns the available range.
	AvailableRange() *opentime.TimeRange

	// SetAvailableRange sets the available range.
	SetAvailableRange(availableRange *opentime.TimeRange)

	// AvailableImageBounds returns the available image bounds.
	AvailableImageBounds() *Box2d

	// SetAvailableImageBounds sets the available image bounds.
	SetAvailableImageBounds(bounds *Box2d)

	// IsMissingReference returns true if this is a missing reference.
	IsMissingReference() bool
}

// MediaReferenceBase is the base implementation of MediaReference.
type MediaReferenceBase struct {
	SerializableObjectWithMetadataBase
	availableRange       *opentime.TimeRange
	availableImageBounds *Box2d
}

// NewMediaReferenceBase creates a new MediaReferenceBase.
func NewMediaReferenceBase(
	name string,
	availableRange *opentime.TimeRange,
	metadata AnyDictionary,
	availableImageBounds *Box2d,
) MediaReferenceBase {
	return MediaReferenceBase{
		SerializableObjectWithMetadataBase: NewSerializableObjectWithMetadataBase(name, metadata),
		availableRange:                     availableRange,
		availableImageBounds:               availableImageBounds,
	}
}

// AvailableRange returns the available range.
func (m *MediaReferenceBase) AvailableRange() *opentime.TimeRange {
	return m.availableRange
}

// SetAvailableRange sets the available range.
func (m *MediaReferenceBase) SetAvailableRange(availableRange *opentime.TimeRange) {
	m.availableRange = availableRange
}

// AvailableImageBounds returns the available image bounds.
func (m *MediaReferenceBase) AvailableImageBounds() *Box2d {
	return m.availableImageBounds
}

// SetAvailableImageBounds sets the available image bounds.
func (m *MediaReferenceBase) SetAvailableImageBounds(bounds *Box2d) {
	m.availableImageBounds = bounds
}

// IsMissingReference returns false by default.
func (m *MediaReferenceBase) IsMissingReference() bool {
	return false
}

// mediaReferenceJSON is the JSON representation.
type mediaReferenceJSON struct {
	Schema               string              `json:"OTIO_SCHEMA"`
	Name                 string              `json:"name"`
	Metadata             AnyDictionary       `json:"metadata"`
	AvailableRange       *opentime.TimeRange `json:"available_range"`
	AvailableImageBounds *Box2d              `json:"available_image_bounds"`
}

// marshalMediaReferenceJSON creates a mediaReferenceJSON.
func (m *MediaReferenceBase) marshalMediaReferenceJSON(schema Schema) mediaReferenceJSON {
	return mediaReferenceJSON{
		Schema:               schema.String(),
		Name:                 m.name,
		Metadata:             m.metadata,
		AvailableRange:       m.availableRange,
		AvailableImageBounds: m.availableImageBounds,
	}
}

// unmarshalMediaReferenceJSON unmarshals JSON.
func (m *MediaReferenceBase) unmarshalMediaReferenceJSON(j *mediaReferenceJSON) {
	m.name = j.Name
	m.metadata = j.Metadata
	if m.metadata == nil {
		m.metadata = make(AnyDictionary)
	}
	m.availableRange = j.AvailableRange
	m.availableImageBounds = j.AvailableImageBounds
}

// cloneAvailableRange creates a copy of a TimeRange pointer.
func cloneAvailableRange(tr *opentime.TimeRange) *opentime.TimeRange {
	if tr == nil {
		return nil
	}
	clone := *tr
	return &clone
}

// cloneBox2d creates a copy of a Box2d pointer.
func cloneBox2d(b *Box2d) *Box2d {
	if b == nil {
		return nil
	}
	return &Box2d{
		Min: b.Min,
		Max: b.Max,
	}
}
