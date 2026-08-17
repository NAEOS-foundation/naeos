package main

import (
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

// ProvisionPlan describes the full server structure from the blueprint:
// categories with channels, plus roles. Idempotent — safe to re-run.
type ProvisionPlan struct {
	Categories []CategoryPlan
	Roles      []RolePlan
	Topics     map[string]string
}

// CategoryPlan is one category with its channels.
type CategoryPlan struct {
	Name     string
	Channels []ChannelPlan
}

// ChannelPlan describes a single channel.
type ChannelPlan struct {
	Name    string
	Voice   bool
	Locked  bool // deny Send Messages for @everyone
	Private bool // only staff (Moderator + Administrator) can see/send
}

// RolePlan describes a role to create.
type RolePlan struct {
	Name  string
	Color int
	Perms int64
}

// staffRoles are the roles that can access private channels.
var staffRoles = []string{"Administrator", "Moderator"}

// defaultPlan returns the blueprint channel/role structure.
func defaultPlan() ProvisionPlan {
	return ProvisionPlan{
		Roles: []RolePlan{
			{Name: "Administrator", Color: colorPurple, Perms: discordgo.PermissionAdministrator},
			{Name: "Moderator", Color: colorViolet, Perms: discordgo.PermissionKickMembers | discordgo.PermissionBanMembers |
				discordgo.PermissionManageChannels | discordgo.PermissionManageMessages | discordgo.PermissionManageThreads |
				discordgo.PermissionVoiceMoveMembers},
			{Name: "Core Contributor", Color: colorBlue},
			{Name: "Contributor", Color: colorCyan},
			{Name: "Launch Champion", Color: 0xffaa00},
			{Name: "Bot", Color: colorInfo},
		},
		Categories: []CategoryPlan{
			{
				Name: "WELCOME",
				Channels: []ChannelPlan{
					{Name: "welcome", Locked: true},
					{Name: "rules", Locked: true},
					{Name: "announcements", Locked: true},
					{Name: "product-hunt"},
				},
			},
			{
				Name: "COMMUNITY",
				Channels: []ChannelPlan{
					{Name: "general"},
					{Name: "introductions"},
					{Name: "showcase"},
					{Name: "off-topic"},
				},
			},
			{
				Name: "ENGINEERING",
				Channels: []ChannelPlan{
					{Name: "spec-language"},
					{Name: "code-generation"},
					{Name: "ai-integration"},
					{Name: "go"},
					{Name: "plugins"},
					{Name: "architecture"},
					{Name: "roadmap"},
				},
			},
			{
				Name: "HELP",
				Channels: []ChannelPlan{
					{Name: "help-requests"},
					{Name: "troubleshooting"},
					{Name: "faq", Locked: true},
				},
			},
			{
				Name: "VOICE",
				Channels: []ChannelPlan{
					{Name: "general-vc", Voice: true},
					{Name: "dev-vc", Voice: true},
					{Name: "launch-party", Voice: true},
				},
			},
			{
				Name: "MODERATION",
				Channels: []ChannelPlan{
					{Name: "mod-chat", Private: true},
					{Name: "mod-log", Private: true, Locked: true},
				},
			},
		},
		Topics: map[string]string{
			"welcome":         "Welcome to the NAEOS Community — specify once, build anywhere. Read #rules and introduce yourself in #introductions.",
			"rules":           "Community rules — be excellent, stay on topic, search before asking, no spam. Full text: launch/discord-server/templates.md",
			"announcements":   "Official announcements: releases and launch updates.",
			"product-hunt":    "Product Hunt launch discussion and upvote link.",
			"general":         "General discussion about NAEOS and everything around it.",
			"introductions":   "New here? Introduce yourself — one line is perfect.",
			"showcase":        "Share projects built with NAEOS.",
			"off-topic":       "Casual chat. Keep it kind.",
			"spec-language":   "Spec Language v2: variables, $ref, $include, $fn, $if, migrations.",
			"code-generation": "Multi-language code generation: Go, TypeScript, Python, Java, Rust.",
			"ai-integration":  "AI compiler, MCP server, and adapters for Copilot/Claude/Cursor/Gemini/Codex/OpenCode.",
			"go":              "Go internals: kernel, LSP server, plugin SDK.",
			"plugins":         "WASM plugin SDK and official example plugins.",
			"architecture":    "NEIR model, patterns, governance and policy.",
			"roadmap":         "Roadmap, RFC and ADR discussions.",
			"help-requests":   "Ask for help. Include your spec and CLI output when relevant.",
			"troubleshooting": "Known issues and setup problems.",
			"faq":             "Frequently asked questions — read-only.",
		},
	}
}

// runProvision connects to Discord and applies the blueprint to a guild.
func runProvision(cfg Config, logger *slog.Logger) error {
	session, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		return fmt.Errorf("create discord session: %w", err)
	}
	if err := session.Open(); err != nil {
		return fmt.Errorf("open gateway: %w", err)
	}
	defer session.Close()

	guildID := cfg.GuildID
	if guildID == "" {
		guildID, err = firstGuildID(session, logger)
		if err != nil {
			return err
		}
	}

	logger.Info("provisioning guild", "guild", guildID)
	plan := defaultPlan()
	if _, err := applyRoles(session, guildID, plan.Roles, logger); err != nil {
		return err
	}
	return applyCategories(session, guildID, plan, logger)
}

func firstGuildID(session *discordgo.Session, logger *slog.Logger) (string, error) {
	guilds, err := session.UserGuilds(100, "", "", false)
	if err != nil {
		return "", fmt.Errorf("list guilds: %w", err)
	}
	if len(guilds) == 0 {
		return "", fmt.Errorf("bot is not in any guild — invite it first: https://discord.com/api/oauth2/authorize?client_id=<APP_ID>&permissions=8&scope=bot%%20applications.commands")
	}
	logger.Warn("no DISCORD_GUILD_ID set, using first guild", "guild", guilds[0].ID, "name", guilds[0].Name)
	return guilds[0].ID, nil
}

// applyRoles creates any missing roles idempotently and returns a map name->roleID.
func applyRoles(session *discordgo.Session, guildID string, roles []RolePlan, logger *slog.Logger) (map[string]string, error) {
	existing, err := session.GuildRoles(guildID)
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	byName := make(map[string]string, len(existing))
	for _, r := range existing {
		byName[r.Name] = r.ID
	}

	created := make(map[string]string, len(roles))
	for _, rp := range roles {
		if id, ok := byName[rp.Name]; ok {
			created[rp.Name] = id
			logger.Info("role exists", "role", rp.Name)
			continue
		}
		hoist := false
		params := &discordgo.RoleParams{
			Name:  rp.Name,
			Color: &rp.Color,
			Hoist: &hoist,
		}
		if rp.Perms != 0 {
			params.Permissions = &rp.Perms
		}
		role, err := session.GuildRoleCreate(guildID, params)
		if err != nil {
			return nil, fmt.Errorf("create role %q: %w", rp.Name, err)
		}
		created[rp.Name] = role.ID
		logger.Info("created role", "role", rp.Name, "color", fmt.Sprintf("#%06X", rp.Color))
	}
	return created, nil
}

// applyCategories creates categories, their channels, topics, and permission
// overwrites. Idempotent: existing channels are left untouched.
func applyCategories(session *discordgo.Session, guildID string, plan ProvisionPlan, logger *slog.Logger) error {
	channels, err := session.GuildChannels(guildID)
	if err != nil {
		return fmt.Errorf("list channels: %w", err)
	}
	existing := make(map[string]*discordgo.Channel, len(channels))
	for _, ch := range channels {
		existing[ch.Name] = ch
	}

	roleIDs := make(map[string]string)
	// Resolve @everyone as the guild ID.
	roleIDs["@everyone"] = guildID
	roles, err := session.GuildRoles(guildID)
	if err == nil {
		for _, r := range roles {
			roleIDs[r.Name] = r.ID
		}
	}

	for _, cat := range plan.Categories {
		parent, ok := existing[cat.Name]
		if !ok {
			parent, err = session.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
				Name: cat.Name,
				Type: discordgo.ChannelTypeGuildCategory,
			})
			if err != nil {
				return fmt.Errorf("create category %q: %w", cat.Name, err)
			}
			logger.Info("created category", "category", cat.Name)
			existing[cat.Name] = parent
		}

		for _, ch := range cat.Channels {
			if _, ok := existing[ch.Name]; ok {
				logger.Info("channel exists", "channel", ch.Name)
				continue
			}
			ctype := discordgo.ChannelTypeGuildText
			if ch.Voice {
				ctype = discordgo.ChannelTypeGuildVoice
			}
			created, err := session.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
				Name:     ch.Name,
				Type:     ctype,
				ParentID: parent.ID,
				Topic:    plan.Topics[ch.Name],
			})
			if err != nil {
				return fmt.Errorf("create channel %q: %w", ch.Name, err)
			}
			existing[ch.Name] = created

			if err := applyOverwrites(session, created.ID, ch, roleIDs); err != nil {
				return err
			}
			logger.Info("created channel", "channel", ch.Name, "category", cat.Name)
		}
	}
	return nil
}

// applyOverwrites sets permission rules for a channel.
func applyOverwrites(session *discordgo.Session, channelID string, ch ChannelPlan, roleIDs map[string]string) error {
	everyone := roleIDs["@everyone"]
	if everyone == "" {
		return nil
	}

	if ch.Locked {
		// @everyone can view but not send messages.
		if err := session.ChannelPermissionSet(channelID, everyone, discordgo.PermissionOverwriteTypeRole, 0, discordgo.PermissionSendMessages); err != nil {
			return fmt.Errorf("lock channel %q: %w", ch.Name, err)
		}
	}
	if ch.Private {
		// @everyone cannot view; staff can.
		if err := session.ChannelPermissionSet(channelID, everyone, discordgo.PermissionOverwriteTypeRole, 0, discordgo.PermissionViewChannel); err != nil {
			return fmt.Errorf("hide channel %q: %w", ch.Name, err)
		}
		allow := int64(discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionReadMessageHistory)
		for _, r := range staffRoles {
			id := roleIDs[r]
			if id == "" {
				continue
			}
			if err := session.ChannelPermissionSet(channelID, id, discordgo.PermissionOverwriteTypeRole, allow, 0); err != nil {
				return fmt.Errorf("grant %q on %q: %w", r, ch.Name, err)
			}
		}
	}
	return nil
}
