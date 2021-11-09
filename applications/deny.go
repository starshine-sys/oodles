package applications

import (
	"fmt"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mozillazg/go-unidecode"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/common"
)

func (bot *Bot) deny(ctx *bcr.Context) (err error) {
	app, err := bot.DB.ChannelApplication(ctx.Message.ChannelID)
	if err != nil {
		return ctx.SendX("This isn't an application channel!")
	}

	if app.UserID == ctx.Author.ID {
		return ctx.SendX("You can't deny yourself!")
	}

	if app.Verified != nil {
		return ctx.SendX("This application is already wrapped up.")
	}

	m, err := ctx.State.Member(ctx.Message.GuildID, app.UserID)
	if err != nil {
		return ctx.SendX("Couldn't find the member associated with this application--did they leave the server?")
	}

	// collect config
	var (
		kick    = bot.DB.Config.Get("kick_on_deny").ToBool()
		confirm = bot.DB.Config.Get("confirm_deny").ToBool()

		ableToDM = false
		dm       = bot.DB.Config.Get("dm_on_deny").ToBool()

		welcCh = bot.DB.Config.Get("welcome_channel").ToChannelID()
		tmpl   = bot.DB.Config.Get("deny_message").ToString()
	)

	if confirm {
		yes, _ := ctx.ConfirmButton(ctx.Author.ID, bcr.ConfirmData{
			Message:   "Are you sure you want to deny?",
			YesPrompt: "Deny",
			YesStyle:  discord.DangerButton,
			NoPrompt:  "Cancel",
			NoStyle:   discord.SecondaryButton,
			Timeout:   2 * time.Minute,
		})
		if !yes {
			return ctx.SendX("Cancelled.")
		}
	}

	reason := ctx.RawArgs
	if reason == "" {
		reason = "No reason specified"
	}

	if tmpl != "" && welcCh.IsValid() {
		s, err := common.ExecTemplate(tmpl, struct {
			Guild  *discord.Guild
			User   discord.User
			Denier *discord.Member
			Reason string
		}{Guild: ctx.Guild, User: m.User, Denier: ctx.Member, Reason: reason})
		if err != nil {
			common.Log.Errorf("Error executing deny message template: %v", err)
		} else {
			ctx.State.SendMessage(welcCh, s)
		}
	}

	if dm {
		ch, err := ctx.State.CreatePrivateChannel(m.User.ID)
		if err != nil {
			ctx.SendX("Note: I wasn't able to DM the user about their denial.")
		} else {
			_, err = ctx.State.SendEmbeds(ch.ID, discord.Embed{
				Title:       "You were denied",
				Description: "Your application in " + ctx.Guild.Name + " was denied.",
				Fields: []discord.EmbedField{{
					Name:  "Reason",
					Value: reason,
				}},
				Color:     bot.Colour,
				Timestamp: discord.NowTimestamp(),
			})
			if err != nil {
				ctx.SendX("Note: I wasn't able to DM the user about their denial.")
			} else {
				ableToDM = true
			}
		}
	}

	bot.deniedMu.Lock()
	bot.denied[m.User.ID] = struct{}{}
	bot.deniedMu.Unlock()

	if kick && (!dm || ableToDM) {
		kickReason := reason
		if len(reason) > 400 {
			kickReason = reason[:397] + "..."
		}

		err = ctx.State.Kick(ctx.Message.GuildID, app.UserID, api.AuditLogReason(fmt.Sprintf("%v (%v): %v", ctx.Author.Tag(), ctx.Author.ID, kickReason)))
		if err != nil {
			ctx.SendfX("I wasn't able to kick the user! Please kick them manually with Carl:\n``!kick %v %v``", m.User.ID, bcr.EscapeBackticks(reason))
		}
	}

	app.Verified = &denied
	app.Moderator = &ctx.Author.ID
	app.DenyReason = &reason
	err = bot.DB.SetVerified(app.ID, ctx.Author.ID, false, &reason)
	if err != nil {
		bot.SendError("Error setting application %v to denied: %v\nMod: %v/verified: false/reason: %v", app.ID, err, ctx.Author.ID, reason)
	}

	_, err = bot.createTranscript(ctx.State, app)
	if err != nil {
		return ctx.SendfX("There was an error saving a transcript:\n> %v", err)
	}

	// edit channel
	cat, err := ctx.State.Channel(ctx.Channel.ParentID)
	if err != nil {
		return ctx.SendfX("Couldn't get this channel's category.")
	}

	return ctx.State.ModifyChannel(app.ChannelID, api.ModifyChannelData{
		Name:           "🔒-app-" + unidecode.Unidecode(m.User.Username),
		Overwrites:     &cat.Overwrites,
		AuditLogReason: "Application completed, user denied",
	})
}
