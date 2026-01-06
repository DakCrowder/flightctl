package fileio

import (
	"fmt"
	"io"
	"os"

	"github.com/flightctl/flightctl/internal/agent/device/errors"
)

const (
	entryTypeFile = "file"
	entryTypeDir  = "dir"
)

// PathExistsOption represents options for PathExists function
type PathExistsOption func(*pathExistsOptions)

type pathExistsOptions struct {
	skipContentCheck bool
}

// WithSkipContentCheck configures PathExists to skip content verification
// and only check if the path can be opened
func WithSkipContentCheck() PathExistsOption {
	return func(opts *pathExistsOptions) {
		opts.skipContentCheck = true
	}
}

// checkPathExists checks if a path exists with optional content validation.
// This is agent-specific as it uses agent error types.
func checkPathExists(path string, options *pathExistsOptions) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("error checking path: %w", err)
	}
	pathType := entryTypeFile
	if info.IsDir() {
		pathType = entryTypeDir
	}

	// Open the file/directory once
	file, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("%s exists but %w: %w", pathType, errors.ErrReadingPath, err)
	}
	defer file.Close()

	// If we only need to check if it can be opened, we're done
	if options.skipContentCheck {
		return true, nil
	}

	if err = validateContents(file, info.IsDir()); err != nil {
		return false, fmt.Errorf("%s exists but %w", pathType, err)
	}
	return true, nil
}

func validateContents(file *os.File, isDir bool) error {
	if isDir {
		// read a single entry from the directory to confirm readability
		_, err := file.Readdirnames(1)
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("%w: %w", errors.ErrReadingPath, err)
		}
		return nil
	}

	// read a single byte from the file to ensure permissions are correct
	buffer := make([]byte, 1)
	_, err := file.Read(buffer)
	if err != nil {
		return fmt.Errorf("%w, %w", errors.ErrReadingPath, err)
	}
	return nil
}
