package version

type VersionInfo struct {
	NEIRVersion   string
	SchemaVersion string
	ProjectVersion string
}

func Default() VersionInfo {
	return VersionInfo{NEIRVersion: "0.1.0", SchemaVersion: "1.0", ProjectVersion: "0.1.0"}
}
