package moderation

import (
	"context"
	"fmt"
	"math"
	"time"

	"1f320.xyz/x/parameters"
	"codeberg.org/eviedelta/detctime/durationparser"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/common"
	"github.com/starshine-sys/oodles/db"
)

func (bot *Bot) mute(ctx *bcr.Context) (err error) {
	muteRole := bot.DB.Config.Get("mute_role").ToRoleID()
	if !muteRole.IsValid() {
		return ctx.SendX("There's no mute role set, so we can't mute members.")
	}

	params := parameters.NewParameters(ctx.RawArgs, false)

	if !params.HasNext() {
		return ctx.SendX("You must give a user to mute.")
	}

	u, err := ctx.ParseMember(params.Pop())
	if err != nil {
		_, err = ctx.Send("User not found.")
		return
	}

	var (
		dur    = time.Duration(math.MaxInt64)
		durStr = "indefinitely"
		reason = "No reason given"
	)

	if params.HasNext() {
		d, err := durationparser.Parse(params.Peek())
		if err == nil {
			dur = d
			durStr = bcr.HumanizeDuration(bcr.DurationPrecisionSeconds, dur)
			params.Pop() // pop this argument so the reason doesn't include duration
		}

		if params.HasNext() {
			reason = params.Remainder(false)
		}
	}

	if u.User.ID == ctx.Bot.ID {
		return ctx.SendX("No.")
	}

	if !bot.aboveUser(ctx, ctx.Member, u) {
		_, err = ctx.Send("You're not high enough in the role hierarchy to do that.")
		return
	}

	auditLogReason := reason
	if len(reason) > 400 {
		auditLogReason = reason[:400] + "..."
	}
	err = ctx.State.AddRole(ctx.Message.GuildID, u.User.ID, muteRole, api.AddRoleData{
		AuditLogReason: api.AuditLogReason(
			fmt.Sprintf("%v (%v): %v", ctx.Author.Tag(), ctx.Author.ID, auditLogReason),
		),
	})
	if err != nil {
		bot.SendError("couldn't mute user %v: %v", u.User.ID, err)
		return ctx.SendX("We couldn't mute that user, please check permissions!")
	}

	entry, err := bot.DB.InsertModLog(context.Background(), db.ModLogEntry{
		GuildID:    ctx.Message.GuildID,
		UserID:     u.User.ID,
		ModID:      ctx.Author.ID,
		ActionType: "mute",
		Reason:     reason,
	})
	if err != nil {
		return bot.Report(ctx, err)
	}

	unmuteReason := fmt.Sprintf("Automatic unmute from mute made %v ago by %v (%v)", durStr, ctx.Author.Tag(), ctx.Author.ID)
	_, err = bot.Scheduler.Add(time.Now().UTC().Add(dur), &changeRoles{
		UserID:         u.User.ID,
		GuildID:        ctx.Message.GuildID,
		RemoveRoles:    []discord.RoleID{muteRole},
		AuditLogReason: unmuteReason,
		SendModLog:     true,
		ModeratorID:    ctx.Author.ID,
		ModLogType:     "unmute",
		ModLogReason:   unmuteReason,
	})
	if err != nil {
		bot.SendError("error scheduling unmute: %v", err)
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

	dm := fmt.Sprintf("You were muted in %v for %v.\nReason: %v", ctx.Guild.Name, durStr, reason)
	if durStr == "indefinitely" {
		dm = fmt.Sprintf("You were muted in %v indefinitely.\nReason: %v", ctx.Guild.Name, reason)
	}

	_, err = ctx.NewDM(u.User.ID).Content(dm).Send()
	if err != nil {
		common.Log.Errorf("error sending mute message to user: %v", err)
	}

	return ctx.SendfX("**%v** muted **%v** %v.\nReason: %v", ctx.Author.Tag(), u.User.Tag(), durStr, reason)
}
