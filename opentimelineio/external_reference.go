// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"encoding/json"

	"github.com/Avalanche-io/gotio/opentime"
)

// ExternalReferenceSchema is the schema for ExternalReference.
var ExternalReferenceSchema = Schema{Name: "ExternalReference", Version: 1}

// ExternalReference represents a URL-based media reference.
type ExternalReference struct {
	MediaReferenceBase
	targetURL string
}

// NewExternalReference creates a new ExternalReference.
func NewExternalReference(
	name string,
	targetURL string,
	availableRange *opentime.TimeRange,
	metadata AnyDictionary,
) *ExternalReference {
	return &ExternalReference{
		MediaReferenceBase: NewMediaReferenceBase(name, availableRange, metadata, nil),
		targetURL:          targetURL,
	}
}

// TargetURL returns the target URL.
func (e *ExternalReference) TargetURL() string {
	return e.targetURL
}

// SetTargetURL sets the target URL.
func (e *ExternalReference) SetTargetURL(targetURL string) {
	e.targetURL = targetURL
}

// SchemaName returns the schema name.
func (e *ExternalReference) SchemaName() string {
	return ExternalReferenceSchema.Name
}

// SchemaVersion returns the schema version.
func (e *ExternalReference) SchemaVersion() int {
	return ExternalReferenceSchema.Version
}

// Clone creates a deep copy.
func (e *ExternalReference) Clone() SerializableObject {
	return &ExternalReference{
		MediaReferenceBase: MediaReferenceBase{
			SerializableObjectWithMetadataBase: SerializableObjectWithMetadataBase{
				name:     e.name,
				metadata: CloneAnyDictionary(e.metadata),
			},
			availableRange:       cloneAvailableRange(e.availableRange),
			availableImageBounds: cloneBox2d(e.availableImageBounds),
		},
		targetURL: e.targetURL,
	}
}

// IsEquivalentTo returns true if equivalent.
func (e *ExternalReference) IsEquivalentTo(other SerializableObject) bool {
	otherE, ok := other.(*ExternalReference)
	if !ok {
		return false
	}
	return e.name == otherE.name && e.targetURL == otherE.targetURL
}

// externalReferenceJSON is the JSON representation.
type externalReferenceJSON struct {
	Schema               string              `json:"OTIO_SCHEMA"`
	Name                 string              `json:"name"`
	Metadata             AnyDictionary       `json:"metadata"`
	AvailableRange       *opentime.TimeRange `json:"available_range"`
	AvailableImageBounds *Box2d              `json:"available_image_bounds"`
	TargetURL            string              `json:"target_url"`
}

// MarshalJSON implements json.Marshaler.
func (e *ExternalReference) MarshalJSON() ([]byte, error) {
	return json.Marshal(&externalReferenceJSON{
		Schema:               ExternalReferenceSchema.String(),
		Name:                 e.name,
		Metadata:             e.metadata,
		AvailableRange:       e.availableRange,
		AvailableImageBounds: e.availableImageBounds,
		TargetURL:            e.targetURL,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (e *ExternalReference) UnmarshalJSON(data []byte) error {
	var j externalReferenceJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	e.name = j.Name
	e.metadata = j.Metadata
	if e.metadata == nil {
		e.metadata = make(AnyDictionary)
	}
	e.availableRange = j.AvailableRange
	e.availableImageBounds = j.AvailableImageBounds
	e.targetURL = j.TargetURL
	return nil
}

func init() {
	RegisterSchema(ExternalReferenceSchema, func() SerializableObject {
		return NewExternalReference("", "", nil, nil)
	})
}
