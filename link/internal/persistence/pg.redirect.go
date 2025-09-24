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

func (repo pg) GetByShortened(shortened string) (redirect.Link, error) {
	row := new(pgLink)
	query := `SELECT * FROM "links" WHERE shortened = $1 LIMIT 1`
	args := []any{shortened}
	if err := repo.db.Get(row, query, args...); err != nil {
		l := redirect.Link{}
		switch {
		case errors.Is(err, sql.ErrNoRows):
			err2 := oops.NotFound{
				Err: err,
				Msg: fmt.Sprintf("link(shortened:%d) not found", shortened)}
			return l, fmt.Errorf("persistence<pgLink.GetByShortened>: %w", err2)
		default:
			return l, fmt.Errorf("persistence<pgLink.GetByShortened>: %w", err)
		}
	}

	return row.toRedirect(), nil
}
