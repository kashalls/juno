-- Hub settings moved to environment variables; drop the table older
-- versions used to store them.
DROP TABLE IF EXISTS voice_hubs;

CREATE TABLE IF NOT EXISTS temp_channels (
    channel_id           TEXT PRIMARY KEY,
    guild_id             TEXT NOT NULL,
    hub_channel_id       TEXT NOT NULL,
    owner_id             TEXT NOT NULL,
    grace_period_seconds INTEGER NOT NULL,
    created_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    empty_since           TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_temp_channels_empty ON temp_channels(empty_since) WHERE empty_since IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_temp_channels_owner_hub ON temp_channels(owner_id, hub_channel_id);
