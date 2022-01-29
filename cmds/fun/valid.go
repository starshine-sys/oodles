package fun

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/bcr"
)

func (bot *Bot) validCmd(ctx *bcr.Context) (err error) {
	vs, err := bot.GetValids()
	if err != nil {
		bot.SendError("error getting valids: %v", err)
		return // silently
	}

	v := vs[rand.Intn(len(vs))]

	return ctx.SendX("", discord.Embed{
		Color:       bot.Colour,
		Description: v.Response,
	})
}

func (bot *Bot) validPlus(ctx *bcr.Context) (err error) {
	var v valid
	if ctx.RawArgs != "" {
		id, err := strconv.Atoi(ctx.RawArgs)
		if err != nil {
			return ctx.SendfX("%v is not a number!", bcr.AsCode(ctx.RawArgs))
		}
		v, err = bot.GetValid(id)
		if err != nil {
			return ctx.SendfX("There doesn't seem to be a valid entry with that ID!")
		}
	} else {
		vs, err := bot.GetValids()
		if err != nil {
			bot.SendError("error getting valids: %v", err)
			return err // silently
		}
		v = vs[rand.Intn(len(vs))]
	}

	u, err := ctx.State.User(v.UserID)
	if err != nil {
		return ctx.SendX("", discord.Embed{
			Color:       bot.Colour,
			Description: v.Response,
			Footer: &discord.EmbedFooter{
				Text: fmt.Sprintf("ID: %v", v.ID),
			},
		})
	}

	return ctx.SendX("", discord.Embed{
		Color:       bot.Colour,
		Description: v.Response,
		Footer: &discord.EmbedFooter{
			Text: fmt.Sprintf("ID: %v | This response was contributed by %v!", v.ID, u.Tag()),
		},
	})
}
