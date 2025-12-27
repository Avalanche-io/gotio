// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"reflect"

	"github.com/mrjoshuak/gotio/internal/jsonenc"
)

// encodeClipFast encodes a Clip to JSON using the streaming encoder.
func encodeClipFast(enc *jsonenc.Encoder, v any) error {
	c := v.(*Clip)
	enc.BeginObject()
	enc.WriteStringField("OTIO_SCHEMA", "Clip.2")
	enc.WriteStringField("name", c.Name())
	if err := jsonenc.EncodeMetadata(enc, "metadata", c.Metadata()); err != nil {
		return err
	}
	if sr := c.SourceRange(); sr != nil {
		enc.WriteKey("source_range")
		if err := jsonenc.EncodeValue(enc, *sr); err != nil {
			return err
		}
	} else {
		enc.WriteNullField("source_range")
	}

	// Effects (polymorphic)
	enc.WriteKey("effects")
	enc.BeginArray()
	for i, eff := range c.Effects() {
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
	for i, m := range c.Markers() {
		if i > 0 {
			enc.WriteComma()
		}
		if err := jsonenc.EncodeValue(enc, m); err != nil {
			return err
		}
	}
	enc.EndArray()

	enc.WriteBoolField("enabled", c.Enabled())

	if color := c.ItemColor(); color != nil {
		enc.WriteKey("color")
		if err := jsonenc.EncodeValue(enc, color); err != nil {
			return err
		}
	} else {
		enc.WriteNullField("color")
	}

	// Media references (polymorphic map)
	enc.WriteKey("media_references")
	enc.BeginObject()
	first := true
	for k, ref := range c.MediaReferences() {
		if !first {
			enc.WriteComma()
		}
		first = false
		enc.WriteKey(k)
		if err := jsonenc.EncodeValue(enc, ref); err != nil {
			return err
		}
	}
	enc.EndObject()

	enc.WriteStringField("active_media_reference_key", c.ActiveMediaReferenceKey())
	enc.EndObject()
	return nil
}

// encodeTrackFast encodes a Track to JSON using the streaming encoder.
func encodeTrackFast(enc *jsonenc.Encoder, v any) error {
	t := v.(*Track)
	enc.BeginObject()
	enc.WriteStringField("OTIO_SCHEMA", "Track.1")
	enc.WriteStringField("name", t.Name())
	if err := jsonenc.EncodeMetadata(enc, "metadata", t.Metadata()); err != nil {
		return err
	}
	if sr := t.SourceRange(); sr != nil {
		enc.WriteKey("source_range")
		if err := jsonenc.EncodeValue(enc, *sr); err != nil {
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
	enc.WriteStringField("kind", string(t.Kind()))

	// Children (polymorphic)
	enc.WriteKey("children")
	enc.BeginArray()
	for i, child := range t.Children() {
		if i > 0 {
			enc.WriteComma()
		}
		if err := jsonenc.EncodeValue(enc, child); err != nil {
			return err
		}
	}
	enc.EndArray()

	enc.EndObject()
	return nil
}

// encodeStackFast encodes a Stack to JSON using the streaming encoder.
func encodeStackFast(enc *jsonenc.Encoder, v any) error {
	s := v.(*Stack)
	enc.BeginObject()
	enc.WriteStringField("OTIO_SCHEMA", "Stack.1")
	enc.WriteStringField("name", s.Name())
	if err := jsonenc.EncodeMetadata(enc, "metadata", s.Metadata()); err != nil {
		return err
	}
	if sr := s.SourceRange(); sr != nil {
		enc.WriteKey("source_range")
		if err := jsonenc.EncodeValue(enc, *sr); err != nil {
			return err
		}
	} else {
		enc.WriteNullField("source_range")
	}

	// Effects (polymorphic)
	enc.WriteKey("effects")
	enc.BeginArray()
	for i, eff := range s.Effects() {
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
	for i, m := range s.Markers() {
		if i > 0 {
			enc.WriteComma()
		}
		if err := jsonenc.EncodeValue(enc, m); err != nil {
			return err
		}
	}
	enc.EndArray()

	enc.WriteBoolField("enabled", s.Enabled())

	// Children (polymorphic)
	enc.WriteKey("children")
	enc.BeginArray()
	for i, child := range s.Children() {
		if i > 0 {
			enc.WriteComma()
		}
		if err := jsonenc.EncodeValue(enc, child); err != nil {
			return err
		}
	}
	enc.EndArray()

	enc.EndObject()
	return nil
}

// encodeTimelineFast encodes a Timeline to JSON using the streaming encoder.
func encodeTimelineFast(enc *jsonenc.Encoder, v any) error {
	t := v.(*Timeline)
	enc.BeginObject()
	enc.WriteStringField("OTIO_SCHEMA", "Timeline.1")
	enc.WriteStringField("name", t.Name())
	if err := jsonenc.EncodeMetadata(enc, "metadata", t.Metadata()); err != nil {
		return err
	}
	if gst := t.GlobalStartTime(); gst != nil {
		enc.WriteKey("global_start_time")
		if err := jsonenc.EncodeValue(enc, *gst); err != nil {
			return err
		}
	} else {
		enc.WriteNullField("global_start_time")
	}

	enc.WriteKey("tracks")
	if tracks := t.Tracks(); tracks != nil {
		if err := encodeStackFast(enc, tracks); err != nil {
			return err
		}
	} else {
		enc.WriteNull()
	}

	enc.EndObject()
	return nil
}

// encodeSerializableCollectionFast encodes a SerializableCollection to JSON using the streaming encoder.
func encodeSerializableCollectionFast(enc *jsonenc.Encoder, v any) error {
	c := v.(*SerializableCollection)
	enc.BeginObject()
	enc.WriteStringField("OTIO_SCHEMA", "SerializableCollection.1")
	enc.WriteStringField("name", c.Name())
	if err := jsonenc.EncodeMetadata(enc, "metadata", c.Metadata()); err != nil {
		return err
	}

	// Children (polymorphic)
	enc.WriteKey("children")
	enc.BeginArray()
	for i, child := range c.Children() {
		if i > 0 {
			enc.WriteComma()
		}
		if err := jsonenc.EncodeValue(enc, child); err != nil {
			return err
		}
	}
	enc.EndArray()

	enc.EndObject()
	return nil
}

// encodeColorFast encodes a Color to JSON using the streaming encoder.
func encodeColorFast(enc *jsonenc.Encoder, v any) error {
	c := v.(*Color)
	enc.BeginObject()
	enc.WriteStringField("OTIO_SCHEMA", "Color.1")
	enc.WriteFloat64Field("r", c.R)
	enc.WriteFloat64Field("g", c.G)
	enc.WriteFloat64Field("b", c.B)
	enc.WriteFloat64Field("a", c.A)
	enc.EndObject()
	return nil
}

func init() {
	jsonenc.Register(jsonenc.TypeInfo{
		SchemaName:    "Clip",
		SchemaVersion: 2,
		GoType:        reflect.TypeOf((*Clip)(nil)),
		Encode:        encodeClipFast,
	})

	jsonenc.Register(jsonenc.TypeInfo{
		SchemaName:    "Track",
		SchemaVersion: 1,
		GoType:        reflect.TypeOf((*Track)(nil)),
		Encode:        encodeTrackFast,
	})

	jsonenc.Register(jsonenc.TypeInfo{
		SchemaName:    "Stack",
		SchemaVersion: 1,
		GoType:        reflect.TypeOf((*Stack)(nil)),
		Encode:        encodeStackFast,
	})

	jsonenc.Register(jsonenc.TypeInfo{
		SchemaName:    "Timeline",
		SchemaVersion: 1,
		GoType:        reflect.TypeOf((*Timeline)(nil)),
		Encode:        encodeTimelineFast,
	})

	jsonenc.Register(jsonenc.TypeInfo{
		SchemaName:    "SerializableCollection",
		SchemaVersion: 1,
		GoType:        reflect.TypeOf((*SerializableCollection)(nil)),
		Encode:        encodeSerializableCollectionFast,
	})

	jsonenc.Register(jsonenc.TypeInfo{
		SchemaName:    "Color",
		SchemaVersion: 1,
		GoType:        reflect.TypeOf((*Color)(nil)),
		Encode:        encodeColorFast,
	})
}
