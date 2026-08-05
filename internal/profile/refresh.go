package profile

import (
	"context"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"
)

const refreshInterval = 10 * time.Minute

// RefreshWorker periodically fetches the tracked user's Discord profile
// (Discord's undocumented GET /users/{id}/profile endpoint - bio, badges,
// banner, connected accounts, etc.) and caches it in Store.
type RefreshWorker struct {
	session *discordgo.Session
	userID  string
	guildID string
	store   *Store
}

func NewRefreshWorker(session *discordgo.Session, userID, guildID string, store *Store) *RefreshWorker {
	return &RefreshWorker{session: session, userID: userID, guildID: guildID, store: store}
}

// Run fetches immediately, then re-fetches every refreshInterval until ctx
// is canceled.
func (w *RefreshWorker) Run(ctx context.Context) {
	w.refresh(ctx)

	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.refresh(ctx)
		}
	}
}

func (w *RefreshWorker) refresh(ctx context.Context) {
	endpoint := discordgo.EndpointUser(w.userID) + "/profile"
	if w.guildID != "" {
		endpoint += "?guild_id=" + w.guildID
	}

	body, err := w.session.Request("GET", endpoint, nil, discordgo.WithContext(ctx))
	if err != nil {
		slog.Error("failed to fetch discord profile", "user_id", w.userID, "err", err)
		return
	}
	w.store.Set(body)
}
