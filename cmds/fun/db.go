package fun

import (
	"context"

	"emperror.dev/errors"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/georgysavva/scany/pgxscan"
)

type valid struct {
	ID       int
	Response string
	UserID   discord.UserID
}

func (bot *Bot) GetValids() (vs []valid, err error) {
	err = pgxscan.Select(context.Background(), bot.DB, &vs, "select * from valid_responses")
	return vs, errors.Cause(err)
}

func (bot *Bot) GetValid(id int) (v valid, err error) {
	err = pgxscan.Get(context.Background(), bot.DB, &v, "select * from valid_responses where id = $1", id)
	return v, errors.Cause(err)
}
