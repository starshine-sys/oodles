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

	b.Router.AddCommand(&bcr.Command{
		Name:              "userinfo",
		Aliases:           []string{"ui"},
		Summary:           "Show user information",
		Usage:             "[user]",
		CustomPermissions: b.Checker,
		Command:           b.memberInfo,
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

	perms := bot.Router.AddCommand(&bcr.Command{
		Name:              "permissions",
		Aliases:           []string{"perms"},
		Summary:           "Show or modify role and user permissions",
		CustomPermissions: b.Checker,
		Command:           b.permsList,
	})

	perms.AddSubcommand(&bcr.Command{
		Name:              "add",
		Summary:           "Add a user or role to a permission level",
		Usage:             "<level> <user or role>",
		Args:              bcr.MinArgs(2),
		CustomPermissions: b.Checker,
		Command:           b.permsAdd,
	})

	perms.AddSubcommand(&bcr.Command{
		Name:              "override",
		Summary:           "Add a permission level override for the given root-level command",
		Usage:             "<level> <root command>",
		Args:              bcr.MinArgs(2),
		CustomPermissions: b.Checker,
		Command:           b.overrideCmdPerms,
	})

	// add other commands
	appCommands(b)
}

func appCommands(b *Bot) {
	app := b.Router.AddCommand(&bcr.Command{
		Name:              "app",
		Aliases:           []string{"application", "applications", "apps"},
		Summary:           "Manage applications",
		CustomPermissions: b.Checker,
		Command: func(ctx *bcr.Context) (err error) {
			return ctx.Help([]string{"app"})
		},
	})

	app.AddSubcommand(&bcr.Command{
		Name:              "setup",
		Summary:           "Send the application trigger message in the current channel.",
		CustomPermissions: b.Checker,
		Command:           b.setupMessage,
	})

	track := app.AddSubcommand(&bcr.Command{
		Name:              "track",
		Aliases:           []string{"tracks"},
		Summary:           "List and manage application tracks.",
		CustomPermissions: b.Checker,
		Command:           b.listAppTracks,
	})

	track.AddSubcommand(&bcr.Command{
		Name:              "create",
		Summary:           "Create an application track",
		Usage:             "<name> <description> <emoji>",
		Args:              bcr.MinArgs(3),
		CustomPermissions: b.Checker,
		Command:           b.createAppTrack,
	})

	questions := app.AddSubcommand(&bcr.Command{
		Name:              "questions",
		Aliases:           []string{"question", "q"},
		Summary:           "List or manage application questions",
		CustomPermissions: b.Checker,
		Command:           b.listQuestions,
	})

	questions.AddSubcommand(&bcr.Command{
		Name:              "add",
		Summary:           "Add questions to the given application track",
		Description:       "Add questions to the given application track. Separate questions with a |",
		Usage:             "<id> <questions...>",
		Args:              bcr.MinArgs(2),
		CustomPermissions: b.Checker,
		Command:           b.addQuestion,
	})
}
