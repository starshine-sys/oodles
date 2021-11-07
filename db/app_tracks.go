package db

import (
	"context"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/georgysavva/scany/pgxscan"
)

// ApplicationTrack is a single list of questions and associated name.
type ApplicationTrack struct {
	ID          int64
	Name        string
	Description string
	// RawEmoji is stored as a?:name:id
	RawEmoji string `db:"emoji"`
}

// Emoji returns the emoji struct version of the stored emoji.
func (t ApplicationTrack) Emoji() discord.Emoji {
	parts := strings.Split(strings.Trim(t.RawEmoji, "<>"), ":")

	switch len(parts) {
	case 0:
		panic("ApplicationTrack.RawEmoji is empty")
	case 1:
		return discord.Emoji{
			ID:   0,
			Name: t.RawEmoji,
		}
	case 2:
		sf, _ := discord.ParseSnowflake(parts[1])

		return discord.Emoji{
			ID:       discord.EmojiID(sf),
			Name:     parts[0],
			Animated: false,
		}
	default:
		sf, _ := discord.ParseSnowflake(parts[2])

		return discord.Emoji{
			ID:       discord.EmojiID(sf),
			Name:     parts[1],
			Animated: true,
		}
	}
}

// ApplicationTracks returns all application tracks sorted by ID.
func (db *DB) ApplicationTracks() (tracks []ApplicationTrack, err error) {
	err = pgxscan.Select(context.Background(), db, &tracks, "select * from application_tracks order by id")
	return tracks, err
}

// AddApplicationTrack adds an application track to the database.
func (db *DB) AddApplicationTrack(t ApplicationTrack) (*ApplicationTrack, error) {
	err := db.QueryRow(context.Background(), "insert into application_tracks (name, emoji, description) values ($1, $2, $3) returning id", t.Name, t.RawEmoji, t.Description).Scan(&t.ID)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// UpdateApplicationTrack updates the given application track.
func (db *DB) UpdateApplicationTrack(t ApplicationTrack) error {
	_, err := db.Exec(context.Background(), "update application_tracks set name = $1, emoji = $2, description = $3 where id = $4", t.Name, t.RawEmoji, t.Description, t.ID)
	return err
}

// AppQuestion is a single application question.
type AppQuestion struct {
	Index    int64
	TrackID  int64
	ID       int64
	Question string
}

// Questions gets all questions for an interview track.
func (db *DB) Questions(id int64) (qs []AppQuestion, err error) {
	err = pgxscan.Select(context.Background(), db, &qs, `select
	row_number() over (order by id) as index,
	track_id, id, question
	from app_questions where track_id = $1
	order by index asc`, id)
	return qs, err
}

// AddQuestion ...
func (db *DB) AddQuestion(trackID int64, question string) (err error) {
	_, err = db.Exec(context.Background(), "insert into app_questions (track_id, question) values ($1, $2)", trackID, question)
	return
}
