package moderation

import (
	"context"
	"fmt"
	"time"

	"1f320.xyz/x/parameters"
	"emperror.dev/errors"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/common"
	"github.com/starshine-sys/oodles/db"
)

type unmuteEntry struct {
	ID        int64
	EventType string
	Expires   time.Time
	Data      changeRoles
}

const unmuteEntrySql = `select * from scheduled_events
where event_type = 'moderation.changeRoles'
and (data->'user_id'->>0)::bigint = $1
and (data->'guild_id'->>0)::bigint = $2
and data->'mod_log_type'->>0 = 'unmute';`

func (bot *Bot) unmute(ctx *bcr.Context) (err error) {
	params := parameters.NewParameters(ctx.RawArgs, false)

	if !params.HasNext() {
		return ctx.SendX("You must give a user to warn.")
	}

	u, err := ctx.ParseMember(params.Pop())
	if err != nil {
		_, err = ctx.Send("User not found.")
		return
	}

	var entry unmuteEntry
	err = pgxscan.Get(context.Background(), bot.DB, &entry, unmuteEntrySql, u.User.ID, ctx.Message.GuildID)
	if err != nil {
		if errors.Cause(err) != pgx.ErrNoRows {
			bot.SendError("error getting pending unmute entries for %v: %v", u.User.ID, err)
		}

		common.Log.Debugf("did not find scheduled unmute for %v", u.User.ID)

		muteRole := bot.DB.Config.Get("mute_role").ToRoleID()
		if !muteRole.IsValid() {
			return ctx.SendX("There's no pending unmute for that user, and there's no mute role set, so we can't unmute them.")
		}

		reason := params.Remainder(false)
		if reason == "" {
			reason = "No reason given"
		}

		err = ctx.State.RemoveRole(ctx.Message.GuildID, u.User.ID, muteRole, api.AuditLogReason(reason))
		if err != nil {
			bot.SendError("error unmuting user: %v", err)
			return ctx.SendX("We couldn't unmute that member!")
		}

		entry, err := bot.DB.InsertModLog(context.Background(), db.ModLogEntry{
			GuildID:    ctx.Message.GuildID,
			UserID:     u.User.ID,
			ModID:      ctx.Author.ID,
			ActionType: "unmute",
			Reason:     reason,
		})
		if err != nil {
			bot.SendError("error inserting mod log entry: %v", err)
		}

		logCh := bot.DB.Config.Get("mod_log").ToChannelID()
		if logCh.IsValid() {
			log, err := ctx.State.SendMessage(logCh, "", entry.Embed(ctx.State))
			if err != nil {
				common.Log.Errorf("error sending mod log message: %v", err)
			} else {
				_, err = bot.DB.UpdateModLogMessage(entry.ID, logCh, log.ID)
				if err != nil {
					common.Log.Errorf("error updating mod log message in db: %v", err)
				}
			}
		}

		return ctx.SendfX("Unmuted %v.", u.User.Tag())
	}

	// basically do the same thing that changeRoles does normally
	rls, err := ctx.State.Roles(ctx.Message.GuildID)
	if err != nil {
		return bot.Report(ctx, err)
	}

	var toSetTo []discord.RoleID
	for _, id := range entry.Data.AddRoles {
		if roleIn(rls, id) {
			toSetTo = append(toSetTo, id)
		}
	}

	for _, id := range u.RoleIDs {
		if !roleIDIn(entry.Data.RemoveRoles, id) && !roleIDIn(toSetTo, id) {
			toSetTo = append(toSetTo, id)
		}
	}

	err = ctx.State.ModifyMember(ctx.Message.GuildID, u.User.ID, api.ModifyMemberData{
		Roles: &toSetTo,
		AuditLogReason: api.AuditLogReason(
			fmt.Sprintf("Unmute command issued by %v (%v)", ctx.Author.Tag(), ctx.Author.ID),
		),
	})
	if err != nil {
		bot.SendError("error unmuting user: %v", err)
		return ctx.SendX("We couldn't unmute that member!")
	}

	common.Log.Debugf("removing scheduled unmute #%d", entry.ID)
	err = bot.Scheduler.Remove(entry.ID)
	if err != nil {
		bot.SendError("error removing scheduled unmute: %v", err)
	}

	reason := params.Remainder(false)
	if reason == "" {
		reason = "No reason given"
	}

	modLog, err := bot.DB.InsertModLog(context.Background(), db.ModLogEntry{
		GuildID:    ctx.Message.GuildID,
		UserID:     u.User.ID,
		ModID:      ctx.Author.ID,
		ActionType: "unmute",
		Reason:     reason,
	})
	if err != nil {
		bot.SendError("error inserting mod log entry: %v", err)
	}

	logCh := bot.DB.Config.Get("mod_log").ToChannelID()
	if logCh.IsValid() {
		log, err := ctx.State.SendMessage(logCh, "", modLog.Embed(ctx.State))
		if err != nil {
			common.Log.Errorf("error sending mod log message: %v", err)
		} else {
			_, err = bot.DB.UpdateModLogMessage(modLog.ID, logCh, log.ID)
			if err != nil {
				common.Log.Errorf("error updating mod log message in db: %v", err)
			}
		}
	}

	return ctx.SendfX("Unmuted %v.", u.User.Tag())
}
