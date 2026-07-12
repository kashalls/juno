package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kashalls/juno/internal/api"
	"github.com/kashalls/juno/internal/config"
	"github.com/kashalls/juno/internal/discord"
	"github.com/kashalls/juno/internal/lanyard"
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

	store := discord.NewStore()

	discordClient, err := discord.NewClient(cfg.DiscordBotToken, cfg.DiscordUserID, cfg.DiscordGuildID, store)
	if err != nil {
		return err
	}
	if err := discordClient.Open(); err != nil {
		return err
	}
	defer discordClient.Close()

	hub := lanyard.NewHub(store)

	router := api.NewRouter(api.RouterConfig{
		Store:             store,
		DiscordUserID:     cfg.DiscordUserID,
		Hub:               hub,
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
