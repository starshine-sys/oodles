package reminders

import (
	"github.com/spf13/pflag"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/bot"
)

// Bot ...
type Bot struct {
	*bot.Bot
}

// Init ...
func Init(b *bot.Bot) {
	bot := &Bot{b}

	bot.Scheduler.AddType(&reminder{})

	rm := bot.Router.AddCommand(&bcr.Command{
		Name:              "remindme",
		Aliases:           []string{"rm", "remind"},
		Summary:           "Set a reminder for yourself",
		Usage:             "<time> [text]",
		Args:              bcr.MinArgs(1),
		CustomPermissions: bot.Checker,
		Command:           bot.remindme,
	})

	rm.AddSubcommand(&bcr.Command{
		Name:              "timezone",
		Aliases:           []string{"tz"},
		Summary:           "Set your timezone for reminders",
		Usage:             "<timezone>",
		Args:              bcr.MinArgs(1),
		CustomPermissions: bot.Checker,
		Command:           bot.setTz,
	})

	rm.AddSubcommand(&bcr.Command{
		Name:              "list",
		Summary:           "Show your current reminders",
		CustomPermissions: bot.Checker,
		Command:           bot.reminders,
		Flags: func(fs *pflag.FlagSet) *pflag.FlagSet {
			fs.BoolP("channel", "c", false, "Only show reminders in this channel.")
			fs.BoolP("server", "s", false, "Only show reminders in this server.")

			return fs
		},
	})

	rm.AddSubcommand(&bcr.Command{
		Name:              "delete",
		Aliases:           []string{"del", "yeet"},
		Summary:           "Delete a reminder",
		CustomPermissions: bot.Checker,
		Command:           bot.delreminder,
	})
}
