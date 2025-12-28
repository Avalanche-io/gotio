// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package gotio

import (
	"encoding/json"

	"github.com/Avalanche-io/gotio/opentime"
)

// GeneratorReferenceSchema is the schema for GeneratorReference.
var GeneratorReferenceSchema = Schema{Name: "GeneratorReference", Version: 1}

// GeneratorReference represents a media reference generated algorithmically.
type GeneratorReference struct {
	MediaReferenceBase
	generatorKind string
	parameters    AnyDictionary
}

// NewGeneratorReference creates a new GeneratorReference.
func NewGeneratorReference(
	name string,
	generatorKind string,
	parameters AnyDictionary,
	availableRange *opentime.TimeRange,
	metadata AnyDictionary,
) *GeneratorReference {
	if parameters == nil {
		parameters = make(AnyDictionary)
	}
	return &GeneratorReference{
		MediaReferenceBase: NewMediaReferenceBase(name, availableRange, metadata, nil),
		generatorKind:      generatorKind,
		parameters:         parameters,
	}
}

// GeneratorKind returns the generator kind.
func (g *GeneratorReference) GeneratorKind() string {
	return g.generatorKind
}

// SetGeneratorKind sets the generator kind.
func (g *GeneratorReference) SetGeneratorKind(kind string) {
	g.generatorKind = kind
}

// Parameters returns the generator parameters.
func (g *GeneratorReference) Parameters() AnyDictionary {
	return g.parameters
}

// SetParameters sets the generator parameters.
func (g *GeneratorReference) SetParameters(params AnyDictionary) {
	if params == nil {
		params = make(AnyDictionary)
	}
	g.parameters = params
}

// SchemaName returns the schema name.
func (g *GeneratorReference) SchemaName() string {
	return GeneratorReferenceSchema.Name
}

// SchemaVersion returns the schema version.
func (g *GeneratorReference) SchemaVersion() int {
	return GeneratorReferenceSchema.Version
}

// Clone creates a deep copy.
func (g *GeneratorReference) Clone() SerializableObject {
	return &GeneratorReference{
		MediaReferenceBase: MediaReferenceBase{
			SerializableObjectWithMetadataBase: SerializableObjectWithMetadataBase{
				name:     g.name,
				metadata: CloneAnyDictionary(g.metadata),
			},
			availableRange:       cloneAvailableRange(g.availableRange),
			availableImageBounds: cloneBox2d(g.availableImageBounds),
		},
		generatorKind: g.generatorKind,
		parameters:    CloneAnyDictionary(g.parameters),
	}
}

// IsEquivalentTo returns true if equivalent.
func (g *GeneratorReference) IsEquivalentTo(other SerializableObject) bool {
	otherG, ok := other.(*GeneratorReference)
	if !ok {
		return false
	}
	return g.name == otherG.name && g.generatorKind == otherG.generatorKind
}

// generatorReferenceJSON is the JSON representation.
type generatorReferenceJSON struct {
	Schema               string              `json:"OTIO_SCHEMA"`
	Name                 string              `json:"name"`
	Metadata             AnyDictionary       `json:"metadata"`
	AvailableRange       *opentime.TimeRange `json:"available_range"`
	AvailableImageBounds *Box2d              `json:"available_image_bounds"`
	GeneratorKind        string              `json:"generator_kind"`
	Parameters           AnyDictionary       `json:"parameters"`
}

// MarshalJSON implements json.Marshaler.
func (g *GeneratorReference) MarshalJSON() ([]byte, error) {
	return json.Marshal(&generatorReferenceJSON{
		Schema:               GeneratorReferenceSchema.String(),
		Name:                 g.name,
		Metadata:             g.metadata,
		AvailableRange:       g.availableRange,
		AvailableImageBounds: g.availableImageBounds,
		GeneratorKind:        g.generatorKind,
		Parameters:           g.parameters,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *GeneratorReference) UnmarshalJSON(data []byte) error {
	var j generatorReferenceJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	g.name = j.Name
	g.metadata = j.Metadata
	if g.metadata == nil {
		g.metadata = make(AnyDictionary)
	}
	g.availableRange = j.AvailableRange
	g.availableImageBounds = j.AvailableImageBounds
	g.generatorKind = j.GeneratorKind
	g.parameters = j.Parameters
	if g.parameters == nil {
		g.parameters = make(AnyDictionary)
	}
	return nil
}

func init() {
	RegisterSchema(GeneratorReferenceSchema, func() SerializableObject {
		return NewGeneratorReference("", "", nil, nil, nil)
	})
}
