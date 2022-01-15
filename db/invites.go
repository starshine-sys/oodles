package db

import (
	"context"

	"emperror.dev/errors"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
)

type invite struct {
	Code string
	Name string
}

func (db *DB) AllInvites() (invs map[string]string, err error) {
	var slice []invite
	err = pgxscan.Select(context.Background(), db, &slice, "select * from invites")
	if err != nil {
		return nil, errors.Cause(err)
	}

	invs = make(map[string]string, len(slice))
	for _, i := range slice {
		invs[i.Code] = i.Name
	}
	return invs, nil
}

func (db *DB) SetInviteName(code, name string) error {
	ct, err := db.Exec(context.Background(), "insert into invites (code, name) values ($1, $2) on conflict (code) do update set name = $2", code, name)
	if err != nil {
		return errors.Cause(err)
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (db *DB) ClearInviteName(code string) error {
	ct, err := db.Exec(context.Background(), "delete from invites where code = $1", code)
	if err != nil {
		return errors.Cause(err)
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
