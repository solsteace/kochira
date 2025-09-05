package repository

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/solsteace/kochira/link/internal/domain"
)

type pgLinkRow struct {
	Id          uint      `db:"id"`
	UserId      uint      `db:"user_id"`
	Shortened   string    `db:"shortened"`
	Destination string    `db:"destination"`
	IsOpen      bool      `db:"is_open"`
	UpdatedAt   time.Time `db:"updated_at"`
	ExpiredAt   time.Time `db:"expired_at"`
}

func (row pgLinkRow) toLink() (domain.Link, error) {
	return domain.NewLink(
		&row.Id,
		row.UserId,
		row.Shortened,
		row.Destination,
		row.IsOpen,
		row.UpdatedAt,
		row.ExpiredAt)
}

type pgLink struct {
	db *sqlx.DB
}

func NewPgLink(db *sqlx.DB) pgLink {
	return pgLink{db}
}

func (repo pgLink) GetMany(q linkQueryParams) ([]domain.Link, error) {
	query := `
		SELECT * 
		FROM "links"
		LIMIT ?
		OFFSET ?`
	args := []any{
		q.limit,
		q.Offset()}

	rows := new([]pgLinkRow)
	if err := repo.db.Select(rows, query, args); err != nil {
		return []domain.Link{}, err
	}

	links := []domain.Link{}
	for _, r := range *rows {
		l, err := r.toLink()
		if err != nil {
			return []domain.Link{}, err
		}
		links = append(links, l)
	}
	return links, nil
}

func (repo pgLink) GetById(id uint) (domain.Link, error) {
	query := `
		SELECT *
		FROM "links"
		WHERE id = ? `
	args := []any{id}

	row := new(pgLinkRow)
	if err := repo.db.Select(row, query, args); err != nil {
		return domain.Link{}, err
	}
	return row.toLink()
}

func (repo pgLink) GetByShortened(shortened string) (domain.Link, error) {
	query := `
		SELECT *
		FROM "links"
		WHERE shortened = ?`
	args := []any{shortened}

	row := new(pgLinkRow)
	if err := repo.db.Select(row, query, args); err != nil {
		return domain.Link{}, err
	}
	return row.toLink()
}

func (repo pgLink) Create(l domain.Link) error {
	query := `
		INSERT INTO "links"(
			user_id,
			shortened,
			destination,
			is_open,
			updated_at,
			expired_at)
		VALUES (
			:userId, 
			:shortened, 
			:destination, 
			:is_open, 
			:updatedAt, 
			:expiredAt)`

	_, err := repo.db.Exec(query, l)
	if err != nil {
		return err
	}
	return nil
}

func (repo pgLink) Update(l domain.Link) error {
	query := `
		UPDATE "links"
		SET 
			shortened = :shortened,
			destination = :destination,
			is_open = :isOpen,
			updated_at = :updatedAt,
			expired_at = :expiredAt
		WHERE
			id = :id`

	_, err := repo.db.Exec(query, l)
	if err != nil {
		return err
	}
	return nil
}

func (repo pgLink) DeleteById(id uint) error {
	query := `
		DELETE 
		FROM "links"
		WHERE id = ?`
	args := []any{id}

	_, err := repo.db.Exec(query, args)
	if err != nil {
		return err
	}
	return nil
}
