package levels

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/starshine-sys/bcr"
)

func (bot *Bot) leaderboard(ctx *bcr.Context) (err error) {
	full, _ := ctx.Flags.GetBool("full")
	lb, err := bot.getLeaderboard(ctx.Message.GuildID, full)
	if err != nil {
		return bot.Report(ctx, err)
	}

	if len(lb) == 0 {
		_, err = ctx.Sendf("There doesn't seem to be anyone on the leaderboard...")
		return
	}

	var strings []string
	for i, l := range lb {
		strings = append(strings, fmt.Sprintf(
			"%v. %v: `%v` XP, level `%v`\n",
			i+1,
			l.UserID.Mention(),
			humanize.Comma(l.XP),
			LevelFromXP(l.XP),
		))
	}

	name := "Leaderboard for " + ctx.Guild.Name

	_, _, err = ctx.ButtonPages(
		bcr.StringPaginator(name, bot.Colour, strings, 15),
		10*time.Minute,
	)
	return err

}
