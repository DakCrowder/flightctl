package quadlet

const (
	// Default installation paths
	DefaultQuadletDir         = "/usr/share/containers/systemd"
	DefaultReadOnlyConfigDir  = "/usr/share/flightctl"
	DefaultWriteableConfigDir = "/etc/flightctl"
	DefaultSystemdUnitDir     = "/usr/lib/systemd/system"
	DefaultBinDir             = "/usr/bin"

	// Systemd target and network names
	FlightctlTarget  = "flightctl.target"
	FlightctlNetwork = "flightctl"
)

// KnownVolumes lists volume names created by flightctl quadlet files
var KnownVolumes = []string{
	"flightctl-db",
	"flightctl-kv",
	"flightctl-alertmanager",
	"flightctl-ui-certs",
	"flightctl-cli-artifacts-certs",
}

// KnownSecrets lists secret names used by flightctl containers
var KnownSecrets = []string{
	"flightctl-postgresql-password",
	"flightctl-postgresql-master-password",
	"flightctl-postgresql-user-password",
	"flightctl-postgresql-migrator-password",
	"flightctl-kv-password",
}
