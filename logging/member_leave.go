package logging

import (
	"fmt"
	"sort"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/common"
)

func (bot *Bot) guildMemberRemove(ev *gateway.GuildMemberRemoveEvent) {
	if ev.GuildID != bot.DB.BotConfig.GuildID {
		return
	}

	logCh := bot.DB.Config.Get("join_leave_log").ToChannelID()
	if !logCh.IsValid() {
		return
	}

	s, _ := bot.Router.StateFromGuildID(ev.GuildID)

	e := discord.Embed{
		Title:       "A member left the server...",
		Color:       bcr.ColourGold,
		Description: fmt.Sprintf("%v (%v) has left the server", ev.User.Mention(), ev.User.Tag()),
		Thumbnail: &discord.EmbedThumbnail{
			URL: ev.User.AvatarURL(),
		},

		Footer: &discord.EmbedFooter{
			Text: fmt.Sprintf("User ID: %v", ev.User.ID),
		},
		Timestamp: discord.NowTimestamp(),
	}

	bot.membersMu.Lock()
	m, ok := bot.members[ev.User.ID]
	if ok {
		e.Fields = append(e.Fields, discord.EmbedField{
			Name:  "Joined",
			Value: fmt.Sprintf("<t:%v> (%v)", m.Joined.Time().Unix(), bcr.HumanizeTime(bcr.DurationPrecisionSeconds, m.Joined.Time())),
		})

		if len(m.RoleIDs) > 0 {
			var mentions []string
			for _, r := range sortedRoles(s, ev.GuildID, m) {
				mentions = append(mentions, r.Mention())
			}

			v := strings.Join(mentions, ", ")
			if len(v) > 1000 {
				v = v[:1000] + "..."
			}

			if strings.TrimSpace(v) != "" {
				e.Fields = append(e.Fields, discord.EmbedField{
					Name:  "Roles",
					Value: v,
				})
			}
		}
	}

	delete(bot.members, ev.User.ID)
	bot.membersMu.Unlock()

	_, err := s.SendEmbeds(logCh, e)
	if err != nil {
		common.Log.Errorf("Error sending join log: %v", err)
	}
}

const filter = "—"

// sortedRoles sorts the member's roles and filters roles to only those whose names do not contain the `filter` character.
func sortedRoles(s *state.State, gID discord.GuildID, m discord.Member) (filtered []discord.RoleID) {
	rls, err := s.Roles(gID)
	if err != nil {
		return m.RoleIDs
	}

	memberRls := make([]discord.Role, 0, len(m.RoleIDs))
	for _, r := range rls {
		for _, id := range m.RoleIDs {
			if r.ID == id {
				memberRls = append(memberRls, r)
			}
		}
	}

	sort.Sort(bcr.Roles(memberRls))

	for _, r := range memberRls {
		if !strings.Contains(r.Name, filter) {
			filtered = append(filtered, r.ID)
		}
	}

	return filtered
}
