package applications

import (
	"github.com/jackc/pgx/v4"
	"github.com/starshine-sys/bcr"
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

	err = bot.DB.CreateApplication(m.User.ID, ch.ID)
	if err != nil {
		bot.SendError("Error registering application in DB: %v", err)
		return ctx.SendX("I couldn't save the newly opened application!")
	}

	err = bot.sendInitialMessage(ch.ID, *m)
	if err != nil {
		bot.SendError("Error sending initial message: %v", err)
		return ctx.SendX("I couldn't send the initial message!")
	}

	return ctx.SendfX("Application opened in %v!", ch.Mention())
}
