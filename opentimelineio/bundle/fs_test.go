// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package bundle

import (
	"errors"
	"io/fs"
	"os"
	"testing"

	"github.com/absfs/absfs"
	"github.com/absfs/memfs"
)

func TestMemFSAdapter(t *testing.T) {
	// Create an in-memory filesystem
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatalf("Failed to create memfs: %v", err)
	}

	fsys := NewMemFSAdapter(mfs)

	// Test MkdirAll
	if err := fsys.MkdirAll("/test/dir", 0755); err != nil {
		t.Errorf("MkdirAll failed: %v", err)
	}

	// Test WriteFile
	data := []byte("hello world")
	if err := fsys.WriteFile("/test/dir/file.txt", data, 0644); err != nil {
		t.Errorf("WriteFile failed: %v", err)
	}

	// Test ReadFile
	readData, err := fsys.ReadFile("/test/dir/file.txt")
	if err != nil {
		t.Errorf("ReadFile failed: %v", err)
	}
	if string(readData) != string(data) {
		t.Errorf("ReadFile mismatch: got %q, want %q", readData, data)
	}

	// Test Stat
	info, err := fsys.Stat("/test/dir/file.txt")
	if err != nil {
		t.Errorf("Stat failed: %v", err)
	}
	if info.Size() != int64(len(data)) {
		t.Errorf("Stat size: got %d, want %d", info.Size(), len(data))
	}

	// Test Open
	f, err := fsys.Open("/test/dir/file.txt")
	if err != nil {
		t.Errorf("Open failed: %v", err)
	}
	f.Close()

	// Test Create
	f2, err := fsys.Create("/test/dir/new.txt")
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
	f2.Write([]byte("new content"))
	f2.Close()

	// Test Remove
	if err := fsys.Remove("/test/dir/new.txt"); err != nil {
		t.Errorf("Remove failed: %v", err)
	}

	// Test Mkdir
	if err := fsys.Mkdir("/test/single", 0755); err != nil {
		t.Errorf("Mkdir failed: %v", err)
	}
}

func TestCopyFileFSMemFS(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatalf("Failed to create memfs: %v", err)
	}

	fsys := NewMemFSAdapter(mfs)

	// Create source file
	srcContent := []byte("source content")
	if err := fsys.WriteFile("/src.txt", srcContent, 0644); err != nil {
		t.Fatalf("Failed to write source: %v", err)
	}

	// Copy file
	if err := copyFileFS(fsys, "/src.txt", "/dst.txt"); err != nil {
		t.Errorf("copyFileFS failed: %v", err)
	}

	// Verify content
	dstContent, err := fsys.ReadFile("/dst.txt")
	if err != nil {
		t.Errorf("ReadFile dst failed: %v", err)
	}
	if string(dstContent) != string(srcContent) {
		t.Errorf("Content mismatch: got %q, want %q", dstContent, srcContent)
	}
}

func TestCopyFileFSErrors(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatalf("Failed to create memfs: %v", err)
	}

	fsys := NewMemFSAdapter(mfs)

	// Test source not found
	err = copyFileFS(fsys, "/nonexistent.txt", "/dst.txt")
	if err == nil {
		t.Error("Expected error for nonexistent source")
	}
}

func TestDefaultFSOperations(t *testing.T) {
	// Test DefaultFS basic operations
	fsys := DefaultFS

	// Stat on nonexistent file should fail
	_, err := fsys.Stat("/definitely/nonexistent/path/12345")
	if err == nil {
		t.Error("Expected error for nonexistent path")
	}

	// Open nonexistent file should fail
	_, err = fsys.Open("/definitely/nonexistent/path/12345")
	if err == nil {
		t.Error("Expected error for nonexistent path")
	}

	// ReadFile nonexistent should fail
	_, err = fsys.ReadFile("/definitely/nonexistent/path/12345")
	if err == nil {
		t.Error("Expected error for nonexistent path")
	}
}

// errorFS is a filesystem that returns errors for testing error paths.
type errorFS struct {
	openErr      error
	createErr    error
	mkdirErr     error
	mkdirAllErr  error
	statErr      error
	readFileErr  error
	writeFileErr error
	removeErr    error
}

func (e *errorFS) Open(name string) (absfs.File, error) {
	if e.openErr != nil {
		return nil, e.openErr
	}
	return nil, os.ErrNotExist
}

func (e *errorFS) Create(name string) (absfs.File, error) {
	if e.createErr != nil {
		return nil, e.createErr
	}
	return nil, errors.New("create not implemented")
}

func (e *errorFS) Mkdir(name string, perm os.FileMode) error {
	if e.mkdirErr != nil {
		return e.mkdirErr
	}
	return nil
}

func (e *errorFS) MkdirAll(name string, perm os.FileMode) error {
	if e.mkdirAllErr != nil {
		return e.mkdirAllErr
	}
	return nil
}

func (e *errorFS) Stat(name string) (fs.FileInfo, error) {
	if e.statErr != nil {
		return nil, e.statErr
	}
	return nil, os.ErrNotExist
}

func (e *errorFS) ReadFile(name string) ([]byte, error) {
	if e.readFileErr != nil {
		return nil, e.readFileErr
	}
	return nil, os.ErrNotExist
}

func (e *errorFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	if e.writeFileErr != nil {
		return e.writeFileErr
	}
	return nil
}

func (e *errorFS) Remove(name string) error {
	if e.removeErr != nil {
		return e.removeErr
	}
	return nil
}

func TestErrorFSCopyFile(t *testing.T) {
	// Test copyFileFS with an error filesystem
	errFS := &errorFS{
		openErr: errors.New("simulated open error"),
	}

	err := copyFileFS(errFS, "/src.txt", "/dst.txt")
	if err == nil {
		t.Error("Expected error from copyFileFS")
	}
}

func TestMemFSAdapterReadFileError(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatalf("Failed to create memfs: %v", err)
	}

	fsys := NewMemFSAdapter(mfs)

	// Reading nonexistent file should fail
	_, err = fsys.ReadFile("/nonexistent.txt")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestMemFSAdapterWriteFileCreatesDir(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatalf("Failed to create memfs: %v", err)
	}

	fsys := NewMemFSAdapter(mfs)

	// WriteFile should create parent directory
	err = fsys.WriteFile("/deep/nested/path/file.txt", []byte("content"), 0644)
	if err != nil {
		t.Errorf("WriteFile with nested path failed: %v", err)
	}

	// Verify file exists
	data, err := fsys.ReadFile("/deep/nested/path/file.txt")
	if err != nil {
		t.Errorf("ReadFile failed: %v", err)
	}
	if string(data) != "content" {
		t.Errorf("Content mismatch: got %q", data)
	}
}

func TestOSFSCreate(t *testing.T) {
	// Test that DefaultFS.Create works with real filesystem
	fsys := DefaultFS

	tmpFile, err := os.CreateTemp("", "osfs_test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	os.Remove(tmpPath)

	// Create using DefaultFS
	f, err := fsys.Create(tmpPath)
	if err != nil {
		t.Fatalf("DefaultFS.Create failed: %v", err)
	}
	f.Write([]byte("test"))
	f.Close()

	// Cleanup
	os.Remove(tmpPath)
}

func TestOSFSMkdirAndRemove(t *testing.T) {
	fsys := DefaultFS

	tmpDir, err := os.MkdirTemp("", "osfs_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testDir := tmpDir + "/testdir"

	// Test Mkdir
	if err := fsys.Mkdir(testDir, 0755); err != nil {
		t.Errorf("Mkdir failed: %v", err)
	}

	// Test Remove
	if err := fsys.Remove(testDir); err != nil {
		t.Errorf("Remove failed: %v", err)
	}
}

func TestOSFSWriteFile(t *testing.T) {
	fsys := DefaultFS

	tmpDir, err := os.MkdirTemp("", "osfs_write_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := tmpDir + "/test.txt"
	data := []byte("test content")

	// Test WriteFile
	if err := fsys.WriteFile(testFile, data, 0644); err != nil {
		t.Errorf("WriteFile failed: %v", err)
	}

	// Verify content
	readData, err := fsys.ReadFile(testFile)
	if err != nil {
		t.Errorf("ReadFile failed: %v", err)
	}
	if string(readData) != string(data) {
		t.Errorf("Content mismatch: got %q, want %q", readData, data)
	}
}
