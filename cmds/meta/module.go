package meta

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/bot"
)

// Bot ...
type Bot struct {
	*bot.Bot
}

// Init ...
func Init(bot *bot.Bot) {
	b := &Bot{bot}

	b.Router.AddCommand(&bcr.Command{
		Name:              "ping",
		Aliases:           []string{"pong"},
		Summary:           "Ping pong!",
		CustomPermissions: b.Checker,
		Command:           b.ping,
	})

	conf := b.Router.AddCommand(&bcr.Command{
		Name:              "config",
		Summary:           "Configure bot settings",
		CustomPermissions: b.Checker,
		Command: func(ctx *bcr.Context) (err error) {
			return ctx.Help([]string{"config"})
		},
	})

	conf.AddSubcommand(&bcr.Command{
		Name:              "list",
		Summary:           "Show all available settings",
		Usage:             "[setting]",
		CustomPermissions: b.Checker,
		Command:           b.configList,
	})

	conf.AddSubcommand(&bcr.Command{
		Name:              "set",
		Summary:           "Change a setting",
		Usage:             "<setting> <new value>",
		CustomPermissions: b.Checker,
		Command:           b.configSet,
	})

	conf.AddSubcommand(&bcr.Command{
		Name:              "get",
		Summary:           "Get the current value of a setting",
		Usage:             "<setting>",
		CustomPermissions: b.Checker,
		Command:           b.configGet,
	})

	help := b.Router.AddCommand(&bcr.Command{
		Name:              "help",
		Aliases:           []string{"hlep"},
		Summary:           "Show this help!",
		CustomPermissions: b.Checker,
		Command:           b.help,
	})

	help.AddSubcommand(&bcr.Command{
		Name:              "list",
		Aliases:           []string{"commands"},
		Summary:           "Show a list of commands",
		CustomPermissions: b.Checker,
		Command:           b.helpList,
	}).AddSubcommand(&bcr.Command{
		Name:    "all",
		Summary: "Show a list of all commands, including ones you can't use",
		// yes this is the one case where we use discord perms
		GuildPermissions:  discord.PermissionManageMessages,
		CustomPermissions: b.Checker,
		Command:           b.helpAll,
	})
}
