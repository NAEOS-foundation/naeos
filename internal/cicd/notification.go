package cicd

import (
	"fmt"
	"strings"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

type NotificationType string

const (
	NotificationSlack   NotificationType = "slack"
	NotificationEmail   NotificationType = "email"
	NotificationWebhook NotificationType = "webhook"
)

type NotificationConfig struct {
	Type    NotificationType
	Target  string
	Channel string
	Events  []string
}

type NotificationGenerator struct{}

func (n *NotificationGenerator) Name() string {
	return "Notification Generator"
}

func (n *NotificationGenerator) GenerateSteps(config *NotificationConfig) ([]PipelineStep, error) {
	if config == nil {
		return nil, naeoserr.New(naeoserr.ErrPipeline, "notification config is nil")
	}

	switch config.Type {
	case NotificationSlack:
		return n.generateSlackSteps(config)
	case NotificationEmail:
		return n.generateEmailSteps(config)
	case NotificationWebhook:
		return n.generateWebhookSteps(config)
	default:
		return nil, naeoserr.New(naeoserr.ErrPipeline, fmt.Sprintf("unsupported notification type: %s", config.Type))
	}
}

func (n *NotificationGenerator) generateSlackSteps(config *NotificationConfig) ([]PipelineStep, error) {
	channel := config.Channel
	if channel == "" {
		channel = "general"
	}
	if config.Target == "" {
		return nil, naeoserr.New(naeoserr.ErrPipeline, "slack webhook URL is required")
	}

	var steps []PipelineStep
	for _, event := range config.Events {
		steps = append(steps, PipelineStep{
			Name:    fmt.Sprintf("Notify Slack on %s", event),
			Command: fmt.Sprintf("curl -X POST -H 'Content-type: application/json' --data '{\"text\":\"Pipeline %s for %s\"}' %s", event, channel, config.Target),
			Env: map[string]string{
				"SLACK_CHANNEL": channel,
			},
		})
	}
	if len(config.Events) == 0 {
		steps = append(steps, PipelineStep{
			Name:    "Notify Slack on completion",
			Command: fmt.Sprintf("curl -X POST -H 'Content-type: application/json' --data '{\"text\":\"Pipeline completed\"}' %s", config.Target),
			Env: map[string]string{
				"SLACK_CHANNEL": channel,
			},
		})
	}

	return steps, nil
}

func (n *NotificationGenerator) generateEmailSteps(config *NotificationConfig) ([]PipelineStep, error) {
	if config.Target == "" {
		return nil, naeoserr.New(naeoserr.ErrPipeline, "email recipient is required")
	}

	var steps []PipelineStep
	for _, event := range config.Events {
		steps = append(steps, PipelineStep{
			Name:    fmt.Sprintf("Email notification on %s", event),
			Command: fmt.Sprintf("echo 'Pipeline %s' | mail -s 'CI/CD Notification' %s", event, config.Target),
			Env: map[string]string{
				"EMAIL_TO": config.Target,
			},
		})
	}
	if len(config.Events) == 0 {
		steps = append(steps, PipelineStep{
			Name:    "Email notification on completion",
			Command: fmt.Sprintf("echo 'Pipeline completed' | mail -s 'CI/CD Notification' %s", config.Target),
			Env: map[string]string{
				"EMAIL_TO": config.Target,
			},
		})
	}

	return steps, nil
}

func (n *NotificationGenerator) generateWebhookSteps(config *NotificationConfig) ([]PipelineStep, error) {
	if config.Target == "" {
		return nil, naeoserr.New(naeoserr.ErrPipeline, "webhook URL is required")
	}

	var steps []PipelineStep
	for _, event := range config.Events {
		steps = append(steps, PipelineStep{
			Name:    fmt.Sprintf("Webhook notification on %s", event),
			Command: fmt.Sprintf("curl -X POST -H 'Content-Type: application/json' -d '{\"event\":\"%s\",\"status\":\"completed\"}' %s", event, config.Target),
			Env: map[string]string{
				"WEBHOOK_URL": config.Target,
			},
		})
	}
	if len(config.Events) == 0 {
		steps = append(steps, PipelineStep{
			Name:    "Webhook notification on completion",
			Command: fmt.Sprintf("curl -X POST -H 'Content-Type: application/json' -d '{\"event\":\"completed\",\"status\":\"success\"}' %s", config.Target),
			Env: map[string]string{
				"WEBHOOK_URL": config.Target,
			},
		})
	}

	return steps, nil
}

func EmbedNotifications(config *PipelineConfig, notifications []*NotificationConfig) error {
	if config == nil {
		return naeoserr.New(naeoserr.ErrPipeline, "pipeline config is nil")
	}

	gen := &NotificationGenerator{}
	for _, notif := range notifications {
		steps, err := gen.GenerateSteps(notif)
		if err != nil {
			return naeoserr.Wrapf(err, naeoserr.ErrPipeline, "failed to generate notification steps")
		}
		config.Steps = append(config.Steps, steps...)
	}

	return nil
}

func GenerateNotificationBlock(config *PipelineConfig, notifications []*NotificationConfig) (string, error) {
	gen := &NotificationGenerator{}
	var allSteps []PipelineStep

	for _, notif := range notifications {
		steps, err := gen.GenerateSteps(notif)
		if err != nil {
			return "", naeoserr.Wrapf(err, naeoserr.ErrPipeline, "failed to generate notification steps")
		}
		allSteps = append(allSteps, steps...)
	}

	switch config.Platform {
	case GitHubActions:
		return generateGitHubNotificationBlock(allSteps), nil
	case GitLabCI:
		return generateGitLabNotificationBlock(allSteps), nil
	}

	var sb strings.Builder
	for _, step := range allSteps {
		fmt.Fprintf(&sb, "      - name: %s\n", step.Name)
		fmt.Fprintf(&sb, "        run: %s\n\n", step.Command)
	}
	return sb.String(), nil
}

func generateGitHubNotificationBlock(steps []PipelineStep) string {
	var sb strings.Builder
	for _, step := range steps {
		fmt.Fprintf(&sb, "      - name: %s\n", step.Name)
		fmt.Fprintf(&sb, "        run: |\n          %s\n\n", strings.ReplaceAll(step.Command, "\n", "\n          "))
	}
	return sb.String()
}

func generateGitLabNotificationBlock(steps []PipelineStep) string {
	var sb strings.Builder
	sb.WriteString("notify:\n")
	sb.WriteString("  stage: deploy\n")
	sb.WriteString("  script:\n")
	for _, step := range steps {
		fmt.Fprintf(&sb, "    - %s\n", step.Command)
	}
	sb.WriteString("\n")
	return sb.String()
}

func FormatNotificationStepsYAML(steps []PipelineStep, platform CICDPlatform) string {
	var sb strings.Builder

	switch platform {
	case GitHubActions:
		for _, step := range steps {
			fmt.Fprintf(&sb, "      - name: %s\n", step.Name)
			if len(step.Env) > 0 {
				sb.WriteString("        env:\n")
				for k, v := range step.Env {
					fmt.Fprintf(&sb, "          %s: %s\n", k, v)
				}
			}
			fmt.Fprintf(&sb, "        run: %s\n\n", step.Command)
		}
	case GitLabCI:
		for _, step := range steps {
			fmt.Fprintf(&sb, "    - %s\n", step.Command)
		}
	}

	return sb.String()
}
