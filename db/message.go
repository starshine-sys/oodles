package db

import (
	"context"

	"emperror.dev/errors"
	"github.com/Masterminds/squirrel"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/starshine-sys/pkgo/v2"
)

var sq = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

// Message is a single message
type Message struct {
	ID        discord.MessageID
	UserID    discord.UserID
	ChannelID discord.ChannelID
	ServerID  discord.GuildID

	Content  string
	Username string

	// These are only filled if the message was proxied by PluralKit
	Member *string
	System *string
}

// InsertMessage inserts a message
func (db *DB) InsertMessage(m Message) (err error) {
	if m.Content == "" {
		m.Content = "None"
	}

	_, err = db.Exec(context.Background(), `insert into messages
(id, user_id, channel_id, server_id, content, username, member, system) values
($1, $2, $3, $4, $5, $6, $7, $8)
on conflict (id) do update
set content = $5`, m.ID, m.UserID, m.ChannelID, m.ServerID, m.Content, m.Username, m.Member, m.System)
	return err
}

// UpdatePKInfo updates the PluralKit info for the given message, if it exists in the database.
func (db *DB) UpdatePKInfo(msgID discord.MessageID, userID pkgo.Snowflake, system, member string) (err error) {
	sql, args, err := sq.Update("messages").Set("user_id", userID).Set("system", system).Set("member", member).Where(squirrel.Eq{"id": msgID}).ToSql()
	if err != nil {
		return
	}

	_, err = db.Exec(context.Background(), sql, args...)
	return
}

// UpdateUserID updates *just* the user ID for the given message, if it exists in the database.
func (db *DB) UpdateUserID(msgID discord.MessageID, userID discord.UserID) (err error) {
	sql, args, err := sq.Update("messages").Set("user_id", userID).Where(squirrel.Eq{"id": msgID}).ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(context.Background(), sql, args...)
	return
}

// GetMessage gets a single message
func (db *DB) GetMessage(id discord.MessageID) (m *Message, err error) {
	m = &Message{}

	sql, args, err := sq.Select("*").From("messages").Where(squirrel.Eq{"id": id}).ToSql()
	if err != nil {
		return nil, errors.Cause(err)
	}

	err = pgxscan.Get(context.Background(), db, m, sql, args...)
	if err != nil {
		return nil, errors.Cause(err)
	}
	return m, nil
}

// DeleteMessage deletes a message from the database
func (db *DB) DeleteMessage(id discord.MessageID) (err error) {
	_, err = db.Exec(context.Background(), "delete from messages where id = $1", id)
	return
}
