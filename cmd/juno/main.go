package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/kashalls/juno/internal/api"
	"github.com/kashalls/juno/internal/appicons"
	"github.com/kashalls/juno/internal/bot"
	"github.com/kashalls/juno/internal/config"
	"github.com/kashalls/juno/internal/db"
	"github.com/kashalls/juno/internal/lanyard"
	"github.com/kashalls/juno/internal/presence"
	"github.com/kashalls/juno/internal/profile"
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

	store := presence.NewStore()

	b, err := bot.New(cfg.DiscordBotToken)
	if err != nil {
		return err
	}
	if err := b.Use(presence.NewFeature(cfg.DiscordUserID, cfg.DiscordGuildID, store)); err != nil {
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

	hub := lanyard.NewHub(store)

	profileStore := profile.NewStore()
	profileWorker := profile.NewRefreshWorker(b.Session, cfg.DiscordUserID, cfg.DiscordGuildID, profileStore)
	go profileWorker.Run(ctx)

	appIconResolver := appicons.NewResolver()

	router := api.NewRouter(api.RouterConfig{
		Store:             store,
		Hub:               hub,
		ProfileStore:      profileStore,
		AppIconResolver:   appIconResolver,
		TrustedProxyCIDRs: cfg.TrustedProxyCIDRs,
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serveErr <- err
		}
	}()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
	}

	slog.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
