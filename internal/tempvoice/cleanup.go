package tempvoice

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/kashalls/juno/internal/db"
)

const cleanupInterval = 30 * time.Second

// CleanupWorker periodically deletes temp channels that have been empty for
// at least their configured grace period. Deadlines are read from the DB
// (not held in memory) so they survive process restarts.
type CleanupWorker struct {
	session  *discordgo.Session
	db       *db.DB
	interval time.Duration
}

func NewCleanupWorker(session *discordgo.Session, database *db.DB) *CleanupWorker {
	return &CleanupWorker{session: session, db: database, interval: cleanupInterval}
}

// Run sweeps for expired channels every interval until ctx is canceled.
func (w *CleanupWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.sweep(ctx)
		}
	}
}

func (w *CleanupWorker) sweep(ctx context.Context) {
	expired, err := w.db.ListExpired(ctx, time.Now())
	if err != nil {
		slog.Error("failed to list expired temp channels", "err", err)
		return
	}

	for _, tc := range expired {
		_, err := w.session.ChannelDelete(tc.ChannelID, discordgo.WithContext(ctx))
		if err != nil && !isUnknownChannel(err) {
			slog.Error("failed to delete expired temp channel", "channel_id", tc.ChannelID, "err", err)
			continue
		}

		if err := w.db.DeleteTempChannel(ctx, tc.ChannelID); err != nil {
			slog.Error("failed to remove temp channel record", "channel_id", tc.ChannelID, "err", err)
		}
	}
}

func isUnknownChannel(err error) bool {
	var rerr *discordgo.RESTError
	if errors.As(err, &rerr) {
		return rerr.Message != nil && rerr.Message.Code == discordgo.ErrCodeUnknownChannel
	}
	return false
}
