package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/kashalls/juno/internal/bot"
	"github.com/kashalls/juno/internal/config"
	"github.com/kashalls/juno/internal/db"
	"github.com/kashalls/juno/internal/tempvoice"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal error", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.LoadJuno()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if dir := filepath.Dir(cfg.DBPath); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create db directory: %w", err)
		}
	}
	database, err := db.Open(ctx, cfg.DBPath)
	if err != nil {
		return err
	}
	// b.Close() (deferred below, registered after this) stops the gateway
	// session before the database closes, so no in-flight handler can write
	// to a closed DB - Go defers run LIFO.
	defer database.Close()

	b, err := bot.New(cfg.DiscordBotToken)
	if err != nil {
		return err
	}
	if cfg.VoiceHubChannelID != "" {
		if err := b.Use(tempvoice.New(database, tempvoice.Config{
			HubChannelID: cfg.VoiceHubChannelID,
			CategoryID:   cfg.VoiceHubCategoryID,
		})); err != nil {
			return err
		}
	}
	if err := b.Open(ctx); err != nil {
		return err
	}
	defer b.Close()

	// The cleanup worker runs even when the hub is unset so temp channels
	// left over from a previous configuration still get deleted.
	cleanup := tempvoice.NewCleanupWorker(b.Session, database)
	go cleanup.Run(ctx)

	<-ctx.Done()
	slog.Info("shutting down")
	return nil
}
