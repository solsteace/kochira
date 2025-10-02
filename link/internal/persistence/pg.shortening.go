package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/solsteace/go-lib/oops"
	"github.com/solsteace/kochira/link/internal/domain/shortening"
	"github.com/solsteace/kochira/link/internal/domain/shortening/messaging"
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
	Id          uint64    `db:"id"`
	UserId      uint64    `db:"user_id"`
	Shortened   string    `db:"shortened"`
	Destination string    `db:"destination"`
	IsOpen      bool      `db:"is_open"`
	UpdatedAt   time.Time `db:"updated_at"`
	ExpiredAt   time.Time `db:"expired_at"`
}

func (row pgLink) toShortening() (shortening.Link, error) {
	return shortening.NewLink(
		&row.Id,
		row.UserId,
		row.Shortened,
		row.Destination,
		row.IsOpen,
		row.UpdatedAt,
		row.ExpiredAt)
}

func newPgLink(l shortening.Link) pgLink {
	return pgLink{
		Id:          l.Id(),
		UserId:      l.UserId(),
		Shortened:   l.Shortened(),
		Destination: l.Destination(),
		IsOpen:      l.IsOpen(),
		UpdatedAt:   l.UpdatedAt(),
		ExpiredAt:   l.ExpiredAt()}
}

func (repo pg) GetMany(q ShorteningQueryParams) ([]shortening.Link, error) {
	rows := new([]pgLink)
	query := `SELECT * FROM "links" LIMIT $1 OFFSET $2`
	args := []any{q.limit, q.Offset()}
	if err := repo.db.Select(rows, query, args...); err != nil {
		return []shortening.Link{}, fmt.Errorf("persistence<pg.GetMany>: %w", err)
	}

	links := []shortening.Link{}
	for _, r := range *rows {
		link, err := r.toShortening()
		if err != nil {
			return []shortening.Link{}, fmt.Errorf("persistence<pg.GetMany>: %w", err)
		}
		links = append(links, link)
	}
	return links, nil
}

func (repo pg) GetManyByUser(userId uint64, q ShorteningQueryParams) ([]shortening.Link, error) {
	rows := new([]pgLink)
	query := `SELECT * FROM "links" WHERE user_id = $1 LIMIT $2 OFFSET $3`
	args := []any{userId, q.limit, q.Offset()}
	if err := repo.db.Select(rows, query, args...); err != nil {
		return []shortening.Link{}, fmt.Errorf("persistence<pg.GetManyByUser>: %w", err)
	}

	links := []shortening.Link{}
	for _, r := range *rows {
		link, err := r.toShortening()
		if err != nil {
			return []shortening.Link{}, fmt.Errorf("persistence<pg.GetManyByUser>: %w", err)
		}
		links = append(links, link)
	}
	return links, nil
}

func (repo pg) GetById(id uint64) (shortening.Link, error) {
	row := new(pgLink)
	query := `SELECT * FROM links WHERE id = $1 LIMIT 1`
	args := []any{id}
	if err := repo.db.Get(row, query, args...); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			err2 := oops.NotFound{
				Err: err,
				Msg: fmt.Sprintf("link(id:%d) not found", id)}
			return shortening.Link{}, fmt.Errorf("persistence<pg.GetById>: %w", err2)
		default:
			return shortening.Link{}, fmt.Errorf("persistence<pg.GetById>: %w", err)
		}
	}

	s, err := row.toShortening()
	if err != nil {
		return shortening.Link{}, fmt.Errorf("persistence<pg.GetById>: %w", err)
	}
	return s, nil
}

func (repo pg) Create(l shortening.Link) error {
	ctx := context.Background()
	tx, err := repo.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("persistence<pg.Create>: %w", err)
	}
	defer tx.Rollback()

	row := newPgLink(l)
	stmt, err := tx.PrepareNamed(`
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
		RETURNING id`)
	if err != nil {
		return fmt.Errorf("persistence<pg.Create>: %w", err)
	}
	var linkId uint64
	if err := stmt.Get(&linkId, row); err != nil {
		return fmt.Errorf("persistence<pg.Create>: %w", err)
	}

	outboxQuery := `
		INSERT INTO link_shortened_outbox(user_id, link_id) 
		VALUES ($1, $2)`
	outboxArgs := []any{row.UserId, linkId}
	if _, err := tx.Exec(outboxQuery, outboxArgs...); err != nil {
		return fmt.Errorf("persistence<pg.Create>: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("persistence<pg.Create>: %w", err)
	}
	return nil
}

func (repo pg) Configure(l shortening.Link) error {
	ctx := context.Background()
	tx, err := repo.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("persistence<pg.Configure>: %w", err)
	}
	defer tx.Rollback()

	row := newPgLink(l)
	query := `
		UPDATE "links"
		SET is_open = false
		WHERE id = :id`
	if _, err := tx.NamedExec(query, row); err != nil {
		return fmt.Errorf("persistence<pg.Configure>: %w", err)
	}

	outboxQuery := `
		INSERT INTO short_configured_outbox(
			user_id, 
			link_id, 
			destination, 
			shortened) 
		VALUES ($1, $2, $3, $4)`
	outboxArgs := []any{
		row.UserId,
		row.Id,
		row.Destination,
		row.Shortened}
	if _, err := tx.Exec(outboxQuery, outboxArgs...); err != nil {
		return fmt.Errorf("persistence<pg.Configure>: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("persistence<pg.Configure>: %w", err)
	}
	return nil
}

func (repo pg) Update(l shortening.Link) error {
	row := newPgLink(l)
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
	if _, err := repo.db.NamedExec(query, row); err != nil {
		return fmt.Errorf("persistence<pg.Update>: %w", err)
	}

	return nil
}

func (pg pg) DeleteById(id uint64) error {
	query := `DELETE FROM "links" WHERE id = $1`
	args := []any{id}
	_, err := pg.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("persistence<pg.DeleteById>: %w", err)
	}
	return nil
}

func (repo pg) CountByUserIdExcept(userId uint64, linkId uint64) (shortening.Stats, error) {
	query := `
		SELECT COUNT(*) AS n_links 
		FROM links 
		WHERE 
			user_id = $1
			AND id <> $2
			AND CURRENT_TIMESTAMP < expired_at`
	args := []any{userId, linkId}
	result := repo.db.QueryRow(query, args...)
	if result.Err() != nil {
		return shortening.Stats{}, fmt.Errorf("persistence<pg.CountByUserIdExcept>: %w", result.Err())
	}

	var count uint
	if err := result.Scan(&count); err != nil {
		return shortening.Stats{}, fmt.Errorf("persistence<pg.CountByUserIdExcept>: %w", err)
	}
	return shortening.NewStats(count), nil
}

type pgLinkShortened struct {
	Id     uint64 `db:"id"`
	UserId uint64 `db:"user_id"`
	LinkId uint64 `db:"link_id"`
}

func (row pgLinkShortened) toMessage() messaging.LinkShortened {
	return messaging.NewLinkShortened(
		row.Id,
		row.UserId,
		row.LinkId)
}

func (repo pg) GetLinkShortened(maxCount uint) ([]messaging.LinkShortened, error) {
	query := `
		SELECT 
			id,
			user_id,
			link_id
		FROM link_shortened_outbox AS o 
		WHERE is_done = false 
		LIMIT $1`
	args := []any{maxCount}
	rows := new([]pgLinkShortened)
	if err := repo.db.Select(rows, query, args...); err != nil {
		return []messaging.LinkShortened{}, fmt.Errorf("persistence<pg.GetLinkShortened>: %w", err)
	}

	messages := []messaging.LinkShortened{}
	for _, row := range *rows {
		messages = append(messages, row.toMessage())
	}
	return messages, nil
}

func (repo pg) GetLinkShortenedById(id uint64) (messaging.LinkShortened, error) {
	query := `SELECT id, user_id, link_id FROM link_shortened_outbox WHERE id = $1`
	args := []any{id}
	row := new(pgLinkShortened)
	if err := repo.db.Get(row, query, args...); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			err2 := oops.NotFound{
				Err: err,
				Msg: fmt.Sprintf("link_shortened_outbox(id:%d) not found", id)}
			return messaging.LinkShortened{}, fmt.Errorf("persistence<pg.GetLinkShortenedById>: %w", err2)
		default:
			return messaging.LinkShortened{}, fmt.Errorf("persistence<pg.GetLinkShortenedById>: %w", err)
		}
	}
	return row.toMessage(), nil
}

func (repo pg) ResolveLinkShortened(id []uint64) error {
	query, args, err := sqlx.In(`
		UPDATE link_shortened_outbox
		SET is_done = true
		WHERE id IN (?)`, id)
	if err != nil {
		return fmt.Errorf("persistence<pg.ResolveCheckSubscriptions>: %w", err)
	}

	if _, err := repo.db.Exec(repo.db.Rebind(query), args...); err != nil {
		return fmt.Errorf("persistence<pg.ResolveCheckSubscriptions>: %w", err)
	}
	return nil
}

type pgShortConfigured struct {
	Id          uint64 `db:"id"`
	UserId      uint64 `db:"user_id"`
	LinkId      uint64 `db:"link_id"`
	Shortened   string `db:"shortened"`
	Destination string `db:"destination"`
}

func (row pgShortConfigured) toMessage() messaging.ShortConfigured {
	return messaging.NewShortConfigured(
		row.Id,
		row.LinkId,
		row.UserId,
		row.Shortened,
		row.Destination)
}

func (repo pg) GetShortConfigured(maxCount uint) ([]messaging.ShortConfigured, error) {
	query := `
		SELECT
			id,
			user_id,
			link_id,
			destination,
			shortened
		FROM short_configured_outbox 
		WHERE is_done = false 
		LIMIT $1`
	args := []any{maxCount}
	rows := new([]pgShortConfigured)
	if err := repo.db.Select(rows, query, args...); err != nil {
		return []messaging.ShortConfigured{}, fmt.Errorf("persistence<pg.GetShortConfigured>: %w", err)
	}

	messages := []messaging.ShortConfigured{}
	for _, row := range *rows {
		messages = append(messages, row.toMessage())
	}
	return messages, nil
}

func (repo pg) GetShortConfiguredById(id uint64) (messaging.ShortConfigured, error) {
	query := ` 
		SELECT
			id,
			user_id,
			link_id,
			destination,
			shortened
		FROM short_configured_outbox 
		WHERE id =  $1`
	args := []any{id}
	row := new(pgShortConfigured)
	if err := repo.db.Get(row, query, args...); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			err2 := oops.NotFound{
				Err: err,
				Msg: fmt.Sprintf("link(id:%d) not found", id)}
			return messaging.ShortConfigured{}, fmt.Errorf("persistence<pg.GetShortConfiguredById>: %w", err2)
		default:
			return messaging.ShortConfigured{}, fmt.Errorf("persistence<pg.GetShortConfiguredById>: %w", err)
		}
	}
	return row.toMessage(), nil
}

func (repo pg) ResolveShortConfigured(id []uint64) error {
	query, args, err := sqlx.In(`
		UPDATE short_configured_outbox
		SET is_done = true
		WHERE id IN (?)`, id)
	if err != nil {
		return fmt.Errorf("persistence<pg.ResolveShortConfigured>: %w", err)
	}

	if _, err := repo.db.Exec(repo.db.Rebind(query), args...); err != nil {
		return fmt.Errorf("persistence<pg.ResolveShortConfigured>: %w", err)
	}
	return nil
}
