package version

import (
	_ "embed"
	"strings"
)

const (
	// ProductName is the product name used across the codebase.
	ProductName = "naeos"

	// ProductNameUpper is the uppercase product name.
	ProductNameUpper = "NAEOS"

	// DefaultEntryVersion is the default version for marketplace entries.
	DefaultEntryVersion = "1.0.0"

	// DefaultModuleVersion is the default version for generated modules.
	DefaultModuleVersion = "0.1.0"
)

//go:embed VERSION
var versionFile string

// Set by ldflags at build time.
var (
	Version   = ""
	GitCommit = ""
	BuildDate = ""
)

func init() {
	if Version == "" {
		Version = strings.TrimSpace(versionFile)
	}
}

func String() string {
	return Version
}

func Full() string {
	if GitCommit != "" {
		return Version + " (" + GitCommit + ")"
	}
	return Version
}
