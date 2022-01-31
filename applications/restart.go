package applications

import (
	"fmt"
	"strconv"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/dustin/go-humanize/english"
	"github.com/starshine-sys/bcr"
)

func (bot *Bot) restart(ctx *bcr.Context) (err error) {
	app, err := bot.DB.ChannelApplication(ctx.Message.ChannelID)
	if err != nil {
		return ctx.SendfX("This isn't an application channel!")
	}

	err = bot.DB.ResetApplication(app.ID)
	if err != nil {
		return bot.Report(ctx, err)
	}

	tracks, err := bot.DB.ApplicationTracks()
	if err != nil {
		return bot.Report(ctx, err)
	}

	str := "Are you "

	var descs []string
	var buttons discord.ActionRowComponent

	for _, t := range tracks {
		descs = append(descs, fmt.Sprintf("%v (%s)", t.Description, t.Emoji()))
		buttons = append(buttons, &discord.ButtonComponent{
			Label:    t.Name,
			CustomID: discord.ComponentID("app-track:" + strconv.FormatInt(t.ID, 10) + ":restart"),
			Style:    discord.SecondaryButtonStyle(),
			Emoji: &discord.ComponentEmoji{
				Name:     t.Emoji().Name,
				ID:       t.Emoji().ID,
				Animated: t.Emoji().Animated,
			},
		})
	}
	str += english.OxfordWordSeries(descs, "or") + "?"

	_, err = ctx.SendComponents(discord.ContainerComponents{&buttons}, str)
	return
}
