package podman

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// PullArtifact pulls an artifact from the registry with optional retry and authentication.
// Logs progress periodically while the operation is in progress.
func (c *Client) PullArtifact(ctx context.Context, artifact string, opts ...CallOption) (string, error) {
	options := &callOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if c.backoff != nil {
		return logProgress(ctx, c.log, "Pulling artifact, please wait...", func(ctx context.Context) (string, error) {
			return retryWithBackoff(ctx, c.log, *c.backoff, func(ctx context.Context) (string, error) {
				return c.pullArtifact(ctx, artifact, options)
			})
		})
	}
	return c.pullArtifact(ctx, artifact, options)
}

func (c *Client) pullArtifact(ctx context.Context, artifact string, options *callOptions) (string, error) {
	timeout := c.timeout
	if options.timeout > 0 {
		timeout = options.timeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := c.EnsureArtifactSupport(ctx); err != nil {
		return "", err
	}

	pullSecretPath := options.pullSecretPath
	args := []string{"artifact", "pull", artifact}
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
		return "", fmt.Errorf("pull artifact: %w", FromStderr(stderr, exitCode))
	}
	return strings.TrimSpace(stdout), nil
}

// ExtractArtifact extracts an artifact to the given destination directory.
func (c *Client) ExtractArtifact(ctx context.Context, artifact, destination string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	if err := c.EnsureArtifactSupport(ctx); err != nil {
		return "", err
	}

	args := []string{"artifact", "extract", artifact, destination}
	stdout, stderr, exitCode := c.exec.ExecuteWithContext(ctx, podmanCmd, args...)
	if exitCode != 0 {
		return "", fmt.Errorf("artifact extract: %w", FromStderr(stderr, exitCode))
	}
	out := strings.TrimSpace(stdout)
	return out, nil
}

// ArtifactExists returns true if the artifact exists in storage.
func (c *Client) ArtifactExists(ctx context.Context, artifact string) bool {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	args := []string{"artifact", "inspect", artifact}
	_, _, exitCode := c.exec.ExecuteWithContext(ctx, podmanCmd, args...)
	return exitCode == 0
}

func (c *Client) artifactInspect(ctx context.Context, reference string) (*ArtifactInspect, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	args := []string{"artifact", "inspect", reference}
	stdout, stderr, exitCode := c.exec.ExecuteWithContext(ctx, podmanCmd, args...)
	if exitCode != 0 {
		return nil, fmt.Errorf("artifact inspect: %w", FromStderr(stderr, exitCode))
	}

	var inspectResult ArtifactInspect
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &inspectResult); err != nil {
		return nil, fmt.Errorf("unmarshal artifact inspect output: %w", err)
	}

	return &inspectResult, nil
}

// ArtifactDigest returns the digest of the specified artifact.
func (c *Client) ArtifactDigest(ctx context.Context, reference string) (string, error) {
	if err := c.EnsureArtifactSupport(ctx); err != nil {
		return "", err
	}

	inspect, err := c.artifactInspect(ctx, reference)
	if err != nil {
		return "", err
	}

	if inspect.Digest == "" {
		return "", fmt.Errorf("artifact digest empty for %s", reference)
	}

	return inspect.Digest, nil
}

// InspectArtifactAnnotations inspects an OCI artifact and returns its annotations map.
func (c *Client) InspectArtifactAnnotations(ctx context.Context, reference string) (map[string]string, error) {
	if err := c.EnsureArtifactSupport(ctx); err != nil {
		return nil, err
	}

	inspect, err := c.artifactInspect(ctx, reference)
	if err != nil {
		return nil, err
	}

	return extractArtifactAnnotations(inspect), nil
}

// ListArtifacts returns a list of all OCI artifacts stored on the device.
func (c *Client) ListArtifacts(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	if err := c.EnsureArtifactSupport(ctx); err != nil {
		return nil, err
	}

	args := []string{"artifact", "ls", "--format", "{{.Name}}"}
	stdout, stderr, exitCode := c.exec.ExecuteWithContext(ctx, podmanCmd, args...)
	if exitCode != 0 {
		return nil, fmt.Errorf("list artifacts: %w", FromStderr(stderr, exitCode))
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	artifactsSeen := make(map[string]struct{})
	artifacts := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if _, ok := artifactsSeen[line]; !ok {
			artifactsSeen[line] = struct{}{}
			artifacts = append(artifacts, line)
		}
	}

	return artifacts, nil
}

// RemoveArtifact removes the specified OCI artifact from Podman.
// Returns an error if the removal fails. Handles non-existent artifacts gracefully.
func (c *Client) RemoveArtifact(ctx context.Context, artifact string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	if err := c.EnsureArtifactSupport(ctx); err != nil {
		return err
	}

	args := []string{"artifact", "rm", artifact}
	_, stderr, exitCode := c.exec.ExecuteWithContext(ctx, podmanCmd, args...)
	if exitCode != 0 {
		if strings.Contains(stderr, noSuchArtifactErrorSubstring) || strings.Contains(stderr, artifactNotKnownErrorSubstring) {
			return nil
		}
		return fmt.Errorf("remove artifact %s: %w", artifact, FromStderr(stderr, exitCode))
	}
	return nil
}

// extractArtifactAnnotations parses the artifact inspect and extracts annotations.
func extractArtifactAnnotations(inspect *ArtifactInspect) map[string]string {
	annotations := make(map[string]string)
	for _, layer := range inspect.Manifest.Layers {
		for key, value := range layer.Annotations {
			annotations[key] = value
		}
	}
	return annotations
}
