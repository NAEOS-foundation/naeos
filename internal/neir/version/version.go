package version

import (
	"fmt"
	"strconv"
	"strings"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

type VersionInfo struct {
	NEIRVersion    string
	SchemaVersion  string
	ProjectVersion string
}

type SemVer struct {
	Major int
	Minor int
	Patch int
}

func Default() VersionInfo {
	return VersionInfo{NEIRVersion: "0.1.0", SchemaVersion: "1.0", ProjectVersion: "0.1.0"}
}

func ParseSemVer(s string) (SemVer, error) {
	s = strings.TrimPrefix(s, "v")
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return SemVer{}, naeoserr.New(naeoserr.ErrParse, fmt.Sprintf("invalid semver format: %s", s))
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return SemVer{}, naeoserr.New(naeoserr.ErrParse, fmt.Sprintf("invalid major version: %s", parts[0]))
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return SemVer{}, naeoserr.New(naeoserr.ErrParse, fmt.Sprintf("invalid minor version: %s", parts[1]))
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return SemVer{}, naeoserr.New(naeoserr.ErrParse, fmt.Sprintf("invalid patch version: %s", parts[2]))
	}
	if major < 0 || minor < 0 || patch < 0 {
		return SemVer{}, naeoserr.New(naeoserr.ErrValidation, "version components must be non-negative")
	}
	return SemVer{Major: major, Minor: minor, Patch: patch}, nil
}

func (v SemVer) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func Compare(a, b SemVer) int {
	if a.Major != b.Major {
		if a.Major < b.Major {
			return -1
		}
		return 1
	}
	if a.Minor != b.Minor {
		if a.Minor < b.Minor {
			return -1
		}
		return 1
	}
	if a.Patch != b.Patch {
		if a.Patch < b.Patch {
			return -1
		}
		return 1
	}
	return 0
}

func IsCompatible(required, actual SemVer) bool {
	if required.Major == 0 {
		return required.Major == actual.Major && required.Minor == actual.Minor
	}
	return required.Major == actual.Major && actual.Minor >= required.Minor
}

func (vi VersionInfo) Validate() error {
	if vi.NEIRVersion == "" {
		return naeoserr.New(naeoserr.ErrValidation, "NEIRVersion must not be empty")
	}
	if _, err := ParseSemVer(vi.NEIRVersion); err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrParse, "invalid NEIRVersion")
	}
	if vi.SchemaVersion == "" {
		return naeoserr.New(naeoserr.ErrValidation, "SchemaVersion must not be empty")
	}
	if _, err := ParseSemVer(vi.SchemaVersion); err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrParse, "invalid SchemaVersion")
	}
	return nil
}

func (vi VersionInfo) IsCompatibleWith(other VersionInfo) bool {
	neirA, errA := ParseSemVer(vi.NEIRVersion)
	neirB, errB := ParseSemVer(other.NEIRVersion)
	if errA != nil || errB != nil {
		return false
	}
	return IsCompatible(neirA, neirB)
}
