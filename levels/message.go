package levels

import (
	"fmt"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/common"
)

func (bot *Bot) messageCreate(m *gateway.MessageCreateEvent) {
	if !m.GuildID.IsValid() || m.Author.Bot || m.Author.DiscordSystem || m.Member == nil {
		return
	}

	s, _ := bot.Router.StateFromGuildID(m.GuildID)

	sc, err := bot.getGuildConfig(m.GuildID)
	if err != nil {
		common.Log.Errorf("Error getting guild config: %v", err)
		return
	}

	if !sc.LevelsEnabled {
		return
	}

	if bot.isBlacklisted(m.GuildID, m.Author.ID) {
		return
	}

	uc, err := bot.getUser(m.GuildID, m.Author.ID)
	if err != nil {
		common.Log.Errorf("Error getting user: %v", err)
		return
	}

	if uc.LastXP.Add(sc.BetweenXP).After(time.Now().UTC()) {
		return
	}

	// check blocked channels/categories
	for _, blocked := range sc.BlockedChannels {
		if m.ChannelID == discord.ChannelID(blocked) {
			return
		}
	}
	if ch, err := s.Channel(m.ChannelID); err == nil {
		for _, blocked := range sc.BlockedCategories {
			if ch.ParentID == discord.ChannelID(blocked) {
				return
			}
		}
	}

	// check blocked roles
	for _, blocked := range sc.BlockedRoles {
		for _, r := range m.Member.RoleIDs {
			if discord.RoleID(blocked) == r {
				return
			}
		}
	}

	// increment the user's xp!
	newXP, err := bot.incrementXP(m.GuildID, m.Author.ID)
	if err != nil {
		common.Log.Errorf("Error updating XP for user: %v", err)
		return
	}

	// only check for rewards on level up
	oldLvl := LevelFromXP(uc.XP)
	newLvl := LevelFromXP(newXP)

	if oldLvl >= newLvl {
		return
	}

	reward := bot.getReward(m.GuildID, newLvl)
	if reward == nil {
		return
	}

	if !reward.RoleReward.IsValid() {
		return
	}

	// don't announce/log roles the user already has
	for _, r := range m.Member.RoleIDs {
		if r == reward.RoleReward {
			return
		}
	}

	err = s.AddRole(m.GuildID, m.Author.ID, reward.RoleReward, api.AddRoleData{
		AuditLogReason: api.AuditLogReason(fmt.Sprintf("Level reward for reaching level %v", newLvl)),
	})
	if err != nil {
		common.Log.Errorf("Error adding role to user: %v", err)
		return
	}

	if sc.RewardLog.IsValid() {
		e := discord.Embed{
			Title:       "Level reward given",
			Description: fmt.Sprintf("%v reached level `%v`.", m.Author.Mention(), newLvl),
			Fields: []discord.EmbedField{
				{
					Name:  "Reward given",
					Value: reward.RoleReward.Mention(),
				},
				{
					Name:  "Message",
					Value: fmt.Sprintf("https://discord.com/channels/%v/%v/%v", m.GuildID, m.ChannelID, m.ID),
				},
			},
			Color: bcr.ColourBlurple,
		}

		_, err = s.SendEmbeds(sc.RewardLog, e)
		if err != nil {
			bot.SendError("Error sending reward log: %v", err)
		}
	}

	if sc.DMOnReward && sc.RewardText != "" {
		txt := strings.NewReplacer("{lvl}", fmt.Sprint(newLvl)).Replace(sc.RewardText)

		ch, err := s.CreatePrivateChannel(m.Author.ID)
		if err == nil {
			_, err = s.SendMessage(ch.ID, txt)
			if err != nil {
				common.Log.Errorf("Error sending reward message to %v: %v", m.Author.Tag(), err)
			}
		}
	}
}
