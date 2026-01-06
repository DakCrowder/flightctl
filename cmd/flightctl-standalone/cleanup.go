package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/flightctl/flightctl/internal/quadlet"
	"github.com/flightctl/flightctl/pkg/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	// Default location for installed quadlet files
	defaultQuadletDir = "/usr/share/containers/systemd"
	// Systemd target name
	flightctlTarget = "flightctl.target"
	// Network name
	networkName = "flightctl"
	// Timeout for waiting for containers to be removed by systemd
	containerStopTimeout = 60 * time.Second
	// Polling interval for checking container status
	containerPollInterval = 2 * time.Second
)

// Known volume names from .volume files
var knownVolumes = []string{
	"flightctl-db",
	"flightctl-kv",
	"flightctl-alertmanager",
	"flightctl-ui-certs",
	"flightctl-cli-artifacts-certs",
}

// Known secret names from Secret= directives in .container files
var knownSecrets = []string{
	"flightctl-postgresql-password",
	"flightctl-postgresql-master-password",
	"flightctl-postgresql-user-password",
	"flightctl-postgresql-migrator-password",
	"flightctl-kv-password",
}

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
	logger := log.InitLogs()

	if !o.AcceptPrompt {
		if !confirmCleanup() {
			fmt.Println("Cleanup cancelled.")
			return nil
		}
	}

	fmt.Println("Starting Flight Control cleanup...")

	if err := stopServices(logger); err != nil {
		logger.Warnf("Failed to stop services: %v", err)
	}

	if err := disableTarget(logger); err != nil {
		logger.Warnf("Failed to disable target: %v", err)
	}

	if err := waitForContainersToStop(logger); err != nil {
		logger.Warnf("Failed waiting for containers to stop: %v", err)
	}

	if err := removeImages(logger); err != nil {
		logger.Warnf("Failed to remove images: %v", err)
	}

	if err := removeVolumes(logger); err != nil {
		logger.Warnf("Failed to remove volumes: %v", err)
	}

	if err := removeSecrets(logger); err != nil {
		logger.Warnf("Failed to remove secrets: %v", err)
	}

	if err := removeNetwork(logger); err != nil {
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

func stopServices(logger *logrus.Logger) error {
	logger.Info("Stopping Flight Control services...")

	cmd := exec.Command("systemctl", "stop", flightctlTarget)
	if output, err := cmd.CombinedOutput(); err != nil {
		// It's okay if the target doesn't exist or is already stopped
		logger.Debugf("systemctl stop output: %s", string(output))
		return nil
	}

	logger.Info("Services stopped")
	return nil
}

func disableTarget(logger *logrus.Logger) error {
	logger.Info("Disabling Flight Control target...")

	cmd := exec.Command("systemctl", "disable", flightctlTarget)
	if output, err := cmd.CombinedOutput(); err != nil {
		// It's okay if the target doesn't exist
		logger.Debugf("systemctl disable output: %s", string(output))
		return nil
	}

	logger.Info("Target disabled")
	return nil
}

func waitForContainersToStop(logger *logrus.Logger) error {
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
	cmd := exec.Command("sudo", "podman", "ps", "-a", "--format", "{{.Names}}", "--filter", "name=flightctl-")
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

func removeImages(logger *logrus.Logger) error {
	logger.Info("Removing Flight Control images...")

	// Find all .container files in the quadlet directory
	pattern := filepath.Join(defaultQuadletDir, "flightctl*.container")
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

	// Remove each unique image
	for image := range imageSet {
		logger.Infof("Removing image: %s", image)
		cmd := exec.Command("sudo", "podman", "rmi", image)
		if output, err := cmd.CombinedOutput(); err != nil {
			logger.Warnf("Failed to remove image %s (may be in use): %s", image, string(output))
		}
	}

	return nil
}

func removeVolumes(logger *logrus.Logger) error {
	logger.Info("Removing Flight Control volumes...")

	for _, volume := range knownVolumes {
		// Check if volume exists
		inspectCmd := exec.Command("sudo", "podman", "volume", "inspect", volume)
		if err := inspectCmd.Run(); err != nil {
			// Volume doesn't exist, skip
			logger.Debugf("Volume %s does not exist, skipping", volume)
			continue
		}

		logger.Infof("Removing volume: %s", volume)
		rmCmd := exec.Command("sudo", "podman", "volume", "rm", volume)
		if output, err := rmCmd.CombinedOutput(); err != nil {
			logger.Warnf("Failed to remove volume %s: %s", volume, string(output))
		}
	}

	return nil
}

func removeNetwork(logger *logrus.Logger) error {
	logger.Info("Removing Flight Control network...")

	// Check if network exists
	inspectCmd := exec.Command("sudo", "podman", "network", "inspect", networkName)
	if err := inspectCmd.Run(); err != nil {
		logger.Debugf("Network %s does not exist, skipping", networkName)
		return nil
	}

	logger.Infof("Removing network: %s", networkName)
	rmCmd := exec.Command("sudo", "podman", "network", "rm", networkName)
	if output, err := rmCmd.CombinedOutput(); err != nil {
		logger.Warnf("Failed to remove network %s: %s", networkName, string(output))
	}

	return nil
}

func removeSecrets(logger *logrus.Logger) error {
	logger.Info("Removing Flight Control secrets...")

	for _, secret := range knownSecrets {
		// Check if secret exists
		inspectCmd := exec.Command("sudo", "podman", "secret", "inspect", secret)
		if err := inspectCmd.Run(); err != nil {
			// Secret doesn't exist, skip
			logger.Debugf("Secret %s does not exist, skipping", secret)
			continue
		}

		logger.Infof("Removing secret: %s", secret)
		rmCmd := exec.Command("sudo", "podman", "secret", "rm", secret)
		if output, err := rmCmd.CombinedOutput(); err != nil {
			logger.Warnf("Failed to remove secret %s: %s", secret, string(output))
		}
	}

	return nil
}
