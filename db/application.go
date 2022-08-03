package db

import (
	"context"
	"time"

	"emperror.dev/errors"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/rs/xid"
)

// Application is a user's application.
type Application struct {
	ID        xid.ID
	UserID    discord.UserID
	ChannelID discord.ChannelID

	// Can be null before a track is chosen + if the track is deleted
	TrackID *int64
	// Question index
	Question int

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

	// The scheduled event used to notify when the app times out
	ScheduledEventID *int64
	ScheduledCloseID *int64
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
func (db *DB) CloseApplication(id xid.ID) error {
	_, err := db.Exec(context.Background(), "update applications set closed = true, closed_time = $2 where id = $1", id, time.Now().UTC())
	return err
}

// CreateApplication ...
func (db *DB) CreateApplication(userID discord.UserID, chID discord.ChannelID) (a Application, err error) {
	err = pgxscan.Get(context.Background(), db, &a, "insert into applications (id, user_id, channel_id) values ($1, $2, $3) returning *", xid.New(), userID, chID)
	return a, err
}

// SetTrack ...
func (db *DB) SetTrack(appID xid.ID, trackID int64) error {
	_, err := db.Exec(context.Background(), "update applications set track_id = $1, question = 0 where id = $2", trackID, appID)
	return err
}

func (db *DB) ResetApplication(appID xid.ID) error {
	_, err := db.Exec(context.Background(), "update applications set track_id = null, question = 0 where id = $1", appID)
	return err
}

// SetQuestionIndex ...
func (db *DB) SetQuestionIndex(appID xid.ID, index int) error {
	_, err := db.Exec(context.Background(), "update applications set question = $1 where id = $2", index, appID)
	return err
}

// CompleteApp ...
func (db *DB) CompleteApp(appID xid.ID) error {
	_, err := db.Exec(context.Background(), "update applications set completed = true where id = $1", appID)
	return err
}

// SetTranscript ...
func (db *DB) SetTranscript(appID xid.ID, chID discord.ChannelID, msgID discord.MessageID) error {
	_, err := db.Exec(context.Background(), "update applications set transcript_channel = $1, transcript_message = $2 where id = $3", chID, msgID, appID)
	return err
}

func (db *DB) SetEventID(appID xid.ID, eventID int64) error {
	_, err := db.Exec(context.Background(), "update applications set scheduled_event_id = $1 where id = $2", eventID, appID)
	return err
}

func (db *DB) SetCloseID(appID xid.ID, eventID int64) error {
	_, err := db.Exec(context.Background(), "update applications set scheduled_close_id = $1 where id = $2", eventID, appID)
	return err
}

// SetVerified ...
func (db *DB) SetVerified(appID xid.ID, mod discord.UserID, verified bool, denyReason *string) error {
	_, err := db.Exec(context.Background(), "update applications set moderator = $1, verified = $2, deny_reason = $3 where id = $4", mod, verified, denyReason, appID)
	return err
}

// AppResponse ...
type AppResponse struct {
	ApplicationID xid.ID
	MessageID     discord.MessageID
	UserID        discord.UserID
	Username      string
	Discriminator string
	Content       string

	FromBot   bool
	FromStaff bool
}

// AddResponse adds or updates a response to an application.
func (db *DB) AddResponse(appID xid.ID, resp AppResponse) error {
	_, err := db.Exec(context.Background(), "insert into app_responses (application_id, message_id, user_id, username, discriminator, content, from_bot, from_staff) values ($1, $2, $3, $4, $5, $6, $7, $8) on conflict (message_id) do update set content = $6", appID, resp.MessageID, resp.UserID, resp.Username, resp.Discriminator, resp.Content, resp.FromBot, resp.FromStaff)
	return err
}
