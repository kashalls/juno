package profile

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"
)

const refreshInterval = 10 * time.Minute

// RefreshWorker periodically fetches the tracked user's Discord user object
// (GET /users/{id}: banner, accent colour, avatar decoration, etc.) and
// caches it in Store. The richer GET /users/{id}/profile (bio, pronouns,
// connected accounts) rejects bot tokens with code 20001, so those fields
// aren't available.
type RefreshWorker struct {
	session *discordgo.Session
	userID  string
	store   *Store
}

func NewRefreshWorker(session *discordgo.Session, userID string, store *Store) *RefreshWorker {
	return &RefreshWorker{session: session, userID: userID, store: store}
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
	body, err := w.session.Request("GET", discordgo.EndpointUser(w.userID), nil, discordgo.WithContext(ctx))
	if err != nil {
		slog.Error("failed to fetch discord user", "user_id", w.userID, "err", err)
		return
	}

	// Wrap it as {"user": ...} so clients written against the profile
	// endpoint's shape keep reading user.banner etc. unchanged.
	wrapped, err := json.Marshal(struct {
		User json.RawMessage `json:"user"`
	}{User: body})
	if err != nil {
		slog.Error("failed to encode discord user", "user_id", w.userID, "err", err)
		return
	}
	w.store.Set(wrapped)
}
