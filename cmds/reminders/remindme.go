package reminders

import (
	"context"
	"fmt"
	"strings"
	"time"

	"codeberg.org/eviedelta/detctime/durationparser"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/common"
)

func (bot *Bot) remindme(ctx *bcr.Context) (err error) {
	loc := bot.DB.UserTime(context.Background(), ctx.Author.ID)

	t, i, err := ParseTime(ctx.Args, loc)
	if err != nil {
		dur, err := durationparser.Parse(ctx.Args[0])
		if err != nil {
			_, err = ctx.Replyc(bcr.ColourRed, "I couldn't parse your input as a valid time or duration.")
			return err
		}
		i = 0
		t = time.Now().In(loc).Add(dur)
	}

	if t.Before(time.Now().In(loc)) {
		_, err = ctx.Replyc(bcr.ColourRed, "That time is in the past.")
		return
	}

	rm := "N/A"

	if len(ctx.Args) > i+1 {
		rm = ctx.RawArgs
		for n := 0; n <= i; n++ {
			rm = strings.TrimSpace(strings.TrimPrefix(rm, ctx.Args[n]))
		}
	}

	if rm == ctx.RawArgs {
		rm = strings.Join(ctx.Args[i+1:], " ")
	}

	id, err := bot.Scheduler.Add(t, &reminder{
		UserID:       ctx.Author.ID,
		ReminderText: rm,
		SetTime:      time.Now().UTC(),
		GuildID:      ctx.Message.GuildID,
		ChannelID:    ctx.Message.ChannelID,
		MessageID:    ctx.Message.ID,
	})
	if err != nil {
		return bot.Report(ctx, err)
	}

	if len(rm) > 128 {
		rm = rm[:128] + "..."
	}
	if rm == "N/A" {
		rm = "something"
	} else {
		rm = "**" + rm + "**"
	}

	content := fmt.Sprintf("Okay %v, we'll remind you about %v %v! (<t:%v>, #%v)", ctx.DisplayName(), rm, bcr.HumanizeTime(bcr.DurationPrecisionSeconds, t.Add(time.Second)), t.Unix(), id)

	msg, err := ctx.State.SendMessageComplex(ctx.Message.ChannelID, api.SendMessageData{
		Content: content,
		AllowedMentions: &api.AllowedMentions{
			Parse: []api.AllowedMentionType{},
		},
	})
	if err != nil {
		return
	}

	_, err = bot.DB.Pool.Exec(context.Background(), "update scheduled_events set data = jsonb_set(data, array['message_id'], $1) where id = $2", msg.ID, id)
	if err != nil {
		common.Log.Errorf("Error updating reminder %v message ID: %v", id, err)
	}
	return
}
