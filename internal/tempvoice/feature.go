// Package tempvoice implements join-to-create temporary voice channels for
// a single, statically configured hub channel: members who join the hub get
// their own voice channel, which is deleted again after it sits empty for
// the grace period.
package tempvoice

import (
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/kashalls/juno/internal/db"
)

const (
	nameTemplate = "{user}'s Channel"
	gracePeriod  = 5 * time.Minute
)

// Config points the feature at the one hub channel it watches.
type Config struct {
	// HubChannelID is the voice channel members join to get their own
	// temp channel.
	HubChannelID string
	// CategoryID is where temp channels are created. Optional: defaults to
	// the hub channel's own category.
	CategoryID string
}

type Feature struct {
	db  *db.DB
	cfg Config
}

func New(database *db.DB, cfg Config) *Feature {
	return &Feature{db: database, cfg: cfg}
}

func (f *Feature) Name() string { return "tempvoice" }

func (f *Feature) Intents() discordgo.Intent {
	return discordgo.IntentsGuilds | discordgo.IntentsGuildVoiceStates
}

func (f *Feature) Register(s *discordgo.Session) error {
	s.AddHandler(f.onVoiceStateUpdate)
	return nil
}
