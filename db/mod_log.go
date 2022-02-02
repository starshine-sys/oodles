package db

import (
	"context"
	"fmt"
	"time"

	"emperror.dev/errors"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/starshine-sys/bcr"
)

type ModLogEntry struct {
	ID      int64
	GuildID discord.GuildID

	UserID discord.UserID
	ModID  discord.UserID

	ActionType string
	Reason     string

	Time time.Time

	ChannelID *discord.ChannelID
	MessageID *discord.MessageID
}

func (m ModLogEntry) Embed(s *state.State) discord.Embed {
	e := discord.Embed{
		Title: fmt.Sprintf("%v | case %v", m.ActionType, m.ID),
		Description: fmt.Sprintf(
			`**Offender:** %v
**Moderator:** %v
**Reason:** %v`,
			userRepr(s, m.GuildID, m.UserID),
			userRepr(s, m.GuildID, m.ModID),
			m.Reason,
		),
		Footer: &discord.EmbedFooter{
			Text: fmt.Sprintf("User ID: %v", m.UserID),
		},
		Timestamp: discord.NewTimestamp(m.Time),
	}

	switch m.ActionType {
	case "warn", "mute", "hardmute":
		e.Color = bcr.ColourOrange
	case "kick", "ban":
		e.Color = bcr.ColourRed
	case "unban", "unmute":
		e.Color = bcr.ColourGreen
	default:
		e.Color = bcr.ColourBlurple
	}

	return e
}

func userRepr(s *state.State, guildID discord.GuildID, userID discord.UserID) string {
	member, err := s.MemberStore.Member(guildID, userID)
	if err == nil {
		return fmt.Sprintf("%v %v", member.User.Tag(), member.Mention())
	}

	u, err := s.User(userID)
	if err == nil {
		return fmt.Sprintf("%v %v", u.Tag(), u.Mention())
	}

	return fmt.Sprintf("%v (%v)", userID.Mention(), userID)
}

func (db *DB) InsertModLog(ctx context.Context, entry ModLogEntry) (ModLogEntry, error) {
	if entry.Time.IsZero() {
		entry.Time = time.Now()
	}

	err := pgxscan.Get(ctx, db, &entry, `insert into mod_log
	(guild_id, user_id, mod_id, action_type, reason, time)
	values ($1, $2, $3, $4, $5, $6)
	returning *`, entry.GuildID, entry.UserID, entry.ModID, entry.ActionType, entry.Reason, entry.Time.UTC())
	return entry, errors.Cause(err)
}

func (db *DB) UpdateModLogReason(id int64, reason string) (e ModLogEntry, err error) {
	err = pgxscan.Get(context.Background(), db, &e, "update mod_log set reason = $1 where id = $2 returning *", reason, id)
	return e, errors.Cause(err)
}

func (db *DB) UpdateModLogMessage(id int64, chID discord.ChannelID, msgID discord.MessageID) (e ModLogEntry, err error) {
	err = pgxscan.Get(context.Background(), db, &e, "update mod_log set channel_id = $1, message_id = $2 where id = $3 returning *", chID, msgID, id)
	return e, errors.Cause(err)
}

func (db *DB) ModLogFor(guildID discord.GuildID, userID discord.UserID) (es []ModLogEntry, err error) {
	err = pgxscan.Select(context.Background(), db, &es, "select * from mod_log where guild_id = $1 and user_id = $2 order by time desc", guildID, userID)
	return es, errors.Cause(err)
}
