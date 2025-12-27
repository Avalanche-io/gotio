// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"bytes"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/mrjoshuak/gotio/opentime"
)

// SanitizeJSON replaces Python's non-standard JSON values (Inf, NaN, -Infinity) with null.
// Uses fast byte scanning - returns original data if no changes needed.
func SanitizeJSON(data []byte) []byte {
	// Fast path: check if any non-standard values might exist
	if !bytes.Contains(data, []byte("Inf")) && !bytes.Contains(data, []byte("NaN")) {
		return data
	}

	result := make([]byte, 0, len(data))
	i := 0

	for i < len(data) {
		if data[i] != ':' {
			result = append(result, data[i])
			i++
			continue
		}

		result = append(result, ':')
		i++

		// Skip whitespace
		wsStart := len(result)
		for i < len(data) && (data[i] == ' ' || data[i] == '\t' || data[i] == '\n' || data[i] == '\r') {
			result = append(result, data[i])
			i++
		}

		if i >= len(data) {
			break
		}

		replaced := false

		if data[i] == '-' && i+1 < len(data) {
			if i+9 <= len(data) && string(data[i:i+9]) == "-Infinity" {
				i += 9
				replaced = true
			} else if i+4 <= len(data) && string(data[i:i+4]) == "-Inf" && (i+4 >= len(data) || !isAlphaNum(data[i+4])) {
				i += 4
				replaced = true
			}
		} else if data[i] == 'I' {
			if i+8 <= len(data) && string(data[i:i+8]) == "Infinity" {
				i += 8
				replaced = true
			} else if i+3 <= len(data) && string(data[i:i+3]) == "Inf" && (i+3 >= len(data) || !isAlphaNum(data[i+3])) {
				i += 3
				replaced = true
			}
		} else if data[i] == 'N' && i+3 <= len(data) && string(data[i:i+3]) == "NaN" && (i+3 >= len(data) || !isAlphaNum(data[i+3])) {
			i += 3
			replaced = true
		}

		if replaced {
			result = result[:wsStart]
			result = append(result, ' ', 'n', 'u', 'l', 'l')
		}
	}

	return result
}

func isAlphaNum(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// decodeSonicMetadata extracts metadata from a map.
func decodeSonicMetadata(m map[string]any) AnyDictionary {
	if md, ok := m["metadata"].(map[string]any); ok {
		return md
	}
	return nil
}

// decodeSonicEffects decodes effects array from a map.
func decodeSonicEffects(m map[string]any) []Effect {
	effs, ok := m["effects"].([]any)
	if !ok {
		return nil
	}
	var effects []Effect
	for _, effAny := range effs {
		if effMap, ok := effAny.(map[string]any); ok {
			if eff := decodeSonicEffect(effMap); eff != nil {
				effects = append(effects, eff)
			}
		}
	}
	return effects
}

// decodeSonicMarkers decodes markers array from a map.
func decodeSonicMarkers(m map[string]any) []*Marker {
	marks, ok := m["markers"].([]any)
	if !ok {
		return nil
	}
	var markers []*Marker
	for _, markAny := range marks {
		if markMap, ok := markAny.(map[string]any); ok {
			if marker := decodeSonicMarker(markMap); marker != nil {
				markers = append(markers, marker)
			}
		}
	}
	return markers
}

// decodeSonicRationalTime decodes a RationalTime from a map.
func decodeSonicRationalTime(v any) *opentime.RationalTime {
	m, ok := v.(map[string]any)
	if !ok || m == nil {
		return nil
	}
	value, _ := m["value"].(float64)
	rate, _ := m["rate"].(float64)
	rt := opentime.NewRationalTime(value, rate)
	return &rt
}

// decodeSonicTimeline decodes a Timeline from a sonic-parsed map.
func decodeSonicTimeline(m map[string]any) (*Timeline, error) {
	name, _ := m["name"].(string)
	metadata := decodeSonicMetadata(m)
	globalStartTime := decodeSonicRationalTime(m["global_start_time"])

	timeline := NewTimeline(name, globalStartTime, metadata)

	if tracks, ok := m["tracks"].(map[string]any); ok {
		if stack, err := decodeSonicStack(tracks); err == nil {
			timeline.SetTracks(stack)
		}
	}

	return timeline, nil
}

// decodeSonicStack decodes a Stack from a sonic-parsed map.
func decodeSonicStack(m map[string]any) (*Stack, error) {
	name, _ := m["name"].(string)
	enabled, _ := m["enabled"].(bool)
	sourceRange := decodeSonicTimeRange(m["source_range"])
	color := decodeSonicColor(m["color"])
	metadata := decodeSonicMetadata(m)
	effects := decodeSonicEffects(m)
	markers := decodeSonicMarkers(m)

	stack := NewStack(name, sourceRange, metadata, effects, markers, color)
	stack.SetEnabled(enabled)

	// Decode children (Tracks)
	if children, ok := m["children"].([]any); ok {
		for _, childAny := range children {
			if childMap, ok := childAny.(map[string]any); ok {
				schema, _ := childMap["OTIO_SCHEMA"].(string)
				if schema == "Track.1" {
					if track, err := decodeSonicTrack(childMap); err == nil {
						stack.AppendChild(track)
					}
				}
			}
		}
	}

	return stack, nil
}

// decodeSonicTrack decodes a Track from a sonic-parsed map.
func decodeSonicTrack(m map[string]any) (*Track, error) {
	name, _ := m["name"].(string)
	kind, _ := m["kind"].(string)
	enabled, _ := m["enabled"].(bool)
	sourceRange := decodeSonicTimeRange(m["source_range"])
	color := decodeSonicColor(m["color"])
	metadata := decodeSonicMetadata(m)

	track := NewTrack(name, sourceRange, kind, metadata, color)
	track.SetEnabled(enabled)
	track.effects = decodeSonicEffects(m)
	track.markers = decodeSonicMarkers(m)

	// Decode children (Clips, Gaps, Transitions)
	if children, ok := m["children"].([]any); ok {
		for _, childAny := range children {
			if childMap, ok := childAny.(map[string]any); ok {
				schema, _ := childMap["OTIO_SCHEMA"].(string)
				switch schema {
				case "Clip.2":
					if clip, err := decodeSonicClip(childMap); err == nil {
						track.AppendChild(clip)
					}
				case "Gap.1":
					if gap := decodeSonicGap(childMap); gap != nil {
						track.AppendChild(gap)
					}
				case "Transition.1":
					if trans := decodeSonicTransition(childMap); trans != nil {
						track.AppendChild(trans)
					}
				}
			}
		}
	}

	return track, nil
}

// decodeSonicClip decodes a Clip from a sonic-parsed map.
func decodeSonicClip(m map[string]any) (*Clip, error) {
	name, _ := m["name"].(string)
	enabled, _ := m["enabled"].(bool)
	sourceRange := decodeSonicTimeRange(m["source_range"])
	color := decodeSonicColor(m["color"])
	activeKey, _ := m["active_media_reference_key"].(string)
	if activeKey == "" {
		activeKey = DefaultMediaKey
	}
	metadata := decodeSonicMetadata(m)
	effects := decodeSonicEffects(m)
	markers := decodeSonicMarkers(m)

	// Decode media references
	mediaRefs := make(map[string]MediaReference)
	if refs, ok := m["media_references"].(map[string]any); ok {
		for key, refAny := range refs {
			if refMap, ok := refAny.(map[string]any); ok {
				if ref := decodeSonicMediaReference(refMap); ref != nil {
					mediaRefs[key] = ref
				}
			}
		}
	}

	var initialRef MediaReference
	if ref, ok := mediaRefs[activeKey]; ok {
		initialRef = ref
	}

	clip := NewClip(name, initialRef, sourceRange, metadata, effects, markers, activeKey, color)

	if len(mediaRefs) > 1 {
		clip.SetMediaReferences(mediaRefs, activeKey)
	}

	clip.SetEnabled(enabled)
	return clip, nil
}

// decodeSonicGap decodes a Gap from a sonic-parsed map.
func decodeSonicGap(m map[string]any) *Gap {
	name, _ := m["name"].(string)
	enabled, _ := m["enabled"].(bool)
	sourceRange := decodeSonicTimeRange(m["source_range"])
	metadata := decodeSonicMetadata(m)

	gap := NewGap(name, sourceRange, metadata, nil, nil, nil)
	gap.SetEnabled(enabled)
	return gap
}

// decodeSonicTransition decodes a Transition from a sonic-parsed map.
func decodeSonicTransition(m map[string]any) *Transition {
	name, _ := m["name"].(string)
	transitionType, _ := m["transition_type"].(string)
	metadata := decodeSonicMetadata(m)

	var inOffset, outOffset opentime.RationalTime
	if rt := decodeSonicRationalTime(m["in_offset"]); rt != nil {
		inOffset = *rt
	}
	if rt := decodeSonicRationalTime(m["out_offset"]); rt != nil {
		outOffset = *rt
	}

	return NewTransition(name, TransitionType(transitionType), inOffset, outOffset, metadata)
}

// decodeSonicMediaReference decodes a MediaReference from a sonic-parsed map.
func decodeSonicMediaReference(m map[string]any) MediaReference {
	schema, _ := m["OTIO_SCHEMA"].(string)
	name, _ := m["name"].(string)
	metadata := decodeSonicMetadata(m)
	availRange := decodeSonicTimeRange(m["available_range"])

	switch schema {
	case "ExternalReference.1":
		targetURL, _ := m["target_url"].(string)
		return NewExternalReference(name, targetURL, availRange, metadata)
	case "MissingReference.1":
		return NewMissingReference(name, availRange, metadata)
	case "GeneratorReference.1":
		generatorKind, _ := m["generator_kind"].(string)
		var parameters AnyDictionary
		if p, ok := m["parameters"].(map[string]any); ok {
			parameters = p
		}
		return NewGeneratorReference(name, generatorKind, parameters, availRange, metadata)
	}
	return nil
}

// decodeSonicEffect decodes an Effect from a sonic-parsed map.
func decodeSonicEffect(m map[string]any) Effect {
	schema, _ := m["OTIO_SCHEMA"].(string)
	name, _ := m["name"].(string)
	effectName, _ := m["effect_name"].(string)
	metadata := decodeSonicMetadata(m)

	switch schema {
	case "Effect.1":
		return NewEffect(name, effectName, metadata)
	case "LinearTimeWarp.1":
		timeScalar, _ := m["time_scalar"].(float64)
		return NewLinearTimeWarp(name, effectName, timeScalar, metadata)
	case "FreezeFrame.1":
		return NewFreezeFrame(name, metadata)
	}
	return nil
}

// decodeSonicExternalReference decodes an ExternalReference for top-level decoding.
func decodeSonicExternalReference(m map[string]any) *ExternalReference {
	name, _ := m["name"].(string)
	targetURL, _ := m["target_url"].(string)
	metadata := decodeSonicMetadata(m)
	availRange := decodeSonicTimeRange(m["available_range"])
	return NewExternalReference(name, targetURL, availRange, metadata)
}

// decodeSonicMissingReference decodes a MissingReference for top-level decoding.
func decodeSonicMissingReference(m map[string]any) *MissingReference {
	name, _ := m["name"].(string)
	metadata := decodeSonicMetadata(m)
	availRange := decodeSonicTimeRange(m["available_range"])
	return NewMissingReference(name, availRange, metadata)
}

// decodeSonicGeneratorReference decodes a GeneratorReference for top-level decoding.
func decodeSonicGeneratorReference(m map[string]any) *GeneratorReference {
	name, _ := m["name"].(string)
	generatorKind, _ := m["generator_kind"].(string)
	metadata := decodeSonicMetadata(m)
	var parameters AnyDictionary
	if p, ok := m["parameters"].(map[string]any); ok {
		parameters = p
	}
	availRange := decodeSonicTimeRange(m["available_range"])
	return NewGeneratorReference(name, generatorKind, parameters, availRange, metadata)
}

// decodeSonicEffectImpl decodes an Effect for top-level decoding.
func decodeSonicEffectImpl(m map[string]any) *EffectImpl {
	name, _ := m["name"].(string)
	effectName, _ := m["effect_name"].(string)
	metadata := decodeSonicMetadata(m)
	return NewEffect(name, effectName, metadata)
}

// decodeSonicLinearTimeWarp decodes a LinearTimeWarp for top-level decoding.
func decodeSonicLinearTimeWarp(m map[string]any) *LinearTimeWarp {
	name, _ := m["name"].(string)
	effectName, _ := m["effect_name"].(string)
	timeScalar, _ := m["time_scalar"].(float64)
	metadata := decodeSonicMetadata(m)
	return NewLinearTimeWarp(name, effectName, timeScalar, metadata)
}

// decodeSonicFreezeFrame decodes a FreezeFrame for top-level decoding.
func decodeSonicFreezeFrame(m map[string]any) *FreezeFrame {
	name, _ := m["name"].(string)
	metadata := decodeSonicMetadata(m)
	return NewFreezeFrame(name, metadata)
}

// decodeSonicTimeEffect decodes a TimeEffect for top-level decoding.
func decodeSonicTimeEffect(m map[string]any) *TimeEffectImpl {
	name, _ := m["name"].(string)
	effectName, _ := m["effect_name"].(string)
	metadata := decodeSonicMetadata(m)
	return NewTimeEffect(name, effectName, metadata)
}

// decodeSonicImageSequenceReference decodes an ImageSequenceReference.
func decodeSonicImageSequenceReference(m map[string]any) *ImageSequenceReference {
	name, _ := m["name"].(string)
	targetURLBase, _ := m["target_url_base"].(string)
	namePrefix, _ := m["name_prefix"].(string)
	nameSuffix, _ := m["name_suffix"].(string)
	startFrame, _ := m["start_frame"].(float64)
	frameStep, _ := m["frame_step"].(float64)
	rate, _ := m["rate"].(float64)
	frameZeroPadding, _ := m["frame_zero_padding"].(float64)
	missingFramePolicy, _ := m["missing_frame_policy"].(string)
	metadata := decodeSonicMetadata(m)
	availRange := decodeSonicTimeRange(m["available_range"])

	return NewImageSequenceReference(
		name,
		targetURLBase,
		namePrefix,
		nameSuffix,
		int(startFrame),
		int(frameStep),
		rate,
		int(frameZeroPadding),
		availRange,
		metadata,
		MissingFramePolicy(missingFramePolicy),
	)
}

// decodeSonicUnknownSchema handles unknown schema types for forward compatibility.
func decodeSonicUnknownSchema(schema string, m map[string]any) *UnknownSchema {
	return &UnknownSchema{
		schema: schema,
		data:   m,
	}
}

// decodeSonicMarker decodes a Marker from a sonic-parsed map.
func decodeSonicMarker(m map[string]any) *Marker {
	name, _ := m["name"].(string)
	comment, _ := m["comment"].(string)
	color, _ := m["color"].(string)
	metadata := decodeSonicMetadata(m)

	var markedRange opentime.TimeRange
	if tr := decodeSonicTimeRange(m["marked_range"]); tr != nil {
		markedRange = *tr
	}

	return NewMarker(name, markedRange, MarkerColor(color), comment, metadata)
}

// decodeSonicTimeRange decodes a TimeRange from a sonic-parsed map.
func decodeSonicTimeRange(v any) *opentime.TimeRange {
	m, ok := v.(map[string]any)
	if !ok || m == nil {
		return nil
	}

	st, ok := m["start_time"].(map[string]any)
	if !ok {
		return nil
	}
	stValue, _ := st["value"].(float64)
	stRate, _ := st["rate"].(float64)

	dur, ok := m["duration"].(map[string]any)
	if !ok {
		return nil
	}
	durValue, _ := dur["value"].(float64)
	durRate, _ := dur["rate"].(float64)

	tr := opentime.NewTimeRange(
		opentime.NewRationalTime(stValue, stRate),
		opentime.NewRationalTime(durValue, durRate),
	)
	return &tr
}

// decodeSonicColor decodes a Color from a sonic-parsed map.
func decodeSonicColor(v any) *Color {
	m, ok := v.(map[string]any)
	if !ok || m == nil {
		return nil
	}
	r, _ := m["r"].(float64)
	g, _ := m["g"].(float64)
	b, _ := m["b"].(float64)
	a, _ := m["a"].(float64)
	return &Color{R: r, G: g, B: b, A: a}
}

// FromJSONBytesSonic parses JSON using sonic for high performance.
func FromJSONBytesSonic(data []byte) (SerializableObject, error) {
	// Sanitize non-standard JSON values (Inf, NaN) from Python
	data = SanitizeJSON(data)

	var m map[string]any
	if err := sonic.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("sonic unmarshal: %w", err)
	}

	return decodeSonicObject(m)
}

// decodeSonicObject decodes a map into a SerializableObject based on schema.
func decodeSonicObject(m map[string]any) (SerializableObject, error) {
	schema, _ := m["OTIO_SCHEMA"].(string)

	switch schema {
	// Container types
	case "Timeline.1":
		return decodeSonicTimeline(m)
	case "Stack.1":
		return decodeSonicStack(m)
	case "Track.1", "Sequence.1": // Sequence is legacy name for Track
		return decodeSonicTrack(m)
	case "Clip.2":
		return decodeSonicClip(m)
	case "SerializableCollection.1":
		return decodeSonicSerializableCollection(m)
	case "Gap.1":
		return decodeSonicGap(m), nil
	case "Transition.1":
		return decodeSonicTransition(m), nil

	// Leaf types
	case "Marker.2":
		return decodeSonicMarker(m), nil
	case "ExternalReference.1":
		return decodeSonicExternalReference(m), nil
	case "MissingReference.1":
		return decodeSonicMissingReference(m), nil
	case "GeneratorReference.1":
		return decodeSonicGeneratorReference(m), nil
	case "Effect.1":
		return decodeSonicEffectImpl(m), nil
	case "LinearTimeWarp.1":
		return decodeSonicLinearTimeWarp(m), nil
	case "FreezeFrame.1":
		return decodeSonicFreezeFrame(m), nil
	case "TimeEffect.1":
		return decodeSonicTimeEffect(m), nil
	case "ImageSequenceReference.1":
		return decodeSonicImageSequenceReference(m), nil

	default:
		// Handle unknown schemas for forward compatibility
		return decodeSonicUnknownSchema(schema, m), nil
	}
}

// decodeSonicSerializableCollection decodes a SerializableCollection.
func decodeSonicSerializableCollection(m map[string]any) (*SerializableCollection, error) {
	name, _ := m["name"].(string)
	metadata := decodeSonicMetadata(m)

	var children []SerializableObject
	if childs, ok := m["children"].([]any); ok {
		for _, childAny := range childs {
			if childMap, ok := childAny.(map[string]any); ok {
				if child, err := decodeSonicObject(childMap); err == nil {
					children = append(children, child)
				}
			}
		}
	}

	return NewSerializableCollection(name, children, metadata), nil
}
