package tempvoice

import "github.com/bwmarrin/discordgo"

var manageGuildPermission int64 = discordgo.PermissionManageGuild

// voiceHubCommand is the single admin-facing entry point for configuring
// join-to-create hubs. DefaultMemberPermissions gates every subcommand on
// Manage Server - Discord enforces this itself before the interaction ever
// reaches the bot.
var voiceHubCommand = &discordgo.ApplicationCommand{
	Name:                     "voice-hub",
	Description:              "Configure join-to-create voice channel hubs",
	DefaultMemberPermissions: &manageGuildPermission,
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "add",
			Description: "Add (or update) a hub channel that spawns temp voice channels on join",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:         discordgo.ApplicationCommandOptionChannel,
					Name:         "hub-channel",
					Description:  "The voice channel members join to create their own channel",
					ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildVoice},
					Required:     true,
				},
				{
					Type:         discordgo.ApplicationCommandOptionChannel,
					Name:         "category",
					Description:  "Category new channels are created in (defaults to the hub's own category)",
					ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildCategory},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name-template",
					Description: "Channel name template; {user} is replaced with the owner's display name",
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "grace-period-seconds",
					Description: "Seconds to wait after a channel empties before deleting it (default 300)",
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "remove",
			Description: "Remove a hub channel",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:         discordgo.ApplicationCommandOptionChannel,
					Name:         "hub-channel",
					Description:  "The hub channel to remove",
					ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildVoice},
					Required:     true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "list",
			Description: "List configured hub channels",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "grace-period",
			Description: "Change how long an empty channel waits before deletion",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:         discordgo.ApplicationCommandOptionChannel,
					Name:         "hub-channel",
					Description:  "The hub channel to update",
					ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildVoice},
					Required:     true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "seconds",
					Description: "New grace period in seconds",
					Required:    true,
				},
			},
		},
	},
}
