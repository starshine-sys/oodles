package moderation

import (
	"fmt"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/bcr"
)

func (bot *Bot) modlogs(ctx *bcr.Context) (err error) {
	u, err := ctx.ParseUser(ctx.RawArgs)
	if err != nil {
		_, err = ctx.Send("Couldn't find that user.")
		return
	}

	entries, err := bot.DB.ModLogFor(ctx.Message.GuildID, u.ID)
	if err != nil {
		return bot.Report(ctx, err)
	}

	modCache := map[discord.UserID]*discord.User{}

	var fields []discord.EmbedField

	for _, entry := range entries {
		mod := modCache[entry.ModID]
		if mod == nil {
			mod, err = ctx.State.User(entry.ModID)
			if err != nil {
				return bot.Report(ctx, err)
			}
			modCache[entry.ModID] = mod
		}

		f := discord.EmbedField{
			Name:  fmt.Sprintf("#%v | %v | <t:%v>", entry.ID, entry.ActionType, entry.Time.Unix()),
			Value: fmt.Sprintf("**Responsible moderator:** %v\n%v", mod.Tag(), entry.Reason),
		}

		if len(f.Value) > 1020 {
			f.Value = f.Value[:1020] + "..."
		}

		fields = append(fields, f)
	}

	embeds := bcr.FieldPaginator("Mod logs", "", bot.Colour, fields, 5)
	for i := range embeds {
		embeds[i].Author = &discord.EmbedAuthor{
			Name: u.Tag() + " (" + u.ID.String() + ")",
			Icon: u.AvatarURL(),
		}
	}

	_, _, err = ctx.ButtonPages(embeds, 30*time.Minute)
	return err
}
