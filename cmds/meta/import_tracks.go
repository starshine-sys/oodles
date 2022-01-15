package meta

import (
	"context"
	"net/http"
	"time"

	"emperror.dev/errors"
	"github.com/jackc/pgx/v4"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/db"
	"gopkg.in/yaml.v3"
)

func (bot *Bot) importTracks(ctx *bcr.Context) (err error) {
	var url string
	if len(ctx.Message.Attachments) > 0 {
		url = ctx.Message.Attachments[0].URL
	} else if ctx.RawArgs != "" {
		url = ctx.RawArgs
	}

	if url == "" {
		return ctx.SendX("You must attach an export file, or give a URL to an export file.")
	}

	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(c, "GET", url, nil)
	if err != nil {
		return bot.Report(ctx, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		bot.SendError("Error downloading track export file: %v", err)
		return ctx.SendX("There was an error downloading the export file.")
	}
	defer resp.Body.Close()

	export := map[string]exportTrack{}
	err = yaml.NewDecoder(resp.Body).Decode(&export)
	if err != nil {
		bot.SendError("Error decoding track export file: %v", err)
		return ctx.SendX("Couldn't decode the export file as a YAML file. Please double check your input!")
	}

	tx, err := bot.DB.Begin(context.Background())
	if err != nil {
		return bot.Report(ctx, err)
	}
	defer func() {
		err = tx.Rollback(context.Background())
		if err != nil && err != pgx.ErrTxClosed {
			bot.SendError("Error rolling back transaction: %v", err)
		}
	}()

	for name, track := range export {
		var trackID int64
		err = tx.QueryRow(context.Background(), "select id from application_tracks where name = $1", name).Scan(&trackID)
		if err != nil {
			// if we got an actual error, bail
			if err != pgx.ErrNoRows {
				return bot.Report(ctx, errors.Wrap(err, "get track ID"))
			}
			// otherwise, create a new track
			t, err := bot.DB.AddApplicationTrack(db.ApplicationTrack{
				Name:        name,
				Description: track.Description,
				RawEmoji:    track.Emoji,
			})
			if err != nil {
				return bot.Report(ctx, errors.Wrap(err, "create app track"))
			}
			trackID = t.ID
		}

		// clear existing questions
		_, err = tx.Exec(context.Background(), "delete from app_questions where track_id = $1", trackID)
		if err != nil {
			return bot.Report(ctx, errors.Wrap(err, "delete existing questions"))
		}

		// add new questions in a loop
		for _, q := range track.Questions {
			_, err = tx.Exec(context.Background(), "insert into app_questions (track_id, question) values ($1, $2)", trackID, q)
			if err != nil {
				return bot.Report(ctx, errors.Wrap(err, "add new question"))
			}
		}
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return bot.Report(ctx, errors.Wrap(err, "commit transaction"))
	}

	return ctx.SendfX("Success, imported %v track(s)!", len(export))
}
