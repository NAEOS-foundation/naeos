package serve

import (
	"fmt"
	"strings"
)

// SystemdConfig holds the parameters used to render a systemd unit for the
// NAEOS daemon.
type SystemdConfig struct {
	// BinaryPath is the absolute path to the naeos binary.
	BinaryPath string
	// ConfigPath is the absolute path to the server config file.
	ConfigPath string
	// WorkingDir is the working directory for the service.
	WorkingDir string
	// User is the system user the service runs as. Empty keeps systemd default.
	User string
	// Group is the system group the service runs as. Empty keeps systemd default.
	Group string
	// Environment is an optional list of KEY=VALUE pairs set for the service.
	Environment []string
	// RunAsUser controls whether User= is emitted (empty otherwise).
}

// SystemdUnit renders a fully-formed systemd service unit for the NAEOS server.
func SystemdUnit(sc SystemdConfig) string {
	var b strings.Builder
	b.WriteString("[Unit]\n")
	b.WriteString("Description=NAEOS Declarative Engineering Runtime server\n")
	b.WriteString("After=network-online.target\n")
	b.WriteString("Wants=network-online.target\n\n")

	b.WriteString("[Service]\n")
	b.WriteString("Type=simple\n")
	b.WriteString("Restart=on-failure\n")
	b.WriteString("RestartSec=3\n")
	if sc.User != "" {
		fmt.Fprintf(&b, "User=%s\n", sc.User)
	}
	if sc.Group != "" {
		fmt.Fprintf(&b, "Group=%s\n", sc.Group)
	}
	if sc.WorkingDir != "" {
		fmt.Fprintf(&b, "WorkingDirectory=%s\n", sc.WorkingDir)
	}
	for _, env := range sc.Environment {
		if strings.Contains(env, "=") {
			fmt.Fprintf(&b, "Environment=%s\n", env)
		}
	}
	fmt.Fprintf(&b, "ExecStart=%s serve run --config %s\n", sc.BinaryPath, sc.ConfigPath)
	b.WriteString("ExecReload=/bin/kill -HUP $MAINPID\n")
	b.WriteString("TimeoutStopSec=60\n")
	b.WriteString("KillSignal=SIGTERM\n")
	b.WriteString("LimitNOFILE=65536\n")
	b.WriteString("\n")

	b.WriteString("[Install]\n")
	b.WriteString("WantedBy=multi-user.target\n")

	return b.String()
}
