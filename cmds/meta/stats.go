package meta

import (
	"fmt"
	"runtime"
	"strings"
	"sync/atomic"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/common"
)

func (bot *Bot) stats(ctx *bcr.Context) (err error) {
	e := discord.Embed{
		Title: "Statistics",
		Fields: []discord.EmbedField{
			{
				Name:  "Running since",
				Value: fmt.Sprintf("<t:%v:D> <t:%v:T>\nlast restarted %v", bot.Uptime.Unix(), bot.Uptime.Unix(), bcr.HumanizeTime(bcr.DurationPrecisionSeconds, bot.Uptime)),
			},
			{
				Name:  "Version",
				Value: fmt.Sprintf("[%v](https://github.com/starshine-sys/oodles)\n%v on %v/%v", common.Version, runtime.Version(), runtime.GOOS, runtime.GOARCH),
			},
			{
				Name: "Pointless stats since last restart",
				Value: fmt.Sprintf(`**%v** messages seen
**%v** mentions

The letter "e" has been used **%v** times

Someone with "n" in their name has talked **%v** times`, atomic.LoadInt64(&bot.messagesSeen), atomic.LoadInt64(&bot.mentions), atomic.LoadInt64(&bot.timesEWasUsed), atomic.LoadInt64(&bot.timesSomeoneWithNInNameTalked)),
			},
		},
		Color:     bot.Colour,
		Timestamp: discord.NowTimestamp(),
	}

	return ctx.SendX("", e)
}

func (bot *Bot) pointlessStats(m *gateway.MessageCreateEvent) {
	atomic.AddInt64(&bot.messagesSeen, 1)

	if strings.Contains(m.Content, "<@"+bot.Router.Bot.ID.String()+">") ||
		strings.Contains(m.Content, "<@!"+bot.Router.Bot.ID.String()+">") {

		atomic.AddInt64(&bot.mentions, 1)
	}

	if m.WebhookID.IsValid() {
		return
	}

	if strings.Contains(strings.ToLower(m.Content), "e") {
		atomic.AddInt64(&bot.timesEWasUsed, 1)
	}

	name := m.Author.Username
	if m.Member != nil {
		if m.Member.Nick != "" {
			name = m.Member.Nick
		}
	}

	if strings.Contains(strings.ToLower(name), "n") {
		atomic.AddInt64(&bot.timesSomeoneWithNInNameTalked, 1)
	}
}
