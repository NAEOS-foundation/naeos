package migration

import (
	"fmt"
	"sort"
	"strings"
)

const (
	CurrentVersion = "0.1.0"
	TargetVersion  = "0.3.0"
)

type Version struct {
	Major int
	Minor int
	Patch int
}

func ParseVersion(v string) (Version, error) {
	var ver Version
	_, err := fmt.Sscanf(v, "%d.%d.%d", &ver.Major, &ver.Minor, &ver.Patch)
	if err != nil {
		return Version{}, fmt.Errorf("parse version %s: %w", v, err)
	}
	if ver.Major < 0 || ver.Minor < 0 || ver.Patch < 0 {
		return Version{}, fmt.Errorf("parse version %s: negative component", v)
	}
	return ver, nil
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v Version) Less(other Version) bool {
	if v.Major != other.Major {
		return v.Major < other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor < other.Minor
	}
	return v.Patch < other.Patch
}

type MigrationStep struct {
	FromVersion string
	ToVersion   string
	Description string
	Migrate     func(spec []byte) ([]byte, error)
}

type MigrationPlanner struct {
	steps []MigrationStep
}

func NewPlanner() *MigrationPlanner {
	p := &MigrationPlanner{}
	p.steps = append(p.steps, builtinMigrations()...)
	return p
}

func (p *MigrationPlanner) AddStep(step MigrationStep) {
	p.steps = append(p.steps, step)
}

func (p *MigrationPlanner) Plan(from, to string) ([]MigrationStep, error) {
	fromVer, err := ParseVersion(from)
	if err != nil {
		return nil, err
	}
	toVer, err := ParseVersion(to)
	if err != nil {
		return nil, err
	}

	if !fromVer.Less(toVer) {
		return nil, fmt.Errorf("target version %s is not newer than source %s", to, from)
	}

	sort.Slice(p.steps, func(i, j int) bool {
		vi, _ := ParseVersion(p.steps[i].ToVersion)
		vj, _ := ParseVersion(p.steps[j].ToVersion)
		return vi.Less(vj)
	})

	var plan []MigrationStep
	for _, step := range p.steps {
		stepFrom, _ := ParseVersion(step.FromVersion)
		stepTo, _ := ParseVersion(step.ToVersion)
		if (fromVer.Less(stepTo) || fromVer == stepFrom) && stepFrom.Less(toVer) || stepFrom == fromVer {
			plan = append(plan, step)
		}
	}

	return plan, nil
}

func (p *MigrationPlanner) Migrate(spec []byte, from, to string) ([]byte, error) {
	plan, err := p.Plan(from, to)
	if err != nil {
		return nil, err
	}

	current := spec
	for _, step := range plan {
		var err error
		current, err = step.Migrate(current)
		if err != nil {
			return nil, fmt.Errorf("migration %s -> %s failed: %w", step.FromVersion, step.ToVersion, err)
		}
	}
	return current, nil
}

func builtinMigrations() []MigrationStep {
	return []MigrationStep{
		{
			FromVersion: "0.1.0",
			ToVersion:   "0.2.0",
			Description: "Add generation section",
			Migrate: func(spec []byte) ([]byte, error) {
				content := string(spec)
				if !strings.Contains(content, "generation:") {
					content += "\ngeneration:\n  languages:\n    - go\n  output_dir: ./out\n"
				}
				return []byte(content), nil
			},
		},
		{
			FromVersion: "0.2.0",
			ToVersion:   "0.3.0",
			Description: "Add testing section",
			Migrate: func(spec []byte) ([]byte, error) {
				content := string(spec)
				if !strings.Contains(content, "testing:") {
					content += "\ntesting:\n  strategy: unit\n  coverage: standard\n"
				}
				return []byte(content), nil
			},
		},
	}
}
