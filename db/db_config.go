package db

import (
	"context"

	"github.com/starshine-sys/oodles/common"
)

func (db *DB) createConfig() error {
	ct, err := db.Exec(context.Background(), "insert into guilds (id, config, commands, perms) on conflict do nothing", db.BotConfig.GuildID, db.Config, db.Overrides, db.Perms)
	if err != nil {
		return err
	}

	if ct.RowsAffected() > 0 {
		common.Log.Info("Initialized configuration")
	}
	return nil
}

func (db *DB) fetchConfig() error {
	return db.QueryRow(context.Background(), "select config, commands, perms from guilds where id = $1", db.BotConfig.GuildID).Scan(&db.Config, &db.Overrides, &db.Perms)
}
