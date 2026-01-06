package podman

// InspectResult represents the overall structure of podman inspect output.
type InspectResult struct {
	Restarts int             `json:"RestartCount"`
	State    ContainerState  `json:"State"`
	Config   ContainerConfig `json:"Config"`
}

// ContainerState represents the container state part of the podman inspect output.
type ContainerState struct {
	OciVersion  string `json:"OciVersion"`
	Status      string `json:"Status"`
	Running     bool   `json:"Running"`
	Paused      bool   `json:"Paused"`
	Restarting  bool   `json:"Restarting"`
	OOMKilled   bool   `json:"OOMKilled"`
	Dead        bool   `json:"Dead"`
	Pid         int    `json:"Pid"`
	ExitCode    int    `json:"ExitCode"`
	Error       string `json:"Error"`
	StartedAt   string `json:"StartedAt"`
	FinishedAt  string `json:"FinishedAt"`
	Healthcheck string `json:"Healthcheck"`
}

// ContainerConfig represents container configuration from podman inspect.
type ContainerConfig struct {
	Labels map[string]string `json:"Labels"`
}

// ArtifactInspect represents the structure of artifact inspect output.
type ArtifactInspect struct {
	Manifest ArtifactManifest `json:"Manifest"`
	Name     string           `json:"Name"`
	Digest   string           `json:"Digest"`
}

// ArtifactManifest represents the manifest structure of an OCI artifact.
type ArtifactManifest struct {
	Layers []ArtifactLayer `json:"layers"`
}

// ArtifactLayer represents a layer in an OCI artifact manifest.
type ArtifactLayer struct {
	Annotations map[string]string `json:"annotations,omitempty"`
}

// Event represents the structure of a podman event as produced via a CLI events command.
// It should be noted that the CLI represents events differently from libpod.
// https://github.com/containers/podman/blob/main/cmd/podman/system/events.go#L81-L96
type Event struct {
	ContainerExitCode int               `json:"ContainerExitCode,omitempty"`
	ID                string            `json:"ID"`
	Image             string            `json:"Image"`
	Name              string            `json:"Name"`
	Status            string            `json:"Status"`
	Type              string            `json:"Type"`
	TimeNano          int64             `json:"timeNano"`
	Attributes        map[string]string `json:"Attributes"`
}

// Version represents the parsed podman version.
type Version struct {
	Major int
	Minor int
}

// GreaterOrEqual returns true if this version is greater than or equal to the given major.minor.
func (v Version) GreaterOrEqual(major, minor int) bool {
	if v.Major > major {
		return true
	}
	if v.Major == major && v.Minor >= minor {
		return true
	}
	return false
}

// volume represents a podman volume from JSON output.
type volume struct {
	Name string `json:"Name"`
}
