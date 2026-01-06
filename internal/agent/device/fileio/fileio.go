// Package fileio provides agent-specific file I/O operations that extend pkg/fileio.
package fileio

import (
	"github.com/flightctl/flightctl/api/v1beta1"
	"github.com/flightctl/flightctl/pkg/fileio"
)

type ManagedFile interface {
	Path() string
	Exists() (bool, error)
	IsUpToDate() (bool, error)
	Write() error
}

type ManagedWriter interface {
	fileio.Writer
	// CreateManagedFile creates a managed file with the given spec.
	CreateManagedFile(file v1beta1.FileSpec) (ManagedFile, error)
}

type ManagedReadWriter interface {
	fileio.Reader
	ManagedWriter
}

type managedWriter struct {
	fileio.Writer
}

func NewManagedWriter() ManagedWriter {
	return &managedWriter{Writer: fileio.NewWriter()}
}

func (w *managedWriter) CreateManagedFile(file v1beta1.FileSpec) (ManagedFile, error) {
	return newManagedFile(file, w)
}

type managedReadWriter struct {
	fileio.Reader
	*managedWriter
}

func NewManagedReadWriter(opts ...Option) ManagedReadWriter {
	rw := &managedReadWriter{
		Reader:        fileio.NewReader(),
		managedWriter: &managedWriter{Writer: fileio.NewWriter()},
	}
	for _, opt := range opts {
		opt(rw)
	}
	return rw
}

func (rw *managedReadWriter) SetRootdir(path string) {
	rw.Reader.SetRootdir(path)
	rw.Writer.SetRootdir(path)
}

func (rw *managedReadWriter) PathFor(path string) string {
	return rw.Writer.PathFor(path)
}

type Option func(*managedReadWriter)

// WithTestRootDir sets the root directory for the reader and writer, useful for testing.
func WithTestRootDir(testRootDir string) Option {
	return func(rw *managedReadWriter) {
		if testRootDir != "" {
			rw.SetRootdir(testRootDir)
		}
	}
}
