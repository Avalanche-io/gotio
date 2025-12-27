// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"encoding/json"
	"fmt"

	"github.com/mrjoshuak/gotio/opentime"
)

// MissingFramePolicy defines how to handle missing frames.
type MissingFramePolicy string

const (
	// MissingFramePolicyError raises an error on missing frames.
	MissingFramePolicyError MissingFramePolicy = "error"
	// MissingFramePolicyHold holds the last frame.
	MissingFramePolicyHold MissingFramePolicy = "hold"
	// MissingFramePolicyBlack shows black for missing frames.
	MissingFramePolicyBlack MissingFramePolicy = "black"
)

// ImageSequenceReferenceSchema is the schema for ImageSequenceReference.
var ImageSequenceReferenceSchema = Schema{Name: "ImageSequenceReference", Version: 1}

// ImageSequenceReference represents a reference to a sequence of image files.
type ImageSequenceReference struct {
	MediaReferenceBase
	targetURLBase      string
	namePrefix         string
	nameSuffix         string
	startFrame         int
	frameStep          int
	rate               float64
	frameZeroPadding   int
	missingFramePolicy MissingFramePolicy
}

// NewImageSequenceReference creates a new ImageSequenceReference.
func NewImageSequenceReference(
	name string,
	targetURLBase string,
	namePrefix string,
	nameSuffix string,
	startFrame int,
	frameStep int,
	rate float64,
	frameZeroPadding int,
	availableRange *opentime.TimeRange,
	metadata AnyDictionary,
	missingFramePolicy MissingFramePolicy,
) *ImageSequenceReference {
	if frameStep == 0 {
		frameStep = 1
	}
	if missingFramePolicy == "" {
		missingFramePolicy = MissingFramePolicyError
	}
	return &ImageSequenceReference{
		MediaReferenceBase: NewMediaReferenceBase(name, availableRange, metadata, nil),
		targetURLBase:      targetURLBase,
		namePrefix:         namePrefix,
		nameSuffix:         nameSuffix,
		startFrame:         startFrame,
		frameStep:          frameStep,
		rate:               rate,
		frameZeroPadding:   frameZeroPadding,
		missingFramePolicy: missingFramePolicy,
	}
}

// TargetURLBase returns the base URL for the sequence.
func (i *ImageSequenceReference) TargetURLBase() string {
	return i.targetURLBase
}

// SetTargetURLBase sets the base URL.
func (i *ImageSequenceReference) SetTargetURLBase(url string) {
	i.targetURLBase = url
}

// NamePrefix returns the filename prefix.
func (i *ImageSequenceReference) NamePrefix() string {
	return i.namePrefix
}

// SetNamePrefix sets the filename prefix.
func (i *ImageSequenceReference) SetNamePrefix(prefix string) {
	i.namePrefix = prefix
}

// NameSuffix returns the filename suffix (extension).
func (i *ImageSequenceReference) NameSuffix() string {
	return i.nameSuffix
}

// SetNameSuffix sets the filename suffix.
func (i *ImageSequenceReference) SetNameSuffix(suffix string) {
	i.nameSuffix = suffix
}

// StartFrame returns the starting frame number.
func (i *ImageSequenceReference) StartFrame() int {
	return i.startFrame
}

// SetStartFrame sets the starting frame number.
func (i *ImageSequenceReference) SetStartFrame(frame int) {
	i.startFrame = frame
}

// FrameStep returns the step between frames.
func (i *ImageSequenceReference) FrameStep() int {
	return i.frameStep
}

// SetFrameStep sets the step between frames.
func (i *ImageSequenceReference) SetFrameStep(step int) {
	i.frameStep = step
}

// Rate returns the frame rate.
func (i *ImageSequenceReference) Rate() float64 {
	return i.rate
}

// SetRate sets the frame rate.
func (i *ImageSequenceReference) SetRate(rate float64) {
	i.rate = rate
}

// FrameZeroPadding returns the zero padding width.
func (i *ImageSequenceReference) FrameZeroPadding() int {
	return i.frameZeroPadding
}

// SetFrameZeroPadding sets the zero padding width.
func (i *ImageSequenceReference) SetFrameZeroPadding(padding int) {
	i.frameZeroPadding = padding
}

// MissingFramePolicy returns the missing frame policy.
func (i *ImageSequenceReference) MissingFramePolicy() MissingFramePolicy {
	return i.missingFramePolicy
}

// SetMissingFramePolicy sets the missing frame policy.
func (i *ImageSequenceReference) SetMissingFramePolicy(policy MissingFramePolicy) {
	i.missingFramePolicy = policy
}

// TargetURLForImageNumber returns the URL for a specific frame number.
func (i *ImageSequenceReference) TargetURLForImageNumber(frameNumber int) string {
	format := fmt.Sprintf("%%s%%s%%0%dd%%s", i.frameZeroPadding)
	return fmt.Sprintf(format, i.targetURLBase, i.namePrefix, frameNumber, i.nameSuffix)
}

// TargetURLForFrame is an alias for TargetURLForImageNumber.
func (i *ImageSequenceReference) TargetURLForFrame(frameNumber int) string {
	return i.TargetURLForImageNumber(frameNumber)
}

// FrameForTime converts a RationalTime to a frame number.
func (i *ImageSequenceReference) FrameForTime(time opentime.RationalTime) int {
	// Convert time to frames at the sequence rate
	frameIndex := int(time.Value())
	return i.startFrame + frameIndex*i.frameStep
}

// EndFrame returns the ending frame number based on available range.
func (i *ImageSequenceReference) EndFrame() int {
	if i.availableRange == nil {
		return i.startFrame
	}
	dur := i.availableRange.Duration()
	frames := int(dur.Value() * dur.Rate() / i.rate)
	return i.startFrame + (frames-1)*i.frameStep
}

// NumberOfImagesInSequence returns the number of images in the sequence.
func (i *ImageSequenceReference) NumberOfImagesInSequence() int {
	if i.availableRange == nil {
		return 0
	}
	dur := i.availableRange.Duration()
	return int(dur.Value() * dur.Rate() / i.rate)
}

// SchemaName returns the schema name.
func (i *ImageSequenceReference) SchemaName() string {
	return ImageSequenceReferenceSchema.Name
}

// SchemaVersion returns the schema version.
func (i *ImageSequenceReference) SchemaVersion() int {
	return ImageSequenceReferenceSchema.Version
}

// Clone creates a deep copy.
func (i *ImageSequenceReference) Clone() SerializableObject {
	return &ImageSequenceReference{
		MediaReferenceBase: MediaReferenceBase{
			SerializableObjectWithMetadataBase: SerializableObjectWithMetadataBase{
				name:     i.name,
				metadata: CloneAnyDictionary(i.metadata),
			},
			availableRange:       cloneAvailableRange(i.availableRange),
			availableImageBounds: cloneBox2d(i.availableImageBounds),
		},
		targetURLBase:      i.targetURLBase,
		namePrefix:         i.namePrefix,
		nameSuffix:         i.nameSuffix,
		startFrame:         i.startFrame,
		frameStep:          i.frameStep,
		rate:               i.rate,
		frameZeroPadding:   i.frameZeroPadding,
		missingFramePolicy: i.missingFramePolicy,
	}
}

// IsEquivalentTo returns true if equivalent.
func (i *ImageSequenceReference) IsEquivalentTo(other SerializableObject) bool {
	otherI, ok := other.(*ImageSequenceReference)
	if !ok {
		return false
	}
	return i.name == otherI.name &&
		i.targetURLBase == otherI.targetURLBase &&
		i.namePrefix == otherI.namePrefix &&
		i.nameSuffix == otherI.nameSuffix &&
		i.startFrame == otherI.startFrame &&
		i.frameStep == otherI.frameStep &&
		i.rate == otherI.rate
}

// imageSequenceReferenceJSON is the JSON representation.
type imageSequenceReferenceJSON struct {
	Schema               string              `json:"OTIO_SCHEMA"`
	Name                 string              `json:"name"`
	Metadata             AnyDictionary       `json:"metadata"`
	AvailableRange       *opentime.TimeRange `json:"available_range"`
	AvailableImageBounds *Box2d              `json:"available_image_bounds"`
	TargetURLBase        string              `json:"target_url_base"`
	NamePrefix           string              `json:"name_prefix"`
	NameSuffix           string              `json:"name_suffix"`
	StartFrame           int                 `json:"start_frame"`
	FrameStep            int                 `json:"frame_step"`
	Rate                 float64             `json:"rate"`
	FrameZeroPadding     int                 `json:"frame_zero_padding"`
	MissingFramePolicy   MissingFramePolicy  `json:"missing_frame_policy"`
}

// MarshalJSON implements json.Marshaler.
func (i *ImageSequenceReference) MarshalJSON() ([]byte, error) {
	return json.Marshal(&imageSequenceReferenceJSON{
		Schema:               ImageSequenceReferenceSchema.String(),
		Name:                 i.name,
		Metadata:             i.metadata,
		AvailableRange:       i.availableRange,
		AvailableImageBounds: i.availableImageBounds,
		TargetURLBase:        i.targetURLBase,
		NamePrefix:           i.namePrefix,
		NameSuffix:           i.nameSuffix,
		StartFrame:           i.startFrame,
		FrameStep:            i.frameStep,
		Rate:                 i.rate,
		FrameZeroPadding:     i.frameZeroPadding,
		MissingFramePolicy:   i.missingFramePolicy,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (i *ImageSequenceReference) UnmarshalJSON(data []byte) error {
	var j imageSequenceReferenceJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	i.name = j.Name
	i.metadata = j.Metadata
	if i.metadata == nil {
		i.metadata = make(AnyDictionary)
	}
	i.availableRange = j.AvailableRange
	i.availableImageBounds = j.AvailableImageBounds
	i.targetURLBase = j.TargetURLBase
	i.namePrefix = j.NamePrefix
	i.nameSuffix = j.NameSuffix
	i.startFrame = j.StartFrame
	i.frameStep = j.FrameStep
	if i.frameStep == 0 {
		i.frameStep = 1
	}
	i.rate = j.Rate
	i.frameZeroPadding = j.FrameZeroPadding
	i.missingFramePolicy = j.MissingFramePolicy
	if i.missingFramePolicy == "" {
		i.missingFramePolicy = MissingFramePolicyError
	}
	return nil
}

func init() {
	RegisterSchema(ImageSequenceReferenceSchema, func() SerializableObject {
		return NewImageSequenceReference("", "", "", "", 0, 1, 24, 4, nil, nil, "")
	})
}
