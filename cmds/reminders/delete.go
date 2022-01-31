package reminders

import (
	"context"
	"fmt"
	"strconv"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/bcr"
)

func (bot *Bot) delreminder(ctx *bcr.Context) (err error) {
	id, err := strconv.ParseUint(ctx.Args[0], 10, 0)
	if err != nil {
		_, err = ctx.Send("Couldn't parse your input as a number.")
		return
	}

	var exists bool
	err = bot.DB.Pool.QueryRow(context.Background(), "select exists (select * from reminders where user_id = $1 and id = $2)", ctx.Author.ID, id).Scan(&exists)
	if err != nil {
		return bot.Report(ctx, err)
	}
	if !exists {
		_, err = ctx.Send("", discord.Embed{
			Color:       bcr.ColourRed,
			Description: fmt.Sprintf("No reminder with ID #%v found, or it's not your reminder.", id),
		})
		return
	}

	_, err = bot.DB.Pool.Exec(context.Background(), "delete from reminders where id = $1 and user_id = $2", id, ctx.Author.ID)
	if err != nil {
		return bot.Report(ctx, err)
	}

	_, err = ctx.Send("", discord.Embed{
		Color:       bot.Colour,
		Description: fmt.Sprintf("Successfully deleted reminder #%v.", id),
	})
	return
}
