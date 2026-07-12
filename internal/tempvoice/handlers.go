package tempvoice

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/kashalls/juno/internal/db"
)

func (f *Feature) HandleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	if len(data.Options) == 0 {
		return
	}
	sub := data.Options[0]

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reply string
	switch sub.Name {
	case "add":
		reply = f.handleAdd(ctx, s, i.GuildID, sub)
	case "remove":
		reply = f.handleRemove(ctx, i.GuildID, sub)
	case "list":
		reply = f.handleList(ctx, i.GuildID)
	case "grace-period":
		reply = f.handleGracePeriod(ctx, i.GuildID, sub)
	default:
		return
	}

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: reply,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	}); err != nil {
		slog.Error("failed to respond to interaction", "command", sub.Name, "err", err)
	}
}

func (f *Feature) handleAdd(ctx context.Context, s *discordgo.Session, guildID string, sub *discordgo.ApplicationCommandInteractionDataOption) string {
	hubOpt := sub.GetOption("hub-channel")
	if hubOpt == nil {
		return "hub-channel is required."
	}
	hubChannel := hubOpt.ChannelValue(s)

	categoryID := hubChannel.ParentID
	if categoryOpt := sub.GetOption("category"); categoryOpt != nil {
		categoryID = categoryOpt.ChannelValue(s).ID
	}
	if categoryID == "" {
		return "hub channel has no parent category; specify category explicitly."
	}

	nameTemplate := defaultNameTemplate
	if tmplOpt := sub.GetOption("name-template"); tmplOpt != nil {
		nameTemplate = tmplOpt.StringValue()
	}

	gracePeriod := time.Duration(defaultGracePeriod) * time.Second
	if graceOpt := sub.GetOption("grace-period-seconds"); graceOpt != nil {
		gracePeriod = time.Duration(graceOpt.IntValue()) * time.Second
	}

	err := f.db.CreateHub(ctx, db.Hub{
		GuildID:      guildID,
		HubChannelID: hubChannel.ID,
		CategoryID:   categoryID,
		NameTemplate: nameTemplate,
		GracePeriod:  gracePeriod,
	})
	if err != nil {
		slog.Error("failed to create voice hub", "guild_id", guildID, "err", err)
		return "Failed to save hub configuration."
	}
	return fmt.Sprintf("Hub configured on <#%s>: new channels go under <#%s>, grace period %s.", hubChannel.ID, categoryID, gracePeriod)
}

func (f *Feature) handleRemove(ctx context.Context, guildID string, sub *discordgo.ApplicationCommandInteractionDataOption) string {
	hubOpt := sub.GetOption("hub-channel")
	if hubOpt == nil {
		return "hub-channel is required."
	}
	hubChannelID := hubOpt.Value.(string)

	if err := f.db.DeleteHub(ctx, guildID, hubChannelID); err != nil {
		slog.Error("failed to delete voice hub", "guild_id", guildID, "err", err)
		return "Failed to remove hub configuration."
	}
	return fmt.Sprintf("Removed hub <#%s>. Already-open temp channels from it are unaffected.", hubChannelID)
}

func (f *Feature) handleList(ctx context.Context, guildID string) string {
	hubs, err := f.db.ListHubs(ctx, guildID)
	if err != nil {
		slog.Error("failed to list voice hubs", "guild_id", guildID, "err", err)
		return "Failed to list hub configurations."
	}
	if len(hubs) == 0 {
		return "No hub channels configured."
	}

	reply := "Configured hubs:\n"
	for _, h := range hubs {
		reply += fmt.Sprintf("- <#%s> → category <#%s>, template %q, grace period %s\n", h.HubChannelID, h.CategoryID, h.NameTemplate, h.GracePeriod)
	}
	return reply
}

func (f *Feature) handleGracePeriod(ctx context.Context, guildID string, sub *discordgo.ApplicationCommandInteractionDataOption) string {
	hubOpt := sub.GetOption("hub-channel")
	secondsOpt := sub.GetOption("seconds")
	if hubOpt == nil || secondsOpt == nil {
		return "hub-channel and seconds are required."
	}
	hubChannelID := hubOpt.Value.(string)
	gracePeriod := time.Duration(secondsOpt.IntValue()) * time.Second

	if err := f.db.SetGracePeriod(ctx, guildID, hubChannelID, gracePeriod); err != nil {
		slog.Error("failed to set grace period", "guild_id", guildID, "err", err)
		return "Failed to update grace period (is that hub configured?)."
	}
	return fmt.Sprintf("Grace period for <#%s> set to %s.", hubChannelID, gracePeriod)
}
