package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/febriW/order-processing/common/models"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresOrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(databaseURL string) *PostgresOrderRepository {
	db, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = db.Exec(ctx, `CREATE TABLE IF NOT EXISTS orders (
		id VARCHAR(255) PRIMARY KEY,
		user_id VARCHAR(255) NOT NULL,
		product_id VARCHAR(255) NOT NULL,
		quantity INTEGER NOT NULL,
		amount NUMERIC(12, 2) NOT NULL DEFAULT 0,
		status VARCHAR(50) NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`)
	if err != nil {
		log.Fatalf("Could not create orders table: %v\n", err)
	}

	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) CreateOrder(order models.Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := r.db.Exec(
		ctx,
		`INSERT INTO orders (id, user_id, product_id, quantity, amount, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		order.ID,
		order.UserID,
		order.ProductID,
		order.Quantity,
		order.Amount,
		order.Status,
		order.CreatedAt,
	)
	return err
}

func (r *PostgresOrderRepository) GetOrderByID(id string) (*models.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var order models.Order
	err := pgxscan.Get(
		ctx,
		r.db,
		&order,
		`SELECT id, user_id, product_id, quantity, amount, status, created_at FROM orders WHERE id = $1`,
		id,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, err
	}
	return &order, nil
}

func (r *PostgresOrderRepository) ListOrdersByUser(userID string) ([]models.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	orders := make([]models.Order, 0)
	err := pgxscan.Select(
		ctx,
		r.db,
		&orders,
		`SELECT id, user_id, product_id, quantity, amount, status, created_at
		 FROM orders WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	return orders, nil
}
