package logging

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/sendpart"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/dischtml"
	"github.com/starshine-sys/oodles/common"
	"github.com/starshine-sys/oodles/db"
)

func (bot *Bot) bulkMessageDelete(ev *gateway.MessageDeleteBulkEvent) {
	s, _ := bot.Router.StateFromGuildID(ev.GuildID)

	if ev.GuildID != bot.DB.BotConfig.GuildID {
		return
	}

	logCh := bot.DB.Config.Get("message_log").ToChannelID()
	if !logCh.IsValid() {
		return
	}

	var (
		msgs            []*db.Message
		found, notFound int
		files           []sendpart.File
		users           = map[discord.UserID]*discord.User{}
	)

	for _, id := range ev.IDs {
		// first, try getting the ID of a normal message
		m, err := bot.DB.GetMessage(id)
		if err == nil && m.UserID != 0 {
			if u, ok := users[m.UserID]; ok {
				m.Username = u.Username + "#" + u.Discriminator
			} else {
				u, err := s.User(m.UserID)
				if err == nil {
					m.Username = u.Username + "#" + u.Discriminator
					users[u.ID] = u
				} else {
					m.Username = "unknown#0000"
				}
			}

			msgs = append(msgs, m)
			found++
			continue
		}
		// else add a dummy message with the ID
		msgs = append(msgs, &db.Message{
			ID:        id,
			ChannelID: ev.ChannelID,
			ServerID:  ev.GuildID,
			Content:   "*[message not in database]*",
			Username:  "unknown#0000",
		})
		notFound++
	}

	// now sort the messages
	sort.Slice(msgs, func(i, j int) bool { return msgs[i].ID < msgs[j].ID })

	html, err := bot.bulkHTML(s, ev.GuildID, ev.ChannelID, msgs)
	if err != nil {
		common.Log.Errorf("Error creating HTML output: %v", err)
	} else {
		files = append(files, sendpart.File{
			Name:   fmt.Sprintf("bulk-delete-%v-%v.html", ev.ChannelID, time.Now().UTC().Format("2006-01-02T15-04-05")),
			Reader: strings.NewReader(html),
		})
	}

	var buf string
	for _, m := range msgs {
		s := fmt.Sprintf(`[%v | %v] %v (%v)
--------------------------------------------
%v
--------------------------------------------
`,
			m.ID.Time().Format(time.ANSIC), m.ID, m.Username, m.UserID, m.Content,
		)
		if m.Member != nil && m.System != nil {
			s = fmt.Sprintf(`[%v | %v] %v (%v)
PK system: %v / PK member: %v
--------------------------------------------
%v
--------------------------------------------
`,
				m.ID.Time().Format(time.ANSIC), m.ID, m.Username, m.UserID, *m.System, *m.Member, m.Content,
			)
		}

		buf += s
	}

	files = append(files, sendpart.File{
		Name:   fmt.Sprintf("bulk-delete-%v-%v.txt", ev.ChannelID, time.Now().UTC().Format("2006-01-02T15-04-05")),
		Reader: strings.NewReader(buf),
	})

	_, err = s.SendMessageComplex(logCh, api.SendMessageData{
		Embeds: []discord.Embed{{
			Title:       "Bulk message deletion",
			Description: fmt.Sprintf("%v messages were deleted in %v.\n%v messages archived, %v messages not found.", len(ev.IDs), ev.ChannelID.Mention(), found, notFound),
			Color:       bcr.ColourRed,
			Timestamp:   discord.NowTimestamp(),
		}},
		Files: files,
	})
	if err != nil {
		common.Log.Errorf("Error sending bulk message delete log: %v", err)
	}
}

func (bot *Bot) bulkHTML(s *state.State, guildID discord.GuildID, channelID discord.ChannelID, msgs []*db.Message) (string, error) {
	g, err := s.Guild(guildID)
	if err != nil {
		return "", err
	}

	ch, err := s.Channel(channelID)
	if err != nil {
		return "", err
	}

	chans, err := s.Channels(g.ID)
	if err != nil {
		return "", err
	}

	rls, err := s.Roles(g.ID)
	if err != nil {
		return "", err
	}

	members, err := s.Members(g.ID)
	if err != nil {
		common.Log.Errorf("Error getting members for guild %v: %v", g.ID, err)
	}

	users := make([]discord.User, len(members))
	for i, m := range members {
		users[i] = m.User
	}

	c := dischtml.Converter{
		Guild:         *g,
		Channels:      chans,
		Roles:         rls,
		Members:       members,
		Users:         users,
		ExtraUserInfo: make(map[discord.MessageID]string),
	}

	dm := make([]discord.Message, len(msgs))
	for i, m := range msgs {
		var u discord.User
		var found bool

		for _, user := range users {
			if user.ID == m.UserID {
				u = user
				found = true
				break
			}
		}

		if !found {
			u = discord.User{
				ID:            m.UserID,
				Username:      m.Username,
				Discriminator: "0000",
				Avatar:        "",
			}
		}

		dmsg := discord.Message{
			ID:        m.ID,
			ChannelID: m.ChannelID,
			GuildID:   m.ServerID,
			Content:   m.Content,
			Author:    u,
		}

		dm[i] = dmsg
	}

	str, err := c.ConvertHTML(dm)
	if err != nil {
		return "", err
	}

	return dischtml.Wrap(*g, *ch, str, len(dm))
}
