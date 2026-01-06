// Package fileio provides agent-specific file I/O operations that wrap pkg/fileio.
// This package adds ManagedFile support using v1beta1.FileSpec and agent-specific error handling.
package fileio

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/flightctl/flightctl/api/v1beta1"
	pkgfileio "github.com/flightctl/flightctl/pkg/fileio"
)

// Re-export constants from pkg/fileio
const (
	DefaultDirectoryPermissions  = pkgfileio.DefaultDirectoryPermissions
	DefaultFilePermissions       = pkgfileio.DefaultFilePermissions
	DefaultExecutablePermissions = pkgfileio.DefaultExecutablePermissions
)

// ManagedFile represents a file managed by the agent with spec-based configuration.
type ManagedFile interface {
	Path() string
	Exists() (bool, error)
	IsUpToDate() (bool, error)
	Write() error
}

// Writer extends pkg/fileio.Writer with agent-specific CreateManagedFile method.
type Writer interface {
	pkgfileio.Writer
	// CreateManagedFile creates a managed file with the given spec.
	CreateManagedFile(file v1beta1.FileSpec) (ManagedFile, error)
}

// Reader extends pkg/fileio.Reader with agent-specific PathExists options.
type Reader interface {
	// SetRootdir sets the root directory for the reader, useful for testing
	SetRootdir(path string)
	// PathFor returns the full path for the given filePath, prepending the rootDir if set
	PathFor(filePath string) string
	// ReadFile reads the file at the provided path
	ReadFile(filePath string) ([]byte, error)
	// ReadDir reads the directory at the provided path and returns a slice of fs.DirEntry
	ReadDir(dirPath string) ([]fs.DirEntry, error)
	// PathExists checks if a path exists with optional content validation
	PathExists(path string, opts ...PathExistsOption) (bool, error)
}

// ReadWriter combines Reader and Writer interfaces.
type ReadWriter interface {
	Reader
	Writer
}

// writer wraps pkg/fileio writer and adds CreateManagedFile.
type writer struct {
	pkgWriter pkgfileio.Writer
}

// NewWriter creates a new Writer.
func NewWriter() *writer {
	return &writer{pkgWriter: pkgfileio.NewWriter()}
}

// Delegate all Writer methods to the embedded pkgWriter
func (w *writer) SetRootdir(path string)                       { w.pkgWriter.SetRootdir(path) }
func (w *writer) PathFor(filePath string) string               { return w.pkgWriter.PathFor(filePath) }
func (w *writer) RemoveFile(file string) error                 { return w.pkgWriter.RemoveFile(file) }
func (w *writer) RemoveAll(path string) error                  { return w.pkgWriter.RemoveAll(path) }
func (w *writer) RemoveContents(path string) error             { return w.pkgWriter.RemoveContents(path) }
func (w *writer) MkdirAll(path string, perm fs.FileMode) error { return w.pkgWriter.MkdirAll(path, perm) }
func (w *writer) MkdirTemp(prefix string) (string, error)      { return w.pkgWriter.MkdirTemp(prefix) }
func (w *writer) CopyFile(src, dst string) error               { return w.pkgWriter.CopyFile(src, dst) }
func (w *writer) OverwriteAndWipe(file string) error           { return w.pkgWriter.OverwriteAndWipe(file) }

func (w *writer) WriteFile(name string, data []byte, perm fs.FileMode, opts ...pkgfileio.FileOption) error {
	return w.pkgWriter.WriteFile(name, data, perm, opts...)
}

func (w *writer) CopyDir(src, dst string, opts ...pkgfileio.CopyDirOption) error {
	return w.pkgWriter.CopyDir(src, dst, opts...)
}

func (w *writer) CreateManagedFile(file v1beta1.FileSpec) (ManagedFile, error) {
	return newManagedFile(file, w)
}

// reader wraps pkg/fileio reader with agent-specific PathExists.
type reader struct {
	pkgReader pkgfileio.Reader
}

// NewReader creates a new Reader.
func NewReader() *reader {
	return &reader{pkgReader: pkgfileio.NewReader()}
}

// Delegate Reader methods to the embedded pkgReader
func (r *reader) SetRootdir(path string)                        { r.pkgReader.SetRootdir(path) }
func (r *reader) PathFor(filePath string) string                { return r.pkgReader.PathFor(filePath) }
func (r *reader) ReadFile(filePath string) ([]byte, error)      { return r.pkgReader.ReadFile(filePath) }
func (r *reader) ReadDir(dirPath string) ([]fs.DirEntry, error) { return r.pkgReader.ReadDir(dirPath) }

// PathExists checks if a path exists with optional content validation.
func (r *reader) PathExists(path string, opts ...PathExistsOption) (bool, error) {
	options := &pathExistsOptions{}
	for _, opt := range opts {
		opt(options)
	}
	return checkPathExists(r.PathFor(path), options)
}

// readWriter combines reader and writer.
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

// Re-export FileOption and related functions from pkg/fileio
type FileOption = pkgfileio.FileOption

var (
	WithUid = pkgfileio.WithUid
	WithGid = pkgfileio.WithGid
)

// Re-export CopyDirOption and related functions from pkg/fileio
type CopyDirOption = pkgfileio.CopyDirOption

var (
	WithSkipSymlink               = pkgfileio.WithSkipSymlink
	WithErrorOnSymlink            = pkgfileio.WithErrorOnSymlink
	WithPreserveSymlink           = pkgfileio.WithPreserveSymlink
	WithFollowSymlink             = pkgfileio.WithFollowSymlink
	WithFollowSymlinkWithinRoot   = pkgfileio.WithFollowSymlinkWithinRoot
	WithPreserveSymlinkWithinRoot = pkgfileio.WithPreserveSymlinkWithinRoot
)

// IsNotExist returns a boolean indicating whether the error is known to report that a file or directory does not exist.
func IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

// Re-export utility functions from pkg/fileio
var (
	LookupUID = pkgfileio.LookupUID
	LookupGID = pkgfileio.LookupGID
)

// WriteTmpFile writes the given content to a temporary file with the specified name prefix.
// It returns the path to the tmp file and a cleanup function to remove it.
func WriteTmpFile(rw ReadWriter, prefix, filename string, content []byte, perm os.FileMode) (path string, cleanup func(), err error) {
	tmpDir, err := rw.MkdirTemp(prefix)
	if err != nil {
		return "", nil, fmt.Errorf("creating tmp dir: %w", err)
	}

	tmpPath := filepath.Join(tmpDir, filename)
	if err := rw.WriteFile(tmpPath, content, perm); err != nil {
		_ = rw.RemoveAll(tmpDir)
		return "", nil, fmt.Errorf("writing tmp file: %w", err)
	}

	cleanup = func() {
		_ = rw.RemoveAll(tmpDir)
	}
	return tmpPath, cleanup, nil
}

// AppendFile appends the provided data to the file at the path, creating the file if it doesn't exist.
func AppendFile(w Writer, name string, data []byte, perm os.FileMode, opts ...FileOption) error {
	adapter := &writerAdapter{w: w}
	return pkgfileio.AppendFile(adapter, name, data, perm, opts...)
}

// UnpackTar unpacks a tar or tar.gz file to the destination directory.
func UnpackTar(writer Writer, tarPath, destDir string) error {
	adapter := &writerAdapter{w: writer}
	return pkgfileio.UnpackTar(adapter, tarPath, destDir)
}

// writerAdapter adapts agent Writer to pkg/fileio.Writer
type writerAdapter struct {
	w Writer
}

func (a *writerAdapter) SetRootdir(path string)         { a.w.SetRootdir(path) }
func (a *writerAdapter) PathFor(filePath string) string { return a.w.PathFor(filePath) }
func (a *writerAdapter) WriteFile(name string, data []byte, perm os.FileMode, opts ...pkgfileio.FileOption) error {
	return a.w.WriteFile(name, data, perm, opts...)
}
func (a *writerAdapter) RemoveFile(file string) error                 { return a.w.RemoveFile(file) }
func (a *writerAdapter) RemoveAll(path string) error                  { return a.w.RemoveAll(path) }
func (a *writerAdapter) RemoveContents(path string) error             { return a.w.RemoveContents(path) }
func (a *writerAdapter) MkdirAll(path string, perm os.FileMode) error { return a.w.MkdirAll(path, perm) }
func (a *writerAdapter) MkdirTemp(prefix string) (string, error)      { return a.w.MkdirTemp(prefix) }
func (a *writerAdapter) CopyFile(src, dst string) error               { return a.w.CopyFile(src, dst) }
func (a *writerAdapter) CopyDir(src, dst string, opts ...pkgfileio.CopyDirOption) error {
	return a.w.CopyDir(src, dst, opts...)
}
func (a *writerAdapter) OverwriteAndWipe(file string) error { return a.w.OverwriteAndWipe(file) }
