// Package config loads juno's runtime configuration from environment
// variables.
package config

import (
	"fmt"
	"os"
)

// JunoConfig configures the Discord bot (cmd/juno).
type JunoConfig struct {
	DiscordBotToken string
	DBPath          string

	// VoiceHubChannelID is the join-to-create hub voice channel. Empty
	// disables the temp voice channel feature.
	VoiceHubChannelID string
	// VoiceHubCategoryID is where temp channels are created; defaults to
	// the hub channel's own category.
	VoiceHubCategoryID string
}

func LoadJuno() (*JunoConfig, error) {
	cfg := &JunoConfig{
		DiscordBotToken:    os.Getenv("DISCORD_BOT_TOKEN"),
		DBPath:             getEnvDefault("DB_PATH", "./data/juno.db"),
		VoiceHubChannelID:  os.Getenv("VOICE_HUB_CHANNEL_ID"),
		VoiceHubCategoryID: os.Getenv("VOICE_HUB_CATEGORY_ID"),
	}

	var missing []string
	if cfg.DiscordBotToken == "" {
		missing = append(missing, "DISCORD_BOT_TOKEN")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %v", missing)
	}

	return cfg, nil
}

func getEnvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
