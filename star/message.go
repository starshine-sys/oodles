package star

import (
	"context"
	"fmt"
	"regexp"

	"emperror.dev/errors"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/jackc/pgx/v4"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/common"
	"github.com/starshine-sys/oodles/star/queries"
)

func (bot *Bot) sendOrUpdateMessage(m discord.Message, cfg queries.ChannelConfigRow, count int) error {
	unlock := bot.acquire(m.ID)
	defer unlock()

	sm, err := bot.queries.StarboardMessage(context.Background(), int64(m.ID))
	if err != nil {
		if errors.Cause(err) == pgx.ErrNoRows {
			if cfg.ReactionLimit > count {
				return nil
			}

			content, embed := bot.starboardMessage(m, cfg.Emoji, count)
			msg, err := bot.State.SendMessage(discord.ChannelID(cfg.Starboard), content, embed)
			if err != nil {
				return err
			}

			_, err = bot.queries.InsertStarboard(context.Background(), queries.InsertStarboardParams{
				MessageID:   int64(m.ID),
				ChannelID:   int64(m.ChannelID),
				GuildID:     int64(m.GuildID),
				StarboardID: int64(msg.ID),
			})
			return err
		}

		common.Log.Errorf("error getting starboard entry for %v: %v", m.ID, err)
		return err
	}

	if cfg.ReactionLimit > count {
		return bot.deleteMessage(cfg, sm.StarboardID)
	}

	// update existing message
	content, embed := bot.starboardMessage(m, cfg.Emoji, count)
	_, err = bot.State.EditMessage(discord.ChannelID(cfg.Starboard), discord.MessageID(sm.StarboardID), content, embed)
	if err != nil {
		_, err2 := bot.queries.RemoveStarboard(context.Background(), sm.MessageID)
		return errors.Append(err, err2)
	}

	return nil
}

func (bot *Bot) deleteMessage(cfg queries.ChannelConfigRow, starboardID int64) error {
	_, err := bot.queries.RemoveStarboard(context.Background(), starboardID)
	if err != nil {
		return err
	}

	return bot.State.DeleteMessage(discord.ChannelID(cfg.Starboard), discord.MessageID(starboardID), "Remove starboard message")
}

func (bot *Bot) starboardMessage(m discord.Message, emoji string, count int) (string, discord.Embed) {
	// message content (count, emoji, link to channel)
	content := fmt.Sprintf("**%v** %v <#%v>", count, emoji, m.ChannelID)

	// the rest of this is all for the embed
	username := m.Author.Username
	if !m.WebhookID.IsValid() {
		member, err := bot.State.Member(m.GuildID, m.Author.ID)
		if err == nil && member.Nick != "" {
			username = member.Nick
		}
	}

	var attachmentLink string
	if len(m.Attachments) > 0 {
		match, _ := regexp.MatchString("\\.(png|jpg|jpeg|gif|webp)$", m.Attachments[0].URL)
		if match {
			attachmentLink = m.Attachments[0].URL
		}
	}

	e := discord.Embed{
		Description: m.Content,
		Author: &discord.EmbedAuthor{
			Name: username,
			Icon: m.Author.AvatarURL(),
		},
		Footer: &discord.EmbedFooter{
			Text: fmt.Sprintf("ID: %v", m.ID),
		},
		Timestamp: discord.Timestamp(m.Timestamp.Time()),
		Color:     bcr.ColourGold,
		Image: &discord.EmbedImage{
			URL: attachmentLink,
		},
	}

	// do our best to "translate" the original message's embed (if any) to the starboard message
	if len(m.Embeds) > 0 {
		title := m.Embeds[0].Title
		if title == "" && m.Embeds[0].Author != nil && m.Embeds[0].Author.Name != "" {
			title = m.Embeds[0].Author.Name
		}

		value := m.Embeds[0].Description
		if len(value) > 1000 {
			value = e.Description[:999] + "..."
		}

		if title != "" && value != "" {
			e.Fields = append(e.Fields, discord.EmbedField{Name: title, Value: value})
		}

		for _, f := range m.Embeds[0].Fields {
			if e.Length() > 4000 {
				break
			}

			e.Fields = append(e.Fields, f)
		}
	}

	// add replied to message if the message is a reply
	if m.Reference != nil {
		s, _ := bot.Router.StateFromGuildID(m.GuildID)
		ref, err := s.Message(m.Reference.ChannelID, m.Reference.MessageID)
		if err == nil {
			name := "Replying to " + ref.Author.Tag()
			value := ref.Content
			if ref.Content == "" {
				value = `*\[no content\]*`
			} else if len(ref.Content) > 5600-e.Length() {
				maxLen := 5600 - e.Length()
				value = ref.Content[:maxLen] + "..."
			}

			if name != "" && value != "" {
				e.Fields = append(e.Fields, discord.EmbedField{
					Name:  name,
					Value: fmt.Sprintf("[%v](%v)", value, ref.URL()),
				})
			}
		}
	}

	// add jump link
	e.Fields = append(e.Fields, discord.EmbedField{
		Name:  "Source",
		Value: fmt.Sprintf("[Jump to message](https://discord.com/channels/%v/%v/%v)", m.GuildID, m.ChannelID, m.ID),
	})

	return content, e
}
