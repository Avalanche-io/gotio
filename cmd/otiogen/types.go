// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package main

// FieldKind represents the kind of field for code generation.
type FieldKind int

const (
	FieldKindBasic       FieldKind = iota // string, int, float64, bool
	FieldKindStruct                       // embedded struct (non-polymorphic)
	FieldKindPointer                      // pointer to struct
	FieldKindSlice                        // slice of items
	FieldKindMap                          // map[string]T
	FieldKindInterface                    // polymorphic interface (requires registry dispatch)
	FieldKindRawMessage                   // json.RawMessage for deferred decoding
	FieldKindAnyDict                      // map[string]any for metadata
)

// Field represents a field in an OTIO type.
type Field struct {
	Name       string    // Go field name (e.g., "name")
	JSONName   string    // JSON key name (e.g., "name")
	GoType     string    // Go type (e.g., "string", "opentime.RationalTime")
	Kind       FieldKind // Kind of field for encoding strategy
	ElemType   string    // Element type for slices/maps (e.g., "Effect" for []Effect)
	IsOptional bool      // True if field can be nil/omitted
	IsPointer  bool      // True if field is a pointer
}

// TypeDef represents an OTIO type to generate encoders/decoders for.
type TypeDef struct {
	Name          string  // Go type name (e.g., "RationalTime", "Clip")
	Package       string  // Package name (e.g., "opentime", "opentimelineio")
	SchemaName    string  // OTIO_SCHEMA name (e.g., "RationalTime")
	SchemaVersion int     // OTIO_SCHEMA version
	Fields        []Field // Fields to encode/decode
	IsLeaf        bool    // True if no polymorphic fields (can be fully generated)
	EmbeddedTypes []string // Embedded type names (for inheritance)
}

// openTimeTypes defines the leaf types in the opentime package.
var openTimeTypes = []TypeDef{
	{
		Name:          "RationalTime",
		Package:       "opentime",
		SchemaName:    "RationalTime",
		SchemaVersion: 1,
		IsLeaf:        true,
		Fields: []Field{
			{Name: "value", JSONName: "value", GoType: "float64", Kind: FieldKindBasic},
			{Name: "rate", JSONName: "rate", GoType: "float64", Kind: FieldKindBasic},
		},
	},
	{
		Name:          "TimeRange",
		Package:       "opentime",
		SchemaName:    "TimeRange",
		SchemaVersion: 1,
		IsLeaf:        true,
		Fields: []Field{
			{Name: "startTime", JSONName: "start_time", GoType: "RationalTime", Kind: FieldKindStruct},
			{Name: "duration", JSONName: "duration", GoType: "RationalTime", Kind: FieldKindStruct},
		},
	},
	{
		Name:          "TimeTransform",
		Package:       "opentime",
		SchemaName:    "TimeTransform",
		SchemaVersion: 1,
		IsLeaf:        true,
		Fields: []Field{
			{Name: "offset", JSONName: "offset", GoType: "RationalTime", Kind: FieldKindStruct},
			{Name: "scale", JSONName: "scale", GoType: "float64", Kind: FieldKindBasic},
			{Name: "rate", JSONName: "rate", GoType: "float64", Kind: FieldKindBasic},
		},
	},
}

// otioLeafTypes defines the leaf types in the opentimelineio package.
var otioLeafTypes = []TypeDef{
	{
		Name:          "Marker",
		Package:       "opentimelineio",
		SchemaName:    "Marker",
		SchemaVersion: 2,
		IsLeaf:        true,
		Fields: []Field{
			{Name: "name", JSONName: "name", GoType: "string", Kind: FieldKindBasic},
			{Name: "metadata", JSONName: "metadata", GoType: "AnyDictionary", Kind: FieldKindAnyDict},
			{Name: "markedRange", JSONName: "marked_range", GoType: "opentime.TimeRange", Kind: FieldKindStruct},
			{Name: "color", JSONName: "color", GoType: "MarkerColor", Kind: FieldKindBasic},
			{Name: "comment", JSONName: "comment", GoType: "string", Kind: FieldKindBasic},
		},
	},
	{
		Name:          "ExternalReference",
		Package:       "opentimelineio",
		SchemaName:    "ExternalReference",
		SchemaVersion: 1,
		IsLeaf:        true,
		Fields: []Field{
			{Name: "name", JSONName: "name", GoType: "string", Kind: FieldKindBasic},
			{Name: "metadata", JSONName: "metadata", GoType: "AnyDictionary", Kind: FieldKindAnyDict},
			{Name: "availableRange", JSONName: "available_range", GoType: "*opentime.TimeRange", Kind: FieldKindPointer, IsPointer: true, IsOptional: true},
			{Name: "availableImageBounds", JSONName: "available_image_bounds", GoType: "*Box2d", Kind: FieldKindPointer, IsPointer: true, IsOptional: true},
			{Name: "targetURL", JSONName: "target_url", GoType: "string", Kind: FieldKindBasic},
		},
	},
	{
		Name:          "MissingReference",
		Package:       "opentimelineio",
		SchemaName:    "MissingReference",
		SchemaVersion: 1,
		IsLeaf:        true,
		Fields: []Field{
			{Name: "name", JSONName: "name", GoType: "string", Kind: FieldKindBasic},
			{Name: "metadata", JSONName: "metadata", GoType: "AnyDictionary", Kind: FieldKindAnyDict},
			{Name: "availableRange", JSONName: "available_range", GoType: "*opentime.TimeRange", Kind: FieldKindPointer, IsPointer: true, IsOptional: true},
			{Name: "availableImageBounds", JSONName: "available_image_bounds", GoType: "*Box2d", Kind: FieldKindPointer, IsPointer: true, IsOptional: true},
		},
	},
	{
		Name:          "GeneratorReference",
		Package:       "opentimelineio",
		SchemaName:    "GeneratorReference",
		SchemaVersion: 1,
		IsLeaf:        true,
		Fields: []Field{
			{Name: "name", JSONName: "name", GoType: "string", Kind: FieldKindBasic},
			{Name: "metadata", JSONName: "metadata", GoType: "AnyDictionary", Kind: FieldKindAnyDict},
			{Name: "availableRange", JSONName: "available_range", GoType: "*opentime.TimeRange", Kind: FieldKindPointer, IsPointer: true, IsOptional: true},
			{Name: "availableImageBounds", JSONName: "available_image_bounds", GoType: "*Box2d", Kind: FieldKindPointer, IsPointer: true, IsOptional: true},
			{Name: "generatorKind", JSONName: "generator_kind", GoType: "string", Kind: FieldKindBasic},
			{Name: "parameters", JSONName: "parameters", GoType: "AnyDictionary", Kind: FieldKindAnyDict},
		},
	},
	{
		Name:          "LinearTimeWarp",
		Package:       "opentimelineio",
		SchemaName:    "LinearTimeWarp",
		SchemaVersion: 1,
		IsLeaf:        true,
		Fields: []Field{
			{Name: "name", JSONName: "name", GoType: "string", Kind: FieldKindBasic},
			{Name: "metadata", JSONName: "metadata", GoType: "AnyDictionary", Kind: FieldKindAnyDict},
			{Name: "effectName", JSONName: "effect_name", GoType: "string", Kind: FieldKindBasic},
			{Name: "timeScalar", JSONName: "time_scalar", GoType: "float64", Kind: FieldKindBasic},
		},
	},
	{
		Name:          "FreezeFrame",
		Package:       "opentimelineio",
		SchemaName:    "FreezeFrame",
		SchemaVersion: 1,
		IsLeaf:        true,
		Fields: []Field{
			{Name: "name", JSONName: "name", GoType: "string", Kind: FieldKindBasic},
			{Name: "metadata", JSONName: "metadata", GoType: "AnyDictionary", Kind: FieldKindAnyDict},
			{Name: "effectName", JSONName: "effect_name", GoType: "string", Kind: FieldKindBasic},
			{Name: "timeScalar", JSONName: "time_scalar", GoType: "float64", Kind: FieldKindBasic},
		},
	},
	{
		Name:          "Transition",
		Package:       "opentimelineio",
		SchemaName:    "Transition",
		SchemaVersion: 1,
		IsLeaf:        true,
		Fields: []Field{
			{Name: "name", JSONName: "name", GoType: "string", Kind: FieldKindBasic},
			{Name: "metadata", JSONName: "metadata", GoType: "AnyDictionary", Kind: FieldKindAnyDict},
			{Name: "transitionType", JSONName: "transition_type", GoType: "TransitionType", Kind: FieldKindBasic},
			{Name: "inOffset", JSONName: "in_offset", GoType: "opentime.RationalTime", Kind: FieldKindStruct},
			{Name: "outOffset", JSONName: "out_offset", GoType: "opentime.RationalTime", Kind: FieldKindStruct},
		},
	},
	{
		Name:          "Gap",
		Package:       "opentimelineio",
		SchemaName:    "Gap",
		SchemaVersion: 1,
		IsLeaf:        true,
		Fields: []Field{
			{Name: "name", JSONName: "name", GoType: "string", Kind: FieldKindBasic},
			{Name: "metadata", JSONName: "metadata", GoType: "AnyDictionary", Kind: FieldKindAnyDict},
			{Name: "sourceRange", JSONName: "source_range", GoType: "*opentime.TimeRange", Kind: FieldKindPointer, IsPointer: true, IsOptional: true},
			{Name: "effects", JSONName: "effects", GoType: "[]Effect", Kind: FieldKindSlice, ElemType: "Effect"},
			{Name: "markers", JSONName: "markers", GoType: "[]*Marker", Kind: FieldKindSlice, ElemType: "*Marker"},
			{Name: "enabled", JSONName: "enabled", GoType: "bool", Kind: FieldKindBasic},
			{Name: "itemColor", JSONName: "color", GoType: "*Color", Kind: FieldKindPointer, IsPointer: true, IsOptional: true},
		},
	},
}

// otioContainerTypes defines types with polymorphic fields (hand-written encoders).
var otioContainerTypes = []TypeDef{
	{
		Name:          "Clip",
		Package:       "opentimelineio",
		SchemaName:    "Clip",
		SchemaVersion: 2,
		IsLeaf:        false,
		Fields: []Field{
			{Name: "name", JSONName: "name", GoType: "string", Kind: FieldKindBasic},
			{Name: "metadata", JSONName: "metadata", GoType: "AnyDictionary", Kind: FieldKindAnyDict},
			{Name: "sourceRange", JSONName: "source_range", GoType: "*opentime.TimeRange", Kind: FieldKindPointer, IsPointer: true, IsOptional: true},
			{Name: "effects", JSONName: "effects", GoType: "[]Effect", Kind: FieldKindInterface, ElemType: "Effect"},
			{Name: "markers", JSONName: "markers", GoType: "[]*Marker", Kind: FieldKindSlice, ElemType: "*Marker"},
			{Name: "enabled", JSONName: "enabled", GoType: "bool", Kind: FieldKindBasic},
			{Name: "color", JSONName: "color", GoType: "*Color", Kind: FieldKindPointer, IsPointer: true, IsOptional: true},
			{Name: "mediaReferences", JSONName: "media_references", GoType: "map[string]MediaReference", Kind: FieldKindInterface, ElemType: "MediaReference"},
			{Name: "activeMediaReferenceKey", JSONName: "active_media_reference_key", GoType: "string", Kind: FieldKindBasic},
		},
	},
	{
		Name:          "Track",
		Package:       "opentimelineio",
		SchemaName:    "Track",
		SchemaVersion: 1,
		IsLeaf:        false,
		Fields: []Field{
			{Name: "name", JSONName: "name", GoType: "string", Kind: FieldKindBasic},
			{Name: "metadata", JSONName: "metadata", GoType: "AnyDictionary", Kind: FieldKindAnyDict},
			{Name: "sourceRange", JSONName: "source_range", GoType: "*opentime.TimeRange", Kind: FieldKindPointer, IsPointer: true, IsOptional: true},
			{Name: "effects", JSONName: "effects", GoType: "[]Effect", Kind: FieldKindInterface, ElemType: "Effect"},
			{Name: "markers", JSONName: "markers", GoType: "[]*Marker", Kind: FieldKindSlice, ElemType: "*Marker"},
			{Name: "enabled", JSONName: "enabled", GoType: "bool", Kind: FieldKindBasic},
			{Name: "kind", JSONName: "kind", GoType: "TrackKind", Kind: FieldKindBasic},
			{Name: "children", JSONName: "children", GoType: "[]Composable", Kind: FieldKindInterface, ElemType: "Composable"},
		},
	},
	{
		Name:          "Stack",
		Package:       "opentimelineio",
		SchemaName:    "Stack",
		SchemaVersion: 1,
		IsLeaf:        false,
		Fields: []Field{
			{Name: "name", JSONName: "name", GoType: "string", Kind: FieldKindBasic},
			{Name: "metadata", JSONName: "metadata", GoType: "AnyDictionary", Kind: FieldKindAnyDict},
			{Name: "sourceRange", JSONName: "source_range", GoType: "*opentime.TimeRange", Kind: FieldKindPointer, IsPointer: true, IsOptional: true},
			{Name: "effects", JSONName: "effects", GoType: "[]Effect", Kind: FieldKindInterface, ElemType: "Effect"},
			{Name: "markers", JSONName: "markers", GoType: "[]*Marker", Kind: FieldKindSlice, ElemType: "*Marker"},
			{Name: "enabled", JSONName: "enabled", GoType: "bool", Kind: FieldKindBasic},
			{Name: "children", JSONName: "children", GoType: "[]Composable", Kind: FieldKindInterface, ElemType: "Composable"},
		},
	},
	{
		Name:          "Timeline",
		Package:       "opentimelineio",
		SchemaName:    "Timeline",
		SchemaVersion: 1,
		IsLeaf:        false,
		Fields: []Field{
			{Name: "name", JSONName: "name", GoType: "string", Kind: FieldKindBasic},
			{Name: "metadata", JSONName: "metadata", GoType: "AnyDictionary", Kind: FieldKindAnyDict},
			{Name: "globalStartTime", JSONName: "global_start_time", GoType: "*opentime.RationalTime", Kind: FieldKindPointer, IsPointer: true, IsOptional: true},
			{Name: "tracks", JSONName: "tracks", GoType: "*Stack", Kind: FieldKindPointer},
		},
	},
	{
		Name:          "SerializableCollection",
		Package:       "opentimelineio",
		SchemaName:    "SerializableCollection",
		SchemaVersion: 1,
		IsLeaf:        false,
		Fields: []Field{
			{Name: "name", JSONName: "name", GoType: "string", Kind: FieldKindBasic},
			{Name: "metadata", JSONName: "metadata", GoType: "AnyDictionary", Kind: FieldKindAnyDict},
			{Name: "children", JSONName: "children", GoType: "[]SerializableObject", Kind: FieldKindInterface, ElemType: "SerializableObject"},
		},
	},
}
