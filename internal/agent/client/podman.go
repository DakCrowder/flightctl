// Package client provides agent-specific client implementations that wrap pkg/ packages.
package client

import (
	"context"
	"io/fs"
	"os/exec"
	"time"

	"github.com/flightctl/flightctl/internal/agent/device/fileio"
	"github.com/flightctl/flightctl/pkg/executer"
	pkgfileio "github.com/flightctl/flightctl/pkg/fileio"
	"github.com/flightctl/flightctl/pkg/log"
	"github.com/flightctl/flightctl/pkg/podman"
	"github.com/flightctl/flightctl/pkg/poll"
)

const (
	podmanCmd = "podman"
)

// Re-export types from pkg/podman for backward compatibility
type (
	PodmanInspect         = podman.InspectResult
	PodmanContainerState  = podman.ContainerState
	PodmanContainerConfig = podman.ContainerConfig
	ArtifactInspect       = podman.ArtifactInspect
	ArtifactManifest      = podman.ArtifactManifest
	ArtifactLayer         = podman.ArtifactLayer
	PodmanEvent           = podman.Event
	PodmanVersion         = podman.Version
)

// Re-export functions and variables from pkg/podman
var (
	SanitizePodmanLabel = podman.SanitizePodmanLabel
	IsPodmanRootless    = podman.IsPodmanRootless
)

// Podman wraps pkg/podman.Client with agent-specific functionality.
type Podman struct {
	*podman.Client
	exec       executer.Executer
	timeout    time.Duration
	readWriter fileio.ReadWriter
	log        *log.PrefixLogger
}

// NewPodman creates a new agent-specific Podman client.
func NewPodman(log *log.PrefixLogger, exec executer.Executer, readWriter fileio.ReadWriter, backoff poll.Config) *Podman {
	// Create pkg/fileio ReadWriter adapter for pkg/podman
	pkgRW := &pkgFileioAdapter{rw: readWriter}

	return &Podman{
		Client:     podman.NewClient(log, exec, pkgRW, podman.WithBackoff(backoff)),
		exec:       exec,
		timeout:    defaultPodmanTimeout,
		readWriter: readWriter,
		log:        log,
	}
}

// Pull pulls an image with agent-specific options.
func (p *Podman) Pull(ctx context.Context, image string, opts ...ClientOption) (string, error) {
	options := &clientOptions{}
	for _, opt := range opts {
		opt(options)
	}

	var callOpts []podman.CallOption
	if options.pullSecretPath != "" {
		callOpts = append(callOpts, podman.WithPullSecret(options.pullSecretPath))
	}
	if options.timeout > 0 {
		callOpts = append(callOpts, podman.Timeout(options.timeout))
	}

	return p.Client.Pull(ctx, image, callOpts...)
}

// PullArtifact pulls an artifact with agent-specific options.
func (p *Podman) PullArtifact(ctx context.Context, artifact string, opts ...ClientOption) (string, error) {
	options := &clientOptions{}
	for _, opt := range opts {
		opt(options)
	}

	var callOpts []podman.CallOption
	if options.pullSecretPath != "" {
		callOpts = append(callOpts, podman.WithPullSecret(options.pullSecretPath))
	}
	if options.timeout > 0 {
		callOpts = append(callOpts, podman.Timeout(options.timeout))
	}

	return p.Client.PullArtifact(ctx, artifact, callOpts...)
}

// Version returns the podman version (wraps GetVersion for backward compatibility).
func (p *Podman) Version(ctx context.Context) (*PodmanVersion, error) {
	return p.Client.GetVersion(ctx)
}

// EnsureArtifactSupport verifies the local podman version can execute artifact commands.
func (p *Podman) EnsureArtifactSupport(ctx context.Context) error {
	return p.Client.EnsureArtifactSupport(ctx)
}

// EventsSinceCmd returns a command to get podman events since the given time.
func (p *Podman) EventsSinceCmd(ctx context.Context, events []string, sinceTime string) *exec.Cmd {
	return p.Client.EventsSinceCmd(ctx, events, sinceTime)
}

// CopyContainerData mounts an image and copies its contents to the destination path.
func (p *Podman) CopyContainerData(ctx context.Context, image, destPath string) error {
	return p.Client.CopyContainerData(ctx, image, destPath)
}

// Compose returns a Compose client that uses this Podman client.
func (p *Podman) Compose() *Compose {
	return &Compose{
		Podman: p,
	}
}

// pkgFileioAdapter adapts agent fileio.ReadWriter to pkg/fileio.ReadWriter.
// This is needed because pkg/podman uses pkg/fileio interfaces.
type pkgFileioAdapter struct {
	rw fileio.ReadWriter
}

func (a *pkgFileioAdapter) SetRootdir(path string)                   { a.rw.SetRootdir(path) }
func (a *pkgFileioAdapter) PathFor(filePath string) string           { return a.rw.PathFor(filePath) }
func (a *pkgFileioAdapter) ReadFile(filePath string) ([]byte, error) { return a.rw.ReadFile(filePath) }
func (a *pkgFileioAdapter) ReadDir(dirPath string) ([]fs.DirEntry, error) {
	return a.rw.ReadDir(dirPath)
}
func (a *pkgFileioAdapter) PathExists(path string) (bool, error) { return a.rw.PathExists(path) }
func (a *pkgFileioAdapter) WriteFile(name string, data []byte, perm fs.FileMode, opts ...pkgfileio.FileOption) error {
	return a.rw.WriteFile(name, data, perm, opts...)
}
func (a *pkgFileioAdapter) RemoveFile(file string) error     { return a.rw.RemoveFile(file) }
func (a *pkgFileioAdapter) RemoveAll(path string) error      { return a.rw.RemoveAll(path) }
func (a *pkgFileioAdapter) RemoveContents(path string) error { return a.rw.RemoveContents(path) }
func (a *pkgFileioAdapter) MkdirAll(path string, perm fs.FileMode) error {
	return a.rw.MkdirAll(path, perm)
}
func (a *pkgFileioAdapter) MkdirTemp(prefix string) (string, error) { return a.rw.MkdirTemp(prefix) }
func (a *pkgFileioAdapter) CopyFile(src, dst string) error          { return a.rw.CopyFile(src, dst) }
func (a *pkgFileioAdapter) CopyDir(src, dst string, opts ...pkgfileio.CopyDirOption) error {
	return a.rw.CopyDir(src, dst, opts...)
}
func (a *pkgFileioAdapter) OverwriteAndWipe(file string) error { return a.rw.OverwriteAndWipe(file) }
