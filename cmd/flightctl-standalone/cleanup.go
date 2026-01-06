package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/flightctl/flightctl/internal/quadlet"
	"github.com/flightctl/flightctl/pkg/executer"
	"github.com/flightctl/flightctl/pkg/fileio"
	"github.com/flightctl/flightctl/pkg/log"
	"github.com/flightctl/flightctl/pkg/podman"
	"github.com/spf13/cobra"
)

const (
	// Timeout for waiting for containers to be removed by systemd
	containerStopTimeout = 60 * time.Second
	// Polling interval for checking container status
	containerPollInterval = 2 * time.Second
)

type CleanupOptions struct {
	AcceptPrompt bool
}

func NewCleanupCommand() *cobra.Command {
	opts := &CleanupOptions{}

	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Clean up Flight Control services and artifacts",
		Long: `Stop and disable Flight Control services, then remove all associated
podman artifacts (containers, images, volumes, networks, secrets).

This operation is destructive - use with caution.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.Run()
		},
	}

	cmd.Flags().BoolVarP(&opts.AcceptPrompt, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func (o *CleanupOptions) Run() error {
	logger := log.NewPrefixLogger("cleanup")

	if os.Geteuid() != 0 {
		return fmt.Errorf("cleanup requires root privileges, please run with sudo")
	}

	if !o.AcceptPrompt {
		if !confirmCleanup() {
			fmt.Println("Cleanup cancelled.")
			return nil
		}
	}

	fmt.Println("Starting Flight Control cleanup...")

	// Create podman client (running as root, no sudo needed)
	rw := fileio.NewReadWriter()
	podmanClient := podman.NewClient(logger, &executer.CommonExecuter{}, rw)

	ctx := context.Background()

	if err := stopServices(logger); err != nil {
		logger.Warnf("Failed to stop services: %v", err)
	}

	if err := disableTarget(logger); err != nil {
		logger.Warnf("Failed to disable target: %v", err)
	}

	if err := waitForContainersToStop(logger); err != nil {
		logger.Warnf("Failed waiting for containers to stop: %v", err)
	}

	if err := removeImages(ctx, logger, podmanClient); err != nil {
		logger.Warnf("Failed to remove images: %v", err)
	}

	if err := removeVolumes(ctx, logger, podmanClient); err != nil {
		logger.Warnf("Failed to remove volumes: %v", err)
	}

	if err := removeSecrets(logger); err != nil {
		logger.Warnf("Failed to remove secrets: %v", err)
	}

	if err := removeNetwork(ctx, logger, podmanClient); err != nil {
		logger.Warnf("Failed to remove network: %v", err)
	}

	fmt.Println("Cleanup completed.")
	return nil
}

func confirmCleanup() bool {
	fmt.Println("WARNING: This will remove all Flight Control services and data.")
	fmt.Println("This operation cannot be undone.")
	fmt.Print("Are you sure you want to continue? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

func stopServices(logger *log.PrefixLogger) error {
	logger.Info("Stopping Flight Control services...")

	cmd := exec.Command("systemctl", "stop", quadlet.FlightctlTarget)
	if output, err := cmd.CombinedOutput(); err != nil {
		// It's okay if the target doesn't exist or is already stopped
		logger.Debugf("systemctl stop output: %s", string(output))
		return nil
	}

	logger.Info("Services stopped")
	return nil
}

func disableTarget(logger *log.PrefixLogger) error {
	logger.Info("Disabling Flight Control target...")

	cmd := exec.Command("systemctl", "disable", quadlet.FlightctlTarget)
	if output, err := cmd.CombinedOutput(); err != nil {
		// It's okay if the target doesn't exist
		logger.Debugf("systemctl disable output: %s", string(output))
		return nil
	}

	logger.Info("Target disabled")
	return nil
}

func waitForContainersToStop(logger *log.PrefixLogger) error {
	logger.Info("Waiting for Flight Control containers to be removed...")

	deadline := time.Now().Add(containerStopTimeout)

	for {
		containers, err := getFlightctlContainers()
		if err != nil {
			return fmt.Errorf("failed to check containers: %w", err)
		}

		if len(containers) == 0 {
			logger.Info("All containers removed")
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for containers to be removed: %v", containers)
		}

		logger.Debugf("Waiting for %d containers to be removed: %v", len(containers), containers)
		time.Sleep(containerPollInterval)
	}
}

func getFlightctlContainers() ([]string, error) {
	// Use -a to include stopped containers that haven't been fully removed yet
	cmd := exec.Command("podman", "ps", "-a", "--format", "{{.Names}}", "--filter", "name=flightctl-")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var containers []string
	for _, name := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		name = strings.TrimSpace(name)
		if name != "" {
			containers = append(containers, name)
		}
	}

	return containers, nil
}

func removeImages(ctx context.Context, logger *log.PrefixLogger, client *podman.Client) error {
	logger.Info("Removing Flight Control images...")

	// Find all .container files in the quadlet directory
	pattern := filepath.Join(quadlet.DefaultQuadletDir, "flightctl*.container")
	containerFiles, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to glob container files: %w", err)
	}

	// Track unique images to avoid duplicate removal attempts
	imageSet := make(map[string]struct{})

	for _, containerFile := range containerFiles {
		content, err := os.ReadFile(containerFile)
		if err != nil {
			logger.Warnf("Failed to read container file %s: %v", containerFile, err)
			continue
		}

		unit, err := quadlet.NewUnit(content)
		if err != nil {
			logger.Warnf("Failed to parse container file %s: %v", containerFile, err)
			continue
		}

		image, err := unit.GetImage()
		if err != nil {
			logger.Debugf("No image found in %s: %v", containerFile, err)
			continue
		}

		// Skip template references (contain {{ }})
		if strings.Contains(image, "{{") {
			logger.Debugf("Skipping template image reference in %s: %s", containerFile, image)
			continue
		}

		imageSet[image] = struct{}{}
	}

	// Remove each unique image using the podman client
	for image := range imageSet {
		logger.Infof("Removing image: %s", image)
		if err := client.RemoveImage(ctx, image); err != nil {
			logger.Warnf("Failed to remove image %s (may be in use): %v", image, err)
		}
	}

	return nil
}

func removeVolumes(ctx context.Context, logger *log.PrefixLogger, client *podman.Client) error {
	logger.Info("Removing Flight Control volumes...")

	for _, volume := range quadlet.KnownVolumes {
		// Check if volume exists using the podman client
		if !client.VolumeExists(ctx, volume) {
			logger.Debugf("Volume %s does not exist, skipping", volume)
			continue
		}

		logger.Infof("Removing volume: %s", volume)
		if err := client.RemoveVolumes(ctx, volume); err != nil {
			logger.Warnf("Failed to remove volume %s: %v", volume, err)
		}
	}

	return nil
}

func removeNetwork(ctx context.Context, logger *log.PrefixLogger, client *podman.Client) error {
	logger.Info("Removing Flight Control network...")

	// Check if network exists by trying to list it
	networks, err := client.ListNetworks(ctx, nil, []string{fmt.Sprintf("name=%s", quadlet.FlightctlNetwork)})
	if err != nil {
		logger.Debugf("Failed to check network %s: %v", quadlet.FlightctlNetwork, err)
		return nil
	}

	if len(networks) == 0 {
		logger.Debugf("Network %s does not exist, skipping", quadlet.FlightctlNetwork)
		return nil
	}

	logger.Infof("Removing network: %s", quadlet.FlightctlNetwork)
	if err := client.RemoveNetworks(ctx, quadlet.FlightctlNetwork); err != nil {
		logger.Warnf("Failed to remove network %s: %v", quadlet.FlightctlNetwork, err)
	}

	return nil
}

func removeSecrets(logger *log.PrefixLogger) error {
	logger.Info("Removing Flight Control secrets...")

	for _, secret := range quadlet.KnownSecrets {
		// Check if secret exists
		inspectCmd := exec.Command("podman", "secret", "inspect", secret)
		if err := inspectCmd.Run(); err != nil {
			// Secret doesn't exist, skip
			logger.Debugf("Secret %s does not exist, skipping", secret)
			continue
		}

		logger.Infof("Removing secret: %s", secret)
		rmCmd := exec.Command("podman", "secret", "rm", secret)
		if output, err := rmCmd.CombinedOutput(); err != nil {
			logger.Warnf("Failed to remove secret %s: %s", secret, string(output))
		}
	}

	return nil
}
