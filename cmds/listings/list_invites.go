package listings

import (
	"fmt"
	"strconv"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/bcr"
)

func (bot *Bot) listInvites(ctx *bcr.Context) error {
	invs, err := ctx.State.GuildInvites(ctx.Message.GuildID)
	if err != nil {
		return bot.Report(ctx, err)
	}

	names, err := bot.DB.AllInvites()
	if err != nil {
		return bot.Report(ctx, err)
	}

	fields := make([]discord.EmbedField, 0, len(invs))
	for _, inv := range invs {
		name, ok := names[inv.Code]
		if !ok {
			name = "Unnamed"
		}

		fields = append(fields, discord.EmbedField{
			Name:   inv.Code,
			Value:  fmt.Sprintf("%v\n\nUses: %v", name, inv.Uses),
			Inline: true,
		})
	}

	_, _, err = ctx.ButtonPages(
		bcr.FieldPaginator("Invites ("+strconv.Itoa(len(invs))+")", "", bot.Colour, fields, 6),
		15*time.Minute,
	)
	return err
}
