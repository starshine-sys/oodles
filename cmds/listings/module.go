package listings

import (
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/bot"
)

type Bot struct {
	*bot.Bot
}

func Init(bot *bot.Bot) {
	b := &Bot{bot}

	invites := b.Router.AddCommand(&bcr.Command{
		Name:              "invites",
		Aliases:           []string{"invite"},
		Summary:           "List and manage invites",
		CustomPermissions: b.Checker,
		Command:           b.listInvites,
	})

	invites.AddSubcommand(&bcr.Command{
		Name:              "create",
		Aliases:           []string{"make"},
		Summary:           "Create a new invite",
		Usage:             "<name>",
		CustomPermissions: b.Checker,
		Command:           b.createInvite,
	})
}
