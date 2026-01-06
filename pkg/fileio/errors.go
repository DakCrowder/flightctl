package fileio

import "errors"

var (
	// ErrReadingPath indicates a failure to read a file or directory path.
	ErrReadingPath = errors.New("failed reading path")
	// ErrPathIsDir indicates that a path expected to be a file is actually a directory.
	ErrPathIsDir = errors.New("provided path is a directory")
)
