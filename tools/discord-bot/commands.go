package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// defaultCommands returns the full slash command set for the bot.
func (b *Bot) defaultCommands() []commandDef {
	return []commandDef{
		{
			command: &discordgo.ApplicationCommand{
				Name:        "help",
				Description: "List all available NAEOS bot commands",
			},
			handler: b.cmdHelp,
		},
		{
			command: &discordgo.ApplicationCommand{
				Name:        "docs",
				Description: "Links to NAEOS documentation, whitepaper, and GitHub",
			},
			handler: b.cmdDocs,
		},
		{
			command: &discordgo.ApplicationCommand{
				Name:        "status",
				Description: "Show NAEOS GitHub repository status",
			},
			handler: b.cmdStatus,
		},
		{
			command: &discordgo.ApplicationCommand{
				Name:        "release",
				Description: "Show the latest NAEOS release",
			},
			handler: b.cmdRelease,
		},
		{
			command: &discordgo.ApplicationCommand{
				Name:        "producthunt",
				Description: "NAEOS on Product Hunt — launch info and link",
			},
			handler: b.cmdProductHunt,
		},
		{
			command: &discordgo.ApplicationCommand{
				Name:        "quickstart",
				Description: "Get started with NAEOS in 30 seconds",
			},
			handler: b.cmdQuickStart,
		},
		{
			command: &discordgo.ApplicationCommand{
				Name:        "doctor",
				Description: "Run naeos doctor against the local NAEOS installation",
			},
			handler: b.cmdDoctor,
		},
		{
			command: &discordgo.ApplicationCommand{
				Name:        "setup",
				Description: "Set this channel as the announcement channel for releases and Product Hunt",
			},
			handler: b.cmdSetup,
		},
		{
			command: &discordgo.ApplicationCommand{
				Name:        "config",
				Description: "Show current bot configuration",
			},
			handler: b.cmdConfig,
		},
		{
			command: &discordgo.ApplicationCommand{
				Name:        "ping",
				Description: "Check bot latency",
			},
			handler: b.cmdPing,
		},
	}
}

func (b *Bot) cmdHelp(s *discordgo.Session, i *discordgo.InteractionCreate) {
	lines := []string{"**NAEOS Bot — commands**", ""}
	for _, def := range b.cmds {
		lines = append(lines, fmt.Sprintf("`/%s` — %s", def.command.Name, def.command.Description))
	}
	lines = append(lines, "", "Specify Once. Build Anywhere.")
	_ = respond(s, i, strings.Join(lines, "\n"))
}

func (b *Bot) cmdDocs(s *discordgo.Session, i *discordgo.InteractionCreate) {
	embed := &discordgo.MessageEmbed{
		Title:       "NAEOS Documentation",
		Description: "Everything you need to go from spec to shipped software.",
		Color:       colorCyan,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Documentation", Value: "https://docs.naeos.dev", Inline: true},
			{Name: "Website", Value: "https://naeos.dev", Inline: true},
			{Name: "GitHub", Value: "https://github.com/NAEOS-foundation/naeos", Inline: true},
			{Name: "Whitepaper", Value: "https://naeos.dev/whitepaper", Inline: true},
			{Name: "Quick Start", Value: "https://docs.naeos.dev/quickstart", Inline: true},
			{Name: "Contributing", Value: "https://github.com/NAEOS-foundation/naeos/blob/main/CONTRIBUTING.md", Inline: true},
		},
		Footer: b.footer(),
	}
	_ = respondEmbeds(s, i, []*discordgo.MessageEmbed{embed},
		linkButton("Open docs", "https://docs.naeos.dev", discordgo.LinkButton),
	)
}

func (b *Bot) cmdStatus(s *discordgo.Session, i *discordgo.InteractionCreate) {
	_ = respond(s, i, "Fetching repository status…")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	st, err := b.gh.Status(ctx)
	if err != nil {
		_, _ = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
			Content: "Could not fetch repository status: " + err.Error(),
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		return
	}

	status := "active"
	statusColor := colorGreen
	if st.Archived {
		status = "archived"
		statusColor = colorViolet
	}

	embed := &discordgo.MessageEmbed{
		Title:       b.cfg.Repo,
		URL:         st.HTMLURL,
		Description: st.Description,
		Color:       statusColor,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://raw.githubusercontent.com/NAEOS-foundation/naeos/main/brand/logo-mark.svg",
		},
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Stars", Value: fmt.Sprintf("%d", st.Stars), Inline: true},
			{Name: "Forks", Value: fmt.Sprintf("%d", st.Forks), Inline: true},
			{Name: "Open Issues", Value: fmt.Sprintf("%d", st.OpenIssues), Inline: true},
			{Name: "License", Value: st.License, Inline: true},
			{Name: "Status", Value: status, Inline: true},
			{Name: "Last Updated", Value: st.UpdatedAt.UTC().Format("2006-01-02"), Inline: true},
		},
		Footer: b.footer(),
	}
	_, _ = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed},
		Flags:  discordgo.MessageFlagsEphemeral,
	})
}

func (b *Bot) cmdRelease(s *discordgo.Session, i *discordgo.InteractionCreate) {
	_ = respond(s, i, "Fetching latest release…")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	rel, err := b.gh.LatestRelease(ctx)
	if err != nil {
		_, _ = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
			Content: "Could not fetch latest release: " + err.Error(),
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		return
	}

	name := rel.Name
	if name == "" {
		name = rel.TagName
	}
	body := strings.TrimSpace(rel.Body)
	if len(body) > 500 {
		body = body[:500] + "…"
	}

	embed := &discordgo.MessageEmbed{
		Title:       name,
		URL:         rel.HTMLURL,
		Description: body,
		Color:       colorPurple,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Tag", Value: rel.TagName, Inline: true},
			{Name: "Published", Value: rel.PublishedAt.UTC().Format("2006-01-02 15:04 MST"), Inline: true},
		},
		Footer: b.footer(),
	}
	_, _ = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed},
		Flags:  discordgo.MessageFlagsEphemeral,
	})
}

func (b *Bot) cmdProductHunt(s *discordgo.Session, i *discordgo.InteractionCreate) {
	embed := &discordgo.MessageEmbed{
		Title:       "NAEOS on Product Hunt",
		Description: b.cfg.PHTagline + "\n\n" + b.cfg.PHReleaseNote,
		Color:       colorViolet,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://raw.githubusercontent.com/NAEOS-foundation/naeos/main/brand/logo-mark.svg",
		},
		Footer: b.footer(),
	}
	_ = respondEmbeds(s, i, []*discordgo.MessageEmbed{embed},
		linkButton("View on Product Hunt", b.cfg.PHLaunchURL, discordgo.LinkButton),
	)
}

func (b *Bot) cmdQuickStart(s *discordgo.Session, i *discordgo.InteractionCreate) {
	code := "```bash\n# install\ncurl -fsSL https://naeos.dev/install.sh | sh\n\n# create a project\nnaeos create my-app\n\n# run the pipeline\nnaeos run --config config.yaml --input-file spec.yaml\n```"
	_ = respond(s, i, code)
}

func (b *Bot) cmdDoctor(s *discordgo.Session, i *discordgo.InteractionCreate) {
	bin := b.cfg.NAEOSBin
	if bin == "" {
		bin = "naeos"
	}
	path, err := exec.LookPath(bin)
	if err != nil {
		_ = respond(s, i, "NAEOS binary not found. Install it: `curl -fsSL https://naeos.dev/install.sh | sh`")
		return
	}
	_ = respond(s, i, fmt.Sprintf("Running `%s doctor`…", path))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, "doctor")
	out, err := cmd.CombinedOutput()
	if err != nil {
		_, _ = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
			Content: "`naeos doctor` failed:\n```\n" + truncate(string(out), 1800) + "\n```",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		return
	}
	_, _ = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Content: "```\n" + truncate(string(out), 1800) + "\n```",
		Flags:   discordgo.MessageFlagsEphemeral,
	})
}

func (b *Bot) cmdPing(s *discordgo.Session, i *discordgo.InteractionCreate) {
	latency := s.HeartbeatLatency().Round(time.Millisecond)
	_ = respond(s, i, fmt.Sprintf("Pong! Latency: %s", latency))
}

func (b *Bot) cmdSetup(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		_ = respond(s, i, "`/setup` must be run inside a server text channel.")
		return
	}
	channelID := i.ChannelID

	b.state.SetAnnounceChannel(channelID)
	if err := b.state.Save(); err != nil {
		b.logger.Error("failed to persist setup", "err", err)
		_ = respond(s, i, "Setup failed: could not persist state ("+err.Error()+")")
		return
	}

	b.logger.Info("announcement channel set", "channel", channelID, "guild", i.GuildID)

	channelName := channelID
	if ch, err := s.Channel(channelID); err == nil {
		channelName = ch.Name
	}

	_ = respond(s, i, fmt.Sprintf(
		"Announcement channel set to **#%s**. Releases and Product Hunt updates will be posted there.\n"+
			"Persisted to `%s` — survives restarts.", channelName, b.cfg.StateFile))

	embed := &discordgo.MessageEmbed{
		Title:       "NAEOS announcement channel ready",
		Description: "This channel will receive new GitHub releases and launch updates.",
		Color:       colorGreen,
		Footer:      b.footer(),
	}
	if _, err := s.ChannelMessageSendEmbed(channelID, embed); err != nil {
		b.logger.Warn("failed to send confirmation embed", "err", err)
	}
}

func (b *Bot) cmdConfig(s *discordgo.Session, i *discordgo.InteractionCreate) {
	channel := b.announceChannelID()
	if channel == "" {
		channel = "not set — run /setup in a channel"
	}

	latency := s.HeartbeatLatency().Round(time.Millisecond)
	_ = respond(s, i, fmt.Sprintf(
		"**NAEOS Bot config**\n"+
			"- Repo: `%s`\n"+
			"- Announcement channel: `%s`\n"+
			"- Poll interval: `%s`\n"+
			"- Product Hunt: `%s`\n"+
			"- State file: `%s`\n"+
			"- Latency: `%s`",
		b.cfg.Repo, channel, b.cfg.ReleasePollInterval, b.cfg.PHLaunchURL, b.cfg.StateFile, latency))
}

func (b *Bot) footer() *discordgo.MessageEmbedFooter {
	return &discordgo.MessageEmbedFooter{
		Text:    "NAEOS · Specify Once. Build Anywhere.",
		IconURL: "https://raw.githubusercontent.com/NAEOS-foundation/naeos/main/brand/logo-mark.svg",
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
