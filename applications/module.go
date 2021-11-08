package applications

import (
	"fmt"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/starshine-sys/oodles/bot"
)

// Bot ...
type Bot struct {
	*bot.Bot
}

// Init ...
func Init(bot *bot.Bot) {
	b := &Bot{bot}

	b.Router.AddHandler(b.interactionCreate)
	b.Router.AddHandler(b.messageCreate)
}

func (bot *Bot) initialResponse(ev *gateway.InteractionCreateEvent) error {
	s, _ := bot.Router.StateFromGuildID(bot.DB.BotConfig.GuildID)

	return s.RespondInteraction(ev.ID, ev.Token, api.InteractionResponse{
		Type: api.DeferredMessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Flags: api.EphemeralResponse,
		},
	})
}

func (bot *Bot) respond(ev *gateway.InteractionCreateEvent, tmpl string, args ...interface{}) error {
	s, _ := bot.Router.StateFromGuildID(bot.DB.BotConfig.GuildID)

	return s.RespondInteraction(ev.ID, ev.Token, api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Content: option.NewNullableString(fmt.Sprintf(tmpl, args...)),
			Flags:   api.EphemeralResponse,
			AllowedMentions: &api.AllowedMentions{
				Parse: []api.AllowedMentionType{},
			},
		},
	})
}

func (bot *Bot) updateResponse(ev *gateway.InteractionCreateEvent, tmpl string, args ...interface{}) (*discord.Message, error) {
	s, _ := bot.Router.StateFromGuildID(bot.DB.BotConfig.GuildID)

	return s.EditInteractionResponse(discord.AppID(bot.Router.Bot.ID), ev.Token, api.EditInteractionResponseData{
		Content: option.NewNullableString(fmt.Sprintf(tmpl, args...)),
		AllowedMentions: &api.AllowedMentions{
			Parse: []api.AllowedMentionType{},
		},
	})
}
