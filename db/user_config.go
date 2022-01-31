package db

import (
	"context"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/oodles/common"
)

func (db *DB) ensureUserConfigExists(ctx context.Context, userID discord.UserID) {
	_, _ = db.Exec(ctx, "insert into users (id) values ($1) on conflict (id) do nothing", userID)
}

func (db *DB) UserStringSet(userID discord.UserID, key, val string) error {
	db.ensureUserConfigExists(context.Background(), userID)

	_, err := db.Exec(context.Background(), "update users set config = coalesce(config, ''::hstore) || hstore($1, $2) where id = $3", key, val, userID)
	return err
}

func (db *DB) UserStringGet(ctx context.Context, userID discord.UserID, key string) (string, error) {
	db.ensureUserConfigExists(ctx, userID)

	var val string
	err := db.QueryRow(ctx, "select coalesce(config->$1, '') from users where id = $2", key, userID).Scan(&val)
	return val, err
}

func (db *DB) UserTime(ctx context.Context, userID discord.UserID) *time.Location {
	name, err := db.UserStringGet(ctx, userID, "timezone")
	if err != nil {
		common.Log.Errorf("Error getting user timezone: %v", err)
	}

	if name == "" {
		name = "UTC"
	}

	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return loc
}
