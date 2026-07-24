// Package config loads juno's runtime configuration from environment
// variables.
package config

import (
	"fmt"
	"os"
	"strings"
)

// JunoConfig configures the Discord/Lanyard presence server (cmd/juno).
type JunoConfig struct {
	Port              string
	DiscordBotToken   string
	DiscordUserID     string
	DiscordGuildID    string
	TrustedProxyCIDRs []string
	DBPath            string

	// VoiceHubChannelID is the join-to-create hub voice channel. Empty
	// disables the temp voice channel feature.
	VoiceHubChannelID string
	// VoiceHubCategoryID is where temp channels are created; defaults to
	// the hub channel's own category.
	VoiceHubCategoryID string
}

func LoadJuno() (*JunoConfig, error) {
	cfg := &JunoConfig{
		Port:               getEnvDefault("PORT", "8080"),
		DiscordBotToken:    os.Getenv("DISCORD_BOT_TOKEN"),
		DiscordUserID:      os.Getenv("DISCORD_USER_ID"),
		DiscordGuildID:     os.Getenv("DISCORD_GUILD_ID"),
		TrustedProxyCIDRs:  parseCIDRList(os.Getenv("TRUSTED_PROXY_CIDRS")),
		DBPath:             getEnvDefault("DB_PATH", "./data/juno.db"),
		VoiceHubChannelID:  os.Getenv("VOICE_HUB_CHANNEL_ID"),
		VoiceHubCategoryID: os.Getenv("VOICE_HUB_CATEGORY_ID"),
	}

	var missing []string
	if cfg.DiscordBotToken == "" {
		missing = append(missing, "DISCORD_BOT_TOKEN")
	}
	if cfg.DiscordUserID == "" {
		missing = append(missing, "DISCORD_USER_ID")
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

func parseCIDRList(v string) []string {
	var cidrs []string
	for _, c := range strings.Split(v, ",") {
		if c = strings.TrimSpace(c); c != "" {
			cidrs = append(cidrs, c)
		}
	}
	return cidrs
}
