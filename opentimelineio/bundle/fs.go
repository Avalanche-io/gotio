// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package bundle

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/absfs/absfs"
)

// FileSystem abstracts filesystem operations for testing.
// It uses the absfs.FileSystem interface for compatibility with memfs.
type FileSystem interface {
	// Open opens a file for reading.
	Open(name string) (absfs.File, error)
	// Create creates a file for writing.
	Create(name string) (absfs.File, error)
	// Mkdir creates a directory.
	Mkdir(name string, perm os.FileMode) error
	// MkdirAll creates a directory and all parents.
	MkdirAll(name string, perm os.FileMode) error
	// Stat returns file info.
	Stat(name string) (fs.FileInfo, error)
	// ReadFile reads a file's entire contents.
	ReadFile(name string) ([]byte, error)
	// WriteFile writes data to a file.
	WriteFile(name string, data []byte, perm os.FileMode) error
	// Remove removes a file or empty directory.
	Remove(name string) error
}

// osFS wraps the os package to implement FileSystem.
type osFS struct{}

// DefaultFS is the default filesystem using os package.
var DefaultFS FileSystem = &osFS{}

func (osFS) Open(name string) (absfs.File, error) {
	return os.Open(name)
}

func (osFS) Create(name string) (absfs.File, error) {
	return os.Create(name)
}

func (osFS) Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}

func (osFS) MkdirAll(name string, perm os.FileMode) error {
	return os.MkdirAll(name, perm)
}

func (osFS) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (osFS) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (osFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

func (osFS) Remove(name string) error {
	return os.Remove(name)
}

// memFSAdapter adapts absfs.FileSystem to our FileSystem interface.
type memFSAdapter struct {
	fs absfs.FileSystem
}

// NewMemFSAdapter creates a FileSystem from an absfs.FileSystem (like memfs).
func NewMemFSAdapter(fs absfs.FileSystem) FileSystem {
	return &memFSAdapter{fs: fs}
}

func (m *memFSAdapter) Open(name string) (absfs.File, error) {
	return m.fs.Open(name)
}

func (m *memFSAdapter) Create(name string) (absfs.File, error) {
	return m.fs.Create(name)
}

func (m *memFSAdapter) Mkdir(name string, perm os.FileMode) error {
	return m.fs.Mkdir(name, perm)
}

func (m *memFSAdapter) MkdirAll(name string, perm os.FileMode) error {
	return m.fs.MkdirAll(name, perm)
}

func (m *memFSAdapter) Stat(name string) (fs.FileInfo, error) {
	return m.fs.Stat(name)
}

func (m *memFSAdapter) ReadFile(name string) ([]byte, error) {
	f, err := m.fs.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

func (m *memFSAdapter) WriteFile(name string, data []byte, perm os.FileMode) error {
	// Ensure parent directory exists
	dir := filepath.Dir(name)
	if dir != "" && dir != "." {
		m.fs.MkdirAll(dir, 0755)
	}

	f, err := m.fs.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

func (m *memFSAdapter) Remove(name string) error {
	return m.fs.Remove(name)
}

// copyFileFS copies a file using the given filesystem.
func copyFileFS(fsys FileSystem, src, dst string) error {
	srcFile, err := fsys.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := fsys.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
