package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

// Brand colors used across embeds (integers are Discord's 0xRRGGBB format).
const (
	colorCyan   = 579327   // #08d6ff
	colorBlue   = 3833080  // #3a7cf8
	colorViolet = 8149503  // #7c4dff
	colorPurple = 9647082  // #9333ea
	colorGreen  = 2607423  // #27c93f
	colorInfo   = 6337786  // #60a5fa
	colorText   = 15790320 // #f0f0f0
)

// commandHandler responds to a slash command interaction.
type commandHandler func(s *discordgo.Session, i *discordgo.InteractionCreate)

// commandDef describes a registered slash command.
type commandDef struct {
	command *discordgo.ApplicationCommand
	handler commandHandler
}

// Bot is the NAEOS Discord bot. It registers slash commands and runs a
// release/Product Hunt notification loop.
type Bot struct {
	cfg     Config
	session *discordgo.Session
	gh      *GitHubClient
	state   *State
	logger  *slog.Logger
	cmds    []commandDef
}

// NewBot constructs a Bot, creating and configuring the Discord session.
func NewBot(cfg Config, logger *slog.Logger) (*Bot, error) {
	session, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("create discord session: %w", err)
	}
	session.Identify.Intents = discordgo.IntentGuildMessages | discordgo.IntentGuildMembers

	b := &Bot{
		cfg:     cfg,
		session: session,
		gh:      NewGitHubClient(cfg.Repo, cfg.GitHubToken),
		state:   NewState(cfg.StateFile),
		logger:  logger,
	}
	return b, nil
}

// Start opens the gateway connection, registers slash commands, and starts
// background notification loops. It blocks until ctx is canceled.
func (b *Bot) Start(ctx context.Context) error {
	b.session.AddHandler(b.onReady)
	b.session.AddHandler(b.onInteractionCreate)

	if err := b.session.Open(); err != nil {
		return fmt.Errorf("open discord session: %w", err)
	}

	if err := b.registerCommands(); err != nil {
		return err
	}

	b.logger.Info("bot is running", "repo", b.cfg.Repo)
	go b.runReleaseWatcher(ctx)
	go b.runStateSaver(ctx)

	<-ctx.Done()
	b.logger.Info("shutting down")
	_ = b.session.Close()
	return nil
}

// registerCommands registers all slash commands, guild-scoped when a guild is
// configured, otherwise globally.
func (b *Bot) registerCommands() error {
	for _, def := range b.cmds {
		_, err := b.session.ApplicationCommandCreate(b.cfg.AppID, b.cfg.GuildID, def.command)
		if err != nil {
			return fmt.Errorf("register command %q: %w", def.command.Name, err)
		}
		b.logger.Info("registered command", "name", def.command.Name)
	}
	return nil
}

// announceChannelID returns the configured announcement channel, falling back
// to the channel persisted via /setup.
func (b *Bot) announceChannelID() string {
	if b.cfg.AnnounceChannelID != "" {
		return b.cfg.AnnounceChannelID
	}
	return b.state.AnnounceChannel()
}

// onReady runs once the gateway is connected.
func (b *Bot) onReady(_ *discordgo.Session, event *discordgo.Ready) {
	b.logger.Info("connected to discord", "user", event.User.Username)

	if b.cfg.AnnounceOnStartup && b.announceChannelID() != "" {
		if err := b.announcePHLaunch(b.announceChannelID()); err != nil {
			b.logger.Error("failed to post launch announcement", "err", err)
		}
	}
}

// onInteractionCreate dispatches slash command interactions to their handlers.
func (b *Bot) onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	name := i.ApplicationCommandData().Name
	for _, def := range b.cmds {
		if def.command.Name == name {
			def.handler(s, i)
			return
		}
	}
	b.logger.Warn("unknown command interaction", "name", name)
}

// respond sends an ephemeral text-only reply to an interaction.
func respond(s *discordgo.Session, i *discordgo.InteractionCreate, content string) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

// respondEmbeds sends an ephemeral reply containing embeds (and optional link buttons).
func respondEmbeds(s *discordgo.Session, i *discordgo.InteractionCreate, embeds []*discordgo.MessageEmbed, components ...discordgo.MessageComponent) error {
	data := &discordgo.InteractionResponseData{
		Embeds: embeds,
		Flags:  discordgo.MessageFlagsEphemeral,
	}
	if len(components) > 0 {
		data.Components = components
	}
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: data,
	})
}

// linkButton builds a link-style button component.
func linkButton(label, url string, style discordgo.ButtonStyle) discordgo.MessageComponent {
	return &discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			&discordgo.Button{Label: label, Style: style, URL: url},
		},
	}
}
