package podman

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flightctl/flightctl/pkg/fileio"
)

// Mount mounts an image and returns the mount point.
func (c *Client) Mount(ctx context.Context, image string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	args := []string{
		"image",
		"mount",
		image,
	}
	stdout, stderr, exitCode := c.exec.ExecuteWithContext(ctx, podmanCmd, args...)
	if exitCode != 0 {
		return "", fmt.Errorf("mount image: %s: %w", image, FromStderr(stderr, exitCode))
	}

	out := strings.TrimSpace(stdout)
	return out, nil
}

// Unmount unmounts an image.
func (c *Client) Unmount(ctx context.Context, image string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	args := []string{
		"image",
		"unmount",
		image,
	}
	_, stderr, exitCode := c.exec.ExecuteWithContext(ctx, podmanCmd, args...)
	if exitCode != 0 {
		return fmt.Errorf("unmount image: %s: %w", image, FromStderr(stderr, exitCode))
	}
	return nil
}

// Copy copies files from a container to a destination.
func (c *Client) Copy(ctx context.Context, src, dst string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	args := []string{"cp", src, dst}
	_, stderr, exitCode := c.exec.ExecuteWithContext(ctx, podmanCmd, args...)
	if exitCode != 0 {
		return fmt.Errorf("copy %s to %s: %w", src, dst, FromStderr(stderr, exitCode))
	}
	return nil
}

// Unshare executes a command in the user namespace.
func (c *Client) Unshare(ctx context.Context, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	args = append([]string{"unshare"}, args...)
	stdout, stderr, exitCode := c.exec.ExecuteWithContext(ctx, podmanCmd, args...)
	if exitCode != 0 {
		return "", fmt.Errorf("unshare: %w", FromStderr(stderr, exitCode))
	}
	out := strings.TrimSpace(stdout)
	return out, nil
}

// CopyContainerData mounts an image and copies its contents to the destination path.
func (c *Client) CopyContainerData(ctx context.Context, image, destPath string) error {
	return copyContainerData(ctx, c.log, c.readWriter, c, image, destPath)
}

// IsPodmanRootless returns true if podman is running in rootless mode.
func IsPodmanRootless() bool {
	return os.Geteuid() != 0
}

func copyContainerData(ctx context.Context, log logger, writer fileio.Writer, podman *Client, image, destPath string) (err error) {
	var mountPoint string

	rootless := IsPodmanRootless()
	if rootless {
		log.Warnf("Running in rootless mode this is for testing only")
		mountPoint, err = podman.Unshare(ctx, "podman", "image", "mount", image)
		if err != nil {
			return fmt.Errorf("failed to execute podman share: %w", err)
		}
	} else {
		mountPoint, err = podman.Mount(ctx, image)
		if err != nil {
			return fmt.Errorf("failed to mount image: %w", err)
		}
	}

	if err := writer.MkdirAll(destPath, fileio.DefaultDirectoryPermissions); err != nil {
		return fmt.Errorf("failed to dest create directory: %w", err)
	}

	defer func() {
		if err := podman.Unmount(ctx, image); err != nil {
			log.Errorf("failed to unmount image: %s %v", image, err)
		}
	}()

	// recursively copy image files to agent destination
	if err := copyData(ctx, log, writer, mountPoint, destPath); err != nil {
		return fmt.Errorf("error during copy: %w", err)
	}

	return nil
}

func copyData(ctx context.Context, log logger, writer fileio.Writer, srcRoot, destRoot string) error {
	walkRoot := writer.PathFor(srcRoot)
	return filepath.Walk(walkRoot, func(walkedSrc string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}

		relPath, err := filepath.Rel(walkRoot, walkedSrc)
		if err != nil {
			return fmt.Errorf("computing relative path: %w", err)
		}

		realSrc := filepath.Join(srcRoot, relPath)
		relDest := filepath.Join(destRoot, relPath)

		if info.IsDir() {
			if info.Name() == "merged" {
				log.Tracef("Skipping merged directory: %s", walkedSrc)
				return nil
			}

			// create the directory in the destination
			log.Tracef("Creating directory: %s", relDest)

			// ensure any directories in the image are also created
			return writer.MkdirAll(relDest, fileio.DefaultDirectoryPermissions)
		}

		log.Tracef("Copying file from %s to %s", realSrc, relDest)
		return writer.CopyFile(realSrc, relDest)
	})
}

// logger interface for mount operations.
type logger interface {
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Tracef(format string, args ...interface{})
}
