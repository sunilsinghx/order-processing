package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrEventAlreadyProcessed = errors.New("event has already been processed")

type IdempotencyRepository interface {
	// ClaimEvent attempts to register the event.
	// Returns nil if successful (first time seeing it).
	// Returns ErrEventAlreadyProcessed if the event already exists.
	ClaimEvent(ctx context.Context, eventID string) error
}

type PostgresIdempotencyRepository struct {
	db *pgx.Conn
}

func NewPostgresIdempotencyRepository(db *pgx.Conn) IdempotencyRepository {
	return &PostgresIdempotencyRepository{
		db: db,
	}
}

func (r *PostgresIdempotencyRepository) ClaimEvent(ctx context.Context, eventID string) error {
	_, err := r.db.Exec(
		ctx,
		`INSERT INTO processed_events (event_id) VALUES ($1)`,
		eventID,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return ErrEventAlreadyProcessed
		}
		return err
	}

	return nil
}

func (r *PostgresIdempotencyRepository) DeleteEvent(ctx context.Context, eventID string) error {
	_, err := r.db.Exec(
		ctx,
		`DELETE FROM processed_events WHERE event_id = $1`,
		eventID,
	)
	if err != nil {
		return err
	}

	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
