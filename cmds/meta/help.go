package meta

import (
	"fmt"
	"sort"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/db"
)

func (bot *Bot) help(ctx *bcr.Context) (err error) {
	if len(ctx.Args) > 0 {
		return ctx.Help(ctx.Args)
	}

	e := discord.Embed{
		Title:       bot.DB.BotConfig.Help.Title,
		Description: bot.DB.BotConfig.Help.Description,
		Color:       bot.Colour,
		Thumbnail: &discord.EmbedThumbnail{
			URL: bot.Router.Bot.AvatarURL(),
		},
	}

	if e.Description == "" {
		return bot.helpList(ctx)
	}

	for _, f := range bot.DB.BotConfig.Help.Fields {
		e.Fields = append(e.Fields, discord.EmbedField{
			Name:  f.Name,
			Value: f.Value,
		})
	}

	return ctx.SendX("", e)
}

func (bot *Bot) helpList(ctx *bcr.Context) (err error) {
	lvl := bot.DB.Perms.Level(ctx.Member)

	return bot.helpListInner(ctx, lvl)
}

func (bot *Bot) helpAll(ctx *bcr.Context) (err error) {
	return bot.helpListInner(ctx, db.OwnerLevel)
}

func (bot *Bot) helpListInner(ctx *bcr.Context, lvl db.PermissionLevel) (err error) {
	cmds := bot.Router.Commands()

	sort.Slice(cmds, func(i, j int) bool {
		return cmds[i].Name < cmds[j].Name
	})

	var s []string

	for _, cmd := range cmds {
		cmdLvl := bot.DB.Overrides.For(cmd.Name)
		if cmdLvl > lvl {
			continue
		}

		s = append(s, fmt.Sprintf("`[%d] %s`: %s\n", cmdLvl, cmd.Name, cmd.Summary))
	}

	_, _, err = ctx.ButtonPages(
		bcr.StringPaginator("List of commands", bot.Colour, s, 10),
		15*time.Minute)
	return
}
