package moderation

import (
	"sort"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/bot"
)

type Bot struct {
	*bot.Bot
}

func Init(b *bot.Bot) {
	bot := &Bot{b}

	bot.Scheduler.AddType(&changeRoles{})

	bot.Router.AddCommand(&bcr.Command{
		Name:              "warn",
		Summary:           "Warn a user",
		Usage:             "<user> <reason>",
		CustomPermissions: bot.Checker,
		Command:           bot.warn,
	})

	bot.Router.AddCommand(&bcr.Command{
		Name:              "mute",
		Summary:           "Mute a user",
		Usage:             "<user> [duration] [reason]",
		CustomPermissions: bot.Checker,
		Command:           bot.mute,
	})

	bot.Router.AddCommand(&bcr.Command{
		Name:              "hardmute",
		Summary:           "Hardmute a user",
		Usage:             "<user> [duration] [reason]",
		CustomPermissions: bot.Checker,
		Command:           bot.hardmute,
	})

	bot.Router.AddCommand(&bcr.Command{
		Name:              "unmute",
		Summary:           "Unmute a user",
		Usage:             "<user> [reason]",
		CustomPermissions: bot.Checker,
		Command:           bot.unmute,
	})

	bot.Router.AddCommand(&bcr.Command{
		Name:              "ban",
		Summary:           "Ban a user",
		Usage:             "<user> [reason]",
		CustomPermissions: bot.Checker,
		Command:           bot.ban,
	})

	bot.Router.AddCommand(&bcr.Command{
		Name:              "kick",
		Summary:           "Kick a user",
		Usage:             "<user> [reason]",
		CustomPermissions: bot.Checker,
		Command:           bot.kick,
	})

	bot.Router.AddCommand(&bcr.Command{
		Name:              "unban",
		Summary:           "Unban a user",
		Usage:             "<user> [reason]",
		CustomPermissions: bot.Checker,
		Command:           bot.unban,
	})
}

// aboveUser returns true if mod is above member in the role hierarchy.
func (bot *Bot) aboveUser(ctx *bcr.Context, mod *discord.Member, member *discord.Member) (above bool) {
	if ctx.Guild == nil {
		return false
	}

	if ctx.Guild.OwnerID == mod.User.ID {
		return true
	}

	var modRoles, memberRoles bcr.Roles
	for _, r := range ctx.Guild.Roles {
		for _, id := range mod.RoleIDs {
			if r.ID == id {
				modRoles = append(modRoles, r)
				break
			}
		}
		for _, id := range member.RoleIDs {
			if r.ID == id {
				memberRoles = append(memberRoles, r)
				break
			}
		}
	}

	if len(modRoles) == 0 {
		return false
	}
	if len(memberRoles) == 0 {
		return true
	}

	sort.Sort(modRoles)
	sort.Sort(memberRoles)

	return modRoles[0].Position > memberRoles[0].Position
}
