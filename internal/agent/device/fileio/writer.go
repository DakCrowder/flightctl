package fileio

import (
	"encoding/base64"
	"fmt"
	"os/user"
	"strconv"

	"github.com/flightctl/flightctl/api/v1beta1"
	pkgfileio "github.com/flightctl/flightctl/pkg/fileio"
	"github.com/samber/lo"
)

// DecodeContent decodes the content based on the encoding type and returns the
// decoded content as a byte slice.
func DecodeContent(content string, encoding *v1beta1.EncodingType) ([]byte, error) {
	if encoding == nil || *encoding == "plain" {
		return []byte(content), nil
	}

	switch *encoding {
	case "base64":
		decoded, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 content: %w", err)
		}
		return decoded, nil
	default:
		return nil, fmt.Errorf("unsupported content encoding: %q", *encoding)
	}
}

// getFileOwnership extracts UID and GID from a FileSpec.
// This is essentially ResolveNodeUidAndGid() from Ignition.
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
		uid, err := pkgfileio.LookupUID(user)
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
		gid, err := pkgfileio.LookupGID(group)
		if err != nil {
			return 0, fmt.Errorf("failed to convert group to GID: %w", err)
		}
		return gid, nil
	}
	return groupID, nil
}

// getUserIdentity returns the current user's UID and GID.
// This is re-exported for tests that need to set file ownership.
func getUserIdentity() (int, int, error) {
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
