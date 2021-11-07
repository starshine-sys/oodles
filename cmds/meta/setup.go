package meta

import (
	"fmt"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/bcr"
)

func (bot *Bot) setupMessage(ctx *bcr.Context) (err error) {
	tmpl := bot.DB.Config.Get("application_channel_message").ToString()

	embeds := []discord.Embed{{
		Title:       fmt.Sprintf("Welcome to %v!", ctx.Guild.Name),
		Description: strings.ReplaceAll(tmpl, "{guild}", ctx.Guild.Name),
		Color:       bot.Colour,
		Thumbnail: &discord.EmbedThumbnail{
			URL: ctx.Guild.IconURL() + "?size=512",
		},
	}}

	components := []discord.Component{&discord.ActionRowComponent{
		Components: []discord.Component{&discord.ButtonComponent{
			Label: "Open application",
			Style: discord.SecondaryButton,
			Emoji: &discord.ButtonEmoji{
				Name: "📜",
			},
			CustomID: "open_application",
		}},
	}}

	_, err = ctx.State.SendMessageComplex(ctx.Message.ChannelID, api.SendMessageData{
		Embeds:     embeds,
		Components: components,
		AllowedMentions: &api.AllowedMentions{
			Parse: []api.AllowedMentionType{},
		},
	})
	if err != nil {
		return err
	}

	return ctx.State.React(ctx.Message.ChannelID, ctx.Message.ID, "✅")
}
