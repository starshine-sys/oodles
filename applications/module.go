package applications

import (
	"fmt"
	"sync"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/spf13/pflag"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/bot"
)

// Bot ...
type Bot struct {
	*bot.Bot

	deniedMu sync.RWMutex
	denied   map[discord.UserID]struct{}
}

// Init ...
func Init(bot *bot.Bot) {
	b := &Bot{bot, sync.RWMutex{}, make(map[discord.UserID]struct{})}

	b.Router.AddHandler(b.interactionCreate)
	b.Router.AddHandler(b.messageCreate)
	b.Router.AddHandler(b.guildMemberAdd)
	b.Router.AddHandler(b.guildMemberRemove)

	b.Router.AddCommand(&bcr.Command{
		Name:              "verify",
		Aliases:           []string{"accept", "approve"},
		Summary:           "Verify the current application",
		CustomPermissions: b.Checker,
		Command:           b.verify,
	})

	b.Router.AddCommand(&bcr.Command{
		Name:              "close",
		Summary:           "Close the current application",
		CustomPermissions: b.Checker,
		Command:           b.closeApp,
		Flags: func(fs *pflag.FlagSet) *pflag.FlagSet {
			fs.BoolP("force", "F", false, "Close even if no transcript was made.")
			return fs
		},
	})

	b.Router.AddCommand(&bcr.Command{
		Name:              "deny",
		Summary:           "Deny the current application",
		Usage:             "[reason...]",
		CustomPermissions: b.Checker,
		Command:           b.deny,
	})

	b.Router.AddCommand(&bcr.Command{
		Name:              "logs",
		Summary:           "Show application logs for the given user",
		Usage:             "<user>",
		Args:              bcr.MinArgs(1),
		CustomPermissions: b.Checker,
		Command:           b.logs,
	})

	b.Router.AddCommand(&bcr.Command{
		Name:              "unverified",
		Summary:           "Show users who are unverified with no open application",
		Usage:             "[since]",
		CustomPermissions: b.Checker,
		Command:           b.unverified,
	})

	b.Router.AddCommand(&bcr.Command{
		Name:              "restart",
		Summary:           "Restart an application",
		Usage:             "[since]",
		CustomPermissions: b.Checker,
		Command:           b.restart,
	})
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
