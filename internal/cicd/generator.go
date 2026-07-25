package cicd

import (
	"fmt"
	"strings"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

type CICDPlatform string

const (
	GitHubActions CICDPlatform = "github"
	GitLabCI      CICDPlatform = "gitlab"
	Jenkins       CICDPlatform = "jenkins"
)

type PipelineConfig struct {
	Project   string
	Platform  CICDPlatform
	Languages []string
	Steps     []PipelineStep
	Trigger   TriggerConfig
	Secrets   []string
}

type PipelineStep struct {
	Name    string
	Command string
	Env     map[string]string
}

type TriggerConfig struct {
	OnPush    bool
	OnPR      bool
	OnRelease bool
	Schedule  string
}

type PipelineGenerator interface {
	Name() string
	Generate(config *PipelineConfig) (string, error)
}

func GetGenerator(platform CICDPlatform) (PipelineGenerator, error) {
	switch platform {
	case GitHubActions:
		return &GitHubActionsGenerator{}, nil
	case GitLabCI:
		return &GitLabCIGenerator{}, nil
	case Jenkins:
		return &JenkinsGenerator{}, nil
	default:
		return nil, naeoserr.New(naeoserr.ErrPipeline, fmt.Sprintf("unsupported platform: %s", platform))
	}
}

type LintError struct {
	Field   string
	Message string
}

func (e LintError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type LintResult struct {
	Valid  bool
	Errors []LintError
}

func LintConfig(config *PipelineConfig) *LintResult {
	result := &LintResult{Valid: true}

	if config == nil {
		result.Valid = false
		result.Errors = append(result.Errors, LintError{Field: "config", Message: "pipeline config is nil"})
		return result
	}

	if config.Project == "" {
		result.Valid = false
		result.Errors = append(result.Errors, LintError{Field: "project", Message: "project name is required"})
	}

	if config.Platform == "" {
		result.Valid = false
		result.Errors = append(result.Errors, LintError{Field: "platform", Message: "platform is required"})
	}

	validPlatforms := map[CICDPlatform]bool{
		GitHubActions: true,
		GitLabCI:      true,
		Jenkins:       true,
	}
	if config.Platform != "" && !validPlatforms[config.Platform] {
		result.Valid = false
		result.Errors = append(result.Errors, LintError{Field: "platform", Message: fmt.Sprintf("unsupported platform: %s", config.Platform)})
	}

	if len(config.Languages) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, LintError{Field: "languages", Message: "at least one language is required"})
	}

	validLanguages := map[string]bool{
		"go": true, "node": true, "typescript": true, "python": true, "java": true, "rust": true,
	}
	for _, lang := range config.Languages {
		if !validLanguages[lang] {
			result.Valid = false
			result.Errors = append(result.Errors, LintError{Field: "languages", Message: fmt.Sprintf("unsupported language: %s", lang)})
		}
	}

	if config.Trigger.Schedule != "" {
		if !strings.Contains(config.Trigger.Schedule, " ") || len(strings.Fields(config.Trigger.Schedule)) != 5 {
			result.Valid = false
			result.Errors = append(result.Errors, LintError{Field: "trigger.schedule", Message: "schedule must be a valid cron expression with 5 fields"})
		}
	}

	for i, step := range config.Steps {
		if step.Name == "" {
			result.Valid = false
			result.Errors = append(result.Errors, LintError{Field: fmt.Sprintf("steps[%d].name", i), Message: "step name is required"})
		}
		if step.Command == "" {
			result.Valid = false
			result.Errors = append(result.Errors, LintError{Field: fmt.Sprintf("steps[%d].command", i), Message: "step command is required"})
		}
	}

	return result
}

type ConfigTemplate struct {
	Name        string
	Description string
	Platform    CICDPlatform
	Base        *PipelineConfig
	MergeFields map[string]string
}

func NewConfigTemplate(name, description string, platform CICDPlatform, base *PipelineConfig) *ConfigTemplate {
	return &ConfigTemplate{
		Name:        name,
		Description: description,
		Platform:    platform,
		Base:        base,
		MergeFields: make(map[string]string),
	}
}

func (t *ConfigTemplate) Render() (*PipelineConfig, error) {
	if t.Base == nil {
		return nil, naeoserr.New(naeoserr.ErrPipeline, fmt.Sprintf("template %s has no base config", t.Name))
	}

	config := *t.Base
	config.Platform = t.Platform

	if projectName, ok := t.MergeFields["project"]; ok {
		config.Project = projectName
	}

	return &config, nil
}

func (t *ConfigTemplate) SetField(key, value string) {
	if t.MergeFields == nil {
		t.MergeFields = make(map[string]string)
	}
	t.MergeFields[key] = value
}

func MergeConfigs(base *PipelineConfig, overrides *PipelineConfig) *PipelineConfig {
	if base == nil {
		return overrides
	}
	if overrides == nil {
		return base
	}

	merged := *base

	if overrides.Project != "" {
		merged.Project = overrides.Project
	}
	if overrides.Platform != "" {
		merged.Platform = overrides.Platform
	}
	if len(overrides.Languages) > 0 {
		merged.Languages = append([]string{}, overrides.Languages...)
	}
	if len(overrides.Secrets) > 0 {
		merged.Secrets = append([]string{}, overrides.Secrets...)
	}

	if overrides.Trigger.OnPush || overrides.Trigger.OnPR || overrides.Trigger.OnRelease || overrides.Trigger.Schedule != "" {
		merged.Trigger = overrides.Trigger
	}

	if len(overrides.Steps) > 0 {
		merged.Steps = append([]PipelineStep{}, overrides.Steps...)
	}

	return &merged
}

func MergeConfigsDeep(base *PipelineConfig, overrides *PipelineConfig) *PipelineConfig {
	if base == nil {
		return overrides
	}
	if overrides == nil {
		return base
	}

	merged := *base

	if overrides.Project != "" {
		merged.Project = overrides.Project
	}
	if overrides.Platform != "" {
		merged.Platform = overrides.Platform
	}

	langSet := make(map[string]bool)
	for _, lang := range base.Languages {
		langSet[lang] = true
		merged.Languages = append([]string{}, base.Languages...)
	}
	for _, lang := range overrides.Languages {
		if !langSet[lang] {
			merged.Languages = append(merged.Languages, lang)
		}
	}

	secretSet := make(map[string]bool)
	for _, s := range base.Secrets {
		secretSet[s] = true
		merged.Secrets = append([]string{}, base.Secrets...)
	}
	for _, s := range overrides.Secrets {
		if !secretSet[s] {
			merged.Secrets = append(merged.Secrets, s)
		}
	}

	if overrides.Trigger.OnPush || overrides.Trigger.OnPR || overrides.Trigger.OnRelease || overrides.Trigger.Schedule != "" {
		merged.Trigger = overrides.Trigger
	}

	merged.Steps = append([]PipelineStep{}, base.Steps...)
	merged.Steps = append(merged.Steps, overrides.Steps...)

	return &merged
}
