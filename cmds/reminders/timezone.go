package reminders

import (
	"time"

	"github.com/starshine-sys/bcr"
)

func (bot *Bot) setTz(ctx *bcr.Context) (err error) {
	loc, err := time.LoadLocation(ctx.RawArgs)
	if err != nil {
		return ctx.SendfX("We couldn't find a timezone named %v!\nTimezone should be in `Continent/City` format; to find your timezone, use a tool such as <https://xske.github.io/tz/>.", bcr.AsCode(ctx.RawArgs))
	}

	err = bot.DB.UserStringSet(ctx.Author.ID, "timezone", loc.String())
	if err != nil {
		return bot.Report(ctx, err)
	}

	return ctx.SendfX("Timezone set to %v!", loc.String())
}
