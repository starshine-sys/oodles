package applications

import (
	"fmt"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mozillazg/go-unidecode"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/common"
)

var verified = true
var denied = false

func (bot *Bot) verify(ctx *bcr.Context) (err error) {
	app, err := bot.DB.ChannelApplication(ctx.Message.ChannelID)
	if err != nil {
		return ctx.SendX("This isn't an application channel!")
	}

	if app.UserID == ctx.Author.ID {
		return ctx.SendX("You can't verify yourself!")
	}

	if app.Verified != nil {
		return ctx.SendX("This application is already wrapped up.")
	}

	m, err := ctx.State.Member(ctx.Message.GuildID, app.UserID)
	if err != nil {
		return ctx.SendX("Couldn't find the member associated with this application--did they leave the server?")
	}

	toAdd := []discord.RoleID{bot.DB.Config.Get("verified_role").ToRoleID()}

	minorRole := bot.DB.Config.Get("minor_role").ToRoleID()
	adultRole := bot.DB.Config.Get("adult_role").ToRoleID()

	if !minorRole.IsValid() || !adultRole.IsValid() {
		if minorRole.IsValid() {
			toAdd = append(toAdd, minorRole)
		} else if adultRole.IsValid() {
			toAdd = append(toAdd, adultRole)
		}
	} else {
		var minor bool
		if strings.EqualFold(ctx.RawArgs, "minor") {
			minor = true
		} else if strings.EqualFold(ctx.RawArgs, "adult") {
			minor = false
		} else {
			var timeout bool

			minor, timeout = ctx.ConfirmButton(ctx.Author.ID, bcr.ConfirmData{
				Message:   "Is the new member a bodily minor or an adult?",
				YesPrompt: "Minor",
				YesStyle:  discord.PrimaryButtonStyle(),
				NoPrompt:  "Adult",
				NoStyle:   discord.PrimaryButtonStyle(),
				Timeout:   2 * time.Minute,
			})
			if timeout {
				return ctx.SendX("Prompt timed out.")
			}
		}

		if minor {
			toAdd = append(toAdd, minorRole)
		} else {
			toAdd = append(toAdd, adultRole)
		}
	}

	setRoles := m.RoleIDs
	for _, add := range toAdd {
		hasRole := false
		for _, r := range setRoles {
			if r == add {
				hasRole = true
				break
			}
		}

		if !hasRole {
			setRoles = append(setRoles, add)
		}
	}

	// set user's roles
	err = ctx.State.ModifyMember(bot.DB.BotConfig.GuildID, m.User.ID, api.ModifyMemberData{
		Roles: &setRoles,
		AuditLogReason: api.AuditLogReason(
			fmt.Sprintf("User was verified by %v (%v)", ctx.Author.Tag(), ctx.Author.ID),
		),
	})
	if err != nil {
		return ctx.SendfX("Couldn't update the user's roles:\n> %v", err)
	}

	app.Verified = &verified
	app.Moderator = &ctx.Author.ID
	err = bot.DB.SetVerified(app.ID, ctx.Author.ID, true, nil)
	if err != nil {
		bot.SendError("Error setting application %v to verified: %v\nMod: %v/verified: true", app.ID, err, ctx.Author.ID)
	}

	// send welcome message
	tmpl := bot.DB.Config.Get("welcome_message").ToString()
	welcCh := bot.DB.Config.Get("welcome_channel").ToChannelID()
	if tmpl != "" && welcCh.IsValid() {
		s, err := common.ExecTemplate(tmpl, struct {
			Guild            *discord.Guild
			Member, Approver *discord.Member
		}{Guild: ctx.Guild, Member: m, Approver: ctx.Member})
		if err == nil {
			_, err = ctx.State.SendMessage(welcCh, s)
			if err != nil {
				common.Log.Errorf("Error sending message: %v", err)
			}
		} else {
			bot.SendError("Error executing welcome message template: %v", err)
		}
	}

	// save transcript
	_, err = bot.createTranscript(ctx.State, app)
	if err != nil {
		return ctx.SendfX("There was an error saving a transcript:\n> %v", err)
	}

	// schedule closing
	eventID, err := bot.Scheduler.Add(
		time.Now().Add(ScheduledCloseTime), &scheduledClose{ChannelID: app.ChannelID},
	)
	if err == nil {
		if err := bot.DB.SetCloseID(app.ID, eventID); err != nil {
			common.Log.Errorf("error setting scheduled close id for app %v: %v", app.ID, err)
		}
	} else {
		common.Log.Errorf("error adding scheduled close for app %v: %v", app.ID, err)
	}

	// edit channel
	newCat := discord.ChannelID(bot.DB.Config.Get("finished_application_category").ToSnowflake())
	if !newCat.IsValid() {
		newCat = ctx.Channel.ParentID
	}

	cat, err := ctx.State.Channel(newCat)
	if err != nil {
		return ctx.SendfX("Couldn't get this channel's category.")
	}

	if app.ScheduledEventID != nil {
		err = bot.Scheduler.Remove(*app.ScheduledEventID)
		if err != nil {
			bot.SendError("Error removing schedled timeout message for app %v: %v", app.ID, err)
		}
	}

	return ctx.State.ModifyChannel(app.ChannelID, api.ModifyChannelData{
		Name:           "🔒-app-" + unidecode.Unidecode(m.User.Username),
		CategoryID:     newCat,
		Overwrites:     &cat.Overwrites,
		AuditLogReason: "Application completed, user verified",
	})
}
