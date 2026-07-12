package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// TempChannel is a voice channel juno created on behalf of a hub join.
type TempChannel struct {
	ChannelID    string
	GuildID      string
	HubChannelID string
	OwnerID      string
	GracePeriod  time.Duration
	EmptySince   *time.Time
}

func (d *DB) InsertTempChannel(ctx context.Context, tc TempChannel) error {
	_, err := d.conn.ExecContext(ctx, `
		INSERT INTO temp_channels (channel_id, guild_id, hub_channel_id, owner_id, grace_period_seconds)
		VALUES (?, ?, ?, ?, ?)
	`, tc.ChannelID, tc.GuildID, tc.HubChannelID, tc.OwnerID, int64(tc.GracePeriod.Seconds()))
	if err != nil {
		return fmt.Errorf("insert temp channel: %w", err)
	}
	return nil
}

// GetTempChannel looks up a temp channel by its channel ID. Returns
// (nil, nil) if channelID isn't a bot-managed temp channel.
func (d *DB) GetTempChannel(ctx context.Context, channelID string) (*TempChannel, error) {
	row := d.conn.QueryRowContext(ctx, `
		SELECT channel_id, guild_id, hub_channel_id, owner_id, grace_period_seconds, empty_since
		FROM temp_channels WHERE channel_id = ?
	`, channelID)

	tc, err := scanTempChannel(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get temp channel: %w", err)
	}
	return &tc, nil
}

// GetTempChannelByOwnerAndHub finds a still-tracked temp channel the owner
// already has open from the given hub, if any. Used to dedupe rapid/duplicate
// join events instead of spawning a second channel. Returns (nil, nil) if
// none exists.
func (d *DB) GetTempChannelByOwnerAndHub(ctx context.Context, ownerID, hubChannelID string) (*TempChannel, error) {
	row := d.conn.QueryRowContext(ctx, `
		SELECT channel_id, guild_id, hub_channel_id, owner_id, grace_period_seconds, empty_since
		FROM temp_channels WHERE owner_id = ? AND hub_channel_id = ?
	`, ownerID, hubChannelID)

	tc, err := scanTempChannel(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get temp channel by owner and hub: %w", err)
	}
	return &tc, nil
}

func (d *DB) MarkEmpty(ctx context.Context, channelID string, at time.Time) error {
	_, err := d.conn.ExecContext(ctx,
		`UPDATE temp_channels SET empty_since = ? WHERE channel_id = ?`, at, channelID)
	if err != nil {
		return fmt.Errorf("mark empty: %w", err)
	}
	return nil
}

func (d *DB) ClearEmpty(ctx context.Context, channelID string) error {
	_, err := d.conn.ExecContext(ctx,
		`UPDATE temp_channels SET empty_since = NULL WHERE channel_id = ?`, channelID)
	if err != nil {
		return fmt.Errorf("clear empty: %w", err)
	}
	return nil
}

func (d *DB) DeleteTempChannel(ctx context.Context, channelID string) error {
	_, err := d.conn.ExecContext(ctx, `DELETE FROM temp_channels WHERE channel_id = ?`, channelID)
	if err != nil {
		return fmt.Errorf("delete temp channel: %w", err)
	}
	return nil
}

// ListExpired returns temp channels that have been empty for at least their
// configured grace period as of now.
func (d *DB) ListExpired(ctx context.Context, now time.Time) ([]TempChannel, error) {
	rows, err := d.conn.QueryContext(ctx, `
		SELECT channel_id, guild_id, hub_channel_id, owner_id, grace_period_seconds, empty_since
		FROM temp_channels WHERE empty_since IS NOT NULL
	`)
	if err != nil {
		return nil, fmt.Errorf("list expired: %w", err)
	}
	defer rows.Close()

	var expired []TempChannel
	for rows.Next() {
		tc, err := scanTempChannel(rows)
		if err != nil {
			return nil, fmt.Errorf("list expired: %w", err)
		}
		if tc.EmptySince != nil && now.Sub(*tc.EmptySince) >= tc.GracePeriod {
			expired = append(expired, tc)
		}
	}
	return expired, rows.Err()
}

func scanTempChannel(row rowScanner) (TempChannel, error) {
	var tc TempChannel
	var seconds int64
	var emptySince sql.NullTime
	if err := row.Scan(&tc.ChannelID, &tc.GuildID, &tc.HubChannelID, &tc.OwnerID, &seconds, &emptySince); err != nil {
		return TempChannel{}, err
	}
	tc.GracePeriod = time.Duration(seconds) * time.Second
	if emptySince.Valid {
		tc.EmptySince = &emptySince.Time
	}
	return tc, nil
}
