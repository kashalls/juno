package presence

import (
	"log/slog"
	"sync"

	"github.com/bwmarrin/discordgo"
)

// Feature tracks a single Discord user's presence in a single guild and
// caches it in Store for the REST/websocket API to serve.
type Feature struct {
	userID  string
	guildID string
	store   *Store

	userMu     sync.Mutex
	cachedUser *discordgo.User
}

func NewFeature(userID, guildID string, store *Store) *Feature {
	return &Feature{
		userID:  userID,
		guildID: guildID,
		store:   store,
	}
}

func (f *Feature) Name() string { return "presence" }

func (f *Feature) Intents() discordgo.Intent {
	return discordgo.IntentsGuilds | discordgo.IntentsGuildMembers | discordgo.IntentsGuildPresences
}

func (f *Feature) Register(s *discordgo.Session) error {
	s.AddHandler(f.onGuildCreate)
	s.AddHandler(f.onPresenceUpdate)
	return nil
}

func (f *Feature) onGuildCreate(s *discordgo.Session, g *discordgo.GuildCreate) {
	if f.guildID != "" && g.ID != f.guildID {
		return
	}
	for _, p := range g.Presences {
		if p.User != nil && p.User.ID == f.userID {
			user := f.resolveUser(s, g.ID, p.User)
			f.store.Set(buildPresence(user, p.Status, p.ClientStatus, p.Activities))
			return
		}
	}
}

func (f *Feature) onPresenceUpdate(s *discordgo.Session, p *discordgo.PresenceUpdate) {
	if p.User == nil || p.User.ID != f.userID {
		return
	}
	if f.guildID != "" && p.GuildID != f.guildID {
		return
	}

	// The gateway payload's user object is often partial (id only); the
	// session state merges it with prior data, so read the merged copy back.
	merged, err := s.State.Presence(p.GuildID, f.userID)
	if err != nil {
		slog.Warn("could not read merged presence from state", "err", err)
		merged = &p.Presence
	}

	user := f.resolveUser(s, p.GuildID, merged.User)
	slog.Debug("presence update", "user_id", f.userID, "status", merged.Status)
	f.store.Set(buildPresence(user, merged.Status, merged.ClientStatus, merged.Activities))
}

// resolveUser fills in username/avatar/discriminator for the tracked user.
// Discord's gateway presence payloads (both in GUILD_CREATE and
// PRESENCE_UPDATE) almost always carry a partial user object - id only -
// since those fields are only attached to member events, not presence ones.
// The Guild Member REST endpoint always returns the full nested user, so it
// backs a cache that's used whenever the gateway data is incomplete.
func (f *Feature) resolveUser(s *discordgo.Session, guildID string, partial *discordgo.User) *discordgo.User {
	f.userMu.Lock()
	defer f.userMu.Unlock()

	if partial != nil && partial.Username != "" {
		f.cachedUser = partial
		return partial
	}

	if f.cachedUser != nil {
		return f.cachedUser
	}

	member, err := s.GuildMember(guildID, f.userID)
	if err != nil {
		slog.Warn("could not fetch guild member for user details", "user_id", f.userID, "err", err)
		return partial
	}

	f.cachedUser = member.User
	return member.User
}
