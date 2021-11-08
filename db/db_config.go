package db

import (
	"context"

	"github.com/starshine-sys/oodles/common"
)

func (db *DB) createConfig() error {
	ct, err := db.Exec(context.Background(), "insert into guilds (id, config, commands, perms) values ($1, $2, $3, $4) on conflict do nothing", db.BotConfig.GuildID, db.Config, db.Overrides, db.Perms)
	if err != nil {
		return err
	}

	if ct.RowsAffected() > 0 {
		common.Log.Info("Initialized configuration")
	}
	return nil
}

func (db *DB) fetchConfig() error {
	err := db.QueryRow(context.Background(), "select config, commands, perms from guilds where id = $1", db.BotConfig.GuildID).Scan(&db.Config, &db.Overrides, &db.Perms)
	if err != nil {
		return err
	}

	db.Perms.BotOwners = db.BotConfig.Owners
	return nil
}

// SyncConfig synchronizes configuration with the database.
func (db *DB) SyncConfig() error {
	return db.QueryRow(context.Background(), "update guilds set config = $1 where id = $2 returning config", db.Config, db.BotConfig.GuildID).Scan(&db.Config)
}

// SyncPerms synchronizes role/user permissions with the database.
func (db *DB) SyncPerms() error {
	return db.QueryRow(context.Background(), "update guilds set perms = $1 where id = $2 returning perms", db.Perms, db.BotConfig.GuildID).Scan(&db.Perms)
}

// SyncOverrides synchronizes command overrides with the database.
func (db *DB) SyncOverrides() error {
	return db.QueryRow(context.Background(), "update guilds set commands = $1 where id = $2 returning commands", db.Overrides, db.BotConfig.GuildID).Scan(&db.Overrides)
}
