// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package gotio

import (
	"encoding/json"

	"github.com/Avalanche-io/gotio/opentime"
)

// DefaultMediaKey is the default key for the primary media reference.
const DefaultMediaKey = "DEFAULT_MEDIA"

// ClipSchema is the schema for Clip.
var ClipSchema = Schema{Name: "Clip", Version: 2}

// Clip is a segment of editable media (usually audio or video).
type Clip struct {
	ItemBase
	mediaReferences         map[string]MediaReference
	activeMediaReferenceKey string
}

// NewClip creates a new Clip.
func NewClip(
	name string,
	mediaReference MediaReference,
	sourceRange *opentime.TimeRange,
	metadata AnyDictionary,
	effects []Effect,
	markers []*Marker,
	activeMediaReferenceKey string,
	color *Color,
) *Clip {
	if activeMediaReferenceKey == "" {
		activeMediaReferenceKey = DefaultMediaKey
	}

	mediaReferences := make(map[string]MediaReference)
	if mediaReference != nil {
		mediaReferences[activeMediaReferenceKey] = mediaReference
	} else {
		mediaReferences[activeMediaReferenceKey] = NewMissingReference("", nil, nil)
	}

	clip := &Clip{
		ItemBase:                NewItemBase(name, sourceRange, metadata, effects, markers, true, color),
		mediaReferences:         mediaReferences,
		activeMediaReferenceKey: activeMediaReferenceKey,
	}
	clip.SetSelf(clip)
	return clip
}

// MediaReference returns the active media reference.
func (c *Clip) MediaReference() MediaReference {
	return c.mediaReferences[c.activeMediaReferenceKey]
}

// SetMediaReference sets the active media reference.
func (c *Clip) SetMediaReference(mediaReference MediaReference) {
	if mediaReference == nil {
		mediaReference = NewMissingReference("", nil, nil)
	}
	c.mediaReferences[c.activeMediaReferenceKey] = mediaReference
}

// MediaReferences returns all media references.
func (c *Clip) MediaReferences() map[string]MediaReference {
	return c.mediaReferences
}

// SetMediaReferences sets all media references and the active key.
func (c *Clip) SetMediaReferences(refs map[string]MediaReference, activeKey string) error {
	if _, ok := refs[activeKey]; !ok {
		return ErrMediaReferenceNotFound
	}
	c.mediaReferences = refs
	c.activeMediaReferenceKey = activeKey
	return nil
}

// ActiveMediaReferenceKey returns the active media reference key.
func (c *Clip) ActiveMediaReferenceKey() string {
	return c.activeMediaReferenceKey
}

// SetActiveMediaReferenceKey sets the active media reference key.
func (c *Clip) SetActiveMediaReferenceKey(key string) error {
	if _, ok := c.mediaReferences[key]; !ok {
		return ErrMediaReferenceNotFound
	}
	c.activeMediaReferenceKey = key
	return nil
}

// Duration returns the duration from source range or available range.
func (c *Clip) Duration() (opentime.RationalTime, error) {
	if c.sourceRange != nil {
		return c.sourceRange.Duration(), nil
	}
	ar, err := c.AvailableRange()
	if err != nil {
		return opentime.RationalTime{}, err
	}
	return ar.Duration(), nil
}

// AvailableRange returns the available range from the media reference.
func (c *Clip) AvailableRange() (opentime.TimeRange, error) {
	ref := c.MediaReference()
	if ref == nil {
		return opentime.TimeRange{}, ErrMissingReference
	}
	ar := ref.AvailableRange()
	if ar == nil {
		return opentime.TimeRange{}, ErrCannotComputeAvailableRange
	}
	return *ar, nil
}

// AvailableImageBounds returns the available image bounds from the media reference.
func (c *Clip) AvailableImageBounds() (*Box2d, error) {
	ref := c.MediaReference()
	if ref == nil {
		return nil, ErrMissingReference
	}
	return ref.AvailableImageBounds(), nil
}

// SchemaName returns the schema name.
func (c *Clip) SchemaName() string {
	return ClipSchema.Name
}

// SchemaVersion returns the schema version.
func (c *Clip) SchemaVersion() int {
	return ClipSchema.Version
}

// Clone creates a deep copy.
func (c *Clip) Clone() SerializableObject {
	refs := make(map[string]MediaReference)
	for k, v := range c.mediaReferences {
		refs[k] = v.Clone().(MediaReference)
	}

	clone := &Clip{
		ItemBase: ItemBase{
			ComposableBase: ComposableBase{
				SerializableObjectWithMetadataBase: SerializableObjectWithMetadataBase{
					name:     c.name,
					metadata: CloneAnyDictionary(c.metadata),
				},
			},
			sourceRange: cloneSourceRange(c.sourceRange),
			effects:     cloneEffects(c.effects),
			markers:     cloneMarkers(c.markers),
			enabled:     c.enabled,
			color:       cloneColor(c.color),
		},
		mediaReferences:         refs,
		activeMediaReferenceKey: c.activeMediaReferenceKey,
	}
	clone.SetSelf(clone)
	return clone
}

// IsEquivalentTo returns true if equivalent.
func (c *Clip) IsEquivalentTo(other SerializableObject) bool {
	otherC, ok := other.(*Clip)
	if !ok {
		return false
	}
	if c.name != otherC.name || c.activeMediaReferenceKey != otherC.activeMediaReferenceKey {
		return false
	}
	if len(c.mediaReferences) != len(otherC.mediaReferences) {
		return false
	}
	for k, v := range c.mediaReferences {
		otherV, ok := otherC.mediaReferences[k]
		if !ok || !v.IsEquivalentTo(otherV) {
			return false
		}
	}
	return true
}

// clipJSON is the JSON representation.
type clipJSON struct {
	Schema                  string                     `json:"OTIO_SCHEMA"`
	Name                    string                     `json:"name"`
	Metadata                AnyDictionary              `json:"metadata"`
	SourceRange             *opentime.TimeRange        `json:"source_range"`
	Effects                 []RawMessage          `json:"effects"`
	Markers                 []*Marker                  `json:"markers"`
	Enabled                 bool                       `json:"enabled"`
	Color                   *Color                     `json:"color"`
	MediaReferences         map[string]RawMessage `json:"media_references"`
	ActiveMediaReferenceKey string                     `json:"active_media_reference_key"`
}

// MarshalJSON implements json.Marshaler.
func (c *Clip) MarshalJSON() ([]byte, error) {
	effects := make([]RawMessage, len(c.effects))
	for i, e := range c.effects {
		data, err := json.Marshal(e)
		if err != nil {
			return nil, err
		}
		effects[i] = data
	}

	mediaRefs := make(map[string]RawMessage)
	for k, v := range c.mediaReferences {
		data, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		mediaRefs[k] = data
	}

	return json.Marshal(&clipJSON{
		Schema:                  ClipSchema.String(),
		Name:                    c.name,
		Metadata:                c.metadata,
		SourceRange:             c.sourceRange,
		Effects:                 effects,
		Markers:                 c.markers,
		Enabled:                 c.enabled,
		Color:                   c.color,
		MediaReferences:         mediaRefs,
		ActiveMediaReferenceKey: c.activeMediaReferenceKey,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *Clip) UnmarshalJSON(data []byte) error {
	var j clipJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}

	c.name = j.Name
	c.metadata = j.Metadata
	if c.metadata == nil {
		c.metadata = make(AnyDictionary)
	}
	c.sourceRange = j.SourceRange
	c.enabled = j.Enabled
	c.color = j.Color
	c.activeMediaReferenceKey = j.ActiveMediaReferenceKey
	if c.activeMediaReferenceKey == "" {
		c.activeMediaReferenceKey = DefaultMediaKey
	}

	// Unmarshal effects
	c.effects = make([]Effect, len(j.Effects))
	for i, data := range j.Effects {
		obj, err := FromJSONString(string(data))
		if err != nil {
			return err
		}
		effect, ok := obj.(Effect)
		if !ok {
			return &TypeMismatchError{Expected: "Effect", Got: obj.SchemaName()}
		}
		c.effects[i] = effect
	}

	// Copy markers
	c.markers = j.Markers
	if c.markers == nil {
		c.markers = make([]*Marker, 0)
	}

	// Unmarshal media references
	c.mediaReferences = make(map[string]MediaReference)
	for k, data := range j.MediaReferences {
		obj, err := FromJSONString(string(data))
		if err != nil {
			return err
		}
		ref, ok := obj.(MediaReference)
		if !ok {
			return &TypeMismatchError{Expected: "MediaReference", Got: obj.SchemaName()}
		}
		c.mediaReferences[k] = ref
	}

	c.SetSelf(c)
	return nil
}

func init() {
	RegisterSchema(ClipSchema, func() SerializableObject {
		return NewClip("", nil, nil, nil, nil, nil, "", nil)
	})
}
