package moderation

import (
	"context"
	"fmt"

	"1f320.xyz/x/parameters"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/common"
	"github.com/starshine-sys/oodles/db"
)

func (bot *Bot) ban(ctx *bcr.Context) (err error) {
	var (
		target   *discord.User
		isMember bool
	)

	params := parameters.NewParameters(ctx.RawArgs, false)

	if !params.HasNext() {
		return ctx.SendX("You must give a user to ban.")
	}

	member, err := ctx.ParseMember(params.Peek())
	if err == nil {
		if !bot.aboveUser(ctx, ctx.Member, member) {
			_, err = ctx.Send("You're not high enough in the role hierarchy to do that.")
			return
		}
		isMember = true
		target = &member.User
	} else {
		target, err = ctx.ParseUser(params.Peek())
		if err != nil {
			_, err = ctx.Send("Couldn't find a user with that name.")
			return
		}
	}
	params.Pop() // pop user mention off the parameters

	if target.ID == ctx.Bot.ID {
		return ctx.SendX("No.")
	}

	reason := params.Remainder(false)
	if reason == "" {
		reason = "No reason given."
	}

	auditLogReason := reason
	if len(reason) > 400 {
		auditLogReason = reason[:397] + "..."
	}

	if isMember {
		_, err = ctx.NewDM(target.ID).Content(fmt.Sprintf("You were banned from %v.\nReason: %v", ctx.Guild.Name, reason)).Send()
		if err != nil {
			_, _ = ctx.Send("We were unable to DM the user about their ban!")
		}
	}

	err = ctx.State.Ban(ctx.Message.GuildID, target.ID, api.BanData{
		DeleteDays:     option.NewUint(0),
		AuditLogReason: api.AuditLogReason(fmt.Sprintf("%v (%v): %v", ctx.Author.Tag(), ctx.Author.ID, auditLogReason)),
	})
	if err != nil {
		_, err = ctx.Send("We could not ban that user.")
		return
	}

	entry, err := bot.DB.InsertModLog(context.Background(), db.ModLogEntry{
		GuildID:    ctx.Message.GuildID,
		UserID:     target.ID,
		ModID:      ctx.Author.ID,
		ActionType: "ban",
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

	_, err = ctx.Sendf("Banned **%v#%v**", target.Username, target.Discriminator)
	return
}

func (bot *Bot) unban(ctx *bcr.Context) (err error) {
	params := parameters.NewParameters(ctx.RawArgs, false)

	if !params.HasNext() {
		return ctx.SendX("You must give a user to warn.")
	}

	u, err := ctx.ParseUser(params.Pop())
	if err != nil {
		_, err = ctx.Send("I couldn't find that user.")
		return
	}

	reason := params.Remainder(false)
	if reason == "" {
		reason = "No reason given."
	}

	auditLogReason := reason
	if len(reason) > 400 {
		auditLogReason = reason[:397] + "..."
	}

	bans, err := ctx.State.Bans(ctx.Message.GuildID)
	if err != nil {
		return bot.Report(ctx, err)
	}

	var isBanned bool
	for _, b := range bans {
		if b.User.ID == u.ID {
			isBanned = true
			break
		}
	}

	if !isBanned {
		_, err = ctx.Send("That user is not banned.")
		return
	}

	err = ctx.State.Unban(ctx.Message.GuildID, u.ID, api.AuditLogReason(ctx.Author.Tag()+": "+auditLogReason))
	if err != nil {
		_, err = ctx.Sendf("We were unable to unban %v.", u.Tag())
		return
	}

	entry, err := bot.DB.InsertModLog(context.Background(), db.ModLogEntry{
		GuildID:    ctx.Message.GuildID,
		UserID:     u.ID,
		ModID:      ctx.Author.ID,
		ActionType: "unban",
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

	_, err = ctx.Sendf("Unbanned **%v**", u.Tag())
	return
}
