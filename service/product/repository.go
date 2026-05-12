package main

import (
	"context"
	"fmt"
	"log"

	"github.com/febriW/order-processing/common/models"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresProductRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(databaseURL string) *PostgresProductRepository {
	db, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	_, err = db.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS products (
		id VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		price NUMERIC(12, 2) NOT NULL DEFAULT 0,
		stock INTEGER NOT NULL DEFAULT 0
	)`)
	if err != nil {
		log.Fatalf("Could not create products table: %v\n", err)
	}

	return &PostgresProductRepository{db: db}
}

func (r *PostgresProductRepository) CreateProduct(product models.Product) error {
	_, err := r.db.Exec(
		context.Background(),
		`INSERT INTO products (id, name, description, price, stock) VALUES ($1, $2, $3, $4, $5)`,
		product.ID,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
	)
	return err
}

func (r *PostgresProductRepository) GetProductByID(id string) (*models.Product, error) {
	var product models.Product
	err := pgxscan.Get(
		context.Background(),
		r.db,
		&product,
		`SELECT id, name, description, price, stock FROM products WHERE id = $1`,
		id,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("product not found")
		}
		return nil, err
	}
	return &product, nil
}

func (r *PostgresProductRepository) ListProducts() ([]models.Product, error) {
	products := make([]models.Product, 0)
	err := pgxscan.Select(
		context.Background(),
		r.db,
		&products,
		`SELECT id, name, description, price, stock FROM products ORDER BY name ASC`,
	)
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (r *PostgresProductRepository) UpdateProduct(product models.Product) error {
	tag, err := r.db.Exec(
		context.Background(),
		`UPDATE products SET name = $1, description = $2, price = $3, stock = $4 WHERE id = $5`,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.ID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}
	return nil
}

func (r *PostgresProductRepository) DeleteProduct(id string) error {
	tag, err := r.db.Exec(context.Background(), `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}
	return nil
}
