package applications

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/dustin/go-humanize/english"
	"github.com/jackc/pgx/v4"
	"github.com/starshine-sys/bcr/v2"
	"github.com/starshine-sys/oodles/common"
	"github.com/starshine-sys/oodles/db"
)

const waitTime = 2 * time.Second

func (bot *Bot) createInterview(ctx *bcr.ButtonContext) (err error) {
	if ctx.Member == nil {
		return ctx.ReplyEphemeral("This event didn't have a member associated with it! This is a bug, please report it to the developer (such as by DMing me!)")
	}

	// i don't trust all the API calls to be fast enough
	if err = ctx.DeferEphemeral(); err != nil {
		return err
	}

	// check application
	existing, err := bot.DB.UserApplication(ctx.User.ID)
	if err == nil {
		ch, err := ctx.State.Channel(existing.ChannelID)
		if err == nil {
			return ctx.ReplyEphemeral(fmt.Sprintf("You already have an open application, in %v!", ch.Mention()))
		}

		// no channel, app should've been closed
		err = bot.DB.CloseApplication(existing.ID)
		if err != nil {
			bot.SendError("Error closing existing app: %v", err)
		}
	}

	if err != nil && err != pgx.ErrNoRows {
		bot.SendError("Unknown error fetching app: %v", err)
		return ctx.ReplyEphemeral("There was an unknown error fetching an existing app!")
	}

	ch, err := bot.newApplicationChannel(*ctx.Member)
	if err != nil {
		bot.SendError("Error creating application channel: %v", err)
		return ctx.ReplyEphemeral("I couldn't create an application channel!")
	}

	app, err := bot.DB.CreateApplication(ctx.User.ID, ch.ID)
	if err != nil {
		bot.SendError("Error registering application in DB: %v", err)
		return ctx.ReplyEphemeral("I couldn't save the newly opened application!")
	}

	err = bot.sendInitialMessage(ch.ID, *ctx.Member)
	if err != nil {
		bot.SendError("Error sending initial message: %v", err)
		return ctx.ReplyEphemeral("I couldn't send the initial message!")
	}

	eventID, err := bot.Scheduler.Add(
		time.Now().Add(24*time.Hour), &timeout{ChannelID: ch.ID, UserID: ctx.User.ID},
	)
	if err == nil {
		if err := bot.DB.SetEventID(app.ID, eventID); err != nil {
			common.Log.Errorf("error setting event id for app %v: %v", app.ID, err)
		}
	} else {
		common.Log.Errorf("error adding timeout event for app %v: %v", app.ID, err)
	}

	return ctx.ReplyEphemeral("Application opened in " + ch.Mention() + "!")
}

func (bot *Bot) chooseAppTrack(ctx *bcr.ButtonContext) (err error) {
	return bot.chooseAppTrackInner(ctx, false)
}

func (bot *Bot) chooseAppTrackAlreadyRestarted(ctx *bcr.ButtonContext) (err error) {
	return bot.chooseAppTrackInner(ctx, true)
}

func (bot *Bot) chooseAppTrackInner(ctx *bcr.ButtonContext, isRestart bool) (err error) {
	if ctx.Member == nil {
		return ctx.ReplyEphemeral("This event didn't have a member associated with it! This is a bug, please report it to the developer (such as by DMing me!)")
	}

	if ctx.Event.Message == nil {
		return ctx.ReplyEphemeral("This event didn't have a member associated with it! This is a bug, please report it to the developer (such as by DMing me!)")
	}

	app, err := bot.DB.ChannelApplication(ctx.Event.ChannelID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ctx.ReplyEphemeral("This channel isn't an application channel!")
		}
		return
	}

	if app.TrackID != nil {
		components := ctx.Event.Message.Components
		for _, c := range components {
			v, ok := c.(*discord.ActionRowComponent)
			if ok {
				for i := range *v {
					btn, ok := (*v)[i].(*discord.ButtonComponent)
					if ok {
						btn.Disabled = true
						(*v)[i] = btn
					}
				}
			}
		}

		return ctx.State.RespondInteraction(ctx.InteractionID, ctx.InteractionToken, api.InteractionResponse{
			Type: api.UpdateMessage,
			Data: &api.InteractionResponseData{
				Components: &components,
			},
		})
	}

	if app.UserID != ctx.User.ID {
		return ctx.ReplyEphemeral("You're not the user who this application is for.")
	}

	ids := strings.TrimPrefix(string(ctx.CustomID), "app-track")
	if isRestart {
		ids = strings.TrimPrefix(ids, "-restart")
	}
	ids = strings.TrimPrefix(ids, ":")

	trackID, err := strconv.ParseInt(ids, 10, 10)
	if err != nil {
		return err
	}

	track, err := bot.DB.ApplicationTrack(trackID)
	if err != nil {
		return err
	}

	// we have a track, so finish this interaction
	components := ctx.Event.Message.Components
	hasRestartButton := false
	for _, c := range components {
		v, ok := c.(*discord.ActionRowComponent)
		if ok {
			for i := range *v {
				btn, ok := (*v)[i].(*discord.ButtonComponent)
				if ok {
					if btn.CustomID != "restart-app" {
						btn.Disabled = true
						(*v)[i] = btn
					} else {
						hasRestartButton = true
					}
				}
			}

			if !hasRestartButton && !isRestart {
				*v = append(*v, &discord.ButtonComponent{
					Label:    "Restart",
					Style:    discord.SecondaryButtonStyle(),
					CustomID: "restart-app",
				})
			}
		}
	}

	err = ctx.State.RespondInteraction(ctx.InteractionID, ctx.InteractionToken, api.InteractionResponse{
		Type: api.UpdateMessage,
		Data: &api.InteractionResponseData{
			Components: &components,
		},
	})
	if err != nil {
		common.Log.Errorf("Error responding to interaction: %v", err)
	}

	err = bot.DB.SetTrack(app.ID, track.ID)
	if err != nil {
		bot.SendError("Error setting application track in %v: %v", app.ChannelID.Mention(), err)
		_, err = ctx.State.SendMessage(app.ChannelID, "Something went wrong! Please ask a mod for assistance.")
		return
	}

	_, err = ctx.State.SendMessage(app.ChannelID, fmt.Sprintf("Starting **%v** application!\nIf you picked the wrong type, or want to restart your application for any other reason, please press the \"restart\" button above!", track.Name))
	if err != nil {
		bot.SendError("error sending message: %v", err)
	}

	qs, err := bot.DB.Questions(track.ID)
	if err != nil {
		bot.SendError("Error getting questions: %v", err)
		return
	}

	if len(qs) == 0 {
		bot.SendError("No questions for track ID %v, can't start application!", track.ID)
		return
	}

	err = bot.sendInterviewMessage(app, qs[0].Question)
	if err != nil {
		common.Log.Errorf("Error sending message in app %v: %v", app.ChannelID, err)
		return
	}

	// set question index to 1 (first user message will start question loop)
	err = bot.DB.SetQuestionIndex(app.ID, 1)
	if err != nil {
		return errors.Wrap(err, "setting question index")
	}

	return
}

func (bot *Bot) restartAppInteraction(ctx *bcr.ButtonContext) error {
	if ctx.Member == nil {
		return ctx.ReplyEphemeral("This event didn't have a member associated with it! This is a bug, please report it to the developer (such as by DMing me!)")
	}

	if ctx.Event.Message == nil {
		return ctx.ReplyEphemeral("This event didn't have a member associated with it! This is a bug, please report it to the developer (such as by DMing me!)")
	}

	app, err := bot.DB.ChannelApplication(ctx.Event.ChannelID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ctx.ReplyEphemeral("This channel isn't an application channel!")
		}
		return err
	}

	if app.UserID != ctx.User.ID {
		return ctx.ReplyEphemeral("You're not the user who this application is for.")
	}

	err = bot.DB.ResetApplication(app.ID)
	if err != nil {
		bot.SendError("Error resetting application %v: %v", app.ID, err)
		return ctx.ReplyEphemeral("Internal error occurred! Please ping staff for assistance.")
	}

	tracks, err := bot.DB.ApplicationTracks()
	if err != nil {
		bot.SendError("Error getting app tracks: %v", err)
		return ctx.ReplyEphemeral("Internal error occurred! Please ping staff for assistance.")
	}

	str := "Are you "

	var descs []string
	var buttons discord.ActionRowComponent

	for _, t := range tracks {
		descs = append(descs, fmt.Sprintf("%v (%s)", t.Description, t.Emoji()))
		buttons = append(buttons, &discord.ButtonComponent{
			Label:    t.Name,
			CustomID: discord.ComponentID("app-track-restart:" + strconv.FormatInt(t.ID, 10)),
			Style:    discord.SecondaryButtonStyle(),
			Emoji: &discord.ComponentEmoji{
				Name:     t.Emoji().Name,
				ID:       t.Emoji().ID,
				Animated: t.Emoji().Animated,
			},
		})
	}
	str += english.OxfordWordSeries(descs, "or") + "?"

	return ctx.ReplyComplex(api.InteractionResponseData{
		Content:    option.NewNullableString(str),
		Components: &discord.ContainerComponents{&buttons},
	})
}

func (bot *Bot) sendInterviewMessage(app *db.Application, msg string) error {
	s, _ := bot.Router.StateFromGuildID(bot.DB.BotConfig.GuildID)

	_, err := s.SendMessageComplex(app.ChannelID, api.SendMessageData{
		Content: msg,
		AllowedMentions: &api.AllowedMentions{
			Parse: []api.AllowedMentionType{},
		},
	})
	return err
}
