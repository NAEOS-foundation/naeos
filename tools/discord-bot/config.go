package main

import (
	"errors"
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration for the Discord bot.
// Values are loaded from environment variables; see the README for the full list.
type Config struct {
	// Discord credentials.
	Token   string
	AppID   string
	GuildID string // optional; guild-scoped slash commands when set

	// Announcement channel for release + Product Hunt notifications (optional).
	AnnounceChannelID string

	// GitHub repository to watch (owner/repo).
	Repo string

	// GitHub token (optional, avoids rate limits for the status/release commands).
	GitHubToken string

	// Release notifier settings.
	ReleasePollInterval time.Duration
	StateFile           string
	AnnounceOnStartup   bool

	// Product Hunt launch metadata.
	PHLaunchURL   string
	PHTagline     string
	PHGalleryURL  string
	PHReleaseNote string

	// Local naeos binary path override (optional; default looks up "naeos" in PATH).
	NAEOSBin string
}

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Repo:                "NAEOS-foundation/naeos",
		ReleasePollInterval: 15 * time.Minute,
		StateFile:           ".naeos-bot-state.json",
		PHTagline:           "Specify Once. Build Anywhere.",
		PHGalleryURL:        "https://raw.githubusercontent.com/NAEOS-foundation/naeos/main/launch/producthunt/assets/01-hero-cover.png",
		PHReleaseNote:       "v3.0.0 is live — pipeline profiling, stage caching, NEIR-aware LSP, distributed builds.",
	}
}

// FromEnv loads configuration from environment variables.
func FromEnv() Config {
	cfg := DefaultConfig()

	if v := os.Getenv("DISCORD_TOKEN"); v != "" {
		cfg.Token = v
	}
	if v := os.Getenv("DISCORD_APP_ID"); v != "" {
		cfg.AppID = v
	} else if v := os.Getenv("DISCORD_APPLICATION_ID"); v != "" {
		cfg.AppID = v
	} else if v := os.Getenv("DISCORD_CLIENT_ID"); v != "" {
		cfg.AppID = v
	}
	if v := os.Getenv("DISCORD_GUILD_ID"); v != "" {
		cfg.GuildID = v
	} else if v := os.Getenv("DISCORD_SERVER_ID"); v != "" {
		cfg.GuildID = v
	}
	if v := os.Getenv("DISCORD_ANNOUNCE_CHANNEL"); v != "" {
		cfg.AnnounceChannelID = v
	}
	if v := os.Getenv("NAEOS_REPO"); v != "" {
		cfg.Repo = v
	}
	if v := os.Getenv("GITHUB_TOKEN"); v != "" {
		cfg.GitHubToken = v
	}
	if v := os.Getenv("NAEOS_POLL_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.ReleasePollInterval = d
		}
	}
	if v := os.Getenv("NAEOS_STATE_FILE"); v != "" {
		cfg.StateFile = v
	}
	if v := os.Getenv("NAEOS_ANNOUNCE_ON_STARTUP"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.AnnounceOnStartup = b
		}
	}
	if v := os.Getenv("NAEOS_PH_LAUNCH_URL"); v != "" {
		cfg.PHLaunchURL = v
	}
	if v := os.Getenv("NAEOS_PH_TAGLINE"); v != "" {
		cfg.PHTagline = v
	}
	if v := os.Getenv("NAEOS_PH_GALLERY_URL"); v != "" {
		cfg.PHGalleryURL = v
	}
	if v := os.Getenv("NAEOS_PH_RELEASE_NOTE"); v != "" {
		cfg.PHReleaseNote = v
	}
	if v := os.Getenv("NAEOS_BIN"); v != "" {
		cfg.NAEOSBin = v
	}

	return cfg
}

// Validate checks required fields and returns an error for the first problem found.
func (c Config) Validate() error {
	if c.Token == "" {
		return errors.New("DISCORD_TOKEN is required")
	}
	if c.AppID == "" {
		return errors.New("DISCORD_APP_ID is required")
	}
	if c.Repo == "" {
		return errors.New("NAEOS_REPO is required (format: owner/repo)")
	}
	if c.ReleasePollInterval <= 0 {
		return errors.New("NAEOS_POLL_INTERVAL must be positive")
	}
	if c.StateFile == "" {
		return errors.New("NAEOS_STATE_FILE is required")
	}
	return nil
}
