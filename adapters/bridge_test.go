// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package adapters

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mrjoshuak/gotio/opentimelineio"
)

func TestDetectPython(t *testing.T) {
	path, err := detectPython()
	if err == ErrPythonNotAvailable {
		t.Skip("Python with opentimelineio not available")
	}
	if err != nil {
		t.Fatalf("detectPython error: %v", err)
	}
	if path == "" {
		t.Error("expected non-empty path")
	}
	t.Logf("Found Python at: %s", path)
}

func TestNewBridge(t *testing.T) {
	bridge, err := NewBridge(Config{})
	if err == ErrPythonNotAvailable {
		t.Skip("Python with opentimelineio not available")
	}
	if err != nil {
		t.Fatalf("NewBridge error: %v", err)
	}
	defer bridge.Close()

	// Check formats were discovered
	formats := bridge.AvailableFormats()
	if len(formats) == 0 {
		t.Error("expected at least one format")
	}

	t.Logf("Discovered %d Python adapters", len(formats))
	for _, f := range formats {
		t.Logf("  - %s: suffixes=%v read=%v write=%v",
			f.Name, f.Suffixes, f.CanRead, f.CanWrite)
	}
}

func TestBridgeReadWrite(t *testing.T) {
	bridge, err := NewBridge(Config{})
	if err == ErrPythonNotAvailable {
		t.Skip("Python with opentimelineio not available")
	}
	if err != nil {
		t.Fatalf("NewBridge error: %v", err)
	}
	defer bridge.Close()

	// Create a simple timeline
	timeline := opentimelineio.NewTimeline("Test Timeline", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	timeline.Tracks().AppendChild(track)

	// Write to temp file
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "test.otio")

	err = bridge.WriteWithFormat("otio_json", timeline, outPath)
	if err != nil {
		t.Fatalf("WriteWithFormat error: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Fatal("output file was not created")
	}

	// Read it back
	obj, err := bridge.ReadWithFormat("otio_json", outPath)
	if err != nil {
		t.Fatalf("ReadWithFormat error: %v", err)
	}

	readTimeline, ok := obj.(*opentimelineio.Timeline)
	if !ok {
		t.Fatalf("expected Timeline, got %T", obj)
	}

	if readTimeline.Name() != "Test Timeline" {
		t.Errorf("expected name 'Test Timeline', got %q", readTimeline.Name())
	}
}

func TestBridgeReadWriteAutoDetect(t *testing.T) {
	bridge, err := NewBridge(Config{})
	if err == ErrPythonNotAvailable {
		t.Skip("Python with opentimelineio not available")
	}
	if err != nil {
		t.Fatalf("NewBridge error: %v", err)
	}
	defer bridge.Close()

	// Create a simple timeline
	timeline := opentimelineio.NewTimeline("Auto Detect Test", nil, nil)
	track := opentimelineio.NewTrack("V1", nil, opentimelineio.TrackKindVideo, nil, nil)
	timeline.Tracks().AppendChild(track)

	// Write using auto-detect (based on file extension)
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "test.otio")

	err = bridge.Write(timeline, outPath)
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}

	// Read it back using auto-detect
	obj, err := bridge.Read(outPath)
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}

	readTimeline, ok := obj.(*opentimelineio.Timeline)
	if !ok {
		t.Fatalf("expected Timeline, got %T", obj)
	}

	if readTimeline.Name() != "Auto Detect Test" {
		t.Errorf("expected name 'Auto Detect Test', got %q", readTimeline.Name())
	}
}

func TestBridgeClose(t *testing.T) {
	bridge, err := NewBridge(Config{})
	if err == ErrPythonNotAvailable {
		t.Skip("Python with opentimelineio not available")
	}
	if err != nil {
		t.Fatalf("NewBridge error: %v", err)
	}

	// Close should work
	if err := bridge.Close(); err != nil {
		t.Errorf("Close error: %v", err)
	}

	// Double close should be safe
	if err := bridge.Close(); err != nil {
		t.Errorf("Double close error: %v", err)
	}

	// Operations after close should fail
	_, err = bridge.Read("/path/to/file.otio")
	if err == nil {
		t.Error("expected error after close")
	}
}

func TestBridgeFormatNotFound(t *testing.T) {
	bridge, err := NewBridge(Config{})
	if err == ErrPythonNotAvailable {
		t.Skip("Python with opentimelineio not available")
	}
	if err != nil {
		t.Fatalf("NewBridge error: %v", err)
	}
	defer bridge.Close()

	// Try to read with unsupported format
	_, err = bridge.ReadWithFormat("definitely_not_a_format", "/path/to/file.xyz")
	if err == nil {
		t.Error("expected error for unsupported format")
	}
	t.Logf("Got expected error: %v", err)
}

func TestBridgeWithCustomPythonPath(t *testing.T) {
	// First detect python to get a valid path
	pythonPath, err := detectPython()
	if err == ErrPythonNotAvailable {
		t.Skip("Python with opentimelineio not available")
	}
	if err != nil {
		t.Fatalf("detectPython error: %v", err)
	}

	// Create bridge with explicit path
	bridge, err := NewBridge(Config{PythonPath: pythonPath})
	if err != nil {
		t.Fatalf("NewBridge with custom path error: %v", err)
	}
	defer bridge.Close()

	formats := bridge.AvailableFormats()
	if len(formats) == 0 {
		t.Error("expected at least one format")
	}
}
