// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"bytes"
	"io"
	"os"

	"github.com/Avalanche-io/gotio/internal/jsonenc"
)

// SerializableObject is the base interface for all serializable types.
type SerializableObject interface {
	// SchemaName returns the schema name.
	SchemaName() string

	// SchemaVersion returns the schema version.
	SchemaVersion() int

	// Clone creates a deep copy.
	Clone() SerializableObject

	// IsEquivalentTo returns true if equivalent to another.
	IsEquivalentTo(other SerializableObject) bool
}

// FromJSONString parses JSON into a SerializableObject.
func FromJSONString(jsonStr string) (SerializableObject, error) {
	return FromJSONBytes([]byte(jsonStr))
}

// FromJSONBytes parses JSON bytes into a SerializableObject.
func FromJSONBytes(data []byte) (SerializableObject, error) {
	return FromJSONBytesSonic(data)
}

// FromJSONFile reads a JSON file into a SerializableObject.
func FromJSONFile(filename string) (SerializableObject, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return FromJSONBytes(data)
}

// ToJSONString converts a SerializableObject to JSON string.
// If indent is provided, the output will be pretty-printed.
func ToJSONString(obj SerializableObject, indent string) (string, error) {
	var data []byte
	var err error
	if indent != "" {
		data, err = ToJSONBytesIndent(obj, indent)
	} else {
		data, err = ToJSONBytes(obj)
	}
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToJSONBytes converts a SerializableObject to JSON bytes.
func ToJSONBytes(obj SerializableObject) ([]byte, error) {
	var buf bytes.Buffer
	enc := jsonenc.NewEncoder(&buf)
	defer enc.Release()

	if err := jsonenc.EncodeValue(enc, obj); err != nil {
		return nil, err
	}

	if err := enc.Flush(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ToJSONWriter writes a SerializableObject to an io.Writer.
func ToJSONWriter(obj SerializableObject, w io.Writer) error {
	enc := jsonenc.NewEncoder(w)
	defer enc.Release()

	if err := jsonenc.EncodeValue(enc, obj); err != nil {
		return err
	}

	return enc.Flush()
}

// ToJSONBytesIndent converts a SerializableObject to indented JSON bytes.
func ToJSONBytesIndent(obj SerializableObject, indent string) ([]byte, error) {
	data, err := ToJSONBytes(obj)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := jsonIndent(&buf, data, "", indent); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ToJSONFile writes a SerializableObject to a JSON file.
func ToJSONFile(obj SerializableObject, filename string, indent string) error {
	var data []byte
	var err error

	if indent != "" {
		data, err = ToJSONBytesIndent(obj, indent)
	} else {
		data, err = ToJSONBytes(obj)
	}
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}
