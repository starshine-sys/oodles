package meta

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/sendpart"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/dischtml"
)

func (bot *Bot) transcript(ctx *bcr.Context) (err error) {
	logCh, err := ctx.ParseChannel(ctx.Args[0])
	if err != nil || logCh.GuildID != ctx.Message.GuildID || (logCh.Type != discord.GuildText && logCh.Type != discord.GuildNews) {
		return ctx.SendX("Channel not found, or it's not in this server, or it's not a text channel.")
	}

	limit := 500
	if len(ctx.Args) > 1 {
		limit, err = strconv.Atoi(ctx.Args[1])
		if err != nil {
			return ctx.SendfX("Couldn't parse %v as a number.", bcr.AsCode(ctx.Args[1]))
		}
	}

	g, err := ctx.State.Guild(ctx.Message.GuildID)
	if err != nil {
		return bot.Report(ctx, err)
	}

	chs, err := ctx.State.Channels(ctx.Message.GuildID)
	if err != nil {
		return bot.Report(ctx, err)
	}

	rls, err := ctx.State.Roles(ctx.Message.GuildID)
	if err != nil {
		return bot.Report(ctx, err)
	}

	members, err := ctx.State.Members(ctx.Message.GuildID)
	if err != nil {
		return bot.Report(ctx, err)
	}

	conv := dischtml.Converter{
		Guild:    *g,
		Channels: chs,
		Roles:    rls,
		Members:  members,
	}

	msgs, err := ctx.State.Messages(ctx.Channel.ID, uint(limit))
	if err != nil {
		return bot.Report(ctx, err)
	}

	for _, m := range msgs {
		for _, u := range m.Mentions {
			conv.Users = append(conv.Users, u.User)
		}
	}

	str, err := conv.ConvertHTML(msgs)
	if err != nil {
		return bot.Report(ctx, err)
	}

	html, err := dischtml.Wrap(*g, *ctx.Channel, str, len(msgs))
	if err != nil {
		return bot.Report(ctx, err)
	}

	msg, err := ctx.State.SendMessageComplex(logCh.ID, api.SendMessageData{
		Content: fmt.Sprintf("Transcript of %v (#%v / %v), made at <t:%v>, with %v messages.", ctx.Channel.Mention(), ctx.Channel.Name, ctx.Channel.ID, time.Now().Unix(), len(msgs)),
		Files: []sendpart.File{{
			Name:   "transcript.html",
			Reader: strings.NewReader(html),
		}},
	})
	if err != nil {
		return ctx.SendfX("There was an error saving the transcript: %v", err)
	}
	return ctx.SendfX("Transcript complete! https://discord.com/channels/%v/%v/%v", ctx.Message.GuildID, logCh.ID, msg.ID)
}
