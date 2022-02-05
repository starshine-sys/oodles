package logging

import (
	"fmt"
	"strconv"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/common"
	"github.com/starshine-sys/pkgo/v2"
)

func (bot *Bot) guildMemberAdd(m *gateway.GuildMemberAddEvent) {
	if m.GuildID != bot.DB.BotConfig.GuildID {
		return
	}

	bot.membersMu.Lock()
	bot.members[m.User.ID] = m.Member
	bot.membersMu.Unlock()

	logCh := bot.DB.Config.Get("join_leave_log").ToChannelID()
	if !logCh.IsValid() {
		return
	}
	var embeds []discord.Embed

	s, _ := bot.Router.StateFromGuildID(m.GuildID)

	e := discord.Embed{
		Title: "New member joined!",
		Thumbnail: &discord.EmbedThumbnail{
			URL: m.User.AvatarURL(),
		},

		Color:       bcr.ColourGreen,
		Description: fmt.Sprintf("%v (%v) joined the server!", m.Mention(), m.User.Tag()),

		Fields: []discord.EmbedField{
			{
				Name:   "Account age",
				Value:  fmt.Sprintf("<t:%v> (%v)", m.User.ID.Time().Unix(), bcr.HumanizeTime(bcr.DurationPrecisionMinutes, m.User.ID.Time())),
				Inline: true,
			},
		},

		Footer: &discord.EmbedFooter{
			Text: fmt.Sprintf("ID: %v", m.User.ID),
		},
		Timestamp: discord.NowTimestamp(),
	}

	g, err := s.GuildWithCount(m.GuildID)
	if err == nil {
		e.Description += "\nWe now have **" + strconv.FormatUint(g.ApproximateMembers, 10) + "** members"
	}

	sys, err := bot.PK.Account(pkgo.Snowflake(m.User.ID))
	if err == nil {
		e.Fields = append(e.Fields, discord.EmbedField{
			Name:   "​",
			Value:  "**PluralKit system information**",
			Inline: false,
		})

		if sys.Name != "" {
			e.Fields = append(e.Fields, discord.EmbedField{
				Name:   "Name",
				Value:  sys.Name,
				Inline: true,
			})
		}

		e.Fields = append(e.Fields, discord.EmbedField{
			Name:   "ID",
			Value:  sys.ID,
			Inline: true,
		})

		tag := "(None)"
		if sys.Tag != "" {
			tag = sys.Tag
		}

		e.Fields = append(e.Fields, discord.EmbedField{
			Name:   "Tag",
			Value:  tag,
			Inline: true,
		})

		e.Fields = append(e.Fields, discord.EmbedField{
			Name:   "Created",
			Value:  fmt.Sprintf("<t:%v>\n%v", sys.Created.Unix(), bcr.HumanizeTime(bcr.DurationPrecisionMinutes, sys.Created)),
			Inline: false,
		})
	}

	if !m.User.Bot {
		is, err := s.GuildInvites(m.GuildID)
		if err == nil {
			bot.invitesMu.Lock()
			allExisting := make([]discord.Invite, len(bot.invites))
			for i := range bot.invites {
				allExisting[i] = bot.invites[i]
			}
			bot.invites = is
			bot.invitesMu.Unlock()

			inv, found := checkInvites(allExisting, is)

			if !found {

				if g.VanityURLCode != "" {
					e.Fields = append(e.Fields, discord.EmbedField{
						Name:  "Invite used",
						Value: "Vanity invite (" + bcr.AsCode(g.VanityURLCode) + ")",
					})
				} else {
					e.Fields = append(e.Fields, discord.EmbedField{
						Name:  "Invite used",
						Value: "Could not determine invite.",
					})
				}

			} else {
				name := bot.DB.InviteName(inv.Code)

				e.Fields = append(e.Fields, []discord.EmbedField{
					{
						Name:  "​",
						Value: "**Invite information**",
					},
					{
						Name:   "Code",
						Value:  inv.Code,
						Inline: true,
					},
					{
						Name:   "Uses",
						Value:  fmt.Sprint(inv.Uses),
						Inline: true,
					},
					{
						Name:   "Created at",
						Value:  fmt.Sprintf("<t:%v>", inv.CreatedAt.Time().Unix()),
						Inline: true,
					},
					{
						Name:   "Name",
						Value:  name,
						Inline: false,
					},
				}...)
			}

		} else {
			common.Log.Errorf("Error fetching previous invites for %v: %v", m.GuildID, err)

			e.Fields = append(e.Fields, discord.EmbedField{
				Name:  "Invite used",
				Value: "Could not determine invite.",
			})
		}
	}

	embeds = append(embeds, e)

	if m.User.CreatedAt().After(time.Now().UTC().Add(-168 * time.Hour)) {
		embeds = append(embeds, discord.Embed{
			Title:       "New account",
			Description: fmt.Sprintf("⚠️ This account was only created **%v** (<t:%v>)", bcr.HumanizeTime(bcr.DurationPrecisionSeconds, m.User.CreatedAt()), m.User.CreatedAt().Unix()),
			Color:       bcr.ColourOrange,
		})
	}

	_, err = s.SendEmbeds(logCh, embeds...)
	if err != nil {
		common.Log.Errorf("Error sending join log: %v", err)
	}
}

func checkInvites(old, new []discord.Invite) (inv discord.Invite, found bool) {
	// check invites in both slices
	for _, o := range old {
		for _, n := range new {
			if o.Code == n.Code && o.Uses < n.Uses {
				return n, true
			}
		}
	}

	// check only new invites with 1 use
	for _, n := range new {
		if !invExists(old, n) && n.Uses == 1 {
			return n, true
		}
	}

	// check only old invites with 1 use less than max
	for _, o := range old {
		if !invExists(new, o) && o.MaxUses != 0 && o.MaxUses == o.Uses+1 {
			// this is an *old* invite so we should update the count before returning
			o.Uses = o.Uses + 1
			return o, true
		}
	}

	return inv, false
}

func invExists(invs []discord.Invite, i discord.Invite) bool {
	for _, o := range invs {
		if i.Code == o.Code {
			return true
		}
	}

	return false
}
