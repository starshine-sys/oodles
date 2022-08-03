package applications

import (
	"sync"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/spf13/pflag"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/bot"
	"github.com/starshine-sys/oodles/common"
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

	b.Scheduler.AddType(&timeout{})
	b.Scheduler.AddType(&scheduledClose{})

	b.Interactions.Button(common.OpenApplication).Exec(b.createInterview)
	b.Interactions.Button("restart-app").Exec(b.restartAppInteraction)
	b.Interactions.Button("app-track:*").Exec(b.chooseAppTrack)
	b.Interactions.Button("app-track-restart:*").Exec(b.chooseAppTrackAlreadyRestarted)

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

	close := b.Router.AddCommand(&bcr.Command{
		Name:              "close",
		Summary:           "Close the current application",
		CustomPermissions: b.Checker,
		Command:           b.closeApp,
		Flags: func(fs *pflag.FlagSet) *pflag.FlagSet {
			fs.BoolP("force", "F", false, "Close even if no transcript was made.")
			return fs
		},
	})

	close.AddSubcommand(&bcr.Command{
		Name:              "cancel",
		Summary:           "Cancel closing the current application",
		CustomPermissions: b.Checker,
		Command:           b.closeCancel,
	})

	b.Router.AddCommand(
		b.Router.AliasMust("cc", []string{}, []string{"close", "cancel"}, nil),
	)

	b.Router.AddCommand(&bcr.Command{
		Name:              "deny",
		Summary:           "Deny the current application",
		Usage:             "[reason...]",
		CustomPermissions: b.Checker,
		Command:           b.deny,
	})

	b.Router.AddCommand(&bcr.Command{
		Name:              "open",
		Aliases:           []string{"create"},
		Summary:           "Open an application for the given user",
		Usage:             "<user>",
		Args:              bcr.MinArgs(1),
		CustomPermissions: b.Checker,
		Command:           b.fallbackCreate,
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
