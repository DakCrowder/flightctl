package renderer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/flightctl/flightctl/pkg/fileio"
	"github.com/sirupsen/logrus"
)

type ActionType int

const (
	ActionCopyFile ActionType = iota
	ActionCopyDir
	ActionCopyBinary
	ActionCreateEmptyFile
	ActionCreateEmptyDir
)

const (
	RegularFileMode    os.FileMode = 0644 // Regular files
	ExecutableFileMode os.FileMode = 0755 // Executable files and directories
)

type InstallAction struct {
	Action      ActionType
	Source      string
	Destination string
	Template    bool
	Mode        os.FileMode
}

type ImageConfig struct {
	Image string `mapstructure:"image"`
	Tag   string `mapstructure:"tag"`
}

type RendererConfig struct {
	// Output directories
	ReadOnlyConfigOutputDir  string `mapstructure:"readonly-config-dir"`
	WriteableConfigOutputDir string `mapstructure:"writeable-config-dir"`
	QuadletFilesOutputDir    string `mapstructure:"quadlet-dir"`
	SystemdUnitOutputDir     string `mapstructure:"systemd-dir"`
	BinOutputDir             string `mapstructure:"bin-dir"`

	// Source directories for binary search
	BinSourceDirs []string `mapstructure:"bin-source-dirs"`

	FlightctlServicesTagOverride string `mapstructure:"flightctl-services-tag-override"`
	FlightctlUiTagOverride       bool   `mapstructure:"flightctl-ui-tag-override"`

	// Images
	Api               ImageConfig `mapstructure:"api"`
	Periodic          ImageConfig `mapstructure:"periodic"`
	Worker            ImageConfig `mapstructure:"worker"`
	AlertExporter     ImageConfig `mapstructure:"alert-exporter"`
	CliArtifacts      ImageConfig `mapstructure:"cli-artifacts"`
	AlertmanagerProxy ImageConfig `mapstructure:"alertmanager-proxy"`
	PamIssuer         ImageConfig `mapstructure:"pam-issuer"`
	Ui                ImageConfig `mapstructure:"ui"`
	DbSetup           ImageConfig `mapstructure:"db-setup"`
	Db                ImageConfig `mapstructure:"db"`
	Kv                ImageConfig `mapstructure:"kv"`
	Alertmanager      ImageConfig `mapstructure:"alertmanager"`
	ImagebuilderApi   ImageConfig `mapstructure:"imagebuilder-api"`
}

func NewRendererConfig() *RendererConfig {
	return &RendererConfig{
		ReadOnlyConfigOutputDir:  "/usr/share/flightctl",
		WriteableConfigOutputDir: "/etc/flightctl",
		QuadletFilesOutputDir:    "/usr/share/containers/systemd",
		SystemdUnitOutputDir:     "/usr/lib/systemd/system",
		BinOutputDir:             "/usr/bin",
	}
}

func findBinarySource(rw fileio.ReadWriter, binaryName string, searchDirs []string) (string, error) {
	for _, dir := range searchDirs {
		path := filepath.Join(dir, binaryName)
		exists, err := rw.PathExists(path)
		if err != nil {
			continue
		}
		if exists {
			return path, nil
		}
	}
	return "", fmt.Errorf("binary %q not found in directories: %v", binaryName, searchDirs)
}

func processInstallManifest(rw fileio.ReadWriter, manifest []InstallAction, config *RendererConfig, log logrus.FieldLogger) error {
	for _, action := range manifest {
		switch action.Action {
		case ActionCopyFile:
			if err := processFile(rw, action.Source, action.Destination, action.Template, action.Mode, config); err != nil {
				return fmt.Errorf("failed to process file %s: %w", action.Source, err)
			}
			log.Infof("Processed file: %s -> %s (template=%t)", action.Source, action.Destination, action.Template)

		case ActionCopyDir:
			if err := copyDir(rw, action.Source, action.Destination, action.Mode); err != nil {
				return fmt.Errorf("failed to copy directory %s to %s: %w", action.Source, action.Destination, err)
			}
			log.Infof("Copied directory: %s -> %s", action.Source, action.Destination)

		case ActionCopyBinary:
			sourcePath, err := findBinarySource(rw, action.Source, config.BinSourceDirs)
			if err != nil {
				return fmt.Errorf("failed to find binary %s: %w", action.Source, err)
			}
			if err := processFile(rw, sourcePath, action.Destination, action.Template, action.Mode, config); err != nil {
				return fmt.Errorf("failed to process binary %s: %w", action.Source, err)
			}
			log.Infof("Processed binary: %s -> %s (found at %s)", action.Source, action.Destination, sourcePath)

		case ActionCreateEmptyFile:
			if err := createEmptyFile(rw, action.Destination, action.Mode, log); err != nil {
				return fmt.Errorf("failed to create empty file %s: %w", action.Destination, err)
			}
			log.Infof("Created empty file: %s", action.Destination)

		case ActionCreateEmptyDir:
			if err := createEmptyDirectory(rw, action.Destination, action.Mode, log); err != nil {
				return fmt.Errorf("failed to create empty directory %s: %w", action.Destination, err)
			}
			log.Infof("Created empty directory: %s", action.Destination)

		default:
			return fmt.Errorf("unknown action type: %v", action.Action)
		}
	}
	return nil
}

func processFile(rw fileio.ReadWriter, sourcePath, destPath string, isTemplate bool, mode os.FileMode, config *RendererConfig) error {
	content, err := rw.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	var finalContent []byte
	if isTemplate {
		tmpl, err := template.New(filepath.Base(sourcePath)).Parse(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse template: %w", err)
		}

		var buf strings.Builder
		if err := tmpl.Execute(&buf, config); err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}
		finalContent = []byte(buf.String())
	} else {
		finalContent = content
	}

	if err := rw.WriteFile(destPath, finalContent, mode); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
}

func createEmptyFile(rw fileio.ReadWriter, destPath string, mode os.FileMode, log logrus.FieldLogger) error {
	exists, err := rw.PathExists(destPath, fileio.WithSkipContentCheck())
	if err != nil {
		return fmt.Errorf("failed to check if file exists: %w", err)
	}
	if exists {
		log.Infof("File already exists, skipping: %s", destPath)
		return nil
	}

	if err := rw.WriteFile(destPath, []byte{}, mode); err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	return nil
}

func createEmptyDirectory(rw fileio.ReadWriter, destPath string, mode os.FileMode, log logrus.FieldLogger) error {
	exists, err := rw.PathExists(destPath, fileio.WithSkipContentCheck())
	if err != nil {
		return fmt.Errorf("failed to check if directory exists: %w", err)
	}
	if exists {
		log.Infof("Directory already exists, skipping: %s", destPath)
		return nil
	}

	if err := rw.MkdirAll(destPath, mode); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return nil
}

func copyDir(rw fileio.ReadWriter, src, dst string, mode os.FileMode) error {
	if err := rw.CopyDir(src, dst); err != nil {
		return err
	}

	return filepath.Walk(rw.PathFor(dst), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return os.Chmod(path, ExecutableFileMode)
		}
		return os.Chmod(path, mode)
	})
}

func (config *RendererConfig) ApplyFlightctlServicesTagOverride(log logrus.FieldLogger) {
	if config.FlightctlServicesTagOverride == "" {
		return
	}

	tag := config.FlightctlServicesTagOverride
	log.Infof("Applying flightctl services tag override: %s", tag)

	config.Api.Tag = tag
	config.Periodic.Tag = tag
	config.Worker.Tag = tag
	config.AlertExporter.Tag = tag
	config.CliArtifacts.Tag = tag
	config.AlertmanagerProxy.Tag = tag
	config.PamIssuer.Tag = tag
	config.DbSetup.Tag = tag

	if config.FlightctlUiTagOverride {
		// For release builds, UI tag must be overridden
		log.Infof("Applying tag override to UI service: %s", tag)
		config.Ui.Tag = tag
	} else {
		// For development builds, UI tag is kept as defined in images.yaml
		log.Infof("Skipping UI tag override (keeping value from images.yaml: %s)", config.Ui.Tag)
	}
}

// RenderQuadlets orchestrates all installation operations
func RenderQuadlets(rw fileio.ReadWriter, config *RendererConfig, log logrus.FieldLogger) error {
	log.Info("Starting installation")

	config.ApplyFlightctlServicesTagOverride(log)

	manifest := servicesManifest(config)
	if err := processInstallManifest(rw, manifest, config, log); err != nil {
		return fmt.Errorf("failed to process install manifest: %w", err)
	}

	log.Info("Installation complete")
	return nil
}
