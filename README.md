# juno

A Discord bot with two independent features:

- A [Lanyard](https://github.com/Phineas/lanyard)-compatible REST + WebSocket feed of your live Discord status, scoped to a single Discord user.
- Join-to-create temporary voice channels for a single "hub" channel.

Juno is built for exactly one guild: everything is configured through environment variables, with no in-Discord configuration commands.

## How it works

A bot account joins your server and, with the Presence and Server Members privileged intents enabled, receives real-time presence updates from Discord's Gateway for your user ID. Juno caches your latest presence in memory and serves it over REST and a Lanyard-shaped WebSocket protocol.

Independently, members who join the hub voice channel (`VOICE_HUB_CHANNEL_ID`) get their own temporary voice channel (with permission to manage it), and are moved into it automatically. The channel is deleted again once it's been empty for five minutes. Open temp channels are tracked in an embedded SQLite database so cleanup survives restarts.

## One-time setup

### Discord bot

1. Create an application at the [Discord Developer Portal](https://discord.com/developers/applications) and add a Bot user.
2. Under Bot settings, enable the **Presence Intent** and **Server Members Intent** privileged intents.
3. Invite the bot with the `bot` OAuth2 scope, and the **Manage Channels** and **Move Members** bot permissions (needed for the temp voice channel feature; the presence feature needs no permissions).
4. Copy the bot token into `DISCORD_BOT_TOKEN`.
5. Enable Developer Mode in Discord, right-click your own name, and copy your user ID into `DISCORD_USER_ID`.

## Configuration

Copy `.env.example` to `.env` and fill in the values (see comments in that file for details).

```
PORT=8080
DISCORD_BOT_TOKEN=
DISCORD_USER_ID=
DISCORD_GUILD_ID=
TRUSTED_PROXY_CIDRS=
VOICE_HUB_CHANNEL_ID=
VOICE_HUB_CATEGORY_ID=
DB_PATH=./data/juno.db
```

`TRUSTED_PROXY_CIDRS` is a comma-separated list of CIDRs for reverse proxies you trust to set `X-Forwarded-For` (e.g. `10.0.0.0/8`). Leave blank if the service is reachable directly, with no reverse proxy in front.

## Running

```
docker compose up --build -d
```

By default the service listens on host port 8080, with a healthcheck at `curl http://localhost:8080/healthz`.

## Join-to-create voice channels

Set `VOICE_HUB_CHANNEL_ID` to a voice channel's ID and members who join it get their own temp channel named `{user}'s Channel`, created under the hub's category (or `VOICE_HUB_CATEGORY_ID` if set). Once a temp channel has been empty for five minutes, it is deleted. Leave `VOICE_HUB_CHANNEL_ID` blank to disable the feature.

The channel name template and grace period are constants in `internal/tempvoice/feature.go`.

## API reference

### `GET /healthz`

Liveness check. Returns `{"status":"ok"}`.

### `GET /api/users/me`

Returns your cached presence.

```json
{
  "success": true,
  "data": {
    "discord_user": {"id": "...", "username": "...", "discriminator": "0", "global_name": "...", "avatar": "..."},
    "discord_status": "online",
    "activities": [ /* raw Discord activity objects */ ],
    "listening_to_spotify": false,
    "spotify": null,
    "active_on_discord_desktop": true,
    "active_on_discord_mobile": false,
    "active_on_discord_web": false
  }
}
```

### `GET /api/socket`

WebSocket endpoint using Lanyard's own protocol:

- Server sends `{"op":1,"d":{"heartbeat_interval":30000}}` (Hello) on connect.
- Client may send `{"op":2,"d":{"subscribe_to_id":"<your id>"}}` (Initialize) — optional, since there's only ever one tracked user.
- Server sends `{"op":0,"t":"INIT_STATE","d":<presence>}` immediately after connecting, then `{"op":0,"t":"PRESENCE_UPDATE","d":<presence>}` on every subsequent change.
- Client should send `{"op":3}` (Heartbeat) periodically to keep the connection alive.
