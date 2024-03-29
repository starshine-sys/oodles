package logging

import (
	"fmt"

	"emperror.dev/errors"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/jackc/pgx/v4"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/common"
	"github.com/starshine-sys/oodles/db"
)

func (bot *Bot) messageUpdate(m *gateway.MessageUpdateEvent) {
	if !m.GuildID.IsValid() || !m.Author.ID.IsValid() {
		return
	}

	if m.GuildID != bot.DB.BotConfig.GuildID {
		return
	}

	// sometimes we get message update events without any content
	// so just ignore those
	if m.Content == "" {
		return
	}

	s, _ := bot.Router.StateFromGuildID(m.GuildID)

	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		return
	}

	// try getting the message
	msg, err := bot.DB.GetMessage(m.ID)
	if err != nil {
		if errors.Cause(err) != pgx.ErrNoRows {
			return
		}

		// insert message and return
		err = bot.DB.InsertMessage(db.Message{
			ID:        m.ID,
			UserID:    m.Author.ID,
			ChannelID: m.ChannelID,
			ServerID:  m.GuildID,
			Username:  m.Author.Username + "#" + m.Author.Discriminator,

			Content: m.Content,
		})
		if err != nil {
			common.Log.Errorf("Error inserting message: %v", err)
			return
		}
	}

	mention := fmt.Sprintf("%v\n%v#%v\nID: %v", m.Author.Mention(), m.Author.Username, m.Author.Discriminator, m.Author.ID)
	author := &discord.EmbedAuthor{
		Icon: m.Author.AvatarURL(),
		Name: m.Author.Username + "#" + m.Author.Discriminator,
	}

	e := discord.Embed{
		Author: author,
		Title:  fmt.Sprintf("Message by %v#%v updated", m.Author.Username, m.Author.Discriminator),
		Color:  bcr.ColourPurple,
		Footer: &discord.EmbedFooter{
			Text: fmt.Sprintf("ID: %v", msg.ID),
		},
		Timestamp: discord.NewTimestamp(msg.ID.Time()),
	}

	updated := m.Content
	if updated == "" {
		updated = "None"
	}

	// sometimes we get update events that don't actually change the content
	// including stuff like the message getting pinned
	// so we just ignore those updates
	if updated == msg.Content {
		return
	}

	if len(msg.Content) > 1000 {
		e.Fields = append(e.Fields, discord.EmbedField{
			Name:  "Old content",
			Value: msg.Content[:1000] + "...",
		})
		if len(msg.Content) > 2000 {
			if len(msg.Content) > 3000 {
				val := msg.Content[3000:]
				if len(val) > 500 {
					val = val[:500] + "..."
				}

				e.Fields = append(e.Fields, []discord.EmbedField{
					{
						Name:  "Old content (cont.)",
						Value: "..." + msg.Content[1000:2000] + "...",
					},
					{
						Name:  "Old content (cont.)",
						Value: "..." + msg.Content[2000:3000],
					},
					{
						Name:  "Old content (cont.)",
						Value: "..." + val,
					},
				}...)
			} else {
				e.Fields = append(e.Fields, []discord.EmbedField{
					{
						Name:  "Old content (cont.)",
						Value: "..." + msg.Content[1000:2000] + "...",
					},
					{
						Name:  "Old content (cont.)",
						Value: "..." + msg.Content[2000:],
					},
				}...)
			}
		} else {
			e.Fields = append(e.Fields, discord.EmbedField{
				Name:  "Old content (cont.)",
				Value: "..." + msg.Content[1000:],
			})
		}
	} else {
		e.Fields = append(e.Fields, discord.EmbedField{
			Name:  "Old content",
			Value: msg.Content,
		})
	}

	e.Fields = append(e.Fields, discord.EmbedField{Name: "​", Value: "​"})

	if len(updated) > 1000 {
		e.Fields = append(e.Fields, discord.EmbedField{
			Name:  "New content",
			Value: updated[:1000] + "...",
		})
		if len(updated) > 2000 {
			val := updated[1000:]
			if len(val) >= 1024 {
				val = val[:1015] + "..."
			}

			e.Fields = append(e.Fields, discord.EmbedField{
				Name:  "New content (cont.)",
				Value: "..." + val,
			})
		} else {
			e.Fields = append(e.Fields, discord.EmbedField{
				Name:  "New content (cont.)",
				Value: "..." + updated[1000:],
			})
		}
	} else {
		e.Fields = append(e.Fields, discord.EmbedField{
			Name:  "New content",
			Value: updated,
		})
	}

	value := fmt.Sprintf("%v\nID: %v", msg.ChannelID.Mention(), msg.ChannelID)
	if channel.Type == discord.GuildNewsThread || channel.Type == discord.GuildPrivateThread || channel.Type == discord.GuildPublicThread {
		value = fmt.Sprintf("%v\nID: %v\n\nThread: %v (%v)", channel.ParentID.Mention(), channel.ParentID, channel.Name, channel.Mention())
	}

	e.Fields = append(e.Fields, []discord.EmbedField{
		{
			Name:   "Channel",
			Value:  value,
			Inline: true,
		},
		{
			Name:   "Sender",
			Value:  mention,
			Inline: true,
		},
	}...)

	if msg.System != nil && msg.Member != nil {
		e.Title = fmt.Sprintf("Message by \"%v\" updated", m.Author.Username)

		u, err := s.User(msg.UserID)
		if err == nil {
			e.Fields[len(e.Fields)-1] = discord.EmbedField{
				Name:   "Linked Discord account",
				Value:  fmt.Sprintf("%v\n%v#%v\nID: %v", u.Mention(), u.Username, u.Discriminator, u.ID),
				Inline: true,
			}
		} else {
			e.Fields[len(e.Fields)-1] = discord.EmbedField{
				Name:   "Linked Discord account",
				Value:  fmt.Sprintf("%v\nID: %v", msg.UserID.Mention(), msg.UserID),
				Inline: true,
			}
		}

		e.Fields = append(e.Fields, []discord.EmbedField{
			{
				Name:  "​",
				Value: "**PluralKit information**",
			},
			{
				Name:   "System ID",
				Value:  *msg.System,
				Inline: true,
			},
			{
				Name:   "Member ID",
				Value:  *msg.Member,
				Inline: true,
			},
		}...)
	}

	e.Fields = append(e.Fields, discord.EmbedField{
		Name:  "Link",
		Value: fmt.Sprintf("https://discord.com/channels/%v/%v/%v", m.GuildID, m.ChannelID, m.ID),
	})

	// update the message
	username := m.Author.Username
	if msg.System == nil {
		username += "#" + m.Author.Discriminator
	}

	err = bot.DB.InsertMessage(db.Message{
		ID:        m.ID,
		UserID:    m.Author.ID,
		ChannelID: m.ChannelID,
		ServerID:  m.GuildID,
		Username:  username,
		Member:    msg.Member,
		System:    msg.System,

		Content: m.Content,
	})
	if err != nil {
		common.Log.Errorf("error inserting message: %v", err)
	}

	logCh := bot.DB.Config.Get("message_log").ToChannelID()
	if !logCh.IsValid() {
		return
	}

	_, err = s.SendEmbeds(logCh, e)
	if err != nil {
		common.Log.Errorf("Error sending log message: %v", err)
	}
}
