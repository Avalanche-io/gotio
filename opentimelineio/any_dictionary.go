// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

// AnyDictionary is a map of string keys to any values.
type AnyDictionary map[string]any

// CloneAnyDictionary creates a shallow copy of an AnyDictionary.
func CloneAnyDictionary(d AnyDictionary) AnyDictionary {
	if d == nil {
		return nil
	}
	result := make(AnyDictionary, len(d))
	for k, v := range d {
		result[k] = v
	}
	return result
}

// areMetadataEqual compares two AnyDictionary values for equality.
func areMetadataEqual(a, b AnyDictionary) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || v != bv {
			return false
		}
	}
	return true
}
