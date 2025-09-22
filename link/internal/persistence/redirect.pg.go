package persistence

import (
	"errors"
	"fmt"

	"github.com/solsteace/go-lib/oops"
	"github.com/solsteace/kochira/link/internal/domain/redirect"
)

func (row pgShorteningRow) toRedirect() redirect.Link {
	return redirect.Link{
		Id:          row.Id,
		Shortened:   row.Shortened,
		Destination: row.Destination,
		IsOpen:      row.IsOpen,
		ExpiredAt:   row.ExpiredAt}
}

func (repo pgLink) GetByShortened(shortened string) (redirect.Link, error) {
	rows := new([]pgShorteningRow)
	query := `
		SELECT * FROM "links"
		WHERE shortened = $1
		LIMIT 1`
	args := []any{shortened}
	if err := repo.db.Select(rows, query, args...); err != nil {
		return redirect.Link{}, fmt.Errorf("repository<pgLink.GetByShortened>: %w", err)
	}

	if len(*rows) != 1 {
		err := oops.NotFound{
			Err: errors.New(
				fmt.Sprintf("link(shortened: %s) not found", shortened))}
		return redirect.Link{}, fmt.Errorf("repository<pgLink.GetByShortened>: %w", err)
	}
	return (*rows)[0].toRedirect(), nil
}
