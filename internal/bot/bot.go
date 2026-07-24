// Package bot owns the discordgo session lifecycle and lets independent
// features (presence tracking, temp voice channels, ...) register their own
// gateway handlers against a single session.
package bot

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

// Feature is a self-contained bot capability.
type Feature interface {
	Name() string
	Intents() discordgo.Intent
	// Register adds this feature's gateway event handlers to the session.
	// Called once, before the session is opened.
	Register(s *discordgo.Session) error
}

type Bot struct {
	Session *discordgo.Session
}

func New(token string) (*Bot, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("create discord session: %w", err)
	}
	return &Bot{Session: session}, nil
}

// Use registers a feature: its intents are OR'd into the session and its
// gateway handlers are added. Must be called before Open.
func (b *Bot) Use(f Feature) error {
	b.Session.Identify.Intents |= f.Intents()

	if err := f.Register(b.Session); err != nil {
		return fmt.Errorf("register feature %q: %w", f.Name(), err)
	}
	return nil
}

// Open wires up housekeeping handlers and opens the gateway session.
func (b *Bot) Open(ctx context.Context) error {
	b.Session.AddHandler(b.clearCommands(ctx))
	return b.Session.Open()
}

func (b *Bot) Close() error {
	return b.Session.Close()
}

// clearCommands removes any slash commands registered by older versions of
// the bot (configuration now comes from the environment, so juno has none).
func (b *Bot) clearCommands(ctx context.Context) func(s *discordgo.Session, g *discordgo.GuildCreate) {
	return func(s *discordgo.Session, g *discordgo.GuildCreate) {
		if _, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, g.ID, nil, discordgo.WithContext(ctx)); err != nil {
			slog.Error("failed to clear slash commands for guild", "guild_id", g.ID, "err", err)
		}
	}
}
