package fileio

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/google/renameio"
)

// writer is responsible for writing files to the filesystem.
type writer struct {
	// rootDir is the root directory, useful for testing
	rootDir string
}

// NewWriter creates a new writer.
func NewWriter() *writer {
	return &writer{}
}

// SetRootdir sets the root directory for the writer, useful for testing.
func (w *writer) SetRootdir(path string) {
	w.rootDir = path
}

// PathFor returns the full path for the given filePath.
func (w *writer) PathFor(filePath string) string {
	return path.Join(w.rootDir, filePath)
}

// WriteFile writes the provided data to the file at the path with the provided permissions.
func (w *writer) WriteFile(name string, data []byte, perm os.FileMode, opts ...FileOption) error {
	fopts := &fileOptions{uid: -1, gid: -1}
	for _, opt := range opts {
		opt(fopts)
	}

	var uid, gid int
	// if rootDir is set use the default UID and GID
	if w.rootDir != "" {
		defaultUID, defaultGID, err := GetUserIdentity()
		if err != nil {
			return err
		}
		uid = defaultUID
		gid = defaultGID
	} else {
		uid = fopts.uid
		gid = fopts.gid
	}

	return writeFileAtomically(filepath.Join(w.rootDir, name), data, DefaultDirectoryPermissions, perm, uid, gid)
}

// RemoveFile removes the file at the given path.
func (w *writer) RemoveFile(file string) error {
	if err := os.Remove(filepath.Join(w.rootDir, file)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove file %q: %w", file, err)
	}
	return nil
}

// RemoveAll removes the file or directory at the given path.
func (w *writer) RemoveAll(path string) error {
	if err := os.RemoveAll(filepath.Join(w.rootDir, path)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove path %q: %w", path, err)
	}
	return nil
}

// RemoveContents removes all files and subdirectories within the given path,
// but leaves the directory itself intact.
func (w *writer) RemoveContents(path string) error {
	fullPath := filepath.Join(w.rootDir, path)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read contents of %q: %w", fullPath, err)
	}

	for _, entry := range entries {
		entryPath := filepath.Join(fullPath, entry.Name())
		if err := os.RemoveAll(entryPath); err != nil {
			return fmt.Errorf("remove entry %q: %w", entryPath, err)
		}
	}

	return nil
}

// MkdirAll creates a directory at the given path with the specified permissions.
func (w *writer) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(filepath.Join(w.rootDir, path), perm)
}

// MkdirTemp creates a temporary directory with the given prefix and returns its path.
func (w *writer) MkdirTemp(prefix string) (string, error) {
	baseDir := filepath.Join(w.rootDir, os.TempDir())
	if err := os.MkdirAll(baseDir, DefaultDirectoryPermissions); err != nil {
		return "", err
	}
	tmpPath, err := os.MkdirTemp(baseDir, prefix)
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(tmpPath, w.rootDir), nil
}

// CopyFile copies a file from src to dst.
func (w *writer) CopyFile(src, dst string) error {
	return w.copyFile(filepath.Join(w.rootDir, src), filepath.Join(w.rootDir, dst))
}

func (w *writer) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dstTarget := dst
	dstInfo, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to stat destination: %w", err)
		}
	} else {
		if dstInfo.IsDir() {
			dstTarget = filepath.Join(dst, filepath.Base(src))
		}
	}

	dstFile, err := os.Create(dstTarget)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dstTarget, err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	srcFileInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}

	if err := os.Chmod(dstTarget, srcFileInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	stat, ok := srcFileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("failed to retrieve UID and GID")
	}

	if err := os.Chown(dstTarget, int(stat.Uid), int(stat.Gid)); err != nil {
		return fmt.Errorf("failed to set UID and GID: %w", err)
	}

	return nil
}

// CopyDir recursively copies a directory from src to dst.
func (w *writer) CopyDir(src, dst string, opts ...CopyDirOption) error {
	options := &copyDirOptions{
		symlinkBehavior: symlinkSkip,
	}
	for _, opt := range opts {
		opt(options)
	}
	fullSrc := filepath.Join(w.rootDir, src)
	fullDst := filepath.Join(w.rootDir, dst)
	absSrc, err := filepath.Abs(fullSrc)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for source: %w", err)
	}
	options.rootDir = absSrc
	return w.copyDirWithVisited(fullSrc, fullDst, options, make(map[string]bool))
}

func (w *writer) copyDirWithVisited(src, dst string, opts *copyDirOptions, visited map[string]bool) error {
	srcInfo, err := os.Lstat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source directory: %w", err)
	}

	if srcInfo.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("source is a symlink: %s", src)
	}

	if !srcInfo.IsDir() {
		return fmt.Errorf("source is not a directory: %s", src)
	}

	absSrc, err := filepath.Abs(src)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for %s: %w", src, err)
	}

	if visited[absSrc] {
		return fmt.Errorf("circular symlink detected: %s (already being processed)", src)
	}
	visited[absSrc] = true
	defer delete(visited, absSrc)

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.Type()&os.ModeSymlink != 0 {
			if err := w.handleSymlink(srcPath, dstPath, opts, visited); err != nil {
				return err
			}
			continue
		}

		if entry.IsDir() {
			if err := w.copyDirWithVisited(srcPath, dstPath, opts, visited); err != nil {
				return err
			}
		} else {
			if err := w.copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *writer) handleSymlink(srcPath, dstPath string, opts *copyDirOptions, visited map[string]bool) error {
	switch opts.symlinkBehavior {
	case symlinkSkip:
		return nil
	case symlinkError:
		return fmt.Errorf("symlink encountered: %s", srcPath)
	case symlinkPreserve:
		return w.preserveSymlink(srcPath, dstPath)
	case symlinkFollow:
		return w.followSymlink(srcPath, dstPath, opts, visited)
	case symlinkFollowWithinRoot:
		return w.followSymlinkWithinRoot(srcPath, dstPath, opts, visited)
	case symlinkPreserveWithinRoot:
		return w.preserveSymlinkWithinRoot(srcPath, dstPath, opts)
	default:
		return fmt.Errorf("unknown symlink behavior: %d", opts.symlinkBehavior)
	}
}

func (w *writer) preserveSymlink(srcPath, dstPath string) error {
	linkTarget, err := os.Readlink(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read symlink %s: %w", srcPath, err)
	}
	if err := os.Symlink(linkTarget, dstPath); err != nil {
		return fmt.Errorf("failed to create symlink %s: %w", dstPath, err)
	}
	return nil
}

func (w *writer) followSymlink(srcPath, dstPath string, opts *copyDirOptions, visited map[string]bool) error {
	resolved, err := filepath.EvalSymlinks(srcPath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlink %s: %w", srcPath, err)
	}

	info, err := os.Stat(resolved)
	if err != nil {
		return fmt.Errorf("failed to stat symlink target %s: %w", resolved, err)
	}

	if info.IsDir() {
		return w.copyDirWithVisited(resolved, dstPath, opts, visited)
	}
	return w.copyFile(resolved, dstPath)
}

func isWithinRoot(targetPath, rootDir string) (bool, error) {
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return false, err
	}

	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return false, err
	}

	if absTarget == absRoot {
		return true, nil
	}

	if strings.HasPrefix(absTarget, fmt.Sprintf("%s/", absRoot)) {
		return true, nil
	}
	return false, nil
}

func (w *writer) followSymlinkWithinRoot(srcPath, dstPath string, opts *copyDirOptions, visited map[string]bool) error {
	resolved, err := filepath.EvalSymlinks(srcPath)
	if err != nil {
		return fmt.Errorf("resolve symlink %s: %w", srcPath, err)
	}

	within, err := isWithinRoot(resolved, opts.rootDir)
	if err != nil {
		return fmt.Errorf("symlink within root: %w", err)
	}
	if !within {
		return nil
	}

	info, err := os.Stat(resolved)
	if err != nil {
		return fmt.Errorf("stat symlink target %s: %w", resolved, err)
	}

	if info.IsDir() {
		return w.copyDirWithVisited(resolved, dstPath, opts, visited)
	}
	return w.copyFile(resolved, dstPath)
}

func (w *writer) preserveSymlinkWithinRoot(srcPath, dstPath string, opts *copyDirOptions) error {
	resolved, err := filepath.EvalSymlinks(srcPath)
	if err != nil {
		return fmt.Errorf("resolve symlink %s: %w", srcPath, err)
	}

	within, err := isWithinRoot(resolved, opts.rootDir)
	if err != nil {
		return fmt.Errorf("symlink within root: %w", err)
	}
	if !within {
		return nil
	}

	return w.preserveSymlink(srcPath, dstPath)
}

// OverwriteAndWipe overwrites the file with random data and then removes it.
func (w *writer) OverwriteAndWipe(file string) error {
	if err := w.overwriteFileWithRandomData(file); err != nil {
		return fmt.Errorf("could not overwrite file %s with random data: %w", file, err)
	}
	if err := w.RemoveFile(file); err != nil {
		return fmt.Errorf("could not remove file %s: %w", file, err)
	}
	return nil
}

func (w *writer) overwriteFileWithRandomData(file string) error {
	f, err := os.OpenFile(filepath.Join(w.rootDir, file), os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	fileInfo, err := f.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	fileSize := fileInfo.Size()

	randomData := make([]byte, fileSize)
	if _, err := rand.Read(randomData); err != nil {
		return fmt.Errorf("failed to generate random data: %w", err)
	}

	if _, err := f.WriteAt(randomData, 0); err != nil {
		return fmt.Errorf("failed to write random data: %w", err)
	}

	return nil
}

// writeFileAtomically uses the renameio package to provide atomic file writing.
func writeFileAtomically(fpath string, b []byte, dirMode, fileMode os.FileMode, uid, gid int) error {
	dir := filepath.Dir(fpath)
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return fmt.Errorf("failed to create directory %q: %w", dir, err)
	}
	t, err := renameio.TempFile(dir, fpath)
	if err != nil {
		return err
	}
	defer func() {
		_ = t.Cleanup()
	}()
	if err := t.Chmod(fileMode); err != nil {
		return err
	}
	w := bufio.NewWriter(t)
	if _, err := w.Write(b); err != nil {
		return err
	}
	if err := w.Flush(); err != nil {
		return err
	}
	if uid != -1 && gid != -1 {
		if err := t.Chown(uid, gid); err != nil {
			return err
		}
	}
	return t.CloseAtomicallyReplace()
}

// GetUserIdentity returns the current user's UID and GID.
func GetUserIdentity() (int, int, error) {
	currentUser, err := user.Current()
	if err != nil {
		return 0, 0, fmt.Errorf("failed retrieving current user: %w", err)
	}
	gid, err := strconv.Atoi(currentUser.Gid)
	if err != nil {
		return 0, 0, fmt.Errorf("failed converting GID to int: %w", err)
	}
	uid, err := strconv.Atoi(currentUser.Uid)
	if err != nil {
		return 0, 0, fmt.Errorf("failed converting UID to int: %w", err)
	}
	return uid, gid, nil
}

// LookupUID looks up a user by username and returns the UID.
func LookupUID(username string) (int, error) {
	osUser, err := user.Lookup(username)
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve UserID for username: %s", username)
	}
	uid, _ := strconv.Atoi(osUser.Uid)
	return uid, nil
}

// LookupGID looks up a group by name and returns the GID.
func LookupGID(group string) (int, error) {
	osGroup, err := user.LookupGroup(group)
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve GroupID for group: %v", group)
	}
	gid, _ := strconv.Atoi(osGroup.Gid)
	return gid, nil
}
