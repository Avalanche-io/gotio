// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package gotio

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// Schema identifies a schema with name and version.
type Schema struct {
	Name    string
	Version int
}

// String returns the schema string representation (e.g., "Clip.2").
func (s Schema) String() string {
	return fmt.Sprintf("%s.%d", s.Name, s.Version)
}

// SchemaFactory creates new instances of a schema type.
type SchemaFactory func() SerializableObject

var (
	schemaRegistry = make(map[string]SchemaFactory)
	schemaAliases  = make(map[string]string) // alias -> canonical name
	schemaLock     sync.RWMutex
)

// RegisterSchema registers a schema factory.
func RegisterSchema(schema Schema, factory SchemaFactory) {
	schemaLock.Lock()
	defer schemaLock.Unlock()
	schemaRegistry[schema.Name] = factory
}

// RegisterSchemaAlias registers an alias name that maps to a canonical schema name.
// This is useful for supporting legacy schema names (e.g., "Sequence" -> "Track").
func RegisterSchemaAlias(alias, canonicalName string) {
	schemaLock.Lock()
	defer schemaLock.Unlock()
	schemaAliases[alias] = canonicalName
}

// resolveSchemaName resolves a schema name, following aliases if necessary.
func resolveSchemaName(name string) string {
	if canonical, ok := schemaAliases[name]; ok {
		return canonical
	}
	return name
}

// CreateSchema creates a new instance of the given schema.
func CreateSchema(schemaName string) (SerializableObject, error) {
	schemaLock.RLock()
	defer schemaLock.RUnlock()

	// Resolve aliases
	resolved := resolveSchemaName(schemaName)
	factory, ok := schemaRegistry[resolved]
	if !ok {
		return nil, &SchemaError{Schema: schemaName, Message: "schema not registered"}
	}
	return factory(), nil
}

// IsSchemaRegistered returns true if the schema is registered.
func IsSchemaRegistered(schemaName string) bool {
	schemaLock.RLock()
	defer schemaLock.RUnlock()
	// Resolve aliases
	resolved := resolveSchemaName(schemaName)
	_, ok := schemaRegistry[resolved]
	return ok
}

// ParseSchema parses a schema string (e.g., "Clip.2") into name and version.
func ParseSchema(schemaStr string) (name string, version int, err error) {
	if schemaStr == "" {
		return "", 0, &SchemaError{Schema: schemaStr, Message: "empty schema string"}
	}

	// Try to split on the last dot
	if idx := strings.LastIndex(schemaStr, "."); idx >= 0 {
		name = schemaStr[:idx]
		versionStr := schemaStr[idx+1:]
		version, err = strconv.Atoi(versionStr)
		if err != nil {
			// If version part isn't a number, treat entire string as name
			name = schemaStr
			version = 1
		}
	} else {
		name = schemaStr
		version = 1
	}

	return name, version, nil
}
