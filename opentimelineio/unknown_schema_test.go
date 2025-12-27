// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"encoding/json"
	"testing"
)

func TestUnknownSchemaVersion(t *testing.T) {
	data := AnyDictionary{"field1": "value1", "field2": 42}
	unknown := NewUnknownSchema("CustomType.2", data)

	if unknown.SchemaVersion() != 2 {
		t.Errorf("SchemaVersion = %d, want 2", unknown.SchemaVersion())
	}
}

func TestUnknownSchemaOriginalSchema(t *testing.T) {
	data := AnyDictionary{}
	unknown := NewUnknownSchema("CustomType.3", data)

	if unknown.OriginalSchema() != "CustomType.3" {
		t.Errorf("OriginalSchema = %s, want CustomType.3", unknown.OriginalSchema())
	}
}

func TestUnknownSchemaClone(t *testing.T) {
	data := AnyDictionary{"key": "value", "number": 123}
	unknown := NewUnknownSchema("MyType.1", data)

	clone := unknown.Clone().(*UnknownSchema)

	if clone.OriginalSchema() != "MyType.1" {
		t.Errorf("Clone OriginalSchema = %s, want MyType.1", clone.OriginalSchema())
	}
	if clone.Data()["key"] != "value" {
		t.Error("Clone data should match")
	}
}

func TestUnknownSchemaIsEquivalentTo(t *testing.T) {
	data := AnyDictionary{"key": "value"}
	u1 := NewUnknownSchema("TypeA.1", data)
	u2 := NewUnknownSchema("TypeA.1", data)
	u3 := NewUnknownSchema("TypeB.1", data)

	if !u1.IsEquivalentTo(u2) {
		t.Error("Identical unknown schemas should be equivalent")
	}
	if u1.IsEquivalentTo(u3) {
		t.Error("Different unknown schemas should not be equivalent")
	}

	// Test with non-UnknownSchema
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)
	if u1.IsEquivalentTo(clip) {
		t.Error("UnknownSchema should not be equivalent to Clip")
	}
}

func TestUnknownSchemaJSON(t *testing.T) {
	data := AnyDictionary{
		"custom_field":  "custom_value",
		"custom_number": float64(42),
	}
	unknown := NewUnknownSchema("CustomWidget.1", data)

	jsonData, err := json.Marshal(unknown)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Verify the JSON contains the original schema
	var m map[string]interface{}
	if err := json.Unmarshal(jsonData, &m); err != nil {
		t.Fatalf("Unmarshal to map error: %v", err)
	}
	if m["OTIO_SCHEMA"] != "CustomWidget.1" {
		t.Errorf("OTIO_SCHEMA = %v, want CustomWidget.1", m["OTIO_SCHEMA"])
	}
}

func TestUnknownSchemaRoundTrip(t *testing.T) {
	// Create JSON for an unknown schema type
	jsonStr := `{
		"OTIO_SCHEMA": "UnknownWidget.5",
		"name": "test_widget",
		"metadata": {},
		"custom_property": "hello",
		"custom_number": 99
	}`

	obj, err := FromJSONString(jsonStr)
	if err != nil {
		t.Fatalf("FromJSONString error: %v", err)
	}

	unknown, ok := obj.(*UnknownSchema)
	if !ok {
		t.Fatalf("Expected UnknownSchema, got %T", obj)
	}

	if unknown.OriginalSchema() != "UnknownWidget.5" {
		t.Errorf("OriginalSchema = %s, want UnknownWidget.5", unknown.OriginalSchema())
	}
	if unknown.Data()["custom_property"] != "hello" {
		t.Error("Custom property should be preserved")
	}
}
