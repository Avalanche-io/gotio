// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

// Package adapters provides access to Python OTIO adapters via a bridge.
//
// This package is a compatibility layer that allows Go code to use Python
// OpenTimelineIO adapters when native Go implementations don't exist.
//
// Basic usage:
//
//	bridge, err := adapters.NewBridge(adapters.Config{})
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer bridge.Close()
//
//	timeline, err := bridge.Read("project.edl")
package adapters

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/Avalanche-io/gotio/opentimelineio"
)

//go:embed bridge.py
var bridgeScript string

// Bridge provides access to Python OTIO adapters.
type Bridge struct {
	pythonPath string
	scriptPath string // temp file containing bridge.py
	cmd        *exec.Cmd
	stdin      io.WriteCloser
	stdout     *bufio.Reader
	stderr     bytes.Buffer
	mu         sync.Mutex // protects cmd, stdin, stdout
	requestID  atomic.Int64
	formats    []FormatInfo
	formatMap  map[string]FormatInfo // name and suffix -> info
	closed     bool
}

// Config configures the Python bridge.
type Config struct {
	PythonPath string // Path to Python interpreter (auto-detected if empty)
}

// FormatInfo describes an available file format adapter.
type FormatInfo struct {
	Name     string   // Adapter name (e.g., "otio_json", "cmx_3600")
	Suffixes []string // File extensions (e.g., [".otio", ".json"])
	CanRead  bool     // Adapter supports reading
	CanWrite bool     // Adapter supports writing
}

// request is a JSON-RPC style request to the Python bridge.
type request struct {
	ID     int64          `json:"id"`
	Method string         `json:"method"`
	Params map[string]any `json:"params"`
}

// response is a JSON-RPC style response from the Python bridge.
type response struct {
	ID     int64           `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *string         `json:"error"`
}

// Common errors.
var (
	ErrPythonNotAvailable = errors.New("Python with opentimelineio not available")
	ErrBridgeClosed       = errors.New("bridge is closed")
	ErrFormatNotFound     = errors.New("format not supported")
)

// NewBridge creates a new Python bridge.
// If cfg.PythonPath is empty, it auto-detects Python.
func NewBridge(cfg Config) (*Bridge, error) {
	pythonPath := cfg.PythonPath
	if pythonPath == "" {
		var err error
		pythonPath, err = detectPython()
		if err != nil {
			return nil, err
		}
	}

	bridge := &Bridge{
		pythonPath: pythonPath,
		formatMap:  make(map[string]FormatInfo),
	}

	if err := bridge.start(); err != nil {
		return nil, err
	}

	// Discover available formats
	if err := bridge.discover(); err != nil {
		bridge.Close()
		return nil, err
	}

	return bridge, nil
}

// detectPython finds a Python interpreter with opentimelineio installed.
func detectPython() (string, error) {
	candidates := []string{"python3", "python"}

	for _, name := range candidates {
		path, err := exec.LookPath(name)
		if err != nil {
			continue
		}

		// Verify opentimelineio is importable
		cmd := exec.Command(path, "-c", "import opentimelineio")
		if err := cmd.Run(); err == nil {
			return path, nil
		}
	}

	return "", ErrPythonNotAvailable
}

// start launches the Python subprocess.
func (b *Bridge) start() error {
	// Write bridge.py to a temp file
	tmpDir := os.TempDir()
	scriptPath := filepath.Join(tmpDir, "gotio_bridge.py")
	if err := os.WriteFile(scriptPath, []byte(bridgeScript), 0600); err != nil {
		return fmt.Errorf("failed to write bridge script: %w", err)
	}
	b.scriptPath = scriptPath

	b.cmd = exec.Command(b.pythonPath, "-u", scriptPath)
	b.cmd.Stderr = &b.stderr

	stdin, err := b.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	b.stdin = stdin

	stdout, err := b.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	b.stdout = bufio.NewReader(stdout)

	if err := b.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start Python: %w", err)
	}

	// Verify the bridge is responsive
	if err := b.ping(); err != nil {
		b.Close()
		return fmt.Errorf("bridge not responsive: %w", err)
	}

	return nil
}

// ping checks if the bridge is alive.
func (b *Bridge) ping() error {
	var result string
	if err := b.call("ping", nil, &result); err != nil {
		return err
	}
	if result != "pong" {
		return errors.New("unexpected ping response")
	}
	return nil
}

// discover fetches the list of available adapters from Python.
func (b *Bridge) discover() error {
	var adapters []struct {
		Name     string   `json:"name"`
		Suffixes []string `json:"suffixes"`
		Features struct {
			Read  bool `json:"read"`
			Write bool `json:"write"`
		} `json:"features"`
	}

	if err := b.call("discover", nil, &adapters); err != nil {
		return err
	}

	b.formats = make([]FormatInfo, 0, len(adapters))
	for _, a := range adapters {
		info := FormatInfo{
			Name:     a.Name,
			Suffixes: a.Suffixes,
			CanRead:  a.Features.Read,
			CanWrite: a.Features.Write,
		}
		b.formats = append(b.formats, info)
		b.formatMap[a.Name] = info

		// Also map by suffix
		for _, suffix := range a.Suffixes {
			s := strings.ToLower(suffix)
			if !strings.HasPrefix(s, ".") {
				s = "." + s
			}
			b.formatMap[s] = info
		}
	}

	return nil
}

// call makes a request to the Python bridge and waits for a response.
func (b *Bridge) call(method string, params map[string]any, result any) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return ErrBridgeClosed
	}

	if params == nil {
		params = make(map[string]any)
	}

	req := request{
		ID:     b.requestID.Add(1),
		Method: method,
		Params: params,
	}

	// Send request
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	if _, err := b.stdin.Write(append(reqBytes, '\n')); err != nil {
		return fmt.Errorf("failed to write request: %w", err)
	}

	// Read response
	line, err := b.stdout.ReadBytes('\n')
	if err != nil {
		stderr := b.stderr.String()
		if stderr != "" {
			return fmt.Errorf("bridge error: %s", stderr)
		}
		return fmt.Errorf("failed to read response: %w", err)
	}

	var resp response
	if err := json.Unmarshal(line, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w (response: %s)", err, string(line))
	}

	if resp.Error != nil {
		return errors.New(*resp.Error)
	}

	if result != nil {
		if err := json.Unmarshal(resp.Result, result); err != nil {
			return fmt.Errorf("failed to parse result: %w", err)
		}
	}

	return nil
}

// Close shuts down the Python subprocess and releases resources.
func (b *Bridge) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}
	b.closed = true

	var errs []error

	if b.stdin != nil {
		if err := b.stdin.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if b.cmd != nil && b.cmd.Process != nil {
		if err := b.cmd.Process.Kill(); err != nil {
			errs = append(errs, err)
		}
		b.cmd.Wait() // ignore error from killed process
	}

	// Clean up temp file
	if b.scriptPath != "" {
		os.Remove(b.scriptPath)
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// AvailableFormats returns information about all available format adapters.
func (b *Bridge) AvailableFormats() []FormatInfo {
	return b.formats
}

// Read reads a file and returns an OTIO object.
// The format is auto-detected based on the file extension.
func (b *Bridge) Read(path string) (opentimelineio.SerializableObject, error) {
	ext := strings.ToLower(filepath.Ext(path))
	return b.ReadWithFormat(ext, path)
}

// Write writes an OTIO object to a file.
// The format is auto-detected based on the file extension.
func (b *Bridge) Write(obj opentimelineio.SerializableObject, path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	return b.WriteWithFormat(ext, obj, path)
}

// ReadWithFormat reads a file using a specific format.
// Format can be either a suffix (e.g., ".edl") or adapter name (e.g., "cmx_3600").
func (b *Bridge) ReadWithFormat(format, path string) (opentimelineio.SerializableObject, error) {
	// Normalize format to look up in map
	f := strings.ToLower(format)
	if !strings.HasPrefix(f, ".") && b.formatMap[f].Name == "" {
		// Try adding a dot prefix
		f = "." + f
	}

	info, ok := b.formatMap[f]
	if !ok {
		return nil, fmt.Errorf("%s: %w", format, ErrFormatNotFound)
	}

	if !info.CanRead {
		return nil, fmt.Errorf("%s: format does not support reading", info.Name)
	}

	params := map[string]any{
		"filepath": path,
	}

	// If format is the adapter name, pass it explicitly
	if format == info.Name {
		params["adapter"] = info.Name
	}

	var otioJSON string
	if err := b.call("read_from_file", params, &otioJSON); err != nil {
		return nil, fmt.Errorf("python adapter error: %w", err)
	}

	return opentimelineio.FromJSONString(otioJSON)
}

// WriteWithFormat writes an OTIO object to a file using a specific format.
// Format can be either a suffix (e.g., ".edl") or adapter name (e.g., "cmx_3600").
func (b *Bridge) WriteWithFormat(format string, obj opentimelineio.SerializableObject, path string) error {
	// Normalize format to look up in map
	f := strings.ToLower(format)
	if !strings.HasPrefix(f, ".") && b.formatMap[f].Name == "" {
		// Try adding a dot prefix
		f = "." + f
	}

	info, ok := b.formatMap[f]
	if !ok {
		return fmt.Errorf("%s: %w", format, ErrFormatNotFound)
	}

	if !info.CanWrite {
		return fmt.Errorf("%s: format does not support writing", info.Name)
	}

	// Serialize to OTIO JSON
	otioJSON, err := opentimelineio.ToJSONString(obj, "")
	if err != nil {
		return fmt.Errorf("failed to serialize OTIO: %w", err)
	}

	params := map[string]any{
		"filepath": path,
		"data":     otioJSON,
	}

	// If format is the adapter name, pass it explicitly
	if format == info.Name {
		params["adapter"] = info.Name
	}

	var success bool
	if err := b.call("write_to_file", params, &success); err != nil {
		return fmt.Errorf("python adapter error: %w", err)
	}

	return nil
}
