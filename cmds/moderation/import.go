package moderation

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/common"
)

type importEntry struct {
	ID          int64          `json:"case_id"`
	ModeratorID discord.UserID `json:"moderator_id"`
	OffenderID  discord.UserID `json:"offender_id"`
	Action      string         `json:"action"`
	Timestamp   importTime     `json:"timestamp"`
	Reason      string         `json:"reason"`
}

type importTime time.Time

func (i *importTime) UnmarshalJSON(src []byte) error {
	if string(src) == "null" {
		return nil
	}

	t, err := time.Parse("2006-01-02T15:04:05.999999", strings.Trim(string(src), `"`))
	if err != nil {
		return err
	}
	*i = importTime(t)
	return nil
}

func (bot *Bot) importCmd(ctx *bcr.Context) (err error) {
	cctx := context.Background()

	url := ctx.RawArgs
	if url == "" {
		if len(ctx.Message.Attachments) == 0 {
			return ctx.SendX("You must give either a link to a Carl-bot export, or attach one.")
		}
		url = ctx.Message.Attachments[0].URL
	}

	c, cancel := context.WithTimeout(cctx, time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(c, "GET", url, nil)
	if err != nil {
		return bot.Report(ctx, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return bot.Report(ctx, err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return bot.Report(ctx, err)
	}

	var entries []importEntry
	err = json.Unmarshal(b, &entries)
	if err != nil {
		return bot.Report(ctx, err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return time.Time(entries[i].Timestamp).Before(time.Time(entries[j].Timestamp))
	})

	tx, err := bot.DB.Begin(cctx)
	if err != nil {
		return bot.Report(ctx, err)
	}

	for _, e := range entries {
		if e.Reason == "" {
			e.Reason = "No reason given."
		}

		_, err = tx.Exec(cctx, `insert into mod_log
		(guild_id, user_id, mod_id, action_type, reason, time)
		values ($1, $2, $3, $4, $5, $6)`, ctx.Message.GuildID, e.OffenderID, e.ModeratorID, e.Action, e.Reason, time.Time(e.Timestamp))
		if err != nil {
			common.Log.Errorf("Error inserting mod log entry %v: %v", e.ID, err)
			errs := errors.Append(err, tx.Rollback(cctx))
			return bot.Report(ctx, errs)
		}
	}

	err = tx.Commit(cctx)
	if err != nil {
		return bot.Report(ctx, err)
	}

	return ctx.SendfX("Imported %v mod log entries!", len(entries))
}
