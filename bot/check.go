package bot

import (
	"errors"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/db"
)

// Checker ...
type Checker struct {
	*db.DB
	*Bot
}

func (c Checker) String(ctx bcr.Contexter) string {
	v, ok := ctx.(*bcr.Context)
	if !ok {
		return "slash command (implement me for those!)"
	}

	switch len(v.FullCommandPath) {
	case 0:
		return "shouldn't get here"
	case 1:
		cmd := c.Router.GetCommand(v.FullCommandPath[0])
		if cmd == nil {
			return db.DisabledLevel.String()
		}

		return c.DB.Overrides.For(cmd.Name).String()
	default:
		if strings.EqualFold(v.FullCommandPath[0], "help") {
			cmd := c.Router.GetCommand(v.FullCommandPath[1])
			if cmd == nil {
				return db.DisabledLevel.String()
			}

			return c.DB.Overrides.For(cmd.Name).String()
		}

		cmd := c.Router.GetCommand(v.FullCommandPath[0])
		if cmd == nil {
			return db.DisabledLevel.String()
		}

		return c.DB.Overrides.For(cmd.Name).String()
	}
}

// Check checks permissions!
func (c *Checker) Check(ctx bcr.Contexter) (bool, error) {
	v, ok := ctx.(*bcr.Context)
	if !ok {
		return false, errors.New("slash command (implement me for those!)")
	}

	m := v.Member
	if m == nil {
		m = &discord.Member{
			User: v.Author,
		}
	}

	if len(v.FullCommandPath) == 0 {
		return false, errors.New("shouldn't get here")
	}

	var required db.PermissionLevel
	cmd := c.Router.GetCommand(v.FullCommandPath[0])
	if cmd == nil {
		required = db.InvalidLevel
	} else {
		required = c.DB.Overrides.For(cmd.Name)
	}

	lvl := c.DB.Perms.Level(m)

	if required > lvl {
		return false, nil
	}

	return true, nil
}
