package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/sunilsinghx/order-processing/internal/models"
)

type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) error
	GetByID(ctx context.Context, id string) (*models.Order, error)
	UpdateStatus(ctx context.Context, id string, status models.OrderStatus) error
}

type PostgresOrderRepository struct {
	db *pgx.Conn
}

func NewPostgresRepository(db *pgx.Conn) OrderRepository {
	return &PostgresOrderRepository{db: db}
}

var ErrOrderNotFound = errors.New("Order not found")

func (r *PostgresOrderRepository) Create(ctx context.Context, order *models.Order) error {

	query := `
		INSERT INTO orders (
			id,
			customer_name,
			amount,
			status,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		order.ID,
		order.CustomerName,
		order.Amount,
		order.Status,
		order.CreatedAt,
	)

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (r *PostgresOrderRepository) GetByID(ctx context.Context, id string) (*models.Order, error) {
	query := `
		SELECT
			id,
			customer_name,
			amount,
			status,
			created_at
		FROM orders
		WHERE id = $1
	`

	order := &models.Order{}

	err := r.db.QueryRow(ctx, query, id).Scan(
		&order.ID,
		&order.CustomerName,
		&order.Amount,
		&order.Status,
		&order.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	return order, nil
}

func (r *PostgresOrderRepository) UpdateStatus(ctx context.Context, id string, status models.OrderStatus) error {
	query := `
		UPDATE orders
		SET status = $1
		WHERE id = $2
	`

	_, err := r.db.Exec(ctx, query, status, id)
	return err
}
