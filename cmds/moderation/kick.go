package moderation

import (
	"context"
	"fmt"

	"1f320.xyz/x/parameters"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/common"
	"github.com/starshine-sys/oodles/db"
)

func (bot *Bot) kick(ctx *bcr.Context) (err error) {
	muteRole := bot.DB.Config.Get("mute_role").ToRoleID()
	if !muteRole.IsValid() {
		return ctx.SendX("There's no mute role set, so we can't mute members.")
	}

	params := parameters.NewParameters(ctx.RawArgs, false)

	if !params.HasNext() {
		return ctx.SendX("You must give a user to kick.")
	}

	u, err := ctx.ParseMember(params.Pop())
	if err != nil {
		_, err = ctx.Send("User not found.")
		return
	}

	if u.User.ID == ctx.Bot.ID {
		return ctx.SendX("No.")
	}

	if !bot.aboveUser(ctx, ctx.Member, u) {
		_, err = ctx.Send("You're not high enough in the hierarchy to do that.")
		return
	}

	reason := "No reason given."
	if params.HasNext() {
		reason = params.Remainder(false)
	}

	auditLogReason := reason
	if len(reason) > 400 {
		auditLogReason = reason[:400] + "..."
	}

	entry, err := bot.DB.InsertModLog(context.Background(), db.ModLogEntry{
		GuildID:    ctx.Message.GuildID,
		UserID:     u.User.ID,
		ModID:      ctx.Author.ID,
		ActionType: "kick",
		Reason:     reason,
	})
	if err != nil {
		return bot.Report(ctx, err)
	}

	_, _ = ctx.NewDM(u.User.ID).Content(
		fmt.Sprintf("You were kicked from %v.\nReason: %v", ctx.Guild.Name, reason),
	).Send()

	err = ctx.State.Kick(ctx.Message.GuildID, u.User.ID, api.AuditLogReason(fmt.Sprintf(
		"%v (%v): %v", ctx.Author.Tag(), ctx.Author.ID, auditLogReason,
	)))
	if err != nil {
		bot.SendError("error kicking user %v: %v", u.User.Tag(), err)
		return ctx.SendfX("We were unable to kick %v!", u.User.Tag())
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

	return ctx.SendfX("Kicked **%v**", u.User.Tag())
}
