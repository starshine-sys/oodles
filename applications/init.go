package applications

import (
	"strconv"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/jackc/pgx/v4"
	"github.com/starshine-sys/oodles/common"
	"github.com/starshine-sys/oodles/db"
)

const waitTime = 2 * time.Second

func (bot *Bot) interactionCreate(ev *gateway.InteractionCreateEvent) {
	data, ok := ev.Data.(*discord.ComponentInteractionData)
	if !ok {
		return
	}

	if data.CustomID == common.OpenApplication {
		err := bot.createInterview(ev, data)
		if err != nil {
			common.Log.Errorf("Error in open application interaction: %v", err)
		}
	}

	if strings.HasPrefix(data.CustomID, "app-track:") {
		err := bot.chooseAppTrack(ev, data)
		if err != nil {
			common.Log.Errorf("Error in choose app track: %v", err)
		}
	}
}

func (bot *Bot) createInterview(ev *gateway.InteractionCreateEvent, data *discord.ComponentInteractionData) (err error) {
	s, _ := bot.Router.StateFromGuildID(bot.DB.BotConfig.GuildID)

	if ev.Member == nil {
		return bot.respond(ev, "This event didn't have a member associated with it! This is a bug, please report it to the developer (such as by DMing me!)")
	}

	// i don't trust all the API calls to be fast enough
	if err = bot.initialResponse(ev); err != nil {
		return err
	}

	// check application
	existing, err := bot.DB.UserApplication(ev.Member.User.ID)
	if err == nil {
		ch, err := s.Channel(existing.ChannelID)
		if err == nil {
			_, err = bot.updateResponse(ev, "You already have an open application, in %v!", ch.Mention())
			return err
		}

		// no channel, app should've been closed
		err = bot.DB.CloseApplication(existing.ID)
		if err != nil {
			bot.SendError("Error closing existing app: %v", err)
		}
	}

	if err != nil && err != pgx.ErrNoRows {
		bot.SendError("Unknown error fetching app: %v", err)
		_, err = bot.updateResponse(ev, "There was an unknown error fetching an existing app!")
		return
	}

	ch, err := bot.newApplicationChannel(*ev.Member)
	if err != nil {
		bot.SendError("Error creating application channel: %v", err)
		_, err = bot.updateResponse(ev, "I couldn't create an application channel!")
		return
	}

	err = bot.DB.CreateApplication(ev.Member.User.ID, ch.ID)
	if err != nil {
		bot.SendError("Error registering application in DB: %v", err)
		_, err = bot.updateResponse(ev, "I couldn't save the newly opened application!")
		return
	}

	err = bot.sendInitialMessage(ch.ID, *ev.Member)
	if err != nil {
		bot.SendError("Error sending initial message: %v", err)
		_, err = bot.updateResponse(ev, "I couldn't send the initial message!")
		return
	}

	_, err = bot.updateResponse(ev, "Application opened in %v!", ch.Mention())
	return
}

func (bot *Bot) chooseAppTrack(ev *gateway.InteractionCreateEvent, data *discord.ComponentInteractionData) (err error) {
	s, _ := bot.Router.StateFromGuildID(bot.DB.BotConfig.GuildID)

	if ev.Member == nil {
		return bot.respond(ev, "This event didn't have a member associated with it! This is a bug, please report it to the developer (such as by DMing me!)")
	}

	if ev.Message == nil {
		return bot.respond(ev, "This event didn't have a member associated with it! This is a bug, please report it to the developer (such as by DMing me!)")
	}

	app, err := bot.DB.ChannelApplication(ev.ChannelID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return bot.respond(ev, "This channel isn't an application channel!")
		}
		return
	}

	if app.TrackID != nil {
		components := discord.UnwrapComponents(ev.Message.Components)
		for _, c := range components {
			v, ok := c.(*discord.ActionRowComponent)
			if ok {
				for i := range v.Components {
					btn, ok := v.Components[i].(*discord.ButtonComponent)
					if ok {
						btn.Disabled = true
						v.Components[i] = btn
					}
				}
			}
		}

		return s.RespondInteraction(ev.ID, ev.Token, api.InteractionResponse{
			Type: api.UpdateMessage,
			Data: &api.InteractionResponseData{
				Components: &components,
			},
		})
	}

	if app.UserID != ev.Member.User.ID {
		return bot.respond(ev, "You're not the user who this application is for.")
	}

	trackID, err := strconv.ParseInt(strings.TrimPrefix(data.CustomID, "app-track:"), 10, 10)
	if err != nil {
		return err
	}

	track, err := bot.DB.ApplicationTrack(trackID)
	if err != nil {
		return err
	}

	// we have a track, so finish this interaction
	components := discord.UnwrapComponents(ev.Message.Components)
	for _, c := range components {
		v, ok := c.(*discord.ActionRowComponent)
		if ok {
			for i := range v.Components {
				btn, ok := v.Components[i].(*discord.ButtonComponent)
				if ok {
					btn.Disabled = true
					v.Components[i] = btn
				}
			}
		}
	}

	err = s.RespondInteraction(ev.ID, ev.Token, api.InteractionResponse{
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
		_, err = s.SendMessage(app.ChannelID, "Something went wrong! Please ask a mod for assistance.")
		return
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
