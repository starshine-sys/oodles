package reminders

import (
	"context"
	"fmt"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/starshine-sys/bcr"
)

func (bot *Bot) reminders(ctx *bcr.Context) (err error) {
	rms := []reminder{}

	err = pgxscan.Select(context.Background(), bot.DB.Pool, &rms, "select * from reminders where user_id = $1 order by expires asc", ctx.Author.ID)
	if err != nil {
		return bot.Report(ctx, err)
	}

	title := "Reminders"

	limitChannel, _ := ctx.Flags.GetBool("channel")
	if limitChannel {
		title = "Reminders in #" + ctx.Channel.Name
		prev := rms
		rms = nil
		for _, r := range prev {
			if r.ChannelID == ctx.Channel.ID {
				rms = append(rms, r)
			}
		}
	}

	limitServer, _ := ctx.Flags.GetBool("server")
	if limitServer && ctx.Guild != nil {
		title = "Reminders in " + ctx.Guild.Name
		prev := rms
		rms = nil
		for _, r := range prev {
			if r.GuildID == ctx.Guild.ID {
				rms = append(rms, r)
			}
		}
	}

	if len(rms) == 0 {
		return ctx.SendfX("You have no reminders. Set some with `%vremindme`!", bot.Prefix())
	}

	var slice []string

	for _, r := range rms {
		text := r.ReminderText
		if len(text) > 100 {
			text = text[:100] + "..."
		}

		linkGuild := r.GuildID.String()
		if !r.GuildID.IsValid() {
			linkGuild = "@me"
		}

		slice = append(slice, fmt.Sprintf(`**#%v**: %v
<t:%v> ([link](https://discord.com/channels/%v/%v/%v))
`, r.ID, text, r.Expires.Unix(), linkGuild, r.ChannelID, r.MessageID))
	}

	embeds := bcr.StringPaginator(fmt.Sprintf("%v (%v)", title, len(rms)), bcr.ColourBlurple, slice, 5)

	_, _, err = ctx.ButtonPages(embeds, 10*time.Minute)
	return
}
