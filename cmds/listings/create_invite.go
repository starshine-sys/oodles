package listings

import (
	"fmt"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/starshine-sys/bcr"
)

func (bot *Bot) createInvite(ctx *bcr.Context) (err error) {
	ch, err := ctx.State.Channel(bot.DB.Config.Get("invite_channel").ToChannelID())
	if err != nil {
		return ctx.SendX(":x: No invite channel is set (or the channel is invalid/deleted), cannot create an invite.")
	}

	if ch.Type != discord.GuildText || ch.GuildID != ctx.Message.GuildID {
		return ctx.SendfX(":x: %v is either not a text channel or not in this guild.", ch.Mention())
	}

	if ctx.RawArgs == "" {
		return ctx.SendX(":x: You must give the new invite a name!")
	}

	inv, err := ctx.State.CreateInvite(ch.ID, api.CreateInviteData{
		Unique:         true,
		MaxAge:         option.NewUint(0),
		AuditLogReason: api.AuditLogReason(ctx.Author.Tag() + ": create invite"),
	})
	if err != nil {
		return bot.Report(ctx, err)
	}

	err = bot.DB.SetInviteName(inv.Code, ctx.RawArgs)
	if err != nil {
		return bot.Report(ctx, err)
	}

	return ctx.SendX(
		fmt.Sprintf("Success! Invite created in %v: https://discord.gg/%v", inv.Channel.Mention(), inv.Code),
		discord.Embed{
			Color:       bot.Colour,
			Description: ctx.RawArgs,
		},
	)
}
