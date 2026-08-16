package main

import (
	"io"
	"log/slog"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestPreLaunchMessagesContainAllChoices(t *testing.T) {
	for _, want := range []string{"welcome", "champion", "walkthrough", "today", "golive"} {
		if _, ok := preLaunchMessages[want]; !ok {
			t.Errorf("preLaunchMessages missing key %q", want)
		}
	}
}

func TestPreLaunchMessagesNonEmpty(t *testing.T) {
	for k, v := range preLaunchMessages {
		if v == "" {
			t.Errorf("preLaunchMessages[%q] is empty", k)
		}
	}
}

func TestIsAdminOrOwner(t *testing.T) {
	b := &Bot{logger: testLogger()}

	admin := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Member: &discordgo.Member{
				User:        &discordgo.User{ID: "u1"},
				Permissions: discordgo.PermissionAdministrator,
			},
		},
	}
	if !b.isAdminOrOwner(nil, admin) {
		t.Error("administrator should pass isAdminOrOwner")
	}

	regular := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Member: &discordgo.Member{
				User:        &discordgo.User{ID: "u2"},
				Permissions: 0,
			},
		},
	}
	if b.isAdminOrOwner(nil, regular) {
		t.Error("regular member should not pass isAdminOrOwner")
	}
}
