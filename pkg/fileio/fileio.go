// Package fileio provides generic file I/O operations with testable root directory support.
// This package contains no agent-specific dependencies and can be used by any component.
package fileio

import (
	"io/fs"
	"os"
)

const (
	// DefaultDirectoryPermissions houses the default mode to use when no directory permissions are provided
	DefaultDirectoryPermissions os.FileMode = 0o755
	// DefaultFilePermissions houses the default mode to use when no file permissions are provided
	DefaultFilePermissions os.FileMode = 0o644
	// DefaultExecutablePermissions houses the default mode to use for executable files
	DefaultExecutablePermissions os.FileMode = 0o755
)

// Reader defines read operations for files and directories.
type Reader interface {
	// SetRootdir sets the root directory for the reader, useful for testing
	SetRootdir(path string)
	// PathFor returns the full path for the given filePath, prepending the rootDir if set
	PathFor(filePath string) string
	// ReadFile reads the file at the provided path
	ReadFile(filePath string) ([]byte, error)
	// ReadDir reads the directory at the provided path and returns a slice of fs.DirEntry
	ReadDir(dirPath string) ([]fs.DirEntry, error)
	// PathExists checks if a path exists and returns a boolean indicating existence
	PathExists(path string) (bool, error)
}

// Writer defines write operations for files and directories.
type Writer interface {
	// SetRootdir sets the root directory for the writer, useful for testing
	SetRootdir(path string)
	// PathFor returns the full path for the given filePath, prepending the rootDir if set
	PathFor(filePath string) string
	// WriteFile writes the provided data to the file at the path with the provided permissions
	WriteFile(name string, data []byte, perm fs.FileMode, opts ...FileOption) error
	// RemoveFile removes the file at the given path
	RemoveFile(file string) error
	// RemoveAll removes the file or directory at the given path
	RemoveAll(path string) error
	// RemoveContents removes all files and subdirectories within the given path,
	// but leaves the directory itself intact
	RemoveContents(path string) error
	// MkdirAll creates a directory at the given path with the specified permissions
	MkdirAll(path string, perm fs.FileMode) error
	// MkdirTemp creates a temporary directory with the given prefix and returns its path
	MkdirTemp(prefix string) (string, error)
	// CopyFile copies a file from src to dst
	CopyFile(src, dst string) error
	// CopyDir recursively copies a directory from src to dst
	CopyDir(src, dst string, opts ...CopyDirOption) error
	// OverwriteAndWipe overwrites the file at the given path with zeros and then deletes it
	OverwriteAndWipe(file string) error
}

// ReadWriter combines Reader and Writer interfaces.
type ReadWriter interface {
	Reader
	Writer
}

type readWriter struct {
	*reader
	*writer
}

// NewReadWriter creates a new ReadWriter instance.
func NewReadWriter(opts ...Option) ReadWriter {
	rw := &readWriter{
		reader: NewReader(),
		writer: NewWriter(),
	}
	for _, opt := range opts {
		opt(rw)
	}
	return rw
}

func (rw *readWriter) SetRootdir(path string) {
	rw.reader.SetRootdir(path)
	rw.writer.SetRootdir(path)
}

func (rw *readWriter) PathFor(path string) string {
	return rw.writer.PathFor(path)
}

// Option is a functional option for configuring ReadWriter.
type Option func(*readWriter)

// WithTestRootDir sets the root directory for the reader and writer, useful for testing.
func WithTestRootDir(testRootDir string) Option {
	return func(rw *readWriter) {
		if testRootDir != "" {
			rw.SetRootdir(testRootDir)
		}
	}
}

// IsNotExist returns a boolean indicating whether the error is known to report that a file or directory does not exist.
func IsNotExist(err error) bool {
	return os.IsNotExist(err)
}
