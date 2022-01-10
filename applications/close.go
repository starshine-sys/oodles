package applications

import (
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/common"
)

func (bot *Bot) closeApp(ctx *bcr.Context) (err error) {
	app, err := bot.DB.ChannelApplication(ctx.Message.ChannelID)
	if err != nil {
		return ctx.SendfX("This isn't an application channel!")
	}

	if app.TranscriptChannel == nil || app.TranscriptMessage == nil {
		if force, _ := ctx.Flags.GetBool("force"); !force {
			return ctx.SendfX("This application doesn't have a transcript, not closing! Make a transcript manually (with `%vtranscript #%v` in %v), then rerun this command with the `--force` flag.", bot.Prefix(), ctx.Channel.Name, bot.DB.Config.Get("transcript_channel").ToChannelID().Mention())
		}
	}

	err = bot.DB.CloseApplication(app.ID)
	if err != nil {
		return ctx.SendfX("Error closing application:\n> %v", err)
	}

	tch := bot.DB.Config.Get("transcript_channel").ToChannelID()
	if tch.IsValid() {
		_, err = ctx.State.SendMessage(tch, "", discord.Embed{
			Author: &discord.EmbedAuthor{
				Name: ctx.Author.Tag(),
				Icon: ctx.Author.AvatarURLWithType(discord.PNGImage),
			},
			Description: "Closed application channel `#" + ctx.Channel.Name + "`",
			Color:       bot.Colour,
			Timestamp:   discord.NowTimestamp(),
			Footer: &discord.EmbedFooter{
				Text: "Channel ID: " + ctx.Message.ChannelID.String(),
			},
		})
		if err != nil {
			common.Log.Errorf("Error sending message: %v", err)
		}
	}

	err = ctx.SendX("Channel will be closed in 10 seconds!")
	if err != nil {
		return err
	}

	time.Sleep(10 * time.Second)

	return ctx.State.DeleteChannel(ctx.Channel.ID, "Close application channel")
}
