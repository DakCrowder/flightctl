package podman

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// GetVersion returns the major and minor versions of podman.
func (c *Client) GetVersion(ctx context.Context) (*Version, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	args := []string{"--version"}
	stdout, stderr, exitCode := c.exec.ExecuteWithContext(ctx, podmanCmd, args...)
	if exitCode != 0 {
		return nil, fmt.Errorf("podman --version: %w", FromStderr(stderr, exitCode))
	}

	// Example: "podman version 5.4.2"
	fields := strings.Fields(stdout)
	if len(fields) < 3 {
		return nil, fmt.Errorf("unexpected podman version output: %q", stdout)
	}

	versionStr := fields[len(fields)-1]
	parts := strings.SplitN(versionStr, ".", 3)

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("parse major version: %w", err)
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("parse minor version: %w", err)
	}

	return &Version{Major: major, Minor: minor}, nil
}

// EnsureArtifactSupport verifies the local podman version can execute artifact commands.
func (c *Client) EnsureArtifactSupport(ctx context.Context) error {
	version, err := c.GetVersion(ctx)
	if err != nil {
		return fmt.Errorf("%w: checking podman version: %w", ErrNoRetry, err)
	}
	if !version.GreaterOrEqual(5, 5) {
		return fmt.Errorf("%w: OCI artifact operations require podman >= 5.5, found %d.%d", ErrNoRetry, version.Major, version.Minor)
	}
	return nil
}

// GetImageCopyTmpDir returns the image copy tmp dir exposed by the podman info API.
func (c *Client) GetImageCopyTmpDir(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	args := []string{"info", "--format", "{{.Store.ImageCopyTmpDir}}"}
	stdout, stderr, exitCode := c.exec.ExecuteWithContext(ctx, podmanCmd, args...)
	if exitCode != 0 {
		return "", fmt.Errorf("get image copy tmpdir: %w", FromStderr(stderr, exitCode))
	}

	tmpDir := strings.TrimSpace(stdout)
	return tmpDir, nil
}
