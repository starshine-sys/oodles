package logging

import (
	"fmt"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/common"
)

func (bot *Bot) guildMemberUpdate(ev *gateway.GuildMemberUpdateEvent) {
	if ev.GuildID != bot.DB.BotConfig.GuildID {
		return
	}

	s, _ := bot.Router.StateFromGuildID(ev.GuildID)

	bot.membersMu.Lock()
	m, ok := bot.members[ev.User.ID]
	if !ok {
		m, err := s.Client.Member(ev.GuildID, ev.User.ID)
		if err != nil {
			bot.membersMu.Unlock()
			common.Log.Errorf("Error fetching member: %v", err)
			return
		}

		ev.UpdateMember(m)
		bot.members[ev.User.ID] = *m
		bot.membersMu.Unlock()
		return
	}
	bot.membersMu.Unlock()

	// update cache
	// copy member struct
	up := m
	up.RoleIDs = append([]discord.RoleID(nil), m.RoleIDs...)
	ev.UpdateMember(&up)

	bot.membersMu.Lock()
	bot.members[ev.User.ID] = up
	bot.membersMu.Unlock()

	logCh := bot.DB.Config.Get("username_role_log").ToChannelID()
	if !logCh.IsValid() {
		return
	}

	if m.Nick != ev.Nick || m.User.Tag() != ev.User.Tag() {
		// username or nickname changed, so run that handler
		bot.guildMemberNickUpdate(ev, m)
		return
	}

	// check for added roles
	var addedRoles, removedRoles []discord.RoleID
	for _, oldRole := range m.RoleIDs {
		if !roleIn(ev.RoleIDs, oldRole) {
			removedRoles = append(removedRoles, oldRole)
		}
	}
	for _, newRole := range ev.RoleIDs {
		if !roleIn(m.RoleIDs, newRole) {
			addedRoles = append(addedRoles, newRole)
		}
	}

	if len(addedRoles) == 0 && len(removedRoles) == 0 {
		return
	}

	e := discord.Embed{
		Author: &discord.EmbedAuthor{
			Icon: m.User.AvatarURL(),
			Name: ev.User.Username + "#" + ev.User.Discriminator,
		},
		Color:       bcr.ColourOrange,
		Title:       "Roles updated",
		Description: ev.User.Mention(),

		Footer: &discord.EmbedFooter{
			Text: fmt.Sprintf("User ID: %v", ev.User.ID),
		},
		Timestamp: discord.NowTimestamp(),
	}

	if len(addedRoles) > 0 {
		var s []string
		for _, r := range addedRoles {
			s = append(s, r.Mention())
		}
		v := strings.Join(s, ", ")
		if len(v) > 1000 {
			v = v[:1000] + "..."
		}

		e.Fields = append(e.Fields, discord.EmbedField{
			Name:  "Added roles",
			Value: v,
		})
	}

	if len(removedRoles) > 0 {
		var s []string
		for _, r := range removedRoles {
			s = append(s, r.Mention())
		}
		v := strings.Join(s, ", ")
		if len(v) > 1000 {
			v = v[:1000] + "..."
		}

		e.Fields = append(e.Fields, discord.EmbedField{
			Name:  "Removed roles",
			Value: v,
		})
	}

	_, err := s.SendEmbeds(logCh, e)
	if err != nil {
		common.Log.Errorf("error sending log embed: %v", err)
	}
}

func (bot *Bot) guildMemberNickUpdate(ev *gateway.GuildMemberUpdateEvent, m discord.Member) {
	logCh := bot.DB.Config.Get("username_role_log").ToChannelID()
	if !logCh.IsValid() {
		return
	}

	e := discord.Embed{
		Title: "Changed nickname",
		Author: &discord.EmbedAuthor{
			Icon: ev.User.AvatarURL(),
			Name: ev.User.Username + "#" + ev.User.Discriminator,
		},
		Thumbnail: &discord.EmbedThumbnail{
			URL: ev.User.AvatarURL() + "?size=1024",
		},
		Color: bcr.ColourGreen,
		Footer: &discord.EmbedFooter{
			Text: fmt.Sprintf("User ID: %v", m.User.ID),
		},
		Timestamp: discord.NowTimestamp(),
	}

	if m.User.Username+"#"+m.User.Discriminator != ev.User.Username+"#"+ev.User.Discriminator {
		e.Title = "Changed username"
	}

	oldNick := m.Nick
	newNick := ev.Nick
	if oldNick == "" {
		oldNick = "<none>"
	}
	if newNick == "" {
		newNick = "<none>"
	}

	if oldNick == newNick {
		oldNick = m.User.Tag()
		newNick = ev.User.Tag()
	}

	e.Description = fmt.Sprintf("**Before:** %v\n**After:** %v", oldNick, newNick)

	s, _ := bot.Router.StateFromGuildID(ev.GuildID)

	_, err := s.SendEmbeds(logCh, e)
	if err != nil {
		common.Log.Errorf("error sending log embed: %v", err)
	}
}

func roleIn(s []discord.RoleID, id discord.RoleID) (exists bool) {
	for _, r := range s {
		if id == r {
			return true
		}
	}
	return false
}
