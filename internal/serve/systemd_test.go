package serve

import (
	"strings"
	"testing"
)

func TestSystemdUnitContainsExpectedSections(t *testing.T) {
	unit := SystemdUnit(SystemdConfig{
		BinaryPath: "/usr/local/bin/naeos",
		ConfigPath: "/etc/naeos/server.yaml",
		User:       "naeos",
		Group:      "naeos",
		WorkingDir: "/opt/naeos",
		Environment: []string{
			"NAEOS_ENV=production",
			"MALFORMED_WITHOUT_EQUALS",
		},
	})

	for _, want := range []string{
		"[Unit]",
		"Description=NAEOS",
		"After=network-online.target",
		"[Service]",
		"Type=simple",
		"Restart=on-failure",
		"User=naeos",
		"Group=naeos",
		"WorkingDirectory=/opt/naeos",
		"Environment=NAEOS_ENV=production",
		"ExecStart=/usr/local/bin/naeos serve run --config /etc/naeos/server.yaml",
		"ExecReload=/bin/kill -HUP $MAINPID",
		"TimeoutStopSec=60",
		"LimitNOFILE=65536",
		"[Install]",
		"WantedBy=multi-user.target",
	} {
		if !strings.Contains(unit, want) {
			t.Fatalf("systemd unit missing %q\n---\n%s", want, unit)
		}
	}

	if strings.Contains(unit, "MALFORMED_WITHOUT_EQUALS") {
		t.Fatal("malformed environment entry (no =) should be skipped")
	}
}

func TestSystemdUnitOmitsEmptyUserGroupWorkingDir(t *testing.T) {
	unit := SystemdUnit(SystemdConfig{
		BinaryPath: "/usr/local/bin/naeos",
		ConfigPath: "/etc/naeos/server.yaml",
	})
	for _, forbidden := range []string{"User=", "Group=", "WorkingDirectory=", "Environment="} {
		if strings.Contains(unit, forbidden) {
			t.Fatalf("systemd unit should not contain %q when unset\n---\n%s", forbidden, unit)
		}
	}
}
