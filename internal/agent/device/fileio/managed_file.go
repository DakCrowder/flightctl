package fileio

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"syscall"

	"github.com/ccoveille/go-safecast"
	"github.com/flightctl/flightctl/api/v1beta1"
	"github.com/flightctl/flightctl/pkg/fileio"
	"github.com/samber/lo"
)

type managedFile struct {
	file     v1beta1.FileSpec
	exists   bool
	size     int64
	perms    os.FileMode
	uid      int
	gid      int
	contents []byte
	writer   ManagedWriter
}

func newManagedFile(f v1beta1.FileSpec, writer ManagedWriter) (ManagedFile, error) {
	mf := &managedFile{
		file:   f,
		writer: writer,
	}
	if err := mf.initExistingFileMetadata(); err != nil {
		return nil, err
	}
	return mf, nil
}

// initExistingFileMetadata initializes the exists and size fields of the on disk managedFile.
func (m *managedFile) initExistingFileMetadata() error {
	path := m.writer.PathFor(m.Path())
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if fileInfo.IsDir() {
		return fmt.Errorf("%w: %s", fileio.ErrPathIsDir, path)
	}
	m.exists = true
	m.size = fileInfo.Size()
	return nil
}

func (m *managedFile) decodeFile() error {
	if m.contents != nil {
		return nil
	}
	contents, err := DecodeContent(m.file.Content, m.file.ContentEncoding)
	if err != nil {
		return err
	}
	m.contents = contents

	m.uid, m.gid, err = getFileOwnership(m.file)
	if err != nil {
		return fmt.Errorf("failed to retrieve file ownership for file %q: %w", m.Path(), err)
	}

	m.perms, err = intToFileMode(m.file.Mode)
	if err != nil {
		return fmt.Errorf("failed to retrieve file permissions for file %q: %w", m.Path(), err)
	}

	return nil
}

func (m *managedFile) isUpToDate() (bool, error) {
	if err := m.decodeFile(); err != nil {
		return false, err
	}
	currentContent, err := os.ReadFile(m.writer.PathFor(m.Path()))
	if err != nil {
		return false, err
	}
	if !bytes.Equal(currentContent, m.contents) {
		return false, nil
	}

	fileInfo, err := os.Stat(m.writer.PathFor(m.Path()))
	if err != nil {
		return false, err
	}
	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return false, fmt.Errorf("failed to retrieve UID and GID")
	}

	uid, err := safecast.ToUint32(m.uid)
	if err != nil {
		return false, err
	}

	gid, err := safecast.ToUint32(m.gid)
	if err != nil {
		return false, err
	}

	// compare file ownership
	if stat.Uid != uid || stat.Gid != gid {
		return false, nil
	}

	// compare file permissions
	if fileInfo.Mode().Perm() != m.perms.Perm() {
		return false, nil
	}

	return true, nil
}

func (m *managedFile) Path() string {
	return m.file.Path
}

func (m *managedFile) Exists() (bool, error) {
	return m.exists, nil
}

func (m *managedFile) IsUpToDate() (bool, error) {
	if err := m.decodeFile(); err != nil {
		return false, err
	}
	if m.exists && m.size == int64(len(m.contents)) {
		isUpToDate, err := m.isUpToDate()
		if err != nil {
			return false, err
		}
		if isUpToDate {
			return true, nil
		}
	}
	return false, nil
}

func (m *managedFile) Write() error {
	if err := m.decodeFile(); err != nil {
		return err
	}

	mode, err := intToFileMode(m.file.Mode)
	if err != nil {
		return fmt.Errorf("failed to retrieve file permissions for file %q: %w", m.Path(), err)
	}

	// set chown if file information is provided
	uid, gid, err := getFileOwnership(m.file)
	if err != nil {
		return fmt.Errorf("failed to retrieve file ownership for file %q: %w", m.Path(), err)
	}

	return m.writer.WriteFile(m.Path(), m.contents, mode, fileio.WithGid(gid), fileio.WithUid(uid))
}

func intToFileMode(i *int) (os.FileMode, error) {
	mode := fileio.DefaultFilePermissions
	if i != nil {
		filemode, err := safecast.ToUint32(*i)
		if err != nil {
			return 0, err
		}

		// Go stores setuid/setgid/sticky differently, so we
		// strip them off and then add them back
		mode = os.FileMode(filemode).Perm()
		if *i&0o1000 != 0 {
			mode = mode | os.ModeSticky
		}
		if *i&0o2000 != 0 {
			mode = mode | os.ModeSetgid
		}
		if *i&0o4000 != 0 {
			mode = mode | os.ModeSetuid
		}
	}
	return mode, nil
}

// This is essentially ResolveNodeUidAndGid() from Ignition; XXX should dedupe
func getFileOwnership(file v1beta1.FileSpec) (int, int, error) {
	uid, gid := 0, 0 // default to root
	var err error
	user := lo.FromPtr(file.User)
	if user != "" {
		uid, err = userToUID(user)
		if err != nil {
			return uid, gid, err
		}
	}

	group := lo.FromPtr(file.Group)
	if group != "" {
		gid, err = groupToGID(*file.Group)
		if err != nil {
			return uid, gid, err
		}
	}
	return uid, gid, nil
}

func userToUID(user string) (int, error) {
	userID, err := strconv.Atoi(user)
	if err != nil {
		uid, err := fileio.LookupUID(user)
		if err != nil {
			return 0, fmt.Errorf("failed to convert user to UID: %w", err)
		}
		return uid, nil
	}
	return userID, nil
}

func groupToGID(group string) (int, error) {
	groupID, err := strconv.Atoi(group)
	if err != nil {
		gid, err := fileio.LookupGID(group)
		if err != nil {
			return 0, fmt.Errorf("failed to convert group to GID: %w", err)
		}
		return gid, nil
	}
	return groupID, nil
}
