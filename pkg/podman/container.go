package podman

import (
	"context"
	"fmt"
)

// StopContainers stops containers matching the given labels.
func (c *Client) StopContainers(ctx context.Context, labels []string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	args := []string{"stop"}
	for _, label := range labels {
		args = append(args, "--filter", fmt.Sprintf("label=%s", label))
	}
	_, stderr, exitCode := c.exec.ExecuteWithContext(ctx, podmanCmd, args...)
	if exitCode != 0 {
		return fmt.Errorf("stop containers: %w", FromStderr(stderr, exitCode))
	}
	return nil
}

// RemoveContainer removes containers matching the given labels.
func (c *Client) RemoveContainer(ctx context.Context, labels []string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	args := []string{"rm"}
	for _, label := range labels {
		args = append(args, "--filter", fmt.Sprintf("label=%s", label))
	}
	_, stderr, exitCode := c.exec.ExecuteWithContext(ctx, podmanCmd, args...)
	if exitCode != 0 {
		return fmt.Errorf("remove containers: %w", FromStderr(stderr, exitCode))
	}
	return nil
}
