// Package bot owns the discordgo session lifecycle and lets independent
// features (presence tracking, temp voice channels, ...) register their own
// gateway handlers and slash commands against a single session.
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
	// Commands returns this feature's top-level slash commands, or nil.
	Commands() []*discordgo.ApplicationCommand
}

// CommandHandler is implemented by features that return non-nil Commands().
type CommandHandler interface {
	HandleCommand(s *discordgo.Session, i *discordgo.InteractionCreate)
}

type Bot struct {
	Session *discordgo.Session

	commandOwners map[string]CommandHandler
	allCommands   []*discordgo.ApplicationCommand
}

func New(token string) (*Bot, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("create discord session: %w", err)
	}
	return &Bot{
		Session:       session,
		commandOwners: make(map[string]CommandHandler),
	}, nil
}

// Use registers a feature: its intents are OR'd into the session, its
// gateway handlers are added, and its slash commands (if any) are recorded
// for registration on Open. Must be called before Open.
func (b *Bot) Use(f Feature) error {
	b.Session.Identify.Intents |= f.Intents()

	if err := f.Register(b.Session); err != nil {
		return fmt.Errorf("register feature %q: %w", f.Name(), err)
	}

	commands := f.Commands()
	if len(commands) == 0 {
		return nil
	}

	handler, ok := f.(CommandHandler)
	if !ok {
		return fmt.Errorf("feature %q returned commands but does not implement CommandHandler", f.Name())
	}

	for _, cmd := range commands {
		if _, exists := b.commandOwners[cmd.Name]; exists {
			return fmt.Errorf("command %q already registered by another feature", cmd.Name)
		}
		b.commandOwners[cmd.Name] = handler
		b.allCommands = append(b.allCommands, cmd)
	}
	return nil
}

// Open wires up command registration/routing and opens the gateway session.
func (b *Bot) Open(ctx context.Context) error {
	b.Session.AddHandler(b.onGuildCreate(ctx))
	b.Session.AddHandler(b.onInteractionCreate)
	return b.Session.Open()
}

func (b *Bot) Close() error {
	return b.Session.Close()
}

// onGuildCreate bulk-overwrites this bot's slash commands for every guild it
// connects to. GuildCreate fires for each guild on initial connect and for
// any guild the bot joins afterward, so this alone keeps commands in sync
// across every guild without a separate join hook.
func (b *Bot) onGuildCreate(ctx context.Context) func(s *discordgo.Session, g *discordgo.GuildCreate) {
	return func(s *discordgo.Session, g *discordgo.GuildCreate) {
		if len(b.allCommands) == 0 {
			return
		}
		if _, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, g.ID, b.allCommands, discordgo.WithContext(ctx)); err != nil {
			slog.Error("failed to register slash commands for guild", "guild_id", g.ID, "err", err)
		}
	}
}

func (b *Bot) onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	handler, ok := b.commandOwners[i.ApplicationCommandData().Name]
	if !ok {
		return
	}
	handler.HandleCommand(s, i)
}
