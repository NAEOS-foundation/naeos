package parser

import (
	"fmt"
	"strconv"
	"strings"
)

type SchemaVersion struct {
	Major int
	Minor int
	Patch int
}

func ParseSchemaVersion(version string) (SchemaVersion, error) {
	version = strings.TrimPrefix(strings.TrimSpace(version), "v")
	parts := strings.Split(version, ".")
	if len(parts) < 1 || len(parts) > 3 {
		return SchemaVersion{}, fmt.Errorf("invalid version format: %s", version)
	}

	var sv SchemaVersion
	var err error

	sv.Major, err = strconv.Atoi(parts[0])
	if err != nil {
		return SchemaVersion{}, fmt.Errorf("invalid major version: %w", err)
	}
	if sv.Major < 0 {
		return SchemaVersion{}, fmt.Errorf("major version cannot be negative")
	}

	if len(parts) > 1 {
		sv.Minor, err = strconv.Atoi(parts[1])
		if err != nil {
			return SchemaVersion{}, fmt.Errorf("invalid minor version: %w", err)
		}
		if sv.Minor < 0 {
			return SchemaVersion{}, fmt.Errorf("minor version cannot be negative")
		}
	}

	if len(parts) > 2 {
		sv.Patch, err = strconv.Atoi(parts[2])
		if err != nil {
			return SchemaVersion{}, fmt.Errorf("invalid patch version: %w", err)
		}
		if sv.Patch < 0 {
			return SchemaVersion{}, fmt.Errorf("patch version cannot be negative")
		}
	}

	return sv, nil
}

func (v SchemaVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v SchemaVersion) GreaterThan(other SchemaVersion) bool {
	if v.Major != other.Major {
		return v.Major > other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor > other.Minor
	}
	return v.Patch > other.Patch
}

func (v SchemaVersion) LessThan(other SchemaVersion) bool {
	return other.GreaterThan(v)
}

func (v SchemaVersion) CompatibleWith(minimum SchemaVersion) bool {
	return !v.LessThan(minimum)
}

const (
	MinSpecVersion = "0.1.0"
	CurrentSpecVersion = "0.3.0"
)

type VersionCheckResult struct {
	Valid    bool
	Current  SchemaVersion
	Required SchemaVersion
	Message  string
}

func CheckSpecVersion(versionStr string) *VersionCheckResult {
	if versionStr == "" {
		return &VersionCheckResult{
			Valid:   true,
			Message: "no version specified, assuming compatible",
		}
	}

	current, err := ParseSchemaVersion(versionStr)
	if err != nil {
		return &VersionCheckResult{
			Valid:   false,
			Message: fmt.Sprintf("invalid version: %v", err),
		}
	}

	required, _ := ParseSchemaVersion(MinSpecVersion)

	if current.CompatibleWith(required) {
		return &VersionCheckResult{
			Valid:    true,
			Current:  current,
			Required: required,
			Message:  fmt.Sprintf("spec version %s is compatible (minimum: %s)", current, required),
		}
	}

	return &VersionCheckResult{
		Valid:    false,
		Current:  current,
		Required: required,
		Message:  fmt.Sprintf("spec version %s is below minimum %s", current, required),
	}
}

func ExtractVersionFromData(data any) string {
	if m, ok := data.(map[string]any); ok {
		if v, ok := m["version"].(string); ok {
			return v
		}
	}
	return ""
}

func VersionCompatWarning(versionStr string) string {
	result := CheckSpecVersion(versionStr)
	if result.Valid {
		return ""
	}
	return result.Message
}
