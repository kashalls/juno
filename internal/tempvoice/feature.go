// Package tempvoice implements join-to-create temporary voice channels:
// admins configure "hub" channels via the /voice-hub slash command, and
// members who join a hub get their own voice channel, which is deleted
// again after it sits empty for a configurable grace period.
package tempvoice

import (
	"github.com/bwmarrin/discordgo"

	"github.com/kashalls/juno/internal/db"
)

const (
	defaultNameTemplate = "{user}'s Channel"
	defaultGracePeriod  = 300 // seconds
)

type Feature struct {
	db *db.DB
}

func New(database *db.DB) *Feature {
	return &Feature{db: database}
}

func (f *Feature) Name() string { return "tempvoice" }

func (f *Feature) Intents() discordgo.Intent {
	return discordgo.IntentsGuilds | discordgo.IntentsGuildVoiceStates
}

func (f *Feature) Register(s *discordgo.Session) error {
	s.AddHandler(f.onVoiceStateUpdate)
	return nil
}

func (f *Feature) Commands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{voiceHubCommand}
}
