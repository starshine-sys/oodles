package fun

import (
	"fmt"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/bot"
	"github.com/starshine-sys/oodles/common"
)

// Bot ...
type Bot struct {
	*bot.Bot
}

// Init ...
func Init(bot *bot.Bot) {
	b := &Bot{bot}

	b.Router.AddHandler(b.dmHandler)
}

func (bot *Bot) dmHandler(m *gateway.MessageCreateEvent) {
	if m.Author.Bot || (m.Content == "" && len(m.Attachments) == 0) {
		return
	}

	s, _ := bot.Router.StateFromGuildID(m.GuildID)

	name := m.Author.Username
	if m.Member != nil {
		if m.Member.Nick != "" {
			name = m.Member.Nick
		}
	}

	if strings.Contains(
		strings.ToLower(m.Content), "thank",
	) && strings.Contains(
		strings.ToLower(m.Content), "oodles",
	) {
		_, err := s.SendEmbeds(m.ChannelID, discord.Embed{
			Color:       bot.Colour,
			Description: fmt.Sprintf("You're welcome, %v! <:plural_blob:752827898309771294>", name),
		})
		if err != nil {
			common.Log.Errorf("Error sending message: %v", err)
		}
		return
	}

	// only forward DMs
	if m.GuildID.IsValid() {
		return
	}

	if !bot.DB.BotConfig.LogChannel.IsValid() {
		return
	}

	// this whole thing is hacky but it works
	e := discord.Embed{
		Title:       "Received a DM!",
		Description: m.Content,
		Color:       bot.Colour,
		Timestamp:   discord.NowTimestamp(),
	}

	extraLinks := ""
	if len(m.Attachments) > 0 {
		if bcr.HasAnySuffix(m.Attachments[0].Filename, ".jpg", ".jpeg", ".png", ".gif", ".webp") {
			e.Image = &discord.EmbedImage{
				URL: m.Attachments[0].URL,
			}
		} else {
			extraLinks += fmt.Sprintf("[Attachment #1](%v)\n", m.Attachments[0].URL)
		}
	}

	embeds := []discord.Embed{e}

	if len(m.Attachments) > 1 {
		for i, a := range m.Attachments[1:] {
			// if it's an attachment, add an embed
			if bcr.HasAnySuffix(a.Filename, ".jpg", ".jpeg", ".png", ".gif", ".webp") {
				embeds = append(embeds, discord.Embed{
					Title:     fmt.Sprintf("Attachment #%v", i+2),
					Timestamp: discord.NowTimestamp(),
					Color:     bot.Colour,
					Image: &discord.EmbedImage{
						URL: a.URL,
					},
				})
			} else {
				extraLinks += fmt.Sprintf("[Attachment #%v](%v)\n", i+2, a.URL)
			}
		}
	}

	if extraLinks != "" {
		embeds[0].Fields = append(embeds[0].Fields, discord.EmbedField{
			Name:  "Attachment(s)",
			Value: extraLinks,
		})
	}

	_, err := s.SendMessage(bot.DB.BotConfig.LogChannel,
		fmt.Sprintf("Received a DM from %v/%v!\nUser ID: %v", m.Author.Mention(), m.Author.Tag(), m.Author.ID),
		embeds...)
	if err != nil {
		common.Log.Errorf("Error forwarding DM: %v", err)
	}

	err = s.React(m.ChannelID, m.ID, "✅")
	if err != nil {
		common.Log.Errorf("Error reacting to DM: %v", err)
	}
}
