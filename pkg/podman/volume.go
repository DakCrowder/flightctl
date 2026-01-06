package podman

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// CreateVolume creates a volume with the given name and labels.
// Returns the mount point of the created volume.
func (c *Client) CreateVolume(ctx context.Context, name string, labels []string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	args := []string{"volume", "create", name}
	for _, label := range labels {
		args = append(args, "--label", label)
	}

	stdout, stderr, exitCode := c.exec.ExecuteWithContext(ctx, podmanCmd, args...)
	if exitCode != 0 {
		return "", fmt.Errorf("create volume: %s: %w", strings.TrimSpace(stdout), FromStderr(stderr, exitCode))
	}

	inspectArgs := []string{"volume", "inspect", name, "--format", "{{.Mountpoint}}"}
	mountpointOut, inspectStderr, inspectExit := c.exec.ExecuteWithContext(ctx, podmanCmd, inspectArgs...)
	if inspectExit != 0 {
		return "", fmt.Errorf("inspect volume mountpoint: %w", FromStderr(inspectStderr, inspectExit))
	}

	mountpoint := strings.TrimSpace(mountpointOut)
	return mountpoint, nil
}

// ListVolumes returns a list of volume names matching the given labels and filters.
func (c *Client) ListVolumes(ctx context.Context, labels []string, filters []string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	args := []string{
		"volume",
		"ls",
		"--format",
		"json",
	}
	args = applyFilters(args, labels, filters)
	stdout, stderr, exitCode := c.exec.ExecuteWithContext(ctx, podmanCmd, args...)
	if exitCode != 0 {
		return nil, fmt.Errorf("list volumes: %w", FromStderr(stderr, exitCode))
	}
	var podVols []volume
	err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &podVols)
	if err != nil {
		return nil, fmt.Errorf("unmarshal volumes: %w", err)
	}
	volumesSeen := make(map[string]struct{})
	volumes := make([]string, 0, len(podVols))
	for _, vol := range podVols {
		if _, ok := volumesSeen[vol.Name]; !ok {
			volumesSeen[vol.Name] = struct{}{}
			volumes = append(volumes, vol.Name)
		}
	}
	return volumes, nil
}

// VolumeExists returns true if the volume exists.
func (c *Client) VolumeExists(ctx context.Context, name string) bool {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	args := []string{"volume", "exists", name}
	_, _, exitCode := c.exec.ExecuteWithContext(ctx, podmanCmd, args...)
	return exitCode == 0
}

func (c *Client) inspectVolumeProperty(ctx context.Context, name string, property string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	args := []string{"volume", "inspect", name, "--format", property}
	stdout, stderr, exitCode := c.exec.ExecuteWithContext(ctx, podmanCmd, args...)
	if exitCode != 0 {
		return "", fmt.Errorf("inspect volume property %s: %w", property, FromStderr(stderr, exitCode))
	}

	return strings.TrimSpace(stdout), nil
}

// InspectVolumeDriver returns the driver of the specified volume.
func (c *Client) InspectVolumeDriver(ctx context.Context, name string) (string, error) {
	return c.inspectVolumeProperty(ctx, name, "{{.Driver}}")
}

// InspectVolumeMount returns the mount point of the specified volume.
func (c *Client) InspectVolumeMount(ctx context.Context, name string) (string, error) {
	return c.inspectVolumeProperty(ctx, name, "{{.Mountpoint}}")
}

// RemoveVolumes removes the specified volumes.
func (c *Client) RemoveVolumes(ctx context.Context, volumes ...string) error {
	for _, vol := range volumes {
		nctx, cancel := context.WithTimeout(ctx, c.timeout)
		args := []string{"volume", "rm", vol}
		_, stderr, exitCode := c.exec.ExecuteWithContext(nctx, podmanCmd, args...)
		cancel()
		if exitCode != 0 {
			return fmt.Errorf("remove volumes: %w", FromStderr(stderr, exitCode))
		}
		c.log.Infof("Removed volume %s", vol)
	}
	return nil
}
