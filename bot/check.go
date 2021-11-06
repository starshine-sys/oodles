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
		return c.DB.Overrides.For(v.FullCommandPath[0]).String()
	default:
		if strings.EqualFold(v.FullCommandPath[0], "help") {
			return c.DB.Overrides.For(v.FullCommandPath[1]).String()
		}

		return c.DB.Overrides.For(v.FullCommandPath[0]).String()
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

	required := c.DB.Overrides.For(v.FullCommandPath[0])

	lvl := c.DB.Perms.Level(m)

	if required > lvl {
		return false, nil
	}

	return true, nil
}
