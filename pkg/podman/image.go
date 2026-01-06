package podman

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Pull pulls an image from the registry with optional retry and authentication via a pull secret.
// Logs progress periodically while the operation is in progress.
func (c *Client) Pull(ctx context.Context, image string, opts ...CallOption) (string, error) {
	options := &callOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if c.backoff != nil {
		return logProgress(ctx, c.log, "Pulling image, please wait...", func(ctx context.Context) (string, error) {
			return retryWithBackoff(ctx, c.log, *c.backoff, func(ctx context.Context) (string, error) {
				return c.pullImage(ctx, image, options)
			})
		})
	}
	return c.pullImage(ctx, image, options)
}

func (c *Client) pullImage(ctx context.Context, image string, options *callOptions) (string, error) {
	timeout := c.timeout
	if options.timeout > 0 {
		timeout = options.timeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	pullSecretPath := options.pullSecretPath
	args := []string{"pull", image}
	if pullSecretPath != "" {
		exists, err := c.readWriter.PathExists(pullSecretPath)
		if err != nil {
			return "", fmt.Errorf("check pull secret path: %w", err)
		}
		if !exists {
			c.log.Errorf("Pull secret path %s does not exist", pullSecretPath)
		} else {
			args = append(args, "--authfile", pullSecretPath)
		}
	}
	stdout, stderr, exitCode := c.exec.ExecuteWithContext(ctx, podmanCmd, args...)
	if exitCode != 0 {
		return "", fmt.Errorf("pull image: %w", FromStderr(stderr, exitCode))
	}
	out := strings.TrimSpace(stdout)
	return out, nil
}

// Inspect returns the JSON output of the image inspection.
// The expectation is that the image exists in local container storage.
func (c *Client) Inspect(ctx context.Context, image string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	args := []string{"inspect", image}
	stdout, stderr, exitCode := c.exec.ExecuteWithContext(ctx, podmanCmd, args...)
	if exitCode != 0 {
		return "", fmt.Errorf("inspect image: %s: %w", image, FromStderr(stderr, exitCode))
	}
	out := strings.TrimSpace(stdout)
	return out, nil
}

// ImageExists returns true if the image exists in local storage.
func (c *Client) ImageExists(ctx context.Context, image string) bool {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	args := []string{"image", "exists", image}
	_, _, exitCode := c.exec.ExecuteWithContext(ctx, podmanCmd, args...)
	return exitCode == 0
}

// ImageDigest returns the digest of the specified image.
// Returns empty string and error if the image does not exist or cannot be inspected.
func (c *Client) ImageDigest(ctx context.Context, image string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	args := []string{"image", "inspect", "--format", "{{.Digest}}", image}
	stdout, stderr, exitCode := c.exec.ExecuteWithContext(ctx, podmanCmd, args...)
	if exitCode != 0 {
		return "", fmt.Errorf("get image digest: %s: %w", image, FromStderr(stderr, exitCode))
	}
	digest := strings.TrimSpace(stdout)
	return digest, nil
}

// ListImages returns a list of all container images stored on the device.
// Returns image references in the format "repository:tag" or image ID for untagged images.
func (c *Client) ListImages(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Use a format that handles both tagged and untagged images
	args := []string{"image", "ls", "--format", "{{if and .Repository (ne .Repository \"<none>\")}}{{.Repository}}:{{.Tag}}{{else}}{{.ID}}{{end}}"}
	stdout, stderr, exitCode := c.exec.ExecuteWithContext(ctx, podmanCmd, args...)
	if exitCode != 0 {
		return nil, fmt.Errorf("list images: %w", FromStderr(stderr, exitCode))
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	imagesSeen := make(map[string]struct{})
	images := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if _, ok := imagesSeen[line]; !ok {
			imagesSeen[line] = struct{}{}
			images = append(images, line)
		}
	}

	return images, nil
}

// RemoveImage removes the specified container image from Podman.
// Returns an error if the removal fails. Handles non-existent images gracefully.
func (c *Client) RemoveImage(ctx context.Context, image string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	args := []string{"image", "rm", image}
	_, stderr, exitCode := c.exec.ExecuteWithContext(ctx, podmanCmd, args...)
	if exitCode != 0 {
		// Check if error is due to image not existing
		if strings.Contains(stderr, noSuchImageErrorSubstring) || strings.Contains(stderr, imageNotKnownErrorSubstring) {
			return nil
		}
		return fmt.Errorf("remove image %s: %w", image, FromStderr(stderr, exitCode))
	}
	return nil
}

// InspectLabels returns the labels from an image or container.
func (c *Client) InspectLabels(ctx context.Context, image string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.Inspect(ctx, image)
	if err != nil {
		return nil, err
	}

	var inspectData []InspectResult
	if err := json.Unmarshal([]byte(resp), &inspectData); err != nil {
		return nil, fmt.Errorf("parse image inspect response: %w", err)
	}

	if len(inspectData) == 0 {
		return nil, fmt.Errorf("no image config found")
	}

	return inspectData[0].Config.Labels, nil
}
