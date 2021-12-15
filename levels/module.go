package levels

import (
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/bot"
)

type Bot struct {
	*bot.Bot
}

func Init(b *bot.Bot) {
	bot := &Bot{b}

	bot.Router.AddHandler(bot.messageCreate)

	bot.Router.AddCommand(&bcr.Command{
		Name:              "level",
		Aliases:           []string{"lvl", "rank"},
		Summary:           "Check your, or another user's, level",
		Usage:             "[user]",
		CustomPermissions: b.Checker,
		Command:           bot.levelCmd,
	})

	cfg := bot.Router.AddCommand(&bcr.Command{
		Name:              "levelcfg",
		Aliases:           []string{"levelconfig"},
		Summary:           "Configure levels for this server",
		CustomPermissions: b.Checker,
		Command:           bot.configShow,
	})

	cfg.AddSubcommand(&bcr.Command{
		Name:              "set",
		Summary:           "Set a configuration key",
		Usage:             "<key> <new value>",
		Args:              bcr.MinArgs(2),
		CustomPermissions: b.Checker,
		Command:           bot.setConfig,
	})

	bl := cfg.AddSubcommand(&bcr.Command{
		Name:              "blacklist",
		Summary:           "Manage this server's level blacklist",
		CustomPermissions: b.Checker,
		Command: func(ctx *bcr.Context) error {
			return ctx.Help([]string{"levelcfg", "blacklist"})
		},
	})

	bl.AddSubcommand(&bcr.Command{
		Name:              "add",
		Summary:           "Add a channel, category, or role to the level blacklist.",
		Usage:             "<channel or role>",
		Args:              bcr.MinArgs(1),
		CustomPermissions: b.Checker,
		Command:           bot.blacklistAdd,
	})

	bl.AddSubcommand(&bcr.Command{
		Name:              "remove",
		Summary:           "Remove a channel, category, or role from the level blacklist.",
		Usage:             "<channel or role>",
		Args:              bcr.MinArgs(1),
		CustomPermissions: b.Checker,
		Command:           bot.blacklistRemove,
	})
}
