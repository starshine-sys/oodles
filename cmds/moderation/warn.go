package moderation

import (
	"context"
	"fmt"

	"1f320.xyz/x/parameters"
	"github.com/dustin/go-humanize"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/common"
	"github.com/starshine-sys/oodles/db"
)

func (bot *Bot) warn(ctx *bcr.Context) (err error) {
	params := parameters.NewParameters(ctx.RawArgs, false)

	if !params.HasNext() {
		return ctx.SendX("You must give a user to warn.")
	}

	u, err := ctx.ParseMember(params.Pop())
	if err != nil {
		_, err = ctx.Send("User not found.")
		return
	}

	if !params.HasNext() {
		return ctx.SendX("You must give a reason.")
	}

	reason := params.Remainder(false)

	if u.User.ID == ctx.Bot.ID {
		_, err = ctx.Send("Why would you do that?")
		return
	}

	if !bot.aboveUser(ctx, ctx.Member, u) {
		_, err = ctx.Send("You're not high enough in the hierarchy to do that.")
		return
	}

	entry, err := bot.DB.InsertModLog(context.Background(), db.ModLogEntry{
		GuildID:    ctx.Message.GuildID,
		UserID:     u.User.ID,
		ModID:      ctx.Author.ID,
		ActionType: "warn",
		Reason:     reason,
	})
	if err != nil {
		return bot.Report(ctx, err)
	}

	logCh := bot.DB.Config.Get("mod_log").ToChannelID()
	if !logCh.IsValid() {
		common.Log.Debug("no mod log channel set")
		return nil
	}

	log, err := ctx.State.SendMessage(logCh, "", entry.Embed(ctx.State))
	if err != nil {
		common.Log.Errorf("error sending mod log message: %v", err)
	} else {
		_, err = bot.DB.UpdateModLogMessage(entry.ID, logCh, log.ID)
		if err != nil {
			common.Log.Errorf("error updating mod log message in db: %v", err)
		}
	}

	_, err = ctx.NewDM(u.User.ID).Content(fmt.Sprintf("You were warned in %v.\nReason: %v", ctx.Guild.Name, reason)).Send()
	if err != nil {
		_, err = ctx.Sendf("The warning was logged, but we were unable to DM %v about their warning!", u.User.Tag())
		return
	}

	var count int
	err = bot.DB.Pool.QueryRow(context.Background(), "select count(*) from mod_log where user_id = $1 and guild_id = $2 and action_type = 'warn'", u.User.ID, ctx.Message.GuildID).Scan(&count)
	if err != nil {
		count = 1
	}

	_, err = ctx.NewMessage().Content(fmt.Sprintf("**%v** has been warned, this is their %v warning.", u.User.Tag(), humanize.Ordinal(count))).Send()
	return
}
