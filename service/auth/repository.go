package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/febriW/order-processing/common/models"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepository struct {
	db *pgxpool.Pool
}

func NewAuthRepository(databaseURL string) *AuthRepository {
	db, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	_, err = db.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS users (
        id VARCHAR(255) PRIMARY KEY, email VARCHAR(255) UNIQUE,
        password VARCHAR(255), role VARCHAR(50) NOT NULL DEFAULT 'customer'
    )`)
	if err != nil {
		log.Fatalf("Could not create users table: %v\n", err)
	}

	return &AuthRepository{db: db}
}

func (r *AuthRepository) CreateUser(user models.User) error {
	ctx := context.Background()
	sqlStatement := `INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, sqlStatement, user.ID, user.Email, user.Password)
	if err != nil {
		return err
	}

	if strings.Contains(user.Email, "admin@") {
		_, err = r.db.Exec(ctx, "UPDATE users SET role = 'admin' WHERE email = $1", user.Email)
	}
	return err
}

func (r *AuthRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	ctx := context.Background()

	err := pgxscan.Get(ctx, r.db, &user, `SELECT id, email, role FROM users WHERE email=$1`, email)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	return &user, nil
}
