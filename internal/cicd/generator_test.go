package cicd

import (
	"strings"
	"testing"
)

func testGoConfig() *PipelineConfig {
	return &PipelineConfig{
		Project:   "myproject",
		Platform:  GitHubActions,
		Languages: []string{"go"},
		Steps: []PipelineStep{
			{Name: "Deploy", Command: "make deploy"},
		},
		Trigger: TriggerConfig{
			OnPush: true,
			OnPR:   true,
		},
		Secrets: []string{"AWS_ACCESS_KEY"},
	}
}

func testNodeConfig() *PipelineConfig {
	return &PipelineConfig{
		Project:   "webapp",
		Platform:  GitLabCI,
		Languages: []string{"node"},
		Trigger: TriggerConfig{
			OnPush: true,
		},
	}
}

func testPythonConfig() *PipelineConfig {
	return &PipelineConfig{
		Project:   "mlservice",
		Platform:  Jenkins,
		Languages: []string{"python"},
		Trigger: TriggerConfig{
			OnPush:   true,
			OnPR:     true,
			Schedule: "0 2 * * *",
		},
	}
}

func testJavaConfig() *PipelineConfig {
	return &PipelineConfig{
		Project:   "javaservice",
		Platform:  GitHubActions,
		Languages: []string{"java"},
		Trigger: TriggerConfig{
			OnRelease: true,
		},
	}
}

func testRustConfig() *PipelineConfig {
	return &PipelineConfig{
		Project:   "rustservice",
		Platform:  GitLabCI,
		Languages: []string{"rust"},
		Trigger: TriggerConfig{
			OnPush: true,
		},
	}
}

func TestGitHubActionsGenerator_Name(t *testing.T) {
	g := &GitHubActionsGenerator{}
	if g.Name() != "GitHub Actions" {
		t.Errorf("expected 'GitHub Actions', got %s", g.Name())
	}
}

func TestGitHubActionsGenerator_Generate(t *testing.T) {
	g := &GitHubActionsGenerator{}
	output, err := g.Generate(testGoConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "name: CI/CD Pipeline") {
		t.Error("missing pipeline name")
	}
	if !strings.Contains(output, "actions/checkout@v4") {
		t.Error("missing checkout step")
	}
	if !strings.Contains(output, "go build ./...") {
		t.Error("missing go build")
	}
	if !strings.Contains(output, "go test ./...") {
		t.Error("missing go test")
	}
}

func TestGitHubActionsGenerator_NodeSteps(t *testing.T) {
	g := &GitHubActionsGenerator{}
	output, err := g.Generate(testNodeConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "actions/setup-node@v4") {
		t.Error("missing node setup")
	}
	if !strings.Contains(output, "npm ci") {
		t.Error("missing npm ci")
	}
}

func TestGitHubActionsGenerator_PythonSteps(t *testing.T) {
	g := &GitHubActionsGenerator{}
	output, err := g.Generate(testPythonConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "actions/setup-python@v5") {
		t.Error("missing python setup")
	}
	if !strings.Contains(output, "pytest") {
		t.Error("missing pytest")
	}
}

func TestGitHubActionsGenerator_JavaSteps(t *testing.T) {
	g := &GitHubActionsGenerator{}
	output, err := g.Generate(testJavaConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "actions/setup-java@v4") {
		t.Error("missing java setup")
	}
	if !strings.Contains(output, "mvn clean compile") {
		t.Error("missing maven build")
	}
}

func TestGitHubActionsGenerator_RustSteps(t *testing.T) {
	g := &GitHubActionsGenerator{}
	output, err := g.Generate(testRustConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "dtolnay/rust-toolchain@stable") {
		t.Error("missing rust setup")
	}
	if !strings.Contains(output, "cargo build --release") {
		t.Error("missing cargo build")
	}
}

func TestGitHubActionsGenerator_Triggers(t *testing.T) {
	g := &GitHubActionsGenerator{}
	config := testGoConfig()
	config.Trigger.OnRelease = true
	config.Trigger.Schedule = "0 0 * * *"
	output, err := g.Generate(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "release:") {
		t.Error("missing release trigger")
	}
	if !strings.Contains(output, "schedule:") {
		t.Error("missing schedule trigger")
	}
}

func TestGitHubActionsGenerator_CustomSteps(t *testing.T) {
	g := &GitHubActionsGenerator{}
	output, err := g.Generate(testGoConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "Deploy") {
		t.Error("missing custom step name")
	}
	if !strings.Contains(output, "make deploy") {
		t.Error("missing custom step command")
	}
}

func TestGitLabCIGenerator_Name(t *testing.T) {
	g := &GitLabCIGenerator{}
	if g.Name() != "GitLab CI" {
		t.Errorf("expected 'GitLab CI', got %s", g.Name())
	}
}

func TestGitLabCIGenerator_Generate(t *testing.T) {
	g := &GitLabCIGenerator{}
	output, err := g.Generate(testGoConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "stages:") {
		t.Error("missing stages")
	}
	if !strings.Contains(output, "build:") {
		t.Error("missing build job")
	}
	if !strings.Contains(output, "test:") {
		t.Error("missing test job")
	}
	if !strings.Contains(output, "deploy:") {
		t.Error("missing deploy job")
	}
}

func TestGitLabCIGenerator_GoImage(t *testing.T) {
	g := &GitLabCIGenerator{}
	output, err := g.Generate(testGoConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "image: golang:1.22") {
		t.Error("missing golang image")
	}
}

func TestGitLabCIGenerator_NodeImage(t *testing.T) {
	g := &GitLabCIGenerator{}
	output, err := g.Generate(testNodeConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "image: node:20") {
		t.Error("missing node image")
	}
}

func TestJenkinsGenerator_Name(t *testing.T) {
	g := &JenkinsGenerator{}
	if g.Name() != "Jenkins" {
		t.Errorf("expected 'Jenkins', got %s", g.Name())
	}
}

func TestJenkinsGenerator_Generate(t *testing.T) {
	g := &JenkinsGenerator{}
	output, err := g.Generate(testGoConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "pipeline {") {
		t.Error("missing pipeline block")
	}
	if !strings.Contains(output, "stages {") {
		t.Error("missing stages block")
	}
	if !strings.Contains(output, "post {") {
		t.Error("missing post block")
	}
}

func TestJenkinsGenerator_Secrets(t *testing.T) {
	g := &JenkinsGenerator{}
	output, err := g.Generate(testGoConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "environment {") {
		t.Error("missing environment block")
	}
	if !strings.Contains(output, "AWS_ACCESS_KEY = credentials('AWS_ACCESS_KEY')") {
		t.Error("missing credentials reference")
	}
}

func TestJenkinsGenerator_DeploySteps(t *testing.T) {
	g := &JenkinsGenerator{}
	output, err := g.Generate(testGoConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "make deploy") {
		t.Error("missing deploy command")
	}
}

func TestAzurePipelinesGenerator_Name(t *testing.T) {
	g := &AzurePipelinesGenerator{}
	if g.Name() != "Azure Pipelines" {
		t.Errorf("expected 'Azure Pipelines', got %s", g.Name())
	}
}

func TestAzurePipelinesGenerator_Generate(t *testing.T) {
	g := &AzurePipelinesGenerator{}
	output, err := g.Generate(testGoConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "pool:") {
		t.Error("missing pool definition")
	}
	if !strings.Contains(output, "stages:") {
		t.Error("missing stages")
	}
	if !strings.Contains(output, "stage: Build") {
		t.Error("missing Build stage")
	}
	if !strings.Contains(output, "stage: Test") {
		t.Error("missing Test stage")
	}
	if !strings.Contains(output, "stage: Deploy") {
		t.Error("missing Deploy stage")
	}
}

func TestAzurePipelinesGenerator_Triggers(t *testing.T) {
	g := &AzurePipelinesGenerator{}
	output, err := g.Generate(testPythonConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "schedules:") {
		t.Error("missing schedules")
	}
	if !strings.Contains(output, "0 2 * * *") {
		t.Error("missing cron expression")
	}
}

func TestAzurePipelinesGenerator_Secrets(t *testing.T) {
	g := &AzurePipelinesGenerator{}
	output, err := g.Generate(testGoConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "variables:") {
		t.Error("missing variables block")
	}
	if !strings.Contains(output, "AWS_ACCESS_KEY: $(AWS_ACCESS_KEY)") {
		t.Error("missing secret variable")
	}
}

func TestDockerComposeGenerator_Name(t *testing.T) {
	g := &DockerComposeGenerator{}
	if g.Name() != "Docker Compose" {
		t.Errorf("expected 'Docker Compose', got %s", g.Name())
	}
}

func TestDockerComposeGenerator_GenerateGo(t *testing.T) {
	g := &DockerComposeGenerator{}
	output, err := g.Generate(testGoConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "version: '3.8'") {
		t.Error("missing docker-compose version")
	}
	if !strings.Contains(output, "services:") {
		t.Error("missing services block")
	}
	if !strings.Contains(output, "myproject:") {
		t.Error("missing service name")
	}
	if !strings.Contains(output, "golang:1.22") {
		t.Error("missing go image")
	}
}

func TestDockerComposeGenerator_GenerateNode(t *testing.T) {
	g := &DockerComposeGenerator{}
	output, err := g.Generate(testNodeConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "node:20") {
		t.Error("missing node image")
	}
	if !strings.Contains(output, "redis:") {
		t.Error("missing redis service for node")
	}
}

func TestDockerComposeGenerator_GeneratePython(t *testing.T) {
	g := &DockerComposeGenerator{}
	output, err := g.Generate(testPythonConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "python:3.12") {
		t.Error("missing python image")
	}
	if !strings.Contains(output, "db:") {
		t.Error("missing db service for python")
	}
}

func TestDockerComposeGenerator_Networks(t *testing.T) {
	g := &DockerComposeGenerator{}
	output, err := g.Generate(testGoConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "networks:") {
		t.Error("missing networks")
	}
	if !strings.Contains(output, "app-network:") {
		t.Error("missing app-network")
	}
}

func TestDockerComposeGenerator_CustomSteps(t *testing.T) {
	g := &DockerComposeGenerator{}
	output, err := g.Generate(testGoConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "# Custom steps:") {
		t.Error("missing custom steps comment")
	}
	if !strings.Contains(output, "Deploy: make deploy") {
		t.Error("missing custom step in comments")
	}
}

func TestDockerComposeGenerator_EmptyProject(t *testing.T) {
	g := &DockerComposeGenerator{}
	config := &PipelineConfig{
		Languages: []string{"go"},
	}
	output, err := g.Generate(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "app:") {
		t.Error("expected default service name 'app'")
	}
}

func TestDockerComposeGenerator_Dockerfile(t *testing.T) {
	g := &DockerComposeGenerator{}
	output, err := g.GenerateDockerfile(testGoConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "FROM golang:1.22 AS builder") {
		t.Error("missing go builder stage")
	}
	if !strings.Contains(output, "CGO_ENABLED=0") {
		t.Error("missing CGO_ENABLED=0 for static binary")
	}
}

func TestDockerComposeGenerator_DockerfileNode(t *testing.T) {
	g := &DockerComposeGenerator{}
	output, err := g.GenerateDockerfile(testNodeConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "FROM node:20 AS builder") {
		t.Error("missing node builder stage")
	}
	if !strings.Contains(output, "npm ci") {
		t.Error("missing npm ci in Dockerfile")
	}
}

func TestDockerComposeGenerator_DockerfileNoLangs(t *testing.T) {
	g := &DockerComposeGenerator{}
	config := &PipelineConfig{}
	_, err := g.GenerateDockerfile(config)
	if err == nil {
		t.Error("expected error for empty languages")
	}
}

func TestDockerComposeGenerator_DockerfileUnsupportedLang(t *testing.T) {
	g := &DockerComposeGenerator{}
	config := &PipelineConfig{Languages: []string{"swift"}}
	_, err := g.GenerateDockerfile(config)
	if err == nil {
		t.Error("expected error for unsupported language")
	}
}

func TestNotificationGenerator_Name(t *testing.T) {
	g := &NotificationGenerator{}
	if g.Name() != "Notification Generator" {
		t.Errorf("expected 'Notification Generator', got %s", g.Name())
	}
}

func TestNotificationGenerator_NilConfig(t *testing.T) {
	g := &NotificationGenerator{}
	_, err := g.GenerateSteps(nil)
	if err == nil {
		t.Error("expected error for nil config")
	}
}

func TestNotificationGenerator_SlackSteps(t *testing.T) {
	g := &NotificationGenerator{}
	config := &NotificationConfig{
		Type:   NotificationSlack,
		Target: "https://hooks.slack.com/test",
		Events: []string{"success", "failure"},
	}
	steps, err := g.GenerateSteps(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(steps))
	}
	if steps[0].Name != "Notify Slack on success" {
		t.Errorf("unexpected step name: %s", steps[0].Name)
	}
}

func TestNotificationGenerator_SlackDefaultEvents(t *testing.T) {
	g := &NotificationGenerator{}
	config := &NotificationConfig{
		Type:   NotificationSlack,
		Target: "https://hooks.slack.com/test",
	}
	steps, err := g.GenerateSteps(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(steps) != 1 {
		t.Fatalf("expected 1 default step, got %d", len(steps))
	}
	if steps[0].Name != "Notify Slack on completion" {
		t.Errorf("unexpected step name: %s", steps[0].Name)
	}
}

func TestNotificationGenerator_SlackMissingTarget(t *testing.T) {
	g := &NotificationGenerator{}
	config := &NotificationConfig{
		Type: NotificationSlack,
	}
	_, err := g.GenerateSteps(config)
	if err == nil {
		t.Error("expected error for missing slack target")
	}
}

func TestNotificationGenerator_EmailSteps(t *testing.T) {
	g := &NotificationGenerator{}
	config := &NotificationConfig{
		Type:   NotificationEmail,
		Target: "admin@example.com",
		Events: []string{"success"},
	}
	steps, err := g.GenerateSteps(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(steps))
	}
	if !strings.Contains(steps[0].Command, "admin@example.com") {
		t.Error("missing email recipient in command")
	}
}

func TestNotificationGenerator_EmailMissingTarget(t *testing.T) {
	g := &NotificationGenerator{}
	config := &NotificationConfig{
		Type: NotificationEmail,
	}
	_, err := g.GenerateSteps(config)
	if err == nil {
		t.Error("expected error for missing email target")
	}
}

func TestNotificationGenerator_WebhookSteps(t *testing.T) {
	g := &NotificationGenerator{}
	config := &NotificationConfig{
		Type:   NotificationWebhook,
		Target: "https://example.com/webhook",
		Events: []string{"started", "completed"},
	}
	steps, err := g.GenerateSteps(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(steps))
	}
}

func TestNotificationGenerator_WebhookMissingTarget(t *testing.T) {
	g := &NotificationGenerator{}
	config := &NotificationConfig{
		Type: NotificationWebhook,
	}
	_, err := g.GenerateSteps(config)
	if err == nil {
		t.Error("expected error for missing webhook target")
	}
}

func TestNotificationGenerator_UnsupportedType(t *testing.T) {
	g := &NotificationGenerator{}
	config := &NotificationConfig{
		Type:   NotificationType("pagerduty"),
		Target: "test",
	}
	_, err := g.GenerateSteps(config)
	if err == nil {
		t.Error("expected error for unsupported notification type")
	}
}

func TestEmbedNotifications(t *testing.T) {
	config := testGoConfig()
	notifications := []*NotificationConfig{
		{
			Type:   NotificationSlack,
			Target: "https://hooks.slack.com/test",
			Events: []string{"success"},
		},
	}
	err := EmbedNotifications(config, notifications)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(config.Steps) != 2 {
		t.Fatalf("expected 2 steps (1 original + 1 notification), got %d", len(config.Steps))
	}
}

func TestEmbedNotifications_NilConfig(t *testing.T) {
	err := EmbedNotifications(nil, []*NotificationConfig{})
	if err == nil {
		t.Error("expected error for nil config")
	}
}

func TestGenerateNotificationBlock_GitHub(t *testing.T) {
	config := testGoConfig()
	notifications := []*NotificationConfig{
		{
			Type:   NotificationWebhook,
			Target: "https://example.com/hook",
			Events: []string{"success"},
		},
	}
	block, err := GenerateNotificationBlock(config, notifications)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(block, "name:") {
		t.Error("missing step name in GitHub block")
	}
}

func TestGenerateNotificationBlock_GitLab(t *testing.T) {
	config := testNodeConfig()
	notifications := []*NotificationConfig{
		{
			Type:   NotificationEmail,
			Target: "team@example.com",
			Events: []string{"failure"},
		},
	}
	block, err := GenerateNotificationBlock(config, notifications)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(block, "notify:") {
		t.Error("missing notify job in GitLab block")
	}
}

func TestGenerateNotificationBlock_OtherPlatform(t *testing.T) {
	config := testGoConfig()
	config.Platform = Jenkins
	notifications := []*NotificationConfig{
		{
			Type:   NotificationSlack,
			Target: "https://hooks.slack.com/test",
			Events: []string{"success"},
		},
	}
	block, err := GenerateNotificationBlock(config, notifications)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(block, "name:") {
		t.Error("missing step name in generic block")
	}
}

func TestFormatNotificationStepsYAML_GitHub(t *testing.T) {
	steps := []PipelineStep{
		{
			Name:    "Test Step",
			Command: "echo test",
			Env:     map[string]string{"FOO": "bar"},
		},
	}
	output := FormatNotificationStepsYAML(steps, GitHubActions)
	if !strings.Contains(output, "name: Test Step") {
		t.Error("missing step name")
	}
	if !strings.Contains(output, "FOO: bar") {
		t.Error("missing env var")
	}
	if !strings.Contains(output, "run: echo test") {
		t.Error("missing run command")
	}
}

func TestFormatNotificationStepsYAML_GitLab(t *testing.T) {
	steps := []PipelineStep{
		{Name: "Step", Command: "echo hi"},
	}
	output := FormatNotificationStepsYAML(steps, GitLabCI)
	if !strings.Contains(output, "echo hi") {
		t.Error("missing command")
	}
}

func TestGetGenerator(t *testing.T) {
	tests := []struct {
		platform CICDPlatform
		wantName string
	}{
		{GitHubActions, "GitHub Actions"},
		{GitLabCI, "GitLab CI"},
		{AzurePipelines, "Azure Pipelines"},
		{Jenkins, "Jenkins"},
	}
	for _, tt := range tests {
		gen, err := GetGenerator(tt.platform)
		if err != nil {
			t.Errorf("GetGenerator(%s) error: %v", tt.platform, err)
			continue
		}
		if gen.Name() != tt.wantName {
			t.Errorf("GetGenerator(%s) name = %s, want %s", tt.platform, gen.Name(), tt.wantName)
		}
	}
}

func TestGetGenerator_Unsupported(t *testing.T) {
	_, err := GetGenerator("circleci")
	if err == nil {
		t.Error("expected error for unsupported platform")
	}
}

func TestLintConfig_NilConfig(t *testing.T) {
	result := LintConfig(nil)
	if result.Valid {
		t.Error("expected invalid for nil config")
	}
	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(result.Errors))
	}
}

func TestLintConfig_ValidConfig(t *testing.T) {
	result := LintConfig(testGoConfig())
	if !result.Valid {
		for _, e := range result.Errors {
			t.Logf("lint error: %s", e.Error())
		}
		t.Error("expected valid config")
	}
}

func TestLintConfig_MissingProject(t *testing.T) {
	config := testGoConfig()
	config.Project = ""
	result := LintConfig(config)
	if result.Valid {
		t.Error("expected invalid for missing project")
	}
}

func TestLintConfig_MissingPlatform(t *testing.T) {
	config := testGoConfig()
	config.Platform = ""
	result := LintConfig(config)
	if result.Valid {
		t.Error("expected invalid for missing platform")
	}
}

func TestLintConfig_InvalidPlatform(t *testing.T) {
	config := testGoConfig()
	config.Platform = "circleci"
	result := LintConfig(config)
	if result.Valid {
		t.Error("expected invalid for unsupported platform")
	}
}

func TestLintConfig_NoLanguages(t *testing.T) {
	config := testGoConfig()
	config.Languages = nil
	result := LintConfig(config)
	if result.Valid {
		t.Error("expected invalid for no languages")
	}
}

func TestLintConfig_UnsupportedLanguage(t *testing.T) {
	config := testGoConfig()
	config.Languages = []string{"swift"}
	result := LintConfig(config)
	if result.Valid {
		t.Error("expected invalid for unsupported language")
	}
}

func TestLintConfig_InvalidSchedule(t *testing.T) {
	config := testGoConfig()
	config.Trigger.Schedule = "invalid"
	result := LintConfig(config)
	if result.Valid {
		t.Error("expected invalid for bad cron schedule")
	}
}

func TestLintConfig_ValidSchedule(t *testing.T) {
	config := testGoConfig()
	config.Trigger.Schedule = "0 0 * * *"
	result := LintConfig(config)
	if !result.Valid {
		for _, e := range result.Errors {
			t.Logf("lint error: %s", e.Error())
		}
		t.Error("expected valid with proper cron")
	}
}

func TestLintConfig_EmptyStepName(t *testing.T) {
	config := testGoConfig()
	config.Steps = []PipelineStep{{Name: "", Command: "echo hi"}}
	result := LintConfig(config)
	if result.Valid {
		t.Error("expected invalid for empty step name")
	}
}

func TestLintConfig_EmptyStepCommand(t *testing.T) {
	config := testGoConfig()
	config.Steps = []PipelineStep{{Name: "Test", Command: ""}}
	result := LintConfig(config)
	if result.Valid {
		t.Error("expected invalid for empty step command")
	}
}

func TestMergeConfigs_BaseOnly(t *testing.T) {
	base := testGoConfig()
	merged := MergeConfigs(base, nil)
	if merged.Project != base.Project {
		t.Error("project should match base")
	}
}

func TestMergeConfigs_OverrideProject(t *testing.T) {
	base := testGoConfig()
	override := &PipelineConfig{Project: "newproject"}
	merged := MergeConfigs(base, override)
	if merged.Project != "newproject" {
		t.Errorf("expected 'newproject', got %s", merged.Project)
	}
}

func TestMergeConfigs_OverrideTrigger(t *testing.T) {
	base := testGoConfig()
	override := &PipelineConfig{
		Trigger: TriggerConfig{OnRelease: true},
	}
	merged := MergeConfigs(base, override)
	if !merged.Trigger.OnRelease {
		t.Error("expected OnRelease to be true")
	}
	if merged.Trigger.OnPush {
		t.Error("expected OnPush to be overridden to false")
	}
}

func TestMergeConfigsDeep_NoDuplicateLangs(t *testing.T) {
	base := &PipelineConfig{
		Languages: []string{"go", "node"},
	}
	override := &PipelineConfig{
		Languages: []string{"node", "python"},
	}
	merged := MergeConfigsDeep(base, override)
	if len(merged.Languages) != 3 {
		t.Errorf("expected 3 unique languages, got %d: %v", len(merged.Languages), merged.Languages)
	}
}

func TestMergeConfigsDeep_NoDuplicateSecrets(t *testing.T) {
	base := &PipelineConfig{
		Secrets: []string{"A", "B"},
	}
	override := &PipelineConfig{
		Secrets: []string{"B", "C"},
	}
	merged := MergeConfigsDeep(base, override)
	if len(merged.Secrets) != 3 {
		t.Errorf("expected 3 unique secrets, got %d: %v", len(merged.Secrets), merged.Secrets)
	}
}

func TestMergeConfigsDeep_CombinesSteps(t *testing.T) {
	base := &PipelineConfig{
		Steps: []PipelineStep{{Name: "A", Command: "cmd a"}},
	}
	override := &PipelineConfig{
		Steps: []PipelineStep{{Name: "B", Command: "cmd b"}},
	}
	merged := MergeConfigsDeep(base, override)
	if len(merged.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(merged.Steps))
	}
}

func TestMergeConfigsDeep_NilBase(t *testing.T) {
	override := testGoConfig()
	merged := MergeConfigsDeep(nil, override)
	if merged.Project != override.Project {
		t.Error("should return override when base is nil")
	}
}

func TestMergeConfigsDeep_NilOverride(t *testing.T) {
	base := testGoConfig()
	merged := MergeConfigsDeep(base, nil)
	if merged.Project != base.Project {
		t.Error("should return base when override is nil")
	}
}

func TestConfigTemplate_Render(t *testing.T) {
	base := testGoConfig()
	tmpl := NewConfigTemplate("ci-template", "Standard CI template", GitHubActions, base)
	tmpl.SetField("project", "rendered-project")
	rendered, err := tmpl.Render()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rendered.Project != "rendered-project" {
		t.Errorf("expected 'rendered-project', got %s", rendered.Project)
	}
	if rendered.Platform != GitHubActions {
		t.Errorf("expected GitHubActions platform, got %s", rendered.Platform)
	}
}

func TestConfigTemplate_RenderNilBase(t *testing.T) {
	tmpl := NewConfigTemplate("test", "test", GitHubActions, nil)
	_, err := tmpl.Render()
	if err == nil {
		t.Error("expected error for nil base")
	}
}

func TestConfigTemplate_SetField(t *testing.T) {
	tmpl := NewConfigTemplate("test", "desc", GitHubActions, testGoConfig())
	tmpl.SetField("key1", "value1")
	tmpl.SetField("key2", "value2")
	if tmpl.MergeFields["key1"] != "value1" {
		t.Error("expected key1=value1")
	}
	if tmpl.MergeFields["key2"] != "value2" {
		t.Error("expected key2=value2")
	}
}

func TestLintError_Error(t *testing.T) {
	e := LintError{Field: "project", Message: "required"}
	if e.Error() != "project: required" {
		t.Errorf("unexpected error string: %s", e.Error())
	}
}

func TestGitHubActionsGenerator_MultipleLanguages(t *testing.T) {
	g := &GitHubActionsGenerator{}
	config := &PipelineConfig{
		Project:   "multi",
		Languages: []string{"go", "node"},
		Trigger:   TriggerConfig{OnPush: true},
	}
	output, err := g.Generate(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "go build ./...") {
		t.Error("missing go build")
	}
	if !strings.Contains(output, "npm ci") {
		t.Error("missing npm ci")
	}
}

func TestDockerComposeGenerator_JavaRedis(t *testing.T) {
	g := &DockerComposeGenerator{}
	output, err := g.Generate(testJavaConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "eclipse-temurin:21") {
		t.Error("missing java image")
	}
	if !strings.Contains(output, "redis:") {
		t.Error("missing redis for java")
	}
}

func TestDockerComposeGenerator_RustNoDB(t *testing.T) {
	g := &DockerComposeGenerator{}
	output, err := g.Generate(testRustConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(output, "db:") {
		t.Error("rust should not have db service")
	}
	if strings.Contains(output, "redis:") {
		t.Error("rust should not have redis service")
	}
}

func TestNotificationGenerator_SlackDefaultChannel(t *testing.T) {
	g := &NotificationGenerator{}
	config := &NotificationConfig{
		Type:   NotificationSlack,
		Target: "https://hooks.slack.com/test",
	}
	steps, err := g.GenerateSteps(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if steps[0].Env["SLACK_CHANNEL"] != "general" {
		t.Errorf("expected default channel 'general', got %s", steps[0].Env["SLACK_CHANNEL"])
	}
}

func TestNotificationGenerator_SlackCustomChannel(t *testing.T) {
	g := &NotificationGenerator{}
	config := &NotificationConfig{
		Type:    NotificationSlack,
		Target:  "https://hooks.slack.com/test",
		Channel: "deploys",
		Events:  []string{"success"},
	}
	steps, err := g.GenerateSteps(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if steps[0].Env["SLACK_CHANNEL"] != "deploys" {
		t.Errorf("expected channel 'deploys', got %s", steps[0].Env["SLACK_CHANNEL"])
	}
}

func TestDockerComposeGenerator_SecretsEnvFile(t *testing.T) {
	g := &DockerComposeGenerator{}
	output, err := g.Generate(testGoConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "env_file:") {
		t.Error("expected env_file for secrets")
	}
}
