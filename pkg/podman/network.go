package podman

import (
	"context"
	"fmt"
	"strings"
)

// ListNetworks returns a list of network IDs matching the given labels and filters.
func (c *Client) ListNetworks(ctx context.Context, labels []string, filters []string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	args := []string{
		"network",
		"ls",
		"--format",
		"{{.Network.ID}}",
	}
	args = applyFilters(args, labels, filters)

	stdout, stderr, exitCode := c.exec.ExecuteWithContext(ctx, podmanCmd, args...)
	if exitCode != 0 {
		return nil, fmt.Errorf("list networks: %w", FromStderr(stderr, exitCode))
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	networkSeen := make(map[string]struct{})
	for _, line := range lines {
		// handle multiple networks comma separated
		networks := strings.Split(line, ",")
		for _, network := range networks {
			network = strings.TrimSpace(network)
			if network != "" {
				networkSeen[network] = struct{}{}
			}
		}
	}

	var networks []string
	for network := range networkSeen {
		networks = append(networks, network)
	}
	return networks, nil
}

// RemoveNetworks removes the specified networks.
func (c *Client) RemoveNetworks(ctx context.Context, networks ...string) error {
	for _, network := range networks {
		nctx, cancel := context.WithTimeout(ctx, c.timeout)
		args := []string{"network", "rm", network}
		_, stderr, exitCode := c.exec.ExecuteWithContext(nctx, podmanCmd, args...)
		cancel()
		if exitCode != 0 {
			return fmt.Errorf("remove networks: %w", FromStderr(stderr, exitCode))
		}
		c.log.Infof("Removed network %s", network)
	}
	return nil
}
