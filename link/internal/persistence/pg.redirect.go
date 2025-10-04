package persistence

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/solsteace/go-lib/oops"
	"github.com/solsteace/kochira/link/internal/domain/redirect"
)

func (row pgLink) toRedirect() redirect.Link {
	return redirect.Link{
		Id:          row.Id,
		Shortened:   row.Shortened,
		Destination: row.Destination,
		IsOpen:      row.IsOpen,
		ExpiredAt:   row.ExpiredAt}
}

func (repo pg) GetByAlias(alias string) (redirect.Link, error) {
	row := new(pgLink)
	query := `SELECT * FROM "links" WHERE alias = $1 LIMIT 1`
	args := []any{alias}
	if err := repo.db.Get(row, query, args...); err != nil {
		l := redirect.Link{}
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return l, fmt.Errorf(
				"persistence<pgLink.GetByAlias>: %w",
				oops.NotFound{
					Err: err,
					Msg: fmt.Sprintf("link(shortened:%s) not found", alias)})
		default:
			return l, fmt.Errorf("persistence<pgLink.GetByAlias>: %w", err)
		}
	}

	return row.toRedirect(), nil
}
