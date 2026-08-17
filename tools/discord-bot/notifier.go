package main

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"
)

// runReleaseWatcher polls GitHub for new releases and posts them to the
// announcement channel. It runs until ctx is canceled.
func (b *Bot) runReleaseWatcher(ctx context.Context) {
	if b.announceChannelID() == "" {
		b.logger.Info("release watcher disabled (no announcement channel set)")
		return
	}

	ticker := time.NewTicker(b.cfg.ReleasePollInterval)
	defer ticker.Stop()

	b.checkReleases(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			b.checkReleases(ctx)
		}
	}
}

// checkReleases fetches the latest release and announces it if it is new.
func (b *Bot) checkReleases(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	rel, err := b.gh.LatestRelease(ctx)
	if err != nil {
		b.logger.Warn("failed to fetch latest release", "err", err)
		return
	}

	last := b.state.LastRelease()
	if last != "" && rel.TagName == last {
		return
	}
	if last != "" && rel.TagName != last {
		b.logger.Info("new release detected", "tag", rel.TagName, "previous", last)
	} else {
		b.logger.Info("found latest release", "tag", rel.TagName)
	}

	name := rel.Name
	if name == "" {
		name = rel.TagName
	}
	body := rel.Body
	if len(body) > 800 {
		body = body[:800] + "…"
	}

	embed := &discordgo.MessageEmbed{
		Title:       "🚀 " + name + " is out",
		URL:         rel.HTMLURL,
		Description: body,
		Color:       colorPurple,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Tag", Value: rel.TagName, Inline: true},
			{Name: "Published", Value: rel.PublishedAt.UTC().Format("2006-01-02 15:04 MST"), Inline: true},
		},
		Footer: b.footer(),
	}

	if _, err := b.session.ChannelMessageSendEmbed(b.announceChannelID(), embed); err != nil {
		b.logger.Error("failed to post release announcement", "tag", rel.TagName, "err", err)
		return
	}

	b.state.SetLastRelease(rel.TagName)
	if err := b.state.Save(); err != nil {
		b.logger.Warn("failed to persist release state", "err", err)
	}
}

// announcePHLaunch posts the Product Hunt launch announcement to the given channel.
func (b *Bot) announcePHLaunch(channelID string) error {
	embed := &discordgo.MessageEmbed{
		Title:       "🎉 NAEOS is live on Product Hunt",
		URL:         b.cfg.PHLaunchURL,
		Description: b.cfg.PHTagline + "\n\n" + b.cfg.PHReleaseNote,
		Color:       colorViolet,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://raw.githubusercontent.com/NAEOS-foundation/naeos/main/brand/logo-mark.svg",
		},
		Image:  &discordgo.MessageEmbedImage{URL: b.cfg.PHGalleryURL},
		Footer: b.footer(),
	}

	msg, err := b.session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{embed},
		Components: []discordgo.MessageComponent{
			&discordgo.ActionsRow{Components: []discordgo.MessageComponent{
				&discordgo.Button{Label: "View on Product Hunt", Style: discordgo.LinkButton, URL: b.cfg.PHLaunchURL},
				&discordgo.Button{Label: "Star on GitHub", Style: discordgo.LinkButton, URL: "https://github.com/NAEOS-foundation/naeos"},
			}},
		},
	})
	if err != nil {
		return err
	}
	b.logger.Info("posted Product Hunt announcement", "channel", channelID, "message", msg.ID)
	return nil
}

// runStateSaver periodically persists bot state to disk.
func (b *Bot) runStateSaver(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := b.state.Save(); err != nil {
				b.logger.Warn("failed to save state", "err", err)
			}
		}
	}
}
