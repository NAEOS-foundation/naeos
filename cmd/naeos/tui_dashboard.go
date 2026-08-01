package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	dash "github.com/NAEOS-foundation/naeos/internal/dashboard"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF00")).
			Padding(0, 1)

	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00BFFF")).
			Padding(0, 1)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Padding(0, 1)

	warnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00")).
			Padding(0, 1)

	errStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFA500")).
			Padding(0, 1)

	separator = "────────────────────────────────────────────"
)

type dashModel struct {
	stats    *dash.Stats
	activity []dash.LogEntry
	health   []dash.ComponentInfo
	al       *dash.ActivityLog
	ch       *dash.ComponentHealth
	err      error
	quitting bool
	width    int
	height   int
}

func initialModel() dashModel {
	return dashModel{
		al: dash.NewActivityLog(50),
		ch: dash.NewComponentHealth(),
		stats: &dash.Stats{
			Projects:  0,
			Pipelines: 0,
			Artifacts: 0,
			LastRun:   "never",
		},
		activity: []dash.LogEntry{},
		health:   []dash.ComponentInfo{},
	}
}

func (m dashModel) Init() tea.Cmd {
	return tea.Batch(
		tickEvery(5 * time.Second),
	)
}

func tickEvery(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type tickMsg time.Time

func (m dashModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil

	case tickMsg:
		m.stats = dash.GetStats()
		m.activity = m.al.Entries()
		m.health = m.ch.All()
		return m, tickEvery(5 * time.Second)

	case error:
		m.err = msg
		return m, nil
	}

	return m, nil
}

func (m dashModel) View() string {
	if m.quitting {
		return "\n  Dashboard closed.\n"
	}

	s := "\n"
	s += titleStyle.Render("  NAEOS Live Dashboard")
	s += "\n\n"

	s += headerStyle.Render("  Stats") + "\n"
	s += fmt.Sprintf("  %s %d\n", labelStyle.Render("Projects:"), m.stats.Projects)
	s += fmt.Sprintf("  %s %d\n", labelStyle.Render("Pipelines:"), m.stats.Pipelines)
	s += fmt.Sprintf("  %s %d\n", labelStyle.Render("Artifacts:"), m.stats.Artifacts)
	lastRun := m.stats.LastRun
	if lastRun == "" {
		lastRun = "never"
	}
	s += fmt.Sprintf("  %s %s\n", labelStyle.Render("Last Run:"), lastRun)
	s += "\n\n"

	s += headerStyle.Render("  Component Health")
	s += "\n"
	if len(m.health) == 0 {
		m.ch.Set("Dashboard", dash.Healthy, "running")
		m.ch.Set("API Server", dash.Healthy, "idle")
		m.ch.Set("Pipeline Engine", dash.Healthy, "ready")
		m.health = m.ch.All()
	}
	for _, ci := range m.health {
		statusStyle := infoStyle
		if ci.Status == dash.Degraded {
			statusStyle = warnStyle
		} else if ci.Status == dash.Unhealthy {
			statusStyle = errStyle
		}
		statusText := "healthy"
		if ci.Status == dash.Degraded {
			statusText = "degraded"
		} else if ci.Status == dash.Unhealthy {
			statusText = "unhealthy"
		}
		s += fmt.Sprintf("  %s: %s (%s)", labelStyle.Render(ci.Name), statusStyle.Render(statusText), valueStyle.Render(ci.Message))
		s += "\n"
	}
	s += "\n"

	s += headerStyle.Render("  Recent Activity")
	s += "\n"
	entries := m.activity
	if len(entries) > 10 {
		entries = entries[len(entries)-10:]
	}
	if len(entries) == 0 {
		s += "  No activity yet.\n"
	} else {
		for _, e := range entries {
			levelStyle := infoStyle
			if e.Level == dash.LevelWarn {
				levelStyle = warnStyle
			} else if e.Level == dash.LevelError {
				levelStyle = errStyle
			}
			ts := e.Timestamp.Format("15:04:05")
			s += fmt.Sprintf("  %s %s %s\n", valueStyle.Render(ts), levelStyle.Render(string(e.Level)), e.Message)
		}
	}
	s += "\n"
	s += lipgloss.NewStyle().Faint(true).Render("  [q] quit  |  auto-refresh every 5s")
	s += "\n"

	return s
}

func newTUIDashboardCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "dashboard",
		Short: "Live terminal dashboard",
		Long:  "Display a live-updating terminal dashboard showing NAEOS stats, component health, and activity.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			m := initialModel()
			p := tea.NewProgram(m, tea.WithAltScreen())
			if _, err := p.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error running dashboard: %v\n", err)
				os.Exit(1)
			}
			return nil
		},
	}
}
