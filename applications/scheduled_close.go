package applications

import (
	"context"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/bot"
	"github.com/starshine-sys/oodles/common"
)

const ScheduledCloseTime = 24 * time.Hour

func (bot *Bot) closeCancel(ctx *bcr.Context) (err error) {
	app, err := bot.DB.ChannelApplication(ctx.Message.ChannelID)
	if err != nil {
		return ctx.SendfX("This isn't an application channel!")
	}

	ct, err := bot.DB.Exec(context.Background(), "delete from scheduled_events where event_type = 'applications.scheduledClose' and (data->'channel_id'->>0)::bigint = $1", app.ChannelID)
	if err != nil {
		bot.SendError("cancel closing app: %v", err)
		return ctx.SendfX("Error cancelling closing application: %v", err)
	}

	if ct.RowsAffected() == 0 {
		return ctx.SendfX("This application is not scheduled to be closed.")
	}

	return ctx.SendX("Cancelled scheduled close!")
}

type scheduledClose struct {
	ChannelID discord.ChannelID `json:"channel_id"`
}

func (dat *scheduledClose) Execute(ctx context.Context, id int64, bot *bot.Bot) error {
	app, err := bot.DB.ChannelApplication(dat.ChannelID)
	if err != nil {
		common.Log.Errorf("getting application channel: %v", err)
		return err
	}

	err = bot.DB.CloseApplication(app.ID)
	if err != nil {
		common.Log.Errorf("closing application: %v", err)
		return err
	}

	ch, err := bot.State.Channel(app.ChannelID)
	if err != nil {
		common.Log.Errorf("getting channel: %v", err)
		return err
	}

	tch := bot.DB.Config.Get("transcript_channel").ToChannelID()
	if tch.IsValid() {
		_, err = bot.State.SendMessage(tch, "", discord.Embed{
			Author: &discord.EmbedAuthor{
				Name: "Scheduled close",
			},
			Description: "Closed application channel `#" + ch.Name + "`",
			Color:       bot.Colour,
			Timestamp:   discord.NowTimestamp(),
			Footer: &discord.EmbedFooter{
				Text: "Channel ID: " + ch.ID.String(),
			},
		})
		if err != nil {
			common.Log.Errorf("Error sending message: %v", err)
		}
	}

	err = bot.State.DeleteChannel(app.ChannelID, "Scheduled closing of application channel")
	if err != nil {
		common.Log.Errorf("deleting channel: %v", err)
	}
	return nil
}

func (dat *scheduledClose) Offset() time.Duration { return time.Minute }
