package db

import (
	"context"
	"time"

	"emperror.dev/errors"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/georgysavva/scany/pgxscan"
)

// Application is a user's application.
type Application struct {
	ID        int64
	UserID    discord.UserID
	ChannelID discord.ChannelID

	// Can be null before a track is chosen + if the track is deleted
	TrackID *int64
	// Question index
	Question int

	// When the application was opened
	Opened time.Time
	// Whether the user has completed the automated interview section
	Completed bool
	// Whether the user was verified--null if not decided yet
	Verified *bool
	// If Verified is false, the reason why the user was denied
	DenyReason *string
	// Moderator who verified or denied the user
	Moderator *discord.UserID

	// Whether the interview has been closed (channel deleted)
	Closed     bool
	ClosedTime *time.Time

	TranscriptChannel *discord.ChannelID
	TranscriptMessage *discord.MessageID
}

// AllUserApplications returns all of this user's applications, sorted by ID descending.
func (db *DB) AllUserApplications(userID discord.UserID) (as []Application, err error) {
	err = pgxscan.Select(context.Background(), db, &as, "select * from applications where user_id = $1 order by id desc", userID)
	if err != nil {
		return nil, errors.Cause(err)
	}
	return as, nil
}

// UserApplication returns an open application for the given user.
func (db *DB) UserApplication(userID discord.UserID) (*Application, error) {
	var a Application
	err := pgxscan.Get(context.Background(), db, &a, "select * from applications where user_id = $1 and closed = false order by id desc limit 1", userID)
	if err != nil {
		return nil, errors.Cause(err)
	}
	return &a, nil
}

// ChannelApplication ...
func (db *DB) ChannelApplication(chID discord.ChannelID) (*Application, error) {
	var a Application
	err := pgxscan.Get(context.Background(), db, &a, "select * from applications where channel_id = $1 and closed = false order by id desc limit 1", chID)
	if err != nil {
		return nil, errors.Cause(err)
	}
	return &a, nil
}

// CloseApplication closes the given application.
func (db *DB) CloseApplication(id int64) error {
	_, err := db.Exec(context.Background(), "update applications set closed = true, closed_time = $2 where id = $1", id, time.Now().UTC())
	return err
}

// CreateApplication ...
func (db *DB) CreateApplication(userID discord.UserID, chID discord.ChannelID) error {
	_, err := db.Exec(context.Background(), "insert into applications (user_id, channel_id) values ($1, $2)", userID, chID)
	return err
}

// SetTrack ...
func (db *DB) SetTrack(appID int64, trackID int64) error {
	_, err := db.Exec(context.Background(), "update applications set track_id = $1, question = 0 where id = $2", trackID, appID)
	return err
}

// SetQuestionIndex ...
func (db *DB) SetQuestionIndex(appID int64, index int) error {
	_, err := db.Exec(context.Background(), "update applications set question = $1 where id = $2", index, appID)
	return err
}

// CompleteApp ...
func (db *DB) CompleteApp(appID int64) error {
	_, err := db.Exec(context.Background(), "update applications set completed = true where id = $1", appID)
	return err
}

// SetTranscript ...
func (db *DB) SetTranscript(appID int64, chID discord.ChannelID, msgID discord.MessageID) error {
	_, err := db.Exec(context.Background(), "update applications set transcript_channel = $1, transcript_message = $2 where id = $3", chID, msgID, appID)
	return err
}

// SetVerified ...
func (db *DB) SetVerified(appID int64, mod discord.UserID, verified bool, denyReason *string) error {
	_, err := db.Exec(context.Background(), "update applications set moderator = $1, verified = $2, deny_reason = $3 where id = $4", mod, verified, denyReason, appID)
	return err
}

// AppResponse ...
type AppResponse struct {
	ApplicationID int64
	MessageID     discord.MessageID
	UserID        discord.UserID
	Username      string
	Discriminator string
	Content       string

	FromBot   bool
	FromStaff bool
}

// AddResponse adds or updates a response to an application.
func (db *DB) AddResponse(appID int64, resp AppResponse) error {
	_, err := db.Exec(context.Background(), "insert into app_responses (application_id, message_id, user_id, username, discriminator, content, from_bot, from_staff) values ($1, $2, $3, $4, $5, $6, $7, $8) on conflict (message_id) do update set content = $6", appID, resp.MessageID, resp.UserID, resp.Username, resp.Discriminator, resp.Content, resp.FromBot, resp.FromStaff)
	return err
}
