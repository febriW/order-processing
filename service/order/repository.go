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

type OrderOutboxEvent struct {
	ID            string
	EventType     string
	Payload       []byte
	Attempts      int
	TraceID       string
	CorrelationID string
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

	_, err = db.Exec(ctx, `CREATE TABLE IF NOT EXISTS order_outbox (
		id VARCHAR(255) PRIMARY KEY,
		event_type VARCHAR(100) NOT NULL,
		payload JSONB NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		published_at TIMESTAMPTZ NULL,
		attempts INTEGER NOT NULL DEFAULT 0,
		last_error TEXT NOT NULL DEFAULT '',
		status VARCHAR(20) NOT NULL DEFAULT 'pending',
		next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		claimed_at TIMESTAMPTZ NULL,
		claim_owner VARCHAR(100) NOT NULL DEFAULT '',
		trace_id VARCHAR(100) NOT NULL DEFAULT '',
		correlation_id VARCHAR(100) NOT NULL DEFAULT ''
	)`)
	if err != nil {
		log.Fatalf("Could not create order_outbox table: %v\n", err)
	}
	_, _ = db.Exec(ctx, `ALTER TABLE order_outbox ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'pending'`)
	_, _ = db.Exec(ctx, `ALTER TABLE order_outbox ADD COLUMN IF NOT EXISTS next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`)
	_, _ = db.Exec(ctx, `ALTER TABLE order_outbox ADD COLUMN IF NOT EXISTS claimed_at TIMESTAMPTZ NULL`)
	_, _ = db.Exec(ctx, `ALTER TABLE order_outbox ADD COLUMN IF NOT EXISTS claim_owner VARCHAR(100) NOT NULL DEFAULT ''`)
	_, _ = db.Exec(ctx, `ALTER TABLE order_outbox ADD COLUMN IF NOT EXISTS trace_id VARCHAR(100) NOT NULL DEFAULT ''`)
	_, _ = db.Exec(ctx, `ALTER TABLE order_outbox ADD COLUMN IF NOT EXISTS correlation_id VARCHAR(100) NOT NULL DEFAULT ''`)
	_, _ = db.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_order_outbox_ready ON order_outbox (next_attempt_at, created_at) WHERE published_at IS NULL`)

	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) CreateOrderWithOutbox(order models.Order, eventType string, payload []byte, traceID, correlationID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(
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
	); err != nil {
		return err
	}

	if _, err := tx.Exec(
		ctx,
		`INSERT INTO order_outbox (id, event_type, payload, trace_id, correlation_id) VALUES ($1, $2, $3::jsonb, $4, $5)`,
		order.ID,
		eventType,
		string(payload),
		traceID,
		correlationID,
	); err != nil {
		return err
	}

	return tx.Commit(ctx)
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

func (r *PostgresOrderRepository) ClaimPendingOutboxEvents(limit, maxAttempts int, owner string, claimTTL time.Duration) ([]OrderOutboxEvent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if limit <= 0 {
		limit = 20
	}
	if maxAttempts <= 0 {
		maxAttempts = 8
	}
	if claimTTL <= 0 {
		claimTTL = 30 * time.Second
	}

	events := make([]OrderOutboxEvent, 0)
	rows, err := r.db.Query(
		ctx,
		`WITH picked AS (
			SELECT id
			FROM order_outbox
			WHERE published_at IS NULL
			  AND attempts < $2
			  AND next_attempt_at <= NOW()
			  AND (
				status = 'pending'
				OR (status = 'processing' AND claimed_at < NOW() - $4::interval)
			  )
			ORDER BY created_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT $1
		)
		UPDATE order_outbox o
		SET status = 'processing',
			claimed_at = NOW(),
			claim_owner = $3
		FROM picked
		WHERE o.id = picked.id
		RETURNING o.id, o.event_type, o.payload::text, o.attempts, o.trace_id, o.correlation_id`,
		limit,
		maxAttempts,
		owner,
		fmt.Sprintf("%f seconds", claimTTL.Seconds()),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var event OrderOutboxEvent
		var payload string
		if err := rows.Scan(&event.ID, &event.EventType, &payload, &event.Attempts, &event.TraceID, &event.CorrelationID); err != nil {
			return nil, err
		}
		event.Payload = []byte(payload)
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (r *PostgresOrderRepository) MarkOutboxEventPublished(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := r.db.Exec(
		ctx,
		`UPDATE order_outbox
		 SET published_at = NOW(),
		     status = 'sent',
		     attempts = attempts + 1,
			 last_error = '',
			 claimed_at = NULL,
			 claim_owner = ''
		 WHERE id = $1 AND published_at IS NULL`,
		id,
	)
	return err
}

func (r *PostgresOrderRepository) MarkOutboxEventRetry(id, reason string, nextAttemptAt time.Time, dead bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	status := "pending"
	if dead {
		status = "dead"
	}

	_, err := r.db.Exec(
		ctx,
		`UPDATE order_outbox
		 SET attempts = attempts + 1,
		     last_error = $2,
			 status = $3,
			 next_attempt_at = $4,
			 claimed_at = NULL,
			 claim_owner = ''
		 WHERE id = $1 AND published_at IS NULL`,
		id,
		reason,
		status,
		nextAttemptAt,
	)
	return err
}
