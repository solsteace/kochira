package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/solsteace/go-lib/oops"
	"github.com/solsteace/kochira/link/internal/domain"
	"github.com/solsteace/kochira/link/internal/view"
)

type pgLink struct {
	db *sqlx.DB
}

func NewPgLink(db *sqlx.DB) pgLink {
	return pgLink{db}
}

// ========================================
// Queries
// ========================================

type pgViewLinkRow struct {
	Id          uint64    `db:"id"`
	UserId      uint64    `db:"user_id"`
	Shortened   string    `db:"shortened"`
	Destination string    `db:"destination"`
	IsOpen      bool      `db:"is_open"`
	UpdatedAt   time.Time `db:"updated_at"`
	ExpiredAt   time.Time `db:"expired_at"`
}

func (row pgViewLinkRow) toView() view.Link {
	return view.Link{
		Id:          row.Id,
		UserId:      row.UserId,
		Shortened:   row.Shortened,
		Destination: row.Destination,
		IsOpen:      row.IsOpen,
		UpdatedAt:   row.UpdatedAt,
		ExpiredAt:   row.ExpiredAt}
}

func newPgViewLinkRow(l domain.Link) pgViewLinkRow {
	return pgViewLinkRow{
		Id:          l.Id(),
		UserId:      l.UserId(),
		Shortened:   l.Shortened(),
		Destination: l.Destination(),
		IsOpen:      l.IsOpen(),
		UpdatedAt:   l.UpdatedAt(),
		ExpiredAt:   l.ExpiredAt()}
}

func (repo pgLink) GetMany(q linkQueryParams) ([]view.Link, error) {
	rows := new([]pgViewLinkRow)
	query := `
		SELECT * 
		FROM "links" 
		LIMIT $1 
		OFFSET $2`
	args := []any{
		q.limit,
		q.Offset()}
	if err := repo.db.Select(rows, query, args...); err != nil {
		return []view.Link{}, fmt.Errorf("repository<pgLink.GetMany>: %w", err)
	}

	links := []view.Link{}
	for _, r := range *rows {
		links = append(links, r.toView())
	}
	return links, nil
}

func (repo pgLink) GetManyByUser(userId uint64, q linkQueryParams) ([]view.Link, error) {
	rows := new([]pgViewLinkRow)
	query := `
		SELECT * 
		FROM "links" 
		WHERE user_id = $1
		LIMIT $2
		OFFSET $3`
	args := []any{userId, q.limit, q.Offset()}
	if err := repo.db.Select(rows, query, args...); err != nil {
		return []view.Link{}, fmt.Errorf("repository<pgLink.GetManyByUser>: %w", err)
	}

	links := []view.Link{}
	for _, r := range *rows {
		links = append(links, r.toView())
	}
	return links, nil
}

func (repo pgLink) GetById(id uint64) (view.Link, error) {
	rows := new([]pgViewLinkRow)
	query := `
		SELECT * 
		FROM links 
		WHERE id = $1 
		LIMIT 1`
	args := []any{id}
	if err := repo.db.Select(rows, query, args...); err != nil {
		return view.Link{}, fmt.Errorf("repository<pgLink.GetById>: %w", err)
	}

	if len(*rows) != 1 {
		err := oops.NotFound{
			Err: errors.New(
				fmt.Sprintf("link(id: %d) not found", id))}
		return view.Link{}, fmt.Errorf("repository<pgLink.GetById>: %w", err)
	}
	return (*rows)[0].toView(), nil
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

func (repo pgLink) GetByShortened(shortened string) (view.Link, error) {
	rows := new([]pgViewLinkRow)
	query := `
		SELECT * FROM "links"
		WHERE shortened = $1
		LIMIT 1`
	args := []any{shortened}
	if err := repo.db.Select(rows, query, args...); err != nil {
		return view.Link{}, fmt.Errorf("repository<pgLink.GetByShortened>: %w", err)
	}

	if len(*rows) != 1 {
		err := oops.NotFound{
			Err: errors.New(
				fmt.Sprintf("link(shortened: %s) not found", shortened))}
		return view.Link{}, fmt.Errorf("repository<pgLink.GetByShortened>: %w", err)
	}
	return (*rows)[0].toView(), nil
}

// ========================================
// COMMANDS
// ========================================
type pgDomainLinkRow struct {
	Id          uint64    `db:"id"`
	UserId      uint64    `db:"user_id"`
	Shortened   string    `db:"shortened"`
	Destination string    `db:"destination"`
	IsOpen      bool      `db:"is_open"`
	UpdatedAt   time.Time `db:"updated_at"`
	ExpiredAt   time.Time `db:"expired_at"`
}

func (row pgDomainLinkRow) toLink() (domain.Link, error) {
	return domain.NewLink(
		&row.Id,
		row.UserId,
		row.Shortened,
		row.Destination,
		row.IsOpen,
		row.UpdatedAt,
		row.ExpiredAt)
}

func newPgDomainLinkRow(l domain.Link) pgDomainLinkRow {
	return pgDomainLinkRow{
		Id:          l.Id(),
		UserId:      l.UserId(),
		Shortened:   l.Shortened(),
		Destination: l.Destination(),
		IsOpen:      l.IsOpen(),
		UpdatedAt:   l.UpdatedAt(),
		ExpiredAt:   l.ExpiredAt()}
}

func (repo pgLink) Load(id uint64) (domain.Link, error) {
	rows := new([]pgDomainLinkRow)
	query := `
		SELECT * 
		FROM links 
		WHERE id = $1 
		LIMIT 1`
	args := []any{id}
	if err := repo.db.Select(rows, query, args...); err != nil {
		return domain.Link{}, fmt.Errorf("repository<pgLink.Load>: %w", err)
	}

	if len(*rows) != 1 {
		err := oops.NotFound{
			Err: errors.New(
				fmt.Sprintf("link(id: %d) not found", id))}
		return domain.Link{}, fmt.Errorf("repository<pgLink.Load>: %w", err)
	}

	link, err := (*rows)[0].toLink()
	if err != nil {
		return domain.Link{}, fmt.Errorf("repository<pgLink.Load>: %w", err)
	}
	return link, nil
}

func (repo pgLink) Create(l domain.Link) (uint64, error) {
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
	link := newPgDomainLinkRow(l)
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

func (repo pgLink) Update(l domain.Link) error {
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
	link := newPgDomainLinkRow(l)
	_, err := repo.db.NamedExec(query, link)
	if err != nil {
		return fmt.Errorf("repository<pgLink.Update>: %w", err)
	}
	return nil
}

func (repo pgLink) DeleteById(id uint64) error {
	query := `
		DELETE 
		FROM "links"
		WHERE id = $1`
	args := []any{id}
	_, err := repo.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("repository<pgLink.DeleteById>: %w", err)
	}
	return nil
}
