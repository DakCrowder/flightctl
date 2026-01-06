package podman

import (
	"context"
	"fmt"
	"strings"
)

// ListPods returns a list of pod IDs that contain containers matching the given labels.
func (c *Client) ListPods(ctx context.Context, labels []string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// pods created by podman-compose don't have the compose project label,
	// so we need to get pod IDs from the containers that do have the label
	args := []string{
		"ps",
		"-a",
		"--format",
		"{{.Pod}}",
	}
	args = applyFilters(args, labels, []string{})

	stdout, stderr, exitCode := c.exec.ExecuteWithContext(ctx, podmanCmd, args...)
	if exitCode != 0 {
		return nil, fmt.Errorf("list pods: %w", FromStderr(stderr, exitCode))
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	podSeen := make(map[string]struct{})
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// skip empty lines and containers not in a pod
		if line == "" || line == "--" {
			continue
		}
		podSeen[line] = struct{}{}
	}

	var pods []string
	for pod := range podSeen {
		pods = append(pods, pod)
	}
	return pods, nil
}

// RemovePods removes the specified pods.
func (c *Client) RemovePods(ctx context.Context, pods ...string) error {
	for _, pod := range pods {
		nctx, cancel := context.WithTimeout(ctx, c.timeout)
		args := []string{"pod", "rm", pod}
		_, stderr, exitCode := c.exec.ExecuteWithContext(nctx, podmanCmd, args...)
		cancel()
		if exitCode != 0 {
			return fmt.Errorf("remove pods: %w", FromStderr(stderr, exitCode))
		}
		c.log.Infof("Removed pod %s", pod)
	}
	return nil
}
