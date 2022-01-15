package listings

import (
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/bcr"
)

func (bot *Bot) renameInvite(ctx *bcr.Context) (err error) {
	code := ctx.Args[0]
	name := strings.TrimSpace(strings.TrimPrefix(ctx.RawArgs, ctx.Args[0]))
	if name == ctx.RawArgs {
		name = strings.Join(ctx.Args[1:], " ")
	}

	invs, err := ctx.State.GuildInvites(ctx.Message.GuildID)
	if err != nil {
		return bot.Report(ctx, err)
	}

	found := false
	for _, inv := range invs {
		if inv.Code == code {
			found = true
			break
		}
	}

	if !found {
		return ctx.SendfX("``%v`` doesn't seem to be a valid invite for this server. Please double check your input!", bcr.EscapeBackticks(code))
	}

	err = bot.DB.SetInviteName(code, name)
	if err != nil {
		return bot.Report(ctx, err)
	}

	return ctx.SendX("Invite **"+code+"** renamed!", discord.Embed{
		Color:       bot.Colour,
		Description: name,
	})
}
