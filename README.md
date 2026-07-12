# juno

A [Lanyard](https://github.com/Phineas/lanyard)-compatible REST + WebSocket feed of your live Discord status, scoped to a single Discord user.

## How it works

A bot account joins a server you're also in and, with the Presence and Server Members privileged intents enabled, receives real-time presence updates from Discord's Gateway for your user ID. Juno caches your latest presence in memory and serves it over REST and a Lanyard-shaped WebSocket protocol.

## One-time setup

### Discord bot

1. Create an application at the [Discord Developer Portal](https://discord.com/developers/applications) and add a Bot user.
2. Under Bot settings, enable the **Presence Intent** and **Server Members Intent** privileged intents.
3. Invite the bot to any server you're also a member of (OAuth2 URL Generator, `bot` scope, no permissions needed).
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
```

`TRUSTED_PROXY_CIDRS` is a comma-separated list of CIDRs for reverse proxies you trust to set `X-Forwarded-For` (e.g. `10.0.0.0/8`). Leave blank if the service is reachable directly, with no reverse proxy in front.

## Running

```
docker compose up --build -d
```

By default the service listens on host port 8080, with a healthcheck at `curl http://localhost:8080/healthz`.

## API reference

### `GET /healthz`

Liveness check. Returns `{"status":"ok"}`.

### `GET /v1/users/{discord_user_id}`

Returns your cached presence if `{discord_user_id}` matches `DISCORD_USER_ID`, otherwise `404`.

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

### `GET /socket`

WebSocket endpoint using Lanyard's own protocol:

- Server sends `{"op":1,"d":{"heartbeat_interval":30000}}` (Hello) on connect.
- Client may send `{"op":2,"d":{"subscribe_to_id":"<your id>"}}` (Initialize) — optional, since there's only ever one tracked user.
- Server sends `{"op":0,"t":"INIT_STATE","d":<presence>}` immediately after connecting, then `{"op":0,"t":"PRESENCE_UPDATE","d":<presence>}` on every subsequent change.
- Client should send `{"op":3}` (Heartbeat) periodically to keep the connection alive.
