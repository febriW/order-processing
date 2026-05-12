package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/febriW/order-processing/common/auth"
	"github.com/redis/go-redis/v9"
)

const sessionKeyPrefix = "auth:sessions:"

type AuthSessionStore interface {
	StoreSession(claims *auth.Claims) error
	ValidateSession(claims *auth.Claims) error
	DeleteSession(tokenID string) error
}

type RedisSessionStore struct {
	client *redis.Client
}

type cachedSession struct {
	UserID    string `json:"user_id"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"`
	ExpiresAt int64  `json:"expires_at"`
}

func NewRedisSessionStore(addr, password string, db int) (*RedisSessionStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("unable to connect to Redis: %w", err)
	}

	return &RedisSessionStore{client: client}, nil
}

func (s *RedisSessionStore) StoreSession(claims *auth.Claims) error {
	if claims == nil || claims.ID == "" || claims.ExpiresAt == nil {
		return fmt.Errorf("invalid token claims")
	}

	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		return fmt.Errorf("token is expired")
	}

	payload, err := json.Marshal(cachedSession{
		UserID:    claims.UserID,
		Role:      claims.Role,
		TokenType: claims.TokenType,
		ExpiresAt: claims.ExpiresAt.Time.Unix(),
	})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return s.client.Set(ctx, sessionKey(claims.ID), payload, ttl).Err()
}

func (s *RedisSessionStore) ValidateSession(claims *auth.Claims) error {
	if claims == nil || claims.ID == "" {
		return fmt.Errorf("invalid token claims")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	value, err := s.client.Get(ctx, sessionKey(claims.ID)).Result()
	if err == redis.Nil {
		return fmt.Errorf("session is not active")
	}
	if err != nil {
		return err
	}

	var session cachedSession
	if err := json.Unmarshal([]byte(value), &session); err != nil {
		return err
	}
	if session.UserID != claims.UserID || session.Role != claims.Role || session.TokenType != claims.TokenType {
		return fmt.Errorf("session does not match token")
	}

	return nil
}

func (s *RedisSessionStore) DeleteSession(tokenID string) error {
	if tokenID == "" {
		return fmt.Errorf("token id is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return s.client.Del(ctx, sessionKey(tokenID)).Err()
}

type noopSessionStore struct{}

func (noopSessionStore) StoreSession(claims *auth.Claims) error {
	return nil
}

func (noopSessionStore) ValidateSession(claims *auth.Claims) error {
	return nil
}

func (noopSessionStore) DeleteSession(tokenID string) error {
	return nil
}

func sessionKey(tokenID string) string {
	return sessionKeyPrefix + tokenID
}
