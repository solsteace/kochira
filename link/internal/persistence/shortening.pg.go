package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/solsteace/go-lib/oops"
	"github.com/solsteace/kochira/link/internal/domain/shortening"
)

type ShorteningQueryParams struct {
	page  uint
	limit uint
}

func (param ShorteningQueryParams) Offset() uint {
	if param.page < 1 {
		return 0
	}
	return (param.page - 1) * param.limit
}

func NewShorteningQueryParams(page, limit *uint) ShorteningQueryParams {
	var actualPage uint = 1
	if page != nil && *page > 0 {
		actualPage = *page
	}

	var actualLimit uint = 10 // DEFAULT
	if limit != nil && *limit > 0 {
		actualLimit = *limit
	}

	return ShorteningQueryParams{
		actualPage,
		actualLimit}
}

type pgLink struct {
	db *sqlx.DB
}

func NewPgLink(db *sqlx.DB) pgLink {
	return pgLink{db}
}

type pgShorteningRow struct {
	Id          uint64    `db:"id"`
	UserId      uint64    `db:"user_id"`
	Shortened   string    `db:"shortened"`
	Destination string    `db:"destination"`
	IsOpen      bool      `db:"is_open"`
	UpdatedAt   time.Time `db:"updated_at"`
	ExpiredAt   time.Time `db:"expired_at"`
}

func (row pgShorteningRow) toShortening() (shortening.Link, error) {
	return shortening.NewLink(
		&row.Id,
		row.UserId,
		row.Shortened,
		row.Destination,
		row.IsOpen,
		row.UpdatedAt,
		row.ExpiredAt)
}

func newPgshorteningLinkRow(l shortening.Link) pgShorteningRow {
	return pgShorteningRow{
		Id:          l.Id(),
		UserId:      l.UserId(),
		Shortened:   l.Shortened(),
		Destination: l.Destination(),
		IsOpen:      l.IsOpen(),
		UpdatedAt:   l.UpdatedAt(),
		ExpiredAt:   l.ExpiredAt()}
}

func (repo pgLink) GetMany(q ShorteningQueryParams) ([]shortening.Link, error) {
	rows := new([]pgShorteningRow)
	query := `SELECT * FROM "links" LIMIT $1 OFFSET $2`
	args := []any{q.limit, q.Offset()}
	if err := repo.db.Select(rows, query, args...); err != nil {
		return []shortening.Link{}, fmt.Errorf("repository<pgLink.GetMany>: %w", err)
	}

	links := []shortening.Link{}
	for _, r := range *rows {
		link, err := r.toShortening()
		if err != nil {
			return []shortening.Link{}, fmt.Errorf("repository<pgLink.GetMany>: %w", err)
		}
		links = append(links, link)
	}
	return links, nil
}

func (repo pgLink) GetManyByUser(userId uint64, q ShorteningQueryParams) ([]shortening.Link, error) {
	rows := new([]pgShorteningRow)
	query := `SELECT * FROM "links" WHERE user_id = $1 LIMIT $2 OFFSET $3`
	args := []any{userId, q.limit, q.Offset()}
	if err := repo.db.Select(rows, query, args...); err != nil {
		return []shortening.Link{}, fmt.Errorf("repository<pgLink.GetManyByUser>: %w", err)
	}

	links := []shortening.Link{}
	for _, r := range *rows {
		link, err := r.toShortening()
		if err != nil {
			return []shortening.Link{}, fmt.Errorf("repository<pgLink.GetManyByUser>: %w", err)
		}
		links = append(links, link)
	}
	return links, nil
}

func (repo pgLink) GetById(id uint64) (shortening.Link, error) {
	row := new(pgShorteningRow)
	query := `SELECT * FROM links WHERE id = $1 LIMIT 1`
	args := []any{id}
	if err := repo.db.Get(row, query, args...); err != nil {
		l := shortening.Link{}
		switch {
		case errors.Is(err, sql.ErrNoRows):
			err2 := oops.NotFound{
				Err: err,
				Msg: fmt.Sprintf("link(id:%d) not found", id)}
			return l, fmt.Errorf("persistence<pgLink.GetById>: %w", err2)
		default:
			return l, fmt.Errorf("persistence<pgLink.GetById>: %w", err)
		}
	}

	s, err := row.toShortening()
	if err != nil {
		return shortening.Link{}, fmt.Errorf("persistence<pgLink.GetById>: %w", err)
	}
	return s, nil

}

func (repo pgLink) Create(l shortening.Link) (uint64, error) {
	query := `
		INSERT INTO "links"(
			user_id,
			shortened,
			destination,
			is_open,
			updated_at,
			expired_at)
		VALUES (
			:user_id, 
			:shortened, 
			:destination, 
			:is_open, 
			:updated_at, 
			:expired_at)
		RETURNING id`
	link := newPgshorteningLinkRow(l)
	row, err := repo.db.NamedQuery(query, link)
	if err != nil {
		return 0, fmt.Errorf("repository<pgLink.Create>: %w", err)
	}

	row.Next()
	defer row.Close()
	var insertId uint64 = 0
	if err := row.Scan(&insertId); err != nil {
		return 0, fmt.Errorf("repository<pgLink.Create>: %w", err)
	}

	return insertId, nil
}

func (repo pgLink) Update(l shortening.Link) error {
	query := `
		UPDATE "links"
		SET 
			shortened = :shortened,
			destination = :destination,
			is_open = :is_open,
			updated_at = :updated_at,
			expired_at = :expired_at
		WHERE
			id = :id`
	link := newPgshorteningLinkRow(l)
	_, err := repo.db.NamedExec(query, link)
	if err != nil {
		return fmt.Errorf("repository<pgLink.Update>: %w", err)
	}
	return nil
}

func (repo pgLink) DeleteById(id uint64) error {
	query := `DELETE FROM "links" WHERE id = $1`
	args := []any{id}
	_, err := repo.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("repository<pgLink.DeleteById>: %w", err)
	}
	return nil
}

func (repo pgLink) CountByUserId(userId uint64) (uint, error) {
	query := `SELECT COUNT(*) as n_links FROM links WHERE user_id = ?`
	args := []any{userId}
	result := repo.db.QueryRow(query, args)
	if result.Err() != nil {
		return 0, fmt.Errorf("repository<pgLink.CountByUserId>: %w", result.Err())
	}

	var count uint = 0
	if err := result.Scan(count); err != nil {
		return 0, fmt.Errorf("repository<pgLink.CountByUserId>: %w", err)
	}
	return count, nil
}
