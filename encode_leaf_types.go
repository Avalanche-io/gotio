// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package gotio

import (
	"reflect"

	"github.com/Avalanche-io/gotio/internal/jsonenc"
)

// encodeMarkerFast encodes a Marker to JSON using the streaming encoder.
func encodeMarkerFast(enc *jsonenc.Encoder, v any) error {
	t := v.(*Marker)
	enc.BeginObject()
	enc.WriteStringField("OTIO_SCHEMA", "Marker.2")
	enc.WriteStringField("name", t.Name())
	if err := jsonenc.EncodeMetadata(enc, "metadata", t.Metadata()); err != nil {
		return err
	}
	enc.WriteKey("marked_range")
	if err := jsonenc.EncodeValue(enc, t.MarkedRange()); err != nil {
		return err
	}
	enc.WriteStringField("color", string(t.Color()))
	enc.WriteStringField("comment", t.Comment())
	enc.EndObject()
	return nil
}

// encodeExternalReferenceFast encodes an ExternalReference to JSON using the streaming encoder.
func encodeExternalReferenceFast(enc *jsonenc.Encoder, v any) error {
	t := v.(*ExternalReference)
	enc.BeginObject()
	enc.WriteStringField("OTIO_SCHEMA", "ExternalReference.1")
	enc.WriteStringField("name", t.Name())
	if err := jsonenc.EncodeMetadata(enc, "metadata", t.Metadata()); err != nil {
		return err
	}
	if ptr := t.AvailableRange(); ptr != nil {
		enc.WriteKey("available_range")
		if err := jsonenc.EncodeValue(enc, *ptr); err != nil {
			return err
		}
	} else {
		enc.WriteNullField("available_range")
	}
	if ptr := t.AvailableImageBounds(); ptr != nil {
		enc.WriteKey("available_image_bounds")
		if err := jsonenc.EncodeValue(enc, ptr); err != nil {
			return err
		}
	} else {
		enc.WriteNullField("available_image_bounds")
	}
	enc.WriteStringField("target_url", t.TargetURL())
	enc.EndObject()
	return nil
}

// encodeMissingReferenceFast encodes a MissingReference to JSON using the streaming encoder.
func encodeMissingReferenceFast(enc *jsonenc.Encoder, v any) error {
	t := v.(*MissingReference)
	enc.BeginObject()
	enc.WriteStringField("OTIO_SCHEMA", "MissingReference.1")
	enc.WriteStringField("name", t.Name())
	if err := jsonenc.EncodeMetadata(enc, "metadata", t.Metadata()); err != nil {
		return err
	}
	if ptr := t.AvailableRange(); ptr != nil {
		enc.WriteKey("available_range")
		if err := jsonenc.EncodeValue(enc, *ptr); err != nil {
			return err
		}
	} else {
		enc.WriteNullField("available_range")
	}
	if ptr := t.AvailableImageBounds(); ptr != nil {
		enc.WriteKey("available_image_bounds")
		if err := jsonenc.EncodeValue(enc, ptr); err != nil {
			return err
		}
	} else {
		enc.WriteNullField("available_image_bounds")
	}
	enc.EndObject()
	return nil
}

// encodeGeneratorReferenceFast encodes a GeneratorReference to JSON using the streaming encoder.
func encodeGeneratorReferenceFast(enc *jsonenc.Encoder, v any) error {
	t := v.(*GeneratorReference)
	enc.BeginObject()
	enc.WriteStringField("OTIO_SCHEMA", "GeneratorReference.1")
	enc.WriteStringField("name", t.Name())
	if err := jsonenc.EncodeMetadata(enc, "metadata", t.Metadata()); err != nil {
		return err
	}
	if ptr := t.AvailableRange(); ptr != nil {
		enc.WriteKey("available_range")
		if err := jsonenc.EncodeValue(enc, *ptr); err != nil {
			return err
		}
	} else {
		enc.WriteNullField("available_range")
	}
	if ptr := t.AvailableImageBounds(); ptr != nil {
		enc.WriteKey("available_image_bounds")
		if err := jsonenc.EncodeValue(enc, ptr); err != nil {
			return err
		}
	} else {
		enc.WriteNullField("available_image_bounds")
	}
	enc.WriteStringField("generator_kind", t.GeneratorKind())
	if err := jsonenc.EncodeMetadata(enc, "parameters", t.Parameters()); err != nil {
		return err
	}
	enc.EndObject()
	return nil
}

// encodeLinearTimeWarpFast encodes a LinearTimeWarp to JSON using the streaming encoder.
func encodeLinearTimeWarpFast(enc *jsonenc.Encoder, v any) error {
	t := v.(*LinearTimeWarp)
	enc.BeginObject()
	enc.WriteStringField("OTIO_SCHEMA", "LinearTimeWarp.1")
	enc.WriteStringField("name", t.Name())
	if err := jsonenc.EncodeMetadata(enc, "metadata", t.Metadata()); err != nil {
		return err
	}
	enc.WriteStringField("effect_name", t.EffectName())
	enc.WriteFloat64Field("time_scalar", t.TimeScalar())
	enc.EndObject()
	return nil
}

// encodeFreezeFrameFast encodes a FreezeFrame to JSON using the streaming encoder.
func encodeFreezeFrameFast(enc *jsonenc.Encoder, v any) error {
	t := v.(*FreezeFrame)
	enc.BeginObject()
	enc.WriteStringField("OTIO_SCHEMA", "FreezeFrame.1")
	enc.WriteStringField("name", t.Name())
	if err := jsonenc.EncodeMetadata(enc, "metadata", t.Metadata()); err != nil {
		return err
	}
	enc.WriteStringField("effect_name", t.EffectName())
	enc.WriteFloat64Field("time_scalar", t.TimeScalar())
	enc.EndObject()
	return nil
}

// encodeTransitionFast encodes a Transition to JSON using the streaming encoder.
func encodeTransitionFast(enc *jsonenc.Encoder, v any) error {
	t := v.(*Transition)
	enc.BeginObject()
	enc.WriteStringField("OTIO_SCHEMA", "Transition.1")
	enc.WriteStringField("name", t.Name())
	if err := jsonenc.EncodeMetadata(enc, "metadata", t.Metadata()); err != nil {
		return err
	}
	enc.WriteStringField("transition_type", string(t.TransitionType()))
	enc.WriteKey("in_offset")
	if err := jsonenc.EncodeValue(enc, t.InOffset()); err != nil {
		return err
	}
	enc.WriteKey("out_offset")
	if err := jsonenc.EncodeValue(enc, t.OutOffset()); err != nil {
		return err
	}
	enc.EndObject()
	return nil
}

// encodeUnknownSchemaFast encodes an UnknownSchema to JSON using its MarshalJSON method.
// UnknownSchema has a dynamic schema name, so we fall back to standard marshaling.
func encodeUnknownSchemaFast(enc *jsonenc.Encoder, v any) error {
	u := v.(*UnknownSchema)
	data, err := u.MarshalJSON()
	if err != nil {
		return err
	}
	enc.WriteRawJSON(data)
	return nil
}

// encodeEffectFast encodes an Effect to JSON using the streaming encoder.
func encodeEffectFast(enc *jsonenc.Encoder, v any) error {
	t := v.(*EffectImpl)
	enc.BeginObject()
	enc.WriteStringField("OTIO_SCHEMA", "Effect.1")
	enc.WriteStringField("name", t.Name())
	if err := jsonenc.EncodeMetadata(enc, "metadata", t.Metadata()); err != nil {
		return err
	}
	enc.WriteStringField("effect_name", t.EffectName())
	enc.EndObject()
	return nil
}

// encodeTimeEffectFast encodes a TimeEffect to JSON using the streaming encoder.
func encodeTimeEffectFast(enc *jsonenc.Encoder, v any) error {
	t := v.(*TimeEffectImpl)
	enc.BeginObject()
	enc.WriteStringField("OTIO_SCHEMA", "TimeEffect.1")
	enc.WriteStringField("name", t.Name())
	if err := jsonenc.EncodeMetadata(enc, "metadata", t.Metadata()); err != nil {
		return err
	}
	enc.WriteStringField("effect_name", t.EffectName())
	enc.EndObject()
	return nil
}

// encodeImageSequenceReferenceFast encodes an ImageSequenceReference to JSON.
func encodeImageSequenceReferenceFast(enc *jsonenc.Encoder, v any) error {
	t := v.(*ImageSequenceReference)
	enc.BeginObject()
	enc.WriteStringField("OTIO_SCHEMA", "ImageSequenceReference.1")
	enc.WriteStringField("name", t.Name())
	if err := jsonenc.EncodeMetadata(enc, "metadata", t.Metadata()); err != nil {
		return err
	}
	if ptr := t.AvailableRange(); ptr != nil {
		enc.WriteKey("available_range")
		if err := jsonenc.EncodeValue(enc, *ptr); err != nil {
			return err
		}
	} else {
		enc.WriteNullField("available_range")
	}
	if ptr := t.AvailableImageBounds(); ptr != nil {
		enc.WriteKey("available_image_bounds")
		if err := jsonenc.EncodeValue(enc, ptr); err != nil {
			return err
		}
	} else {
		enc.WriteNullField("available_image_bounds")
	}
	enc.WriteStringField("target_url_base", t.TargetURLBase())
	enc.WriteStringField("name_prefix", t.NamePrefix())
	enc.WriteStringField("name_suffix", t.NameSuffix())
	enc.WriteIntField("start_frame", t.StartFrame())
	enc.WriteIntField("frame_step", t.FrameStep())
	enc.WriteFloat64Field("rate", t.Rate())
	enc.WriteIntField("frame_zero_padding", t.FrameZeroPadding())
	enc.WriteStringField("missing_frame_policy", string(t.MissingFramePolicy()))
	enc.EndObject()
	return nil
}

// encodeGapFast encodes a Gap to JSON using the streaming encoder.
func encodeGapFast(enc *jsonenc.Encoder, v any) error {
	t := v.(*Gap)
	enc.BeginObject()
	enc.WriteStringField("OTIO_SCHEMA", "Gap.1")
	enc.WriteStringField("name", t.Name())
	if err := jsonenc.EncodeMetadata(enc, "metadata", t.Metadata()); err != nil {
		return err
	}
	if ptr := t.SourceRange(); ptr != nil {
		enc.WriteKey("source_range")
		if err := jsonenc.EncodeValue(enc, *ptr); err != nil {
			return err
		}
	} else {
		enc.WriteNullField("source_range")
	}

	// Effects (polymorphic)
	enc.WriteKey("effects")
	enc.BeginArray()
	for i, eff := range t.Effects() {
		if i > 0 {
			enc.WriteComma()
		}
		if err := jsonenc.EncodeValue(enc, eff); err != nil {
			return err
		}
	}
	enc.EndArray()

	// Markers
	enc.WriteKey("markers")
	enc.BeginArray()
	for i, m := range t.Markers() {
		if i > 0 {
			enc.WriteComma()
		}
		if err := jsonenc.EncodeValue(enc, m); err != nil {
			return err
		}
	}
	enc.EndArray()

	enc.WriteBoolField("enabled", t.Enabled())
	if ptr := t.ItemColor(); ptr != nil {
		enc.WriteKey("color")
		if err := jsonenc.EncodeValue(enc, ptr); err != nil {
			return err
		}
	} else {
		enc.WriteNullField("color")
	}
	enc.EndObject()
	return nil
}

func init() {
	jsonenc.Register(jsonenc.TypeInfo{
		SchemaName:    "Marker",
		SchemaVersion: 2,
		GoType:        reflect.TypeOf((*Marker)(nil)),
		Encode:        encodeMarkerFast,
	})

	jsonenc.Register(jsonenc.TypeInfo{
		SchemaName:    "ExternalReference",
		SchemaVersion: 1,
		GoType:        reflect.TypeOf((*ExternalReference)(nil)),
		Encode:        encodeExternalReferenceFast,
	})

	jsonenc.Register(jsonenc.TypeInfo{
		SchemaName:    "MissingReference",
		SchemaVersion: 1,
		GoType:        reflect.TypeOf((*MissingReference)(nil)),
		Encode:        encodeMissingReferenceFast,
	})

	jsonenc.Register(jsonenc.TypeInfo{
		SchemaName:    "GeneratorReference",
		SchemaVersion: 1,
		GoType:        reflect.TypeOf((*GeneratorReference)(nil)),
		Encode:        encodeGeneratorReferenceFast,
	})

	jsonenc.Register(jsonenc.TypeInfo{
		SchemaName:    "LinearTimeWarp",
		SchemaVersion: 1,
		GoType:        reflect.TypeOf((*LinearTimeWarp)(nil)),
		Encode:        encodeLinearTimeWarpFast,
	})

	jsonenc.Register(jsonenc.TypeInfo{
		SchemaName:    "FreezeFrame",
		SchemaVersion: 1,
		GoType:        reflect.TypeOf((*FreezeFrame)(nil)),
		Encode:        encodeFreezeFrameFast,
	})

	jsonenc.Register(jsonenc.TypeInfo{
		SchemaName:    "Transition",
		SchemaVersion: 1,
		GoType:        reflect.TypeOf((*Transition)(nil)),
		Encode:        encodeTransitionFast,
	})

	jsonenc.Register(jsonenc.TypeInfo{
		SchemaName:    "Gap",
		SchemaVersion: 1,
		GoType:        reflect.TypeOf((*Gap)(nil)),
		Encode:        encodeGapFast,
	})

	jsonenc.Register(jsonenc.TypeInfo{
		SchemaName:    "Effect",
		SchemaVersion: 1,
		GoType:        reflect.TypeOf((*EffectImpl)(nil)),
		Encode:        encodeEffectFast,
	})

	jsonenc.Register(jsonenc.TypeInfo{
		SchemaName:    "TimeEffect",
		SchemaVersion: 1,
		GoType:        reflect.TypeOf((*TimeEffectImpl)(nil)),
		Encode:        encodeTimeEffectFast,
	})

	jsonenc.Register(jsonenc.TypeInfo{
		SchemaName:    "ImageSequenceReference",
		SchemaVersion: 1,
		GoType:        reflect.TypeOf((*ImageSequenceReference)(nil)),
		Encode:        encodeImageSequenceReferenceFast,
	})

	// UnknownSchema uses dynamic schema name, so register with empty name
	// and let EncodeValue fall back to MarshalJSON
	jsonenc.Register(jsonenc.TypeInfo{
		SchemaName:    "", // UnknownSchema has dynamic schema
		SchemaVersion: 0,
		GoType:        reflect.TypeOf((*UnknownSchema)(nil)),
		Encode:        encodeUnknownSchemaFast,
	})
}
