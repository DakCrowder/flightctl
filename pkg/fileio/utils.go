package fileio

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	fopts := &fileOptions{uid: -1, gid: -1}
	for _, opt := range opts {
		opt(fopts)
	}

	var uid, gid int
	// Check if we're in test mode by checking if PathFor returns a modified path
	testPath := w.PathFor("")
	if testPath != "" {
		// Test mode: use default UID and GID
		defaultUID, defaultGID, err := GetUserIdentity()
		if err != nil {
			return err
		}
		uid = defaultUID
		gid = defaultGID
	} else {
		// Production mode: use provided or default ownership
		uid = fopts.uid
		gid = fopts.gid
	}

	filePath := w.PathFor(name)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), DefaultDirectoryPermissions); err != nil {
		return err
	}

	// Open file for appending, create if not exists
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	defer file.Close()

	// Set ownership if specified (use -1 as "don't change ownership" sentinel)
	if uid != -1 && gid != -1 {
		if err := file.Chown(uid, gid); err != nil {
			return err
		}
	}

	// Append data
	_, err = file.Write(data)
	return err
}

// UnpackTar unpacks a tar or tar.gz file to the destination directory.
func UnpackTar(writer Writer, tarPath, destDir string) error {
	// Open the tar file for streaming
	file, err := os.Open(writer.PathFor(tarPath))
	if err != nil {
		return fmt.Errorf("opening tar file: %w", err)
	}
	defer file.Close()

	var tarReader *tar.Reader
	if strings.HasSuffix(tarPath, ".gz") || strings.HasSuffix(tarPath, ".tgz") {
		gzr, err := gzip.NewReader(file)
		if err != nil {
			return fmt.Errorf("creating gzip reader: %w", err)
		}
		defer gzr.Close()
		tarReader = tar.NewReader(gzr)
	} else {
		tarReader = tar.NewReader(file)
	}

	// Extract tar contents
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading tar header: %w", err)
		}

		cleanName := filepath.Clean(header.Name)
		if strings.HasPrefix(cleanName, "..") || filepath.IsAbs(cleanName) {
			return fmt.Errorf("invalid file path in tar: %s", header.Name)
		}
		destPath := filepath.Join(destDir, cleanName)

		perm := header.FileInfo().Mode().Perm()

		switch header.Typeflag {
		case tar.TypeDir:
			if perm == 0 {
				perm = DefaultDirectoryPermissions
			}
			if err := writer.MkdirAll(destPath, perm); err != nil {
				return fmt.Errorf("creating directory %s: %w", destPath, err)
			}
		case tar.TypeReg:
			if perm == 0 {
				perm = DefaultFilePermissions
			}
			if err := os.MkdirAll(filepath.Dir(writer.PathFor(destPath)), DefaultDirectoryPermissions); err != nil {
				return fmt.Errorf("creating parent directory: %w", err)
			}
			destFile, err := os.OpenFile(writer.PathFor(destPath), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
			if err != nil {
				return fmt.Errorf("creating file %s: %w", destPath, err)
			}
			// Use LimitReader to prevent decompression bombs
			limitedReader := io.LimitReader(tarReader, header.Size)
			if _, err := io.Copy(destFile, limitedReader); err != nil { // #nosec G110
				destFile.Close()
				return fmt.Errorf("writing file %s: %w", destPath, err)
			}
			destFile.Close()
		}
	}

	return nil
}
