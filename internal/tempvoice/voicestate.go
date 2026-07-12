package tempvoice

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/kashalls/juno/internal/db"
)

const maxChannelNameLength = 100

// onVoiceStateUpdate handles both halves of the join-to-create lifecycle:
// joining a hub spawns (or reuses) a temp channel, and leaving a tracked
// temp channel updates its empty/occupied bookkeeping for the grace-period
// cleanup worker.
func (f *Feature) onVoiceStateUpdate(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	after := v.VoiceState
	before := v.BeforeUpdate

	if after.ChannelID != "" && (before == nil || before.ChannelID != after.ChannelID) {
		f.handleJoin(ctx, s, after)
	}

	if before != nil && before.ChannelID != "" && before.ChannelID != after.ChannelID {
		f.handleLeave(ctx, s, before)
	}
}

func (f *Feature) handleJoin(ctx context.Context, s *discordgo.Session, vs *discordgo.VoiceState) {
	hub, err := f.db.GetHubByChannelID(ctx, vs.GuildID, vs.ChannelID)
	if err != nil {
		slog.Error("failed to look up voice hub", "channel_id", vs.ChannelID, "err", err)
		return
	}
	if hub == nil {
		return
	}

	existing, err := f.db.GetTempChannelByOwnerAndHub(ctx, vs.UserID, hub.HubChannelID)
	if err != nil {
		slog.Error("failed to check existing temp channel", "user_id", vs.UserID, "err", err)
		return
	}
	if existing != nil {
		// Duplicate/reconnect join event for a channel the user already
		// owns from this hub - move them back in rather than spawning a
		// second channel.
		channelID := existing.ChannelID
		if err := s.GuildMemberMove(vs.GuildID, vs.UserID, &channelID); err != nil {
			slog.Error("failed to move member into existing temp channel", "channel_id", channelID, "err", err)
		}
		return
	}

	overwrites, err := f.ownerOverwrites(s, hub.CategoryID, vs.UserID)
	if err != nil {
		slog.Error("failed to build permission overwrites", "category_id", hub.CategoryID, "err", err)
		return
	}

	channel, err := s.GuildChannelCreateComplex(vs.GuildID, discordgo.GuildChannelCreateData{
		Name:                 buildChannelName(hub.NameTemplate, vs.Member),
		Type:                 discordgo.ChannelTypeGuildVoice,
		ParentID:             hub.CategoryID,
		PermissionOverwrites: overwrites,
	})
	if err != nil {
		slog.Error("failed to create temp voice channel", "guild_id", vs.GuildID, "err", err)
		return
	}

	if err := f.db.InsertTempChannel(ctx, db.TempChannel{
		ChannelID:    channel.ID,
		GuildID:      vs.GuildID,
		HubChannelID: hub.HubChannelID,
		OwnerID:      vs.UserID,
		GracePeriod:  hub.GracePeriod,
	}); err != nil {
		slog.Error("failed to persist temp channel", "channel_id", channel.ID, "err", err)
	}

	channelID := channel.ID
	if err := s.GuildMemberMove(vs.GuildID, vs.UserID, &channelID); err != nil {
		slog.Error("failed to move member into new temp channel", "channel_id", channelID, "err", err)
	}
}

func (f *Feature) handleLeave(ctx context.Context, s *discordgo.Session, before *discordgo.VoiceState) {
	tc, err := f.db.GetTempChannel(ctx, before.ChannelID)
	if err != nil {
		slog.Error("failed to look up temp channel", "channel_id", before.ChannelID, "err", err)
		return
	}
	if tc == nil {
		return
	}

	if f.channelOccupied(s, before.GuildID, before.ChannelID) {
		if err := f.db.ClearEmpty(ctx, before.ChannelID); err != nil {
			slog.Error("failed to clear empty temp channel", "channel_id", before.ChannelID, "err", err)
		}
		return
	}

	if err := f.db.MarkEmpty(ctx, before.ChannelID, time.Now()); err != nil {
		slog.Error("failed to mark temp channel empty", "channel_id", before.ChannelID, "err", err)
	}
}

// channelOccupied reports whether any member remains in channelID. On a
// state read failure it fails safe (reports occupied) so a transient error
// can't cause a channel to be deleted out from under its members.
func (f *Feature) channelOccupied(s *discordgo.Session, guildID, channelID string) bool {
	guild, err := s.State.Guild(guildID)
	if err != nil {
		slog.Warn("failed to read guild state for occupancy check", "guild_id", guildID, "err", err)
		return true
	}
	for _, vs := range guild.VoiceStates {
		if vs.ChannelID == channelID {
			return true
		}
	}
	return false
}

// ownerOverwrites copies the target category's permission overwrites (so the
// new channel doesn't unexpectedly become visible outside a restricted
// category) and appends an overwrite granting the owner control of their
// own channel.
func (f *Feature) ownerOverwrites(s *discordgo.Session, categoryID, ownerID string) ([]*discordgo.PermissionOverwrite, error) {
	category, err := s.Channel(categoryID)
	if err != nil {
		return nil, fmt.Errorf("fetch category %s: %w", categoryID, err)
	}

	overwrites := make([]*discordgo.PermissionOverwrite, 0, len(category.PermissionOverwrites)+1)
	overwrites = append(overwrites, category.PermissionOverwrites...)
	overwrites = append(overwrites, &discordgo.PermissionOverwrite{
		ID:   ownerID,
		Type: discordgo.PermissionOverwriteTypeMember,
		Allow: discordgo.PermissionManageChannels |
			discordgo.PermissionVoiceMoveMembers |
			discordgo.PermissionVoiceMuteMembers |
			discordgo.PermissionVoiceDeafenMembers,
	})
	return overwrites, nil
}

func buildChannelName(template string, member *discordgo.Member) string {
	name := template
	if name == "" {
		name = defaultNameTemplate
	}
	if member != nil {
		name = strings.ReplaceAll(name, "{user}", member.DisplayName())
	}
	if runes := []rune(name); len(runes) > maxChannelNameLength {
		name = string(runes[:maxChannelNameLength])
	}
	return name
}
