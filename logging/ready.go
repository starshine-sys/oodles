package logging

import (
	"context"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/starshine-sys/oodles/common"
)

func (bot *Bot) guildCreate(ev *gateway.GuildCreateEvent) {
	if ev.ID != bot.DB.BotConfig.GuildID {
		return
	}

	s, _ := bot.Router.StateFromGuildID(ev.ID)

	invs, err := s.GuildInvites(ev.ID)
	if err != nil {
		common.Log.Errorf("Error fetching guild invites: %v", err)
	}

	bot.invitesMu.Lock()
	bot.invites = invs
	bot.invitesMu.Unlock()

	bot.membersMu.Lock()
	for _, m := range ev.Members {
		bot.members[m.User.ID] = m
	}
	bot.membersMu.Unlock()

	if int(ev.MemberCount) == len(ev.Members) {
		return
	}

	common.Log.Debugf("Didn't get all members for bot guild (%v) in guild create, requesting chunks", ev.Name)
	// otherwise, request all members
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	err = s.Gateway().Send(ctx, &gateway.RequestGuildMembersCommand{
		GuildIDs: []discord.GuildID{ev.ID},
	})
	if err != nil {
		common.Log.Errorf("Error requesting guild members: %v", err)
	}
}

func (bot *Bot) guildMembersChunk(ev *gateway.GuildMembersChunkEvent) {
	// this should never happen, but check anyway
	if ev.GuildID != bot.DB.BotConfig.GuildID {
		return
	}

	bot.membersMu.Lock()
	for _, m := range ev.Members {
		bot.members[m.User.ID] = m
	}
	bot.membersMu.Unlock()
}

func (bot *Bot) inviteCreate(ev *gateway.InviteCreateEvent) {
	// this should never happen, but check anyway
	if ev.GuildID != bot.DB.BotConfig.GuildID {
		return
	}

	s, _ := bot.Router.StateFromGuildID(ev.GuildID)

	invs, err := s.GuildInvites(ev.GuildID)
	if err != nil {
		common.Log.Errorf("Error fetching guild invites: %v", err)
	}

	bot.invitesMu.Lock()
	bot.invites = invs
	bot.invitesMu.Unlock()
}

func (bot *Bot) inviteDelete(ev *gateway.InviteDeleteEvent) {
	// this should never happen, but check anyway
	if ev.GuildID != bot.DB.BotConfig.GuildID {
		return
	}

	s, _ := bot.Router.StateFromGuildID(ev.GuildID)

	invs, err := s.GuildInvites(ev.GuildID)
	if err != nil {
		common.Log.Errorf("Error fetching guild invites: %v", err)
	}

	bot.invitesMu.Lock()
	bot.invites = invs
	bot.invitesMu.Unlock()
}
