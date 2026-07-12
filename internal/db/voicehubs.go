package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Hub is a configured "join to create" voice channel for a guild.
type Hub struct {
	GuildID      string
	HubChannelID string
	CategoryID   string
	NameTemplate string
	GracePeriod  time.Duration
}

func (d *DB) CreateHub(ctx context.Context, h Hub) error {
	_, err := d.conn.ExecContext(ctx, `
		INSERT INTO voice_hubs (guild_id, hub_channel_id, category_id, name_template, grace_period_seconds)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (guild_id, hub_channel_id) DO UPDATE SET
			category_id = excluded.category_id,
			name_template = excluded.name_template,
			grace_period_seconds = excluded.grace_period_seconds,
			updated_at = CURRENT_TIMESTAMP
	`, h.GuildID, h.HubChannelID, h.CategoryID, h.NameTemplate, int64(h.GracePeriod.Seconds()))
	if err != nil {
		return fmt.Errorf("create hub: %w", err)
	}
	return nil
}

func (d *DB) DeleteHub(ctx context.Context, guildID, hubChannelID string) error {
	_, err := d.conn.ExecContext(ctx,
		`DELETE FROM voice_hubs WHERE guild_id = ? AND hub_channel_id = ?`, guildID, hubChannelID)
	if err != nil {
		return fmt.Errorf("delete hub: %w", err)
	}
	return nil
}

func (d *DB) ListHubs(ctx context.Context, guildID string) ([]Hub, error) {
	rows, err := d.conn.QueryContext(ctx, `
		SELECT guild_id, hub_channel_id, category_id, name_template, grace_period_seconds
		FROM voice_hubs WHERE guild_id = ? ORDER BY created_at
	`, guildID)
	if err != nil {
		return nil, fmt.Errorf("list hubs: %w", err)
	}
	defer rows.Close()

	var hubs []Hub
	for rows.Next() {
		h, seconds, err := scanHub(rows)
		if err != nil {
			return nil, fmt.Errorf("list hubs: %w", err)
		}
		h.GracePeriod = time.Duration(seconds) * time.Second
		hubs = append(hubs, h)
	}
	return hubs, rows.Err()
}

// GetHubByChannelID looks up whether channelID is a configured hub in
// guildID. Returns (nil, nil) if it isn't - this is checked on every voice
// join, so a not-found case is expected, not an error.
func (d *DB) GetHubByChannelID(ctx context.Context, guildID, channelID string) (*Hub, error) {
	row := d.conn.QueryRowContext(ctx, `
		SELECT guild_id, hub_channel_id, category_id, name_template, grace_period_seconds
		FROM voice_hubs WHERE guild_id = ? AND hub_channel_id = ?
	`, guildID, channelID)

	h, seconds, err := scanHub(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get hub: %w", err)
	}
	h.GracePeriod = time.Duration(seconds) * time.Second
	return &h, nil
}

func (d *DB) SetGracePeriod(ctx context.Context, guildID, hubChannelID string, d2 time.Duration) error {
	res, err := d.conn.ExecContext(ctx, `
		UPDATE voice_hubs SET grace_period_seconds = ?, updated_at = CURRENT_TIMESTAMP
		WHERE guild_id = ? AND hub_channel_id = ?
	`, int64(d2.Seconds()), guildID, hubChannelID)
	if err != nil {
		return fmt.Errorf("set grace period: %w", err)
	}
	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return fmt.Errorf("set grace period: no hub %s in guild %s", hubChannelID, guildID)
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanHub(row rowScanner) (Hub, int64, error) {
	var h Hub
	var seconds int64
	err := row.Scan(&h.GuildID, &h.HubChannelID, &h.CategoryID, &h.NameTemplate, &seconds)
	return h, seconds, err
}
