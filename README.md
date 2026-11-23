# juno

A Discord bot that provides join-to-create temporary voice channels for a single "hub" channel.

Juno is built for exactly one guild: everything is configured through environment variables, with no in-Discord configuration commands.

## How it works

Members who join the hub voice channel (`VOICE_HUB_CHANNEL_ID`) get their own temporary voice channel (with permission to manage it), and are moved into it automatically. The channel is deleted again once it's been empty for five minutes. Open temp channels are tracked in an embedded SQLite database so cleanup survives restarts.

## One-time setup

### Discord bot

1. Create an application at the [Discord Developer Portal](https://discord.com/developers/applications) and add a Bot user.
2. Invite the bot with the `bot` OAuth2 scope, and the **Manage Channels** and **Move Members** bot permissions.
3. Copy the bot token into `DISCORD_BOT_TOKEN`.

## Configuration

Copy `.env.example` to `.env` and fill in the values (see comments in that file for details).

```
DISCORD_BOT_TOKEN=
VOICE_HUB_CHANNEL_ID=
VOICE_HUB_CATEGORY_ID=
DB_PATH=./data/juno.db
```

## Running

```
docker compose up --build -d
```

## Join-to-create voice channels

Set `VOICE_HUB_CHANNEL_ID` to a voice channel's ID and members who join it get their own temp channel named `{user}'s Channel`, created under the hub's category (or `VOICE_HUB_CATEGORY_ID` if set). Once a temp channel has been empty for five minutes, it is deleted. Leave `VOICE_HUB_CHANNEL_ID` blank to disable the feature.

The channel name template and grace period are constants in `internal/tempvoice/feature.go`.
