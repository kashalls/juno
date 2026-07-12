CREATE TABLE IF NOT EXISTS voice_hubs (
    id                   INTEGER PRIMARY KEY AUTOINCREMENT,
    guild_id             TEXT NOT NULL,
    hub_channel_id       TEXT NOT NULL,
    category_id          TEXT NOT NULL DEFAULT '',
    name_template        TEXT NOT NULL DEFAULT '{user}''s Channel',
    grace_period_seconds INTEGER NOT NULL DEFAULT 300,
    created_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (guild_id, hub_channel_id)
);
CREATE INDEX IF NOT EXISTS idx_voice_hubs_guild ON voice_hubs(guild_id);

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
