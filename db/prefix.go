package db

import (
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/bcr"
)

var _ bcr.Prefixer = (*DB)(nil).Prefixer

// Prefixer ...
func (db *DB) Prefixer(m discord.Message) int {
	p := db.Config.Get("prefix").ToString()

	if strings.HasPrefix(strings.ToLower(m.Content), strings.ToLower(p)) {
		return len(p)
	}
	return -1
}
