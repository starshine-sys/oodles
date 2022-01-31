package applications

import (
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/common"
)

func (bot *Bot) fallbackCreate(ctx *bcr.Context) (err error) {
	m, err := ctx.ParseMember(ctx.RawArgs)
	if err != nil {
		return ctx.SendX("User not found!")
	}

	existing, err := bot.DB.UserApplication(m.User.ID)
	if err == nil {
		ch, err := ctx.State.Channel(existing.ChannelID)
		if err == nil {
			return ctx.SendfX("%v already has an open application, in %v.", m.User.Tag(), ch.Mention())
		}

		// no channel, app should've been closed
		err = bot.DB.CloseApplication(existing.ID)
		if err != nil {
			bot.SendError("Error closing existing app: %v", err)
		}
	}

	if err != nil && err != pgx.ErrNoRows {
		bot.SendError("Unknown error fetching app: %v", err)
		return ctx.SendX("There was an unknown error fetching an existing app!")
	}

	ch, err := bot.newApplicationChannel(*m)
	if err != nil {
		bot.SendError("Error creating application channel: %v", err)
		return ctx.SendX("I couldn't create an application channel!")
	}

	app, err := bot.DB.CreateApplication(m.User.ID, ch.ID)
	if err != nil {
		bot.SendError("Error registering application in DB: %v", err)
		return ctx.SendX("I couldn't save the newly opened application!")
	}

	err = bot.sendInitialMessage(ch.ID, *m)
	if err != nil {
		bot.SendError("Error sending initial message: %v", err)
		return ctx.SendX("I couldn't send the initial message!")
	}

	eventID, err := bot.Scheduler.Add(
		time.Now().Add(24*time.Hour), &timeout{ChannelID: ch.ID, UserID: m.User.ID},
	)
	if err == nil {
		if err := bot.DB.SetEventID(app.ID, eventID); err != nil {
			common.Log.Errorf("error setting event id for app %v: %v", app.ID, err)
		}
	} else {
		common.Log.Errorf("error adding timeout event for app %v: %v", app.ID, err)
	}

	return ctx.SendfX("Application opened in %v!", ch.Mention())
}
