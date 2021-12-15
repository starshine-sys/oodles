package meta

import (
	"strings"

	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/db"
)

func (bot *Bot) overrideCmdPerms(ctx *bcr.Context) (err error) {
	var level db.PermissionLevel
	switch strings.ToLower(ctx.Args[0]) {
	case "user":
		level = db.UserLevel
	case "helper":
		level = db.HelperLevel
	case "staff":
		level = db.StaffLevel
	case "owner":
		level = db.OwnerLevel
	default:
		return ctx.SendfX("``%v`` is not a valid permission level.", bcr.EscapeBackticks(ctx.Args[0]))
	}

	db.OverridesMu.RLock()
	_, ok := db.DefaultPermissions[strings.ToLower(ctx.Args[1])]
	db.OverridesMu.RUnlock()

	if !ok {
		return ctx.SendfX("``%v`` is not a valid root-level command.", bcr.EscapeBackticks(ctx.Args[1]))
	}

	db.OverridesMu.Lock()
	bot.DB.Overrides[strings.ToLower(ctx.Args[1])] = level
	defer db.OverridesMu.Unlock()

	err = bot.DB.SyncOverrides()
	if err != nil {
		return bot.Report(ctx, err)
	}

	return ctx.SendfX("Updated permission level for ``%v`` to `%s`!", bcr.EscapeBackticks(strings.ToLower(ctx.Args[1])), level)
}
