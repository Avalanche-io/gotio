// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package jsonenc

import (
	"fmt"
	"reflect"
	"sync"
)

// EncoderFunc is the function signature for type-specific encoders.
// It encodes the value v to the encoder enc.
type EncoderFunc func(enc *Encoder, v any) error

// TypeInfo holds encoding information for a schema type.
type TypeInfo struct {
	SchemaName    string
	SchemaVersion int
	GoType        reflect.Type
	Encode        EncoderFunc
}

// Registry provides O(1) lookup of encoders by schema name or Go type.
type Registry struct {
	mu       sync.RWMutex
	bySchema map[string]*TypeInfo // "Clip.2" -> TypeInfo
	byType   map[reflect.Type]*TypeInfo
}

// globalRegistry is the default registry used by the package.
var globalRegistry = NewRegistry()

// NewRegistry creates a new empty registry.
func NewRegistry() *Registry {
	return &Registry{
		bySchema: make(map[string]*TypeInfo, 32),
		byType:   make(map[reflect.Type]*TypeInfo, 32),
	}
}

// Register adds a type to the global registry.
// This should be called from init() functions.
func Register(info TypeInfo) {
	globalRegistry.Register(info)
}

// Register adds a type to the registry.
func (r *Registry) Register(info TypeInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%s.%d", info.SchemaName, info.SchemaVersion)
	r.bySchema[key] = &info

	if info.GoType != nil {
		r.byType[info.GoType] = &info
	}
}

// LookupBySchema returns the TypeInfo for a schema string (e.g., "Clip.2").
func LookupBySchema(schema string) (*TypeInfo, bool) {
	return globalRegistry.LookupBySchema(schema)
}

// LookupBySchema returns the TypeInfo for a schema string.
func (r *Registry) LookupBySchema(schema string) (*TypeInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info, ok := r.bySchema[schema]
	return info, ok
}

// LookupByType returns the TypeInfo for a Go type.
func LookupByType(t reflect.Type) (*TypeInfo, bool) {
	return globalRegistry.LookupByType(t)
}

// LookupByType returns the TypeInfo for a Go type.
func (r *Registry) LookupByType(t reflect.Type) (*TypeInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info, ok := r.byType[t]
	return info, ok
}

// SchemaProvider is implemented by types that know their schema info.
type SchemaProvider interface {
	SchemaName() string
	SchemaVersion() int
}

// EncodeValue encodes a value using the appropriate registered encoder.
// It first checks if the value implements SchemaProvider for O(1) dispatch.
func EncodeValue(enc *Encoder, v any) error {
	return globalRegistry.EncodeValue(enc, v)
}

// EncodeValue encodes a value using the appropriate registered encoder.
func (r *Registry) EncodeValue(enc *Encoder, v any) error {
	if v == nil {
		enc.WriteNull()
		return nil
	}

	// Fast path: check if value provides schema info
	if sp, ok := v.(SchemaProvider); ok {
		key := fmt.Sprintf("%s.%d", sp.SchemaName(), sp.SchemaVersion())
		r.mu.RLock()
		info, found := r.bySchema[key]
		r.mu.RUnlock()
		if found && info.Encode != nil {
			return info.Encode(enc, v)
		}
	}

	// Slow path: lookup by reflection type
	t := reflect.TypeOf(v)
	r.mu.RLock()
	info, found := r.byType[t]
	r.mu.RUnlock()
	if found && info.Encode != nil {
		return info.Encode(enc, v)
	}

	// Fallback: encode as basic JSON value
	return encodeBasicValue(enc, v)
}

// encodeBasicValue handles encoding of primitive types and maps.
func encodeBasicValue(enc *Encoder, v any) error {
	switch val := v.(type) {
	case nil:
		enc.WriteNull()
	case bool:
		enc.WriteBool(val)
	case int:
		enc.WriteInt(val)
	case int64:
		enc.WriteInt64(val)
	case float64:
		enc.WriteFloat64(val)
	case string:
		enc.WriteQuotedString(val)
	case []byte:
		enc.WriteRawJSON(val)
	case map[string]any:
		return encodeAnyMap(enc, val)
	case []any:
		return encodeAnySlice(enc, val)
	default:
		return fmt.Errorf("jsonenc: unsupported type %T", v)
	}
	return nil
}

// encodeAnyMap encodes a map[string]any (for metadata).
func encodeAnyMap(enc *Encoder, m map[string]any) error {
	enc.BeginObject()
	first := true
	for k, v := range m {
		if !first {
			enc.needComma = true
		}
		first = false
		enc.WriteKey(k)
		if err := encodeBasicValue(enc, v); err != nil {
			return err
		}
	}
	enc.EndObject()
	return nil
}

// encodeAnySlice encodes a []any.
func encodeAnySlice(enc *Encoder, s []any) error {
	enc.BeginArray()
	for i, v := range s {
		if i > 0 {
			enc.WriteComma()
		}
		if err := encodeBasicValue(enc, v); err != nil {
			return err
		}
	}
	enc.EndArray()
	return nil
}

// EncodeMetadata encodes a metadata map (map[string]any).
// This is a common operation in OTIO types.
func EncodeMetadata(enc *Encoder, key string, metadata map[string]any) error {
	enc.WriteKey(key)
	if metadata == nil {
		enc.BeginObject()
		enc.EndObject()
		return nil
	}
	return encodeAnyMap(enc, metadata)
}
