// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"encoding/json"
	"testing"

	"github.com/Avalanche-io/gotio/opentime"
)

func TestExternalReferenceSetters(t *testing.T) {
	ref := NewExternalReference("", "/path/to/file.mov", nil, nil)

	// Test SetTargetURL
	ref.SetTargetURL("/new/path.mov")
	if ref.TargetURL() != "/new/path.mov" {
		t.Errorf("SetTargetURL: got %s, want /new/path.mov", ref.TargetURL())
	}

	// Test SchemaName
	if ref.SchemaName() != "ExternalReference" {
		t.Errorf("SchemaName = %s, want ExternalReference", ref.SchemaName())
	}

	// Test SchemaVersion
	if ref.SchemaVersion() != 1 {
		t.Errorf("SchemaVersion = %d, want 1", ref.SchemaVersion())
	}
}

func TestExternalReferenceIsEquivalentTo(t *testing.T) {
	ref1 := NewExternalReference("", "/path/to/file.mov", nil, nil)
	ref2 := NewExternalReference("", "/path/to/file.mov", nil, nil)
	ref3 := NewExternalReference("", "/different/file.mov", nil, nil)

	if !ref1.IsEquivalentTo(ref2) {
		t.Error("Identical external references should be equivalent")
	}
	if ref1.IsEquivalentTo(ref3) {
		t.Error("Different external references should not be equivalent")
	}

	// Test with non-ExternalReference
	missing := NewMissingReference("", nil, nil)
	if ref1.IsEquivalentTo(missing) {
		t.Error("ExternalReference should not be equivalent to MissingReference")
	}
}

func TestExternalReferenceJSON(t *testing.T) {
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	ref := NewExternalReference("", "/path/to/file.mov", &ar, AnyDictionary{"author": "test"})

	data, err := json.Marshal(ref)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	ref2 := &ExternalReference{}
	if err := json.Unmarshal(data, ref2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if ref2.TargetURL() != "/path/to/file.mov" {
		t.Errorf("URL mismatch: got %s", ref2.TargetURL())
	}
	if ref2.AvailableRange() == nil {
		t.Error("AvailableRange should not be nil")
	}
}

func TestMissingReferenceSchema(t *testing.T) {
	ref := NewMissingReference("", nil, nil)

	if ref.SchemaName() != "MissingReference" {
		t.Errorf("SchemaName = %s, want MissingReference", ref.SchemaName())
	}
	if ref.SchemaVersion() != 1 {
		t.Errorf("SchemaVersion = %d, want 1", ref.SchemaVersion())
	}
}

func TestMissingReferenceClone(t *testing.T) {
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(50, 24))
	ref := NewMissingReference("missing_ref", &ar, AnyDictionary{"key": "value"})

	clone := ref.Clone().(*MissingReference)

	if clone.Name() != "missing_ref" {
		t.Errorf("Clone name = %s, want missing_ref", clone.Name())
	}
	if !clone.IsMissingReference() {
		t.Error("Clone should be missing reference")
	}
	if clone.AvailableRange() == nil {
		t.Error("Clone available range should not be nil")
	}
}

func TestMissingReferenceJSON(t *testing.T) {
	ref := NewMissingReference("test_missing", nil, nil)

	data, err := json.Marshal(ref)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	ref2 := &MissingReference{}
	if err := json.Unmarshal(data, ref2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if ref2.Name() != "test_missing" {
		t.Errorf("Name mismatch: got %s", ref2.Name())
	}
}

func TestGeneratorReferenceSetters(t *testing.T) {
	ref := NewGeneratorReference("gen", "TestGenerator", nil, nil, nil)

	// Test SetGeneratorKind
	ref.SetGeneratorKind("NewGenerator")
	if ref.GeneratorKind() != "NewGenerator" {
		t.Errorf("SetGeneratorKind: got %s, want NewGenerator", ref.GeneratorKind())
	}

	// Test SetParameters
	params := AnyDictionary{"param1": 100, "param2": "test"}
	ref.SetParameters(params)
	if ref.Parameters()["param1"] != 100 {
		t.Error("SetParameters failed")
	}
}

func TestGeneratorReferenceSchema(t *testing.T) {
	ref := NewGeneratorReference("", "", nil, nil, nil)

	if ref.SchemaName() != "GeneratorReference" {
		t.Errorf("SchemaName = %s, want GeneratorReference", ref.SchemaName())
	}
	if ref.SchemaVersion() != 1 {
		t.Errorf("SchemaVersion = %d, want 1", ref.SchemaVersion())
	}
}

func TestGeneratorReferenceClone(t *testing.T) {
	params := AnyDictionary{"color": "red"}
	ref := NewGeneratorReference("gen", "ColorBars", params, nil, nil)

	clone := ref.Clone().(*GeneratorReference)

	if clone.Name() != "gen" {
		t.Errorf("Clone name = %s, want gen", clone.Name())
	}
	if clone.GeneratorKind() != "ColorBars" {
		t.Errorf("Clone generator kind = %s, want ColorBars", clone.GeneratorKind())
	}
	if clone.Parameters()["color"] != "red" {
		t.Error("Clone parameters should match")
	}
}

func TestGeneratorReferenceIsEquivalentTo(t *testing.T) {
	ref1 := NewGeneratorReference("gen", "ColorBars", nil, nil, nil)
	ref2 := NewGeneratorReference("gen", "ColorBars", nil, nil, nil)
	ref3 := NewGeneratorReference("gen", "BlackBars", nil, nil, nil)

	if !ref1.IsEquivalentTo(ref2) {
		t.Error("Identical generator references should be equivalent")
	}
	if ref1.IsEquivalentTo(ref3) {
		t.Error("Different generator references should not be equivalent")
	}
}

func TestGeneratorReferenceJSON(t *testing.T) {
	params := AnyDictionary{"width": 1920, "height": 1080}
	ref := NewGeneratorReference("bars", "ColorBars", params, nil, nil)

	data, err := json.Marshal(ref)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	ref2 := &GeneratorReference{}
	if err := json.Unmarshal(data, ref2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if ref2.GeneratorKind() != "ColorBars" {
		t.Errorf("GeneratorKind mismatch: got %s", ref2.GeneratorKind())
	}
}

func TestImageSequenceReferenceComplete(t *testing.T) {
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	// NewImageSequenceReference(name, targetURLBase, namePrefix, nameSuffix, startFrame, frameStep, rate, frameZeroPadding, availableRange, metadata, missingFramePolicy)
	ref := NewImageSequenceReference(
		"",
		"/path/to/images/",
		"frame_",
		".exr",
		1001,
		1,
		24,
		4,
		&ar,
		nil,
		MissingFramePolicyError,
	)

	// Test getters
	if ref.TargetURLBase() != "/path/to/images/" {
		t.Errorf("TargetURLBase = %s", ref.TargetURLBase())
	}
	if ref.NamePrefix() != "frame_" {
		t.Errorf("NamePrefix = %s", ref.NamePrefix())
	}
	if ref.NameSuffix() != ".exr" {
		t.Errorf("NameSuffix = %s", ref.NameSuffix())
	}
	if ref.StartFrame() != 1001 {
		t.Errorf("StartFrame = %d", ref.StartFrame())
	}
	if ref.FrameStep() != 1 {
		t.Errorf("FrameStep = %d", ref.FrameStep())
	}
	if ref.Rate() != 24 {
		t.Errorf("Rate = %f", ref.Rate())
	}
	if ref.FrameZeroPadding() != 4 {
		t.Errorf("FrameZeroPadding = %d", ref.FrameZeroPadding())
	}
	if ref.MissingFramePolicy() != MissingFramePolicyError {
		t.Errorf("MissingFramePolicy = %s", ref.MissingFramePolicy())
	}

	// Test setters
	ref.SetTargetURLBase("/new/path/")
	ref.SetNamePrefix("img_")
	ref.SetNameSuffix(".png")
	ref.SetStartFrame(1)
	ref.SetFrameStep(2)
	ref.SetRate(30)
	ref.SetFrameZeroPadding(5)
	ref.SetMissingFramePolicy(MissingFramePolicyHold)

	if ref.TargetURLBase() != "/new/path/" {
		t.Error("SetTargetURLBase failed")
	}
	if ref.NamePrefix() != "img_" {
		t.Error("SetNamePrefix failed")
	}
	if ref.NameSuffix() != ".png" {
		t.Error("SetNameSuffix failed")
	}
	if ref.StartFrame() != 1 {
		t.Error("SetStartFrame failed")
	}
	if ref.FrameStep() != 2 {
		t.Error("SetFrameStep failed")
	}
	if ref.Rate() != 30 {
		t.Error("SetRate failed")
	}
	if ref.FrameZeroPadding() != 5 {
		t.Error("SetFrameZeroPadding failed")
	}
	if ref.MissingFramePolicy() != MissingFramePolicyHold {
		t.Error("SetMissingFramePolicy failed")
	}
}

func TestImageSequenceReferenceFrameNumber(t *testing.T) {
	// NewImageSequenceReference(name, targetURLBase, namePrefix, nameSuffix, startFrame, frameStep, rate, frameZeroPadding, availableRange, metadata, missingFramePolicy)
	ref := NewImageSequenceReference(
		"",
		"/path/",
		"frame_",
		".exr",
		1001,
		1,
		24,
		4,
		nil,
		nil,
		MissingFramePolicyError,
	)

	// Test FrameForTime
	time := opentime.NewRationalTime(10, 24)
	frame := ref.FrameForTime(time)
	expected := 1001 + 10 // Start frame + time value
	if frame != expected {
		t.Errorf("FrameForTime(%v) = %d, want %d", time, frame, expected)
	}

	// Test TargetURLForFrame
	url := ref.TargetURLForFrame(1001)
	expectedURL := "/path/frame_1001.exr"
	if url != expectedURL {
		t.Errorf("TargetURLForFrame(1001) = %s, want %s", url, expectedURL)
	}

	// Test with zero padding
	url = ref.TargetURLForFrame(1)
	expectedURL = "/path/frame_0001.exr"
	if url != expectedURL {
		t.Errorf("TargetURLForFrame(1) with padding = %s, want %s", url, expectedURL)
	}

	// Test EndFrame
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	ref2 := NewImageSequenceReference("", "/path/", "f_", ".exr", 1, 1, 24, 4, &ar, nil, MissingFramePolicyError)
	endFrame := ref2.EndFrame()
	if endFrame != 100 { // 1 + 100 - 1
		t.Errorf("EndFrame = %d, want 100", endFrame)
	}

	// Test NumberOfImagesInSequence
	numImages := ref2.NumberOfImagesInSequence()
	if numImages != 100 {
		t.Errorf("NumberOfImagesInSequence = %d, want 100", numImages)
	}
}

func TestImageSequenceReferenceClone(t *testing.T) {
	ref := NewImageSequenceReference("", "/path/", "f_", ".exr", 1001, 1, 24, 4, nil, nil, MissingFramePolicyError)

	clone := ref.Clone().(*ImageSequenceReference)

	if clone.TargetURLBase() != "/path/" {
		t.Error("Clone TargetURLBase mismatch")
	}
	if clone.StartFrame() != 1001 {
		t.Error("Clone StartFrame mismatch")
	}
}

func TestImageSequenceReferenceSchema(t *testing.T) {
	ref := NewImageSequenceReference("", "", "", "", 0, 1, 24, 0, nil, nil, MissingFramePolicyError)

	if ref.SchemaName() != "ImageSequenceReference" {
		t.Errorf("SchemaName = %s, want ImageSequenceReference", ref.SchemaName())
	}
	if ref.SchemaVersion() != 1 {
		t.Errorf("SchemaVersion = %d, want 1", ref.SchemaVersion())
	}
}

func TestImageSequenceReferenceJSON(t *testing.T) {
	ref := NewImageSequenceReference("", "/images/", "frame_", ".exr", 1001, 1, 24, 4, nil, nil, MissingFramePolicyHold)

	data, err := json.Marshal(ref)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	ref2 := &ImageSequenceReference{}
	if err := json.Unmarshal(data, ref2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if ref2.StartFrame() != 1001 {
		t.Errorf("StartFrame mismatch: got %d", ref2.StartFrame())
	}
	if ref2.MissingFramePolicy() != MissingFramePolicyHold {
		t.Errorf("MissingFramePolicy mismatch: got %s", ref2.MissingFramePolicy())
	}
}

func TestMediaReferenceBaseSetters(t *testing.T) {
	ref := NewExternalReference("", "/path", nil, nil)

	// Test SetAvailableRange
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	ref.SetAvailableRange(&ar)
	if ref.AvailableRange() == nil {
		t.Error("SetAvailableRange failed")
	}

	// Set to nil
	ref.SetAvailableRange(nil)
	if ref.AvailableRange() != nil {
		t.Error("SetAvailableRange to nil failed")
	}
}
