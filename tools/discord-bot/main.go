package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg := FromEnv()
	if err := cfg.Validate(); err != nil {
		logger.Error("invalid configuration", "err", err)
		os.Exit(1)
	}

	cmd := "serve"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "provision":
		if err := runProvision(cfg, logger); err != nil {
			logger.Error("provision failed", "err", err)
			os.Exit(1)
		}
	case "serve":
		bot, err := NewBot(cfg, logger)
		if err != nil {
			logger.Error("failed to create bot", "err", err)
			os.Exit(1)
		}
		bot.cmds = bot.defaultCommands()

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		if err := bot.Start(ctx); err != nil {
			logger.Error("bot stopped with error", "err", err)
			os.Exit(1)
		}
		logger.Info("bot stopped cleanly")
	default:
		logger.Error("unknown command", "command", cmd, "hint", "usage: naeos-discord-bot [serve|provision]")
		os.Exit(2)
	}
}
