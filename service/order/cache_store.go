package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/febriW/order-processing/common/models"
	"github.com/redis/go-redis/v9"
)

const (
	idempotencyKeyPrefix = "orders:idempotency:"
	orderCacheKeyPrefix  = "orders:by-id:"
)

type OrderCacheStore interface {
	StoreCreatedOrder(idempotencyKey string, order models.Order) error
	GetCreatedOrder(idempotencyKey string) (*models.Order, error)
	CacheOrder(order models.Order) error
	GetCachedOrder(id string) (*models.Order, error)
}

type RedisOrderStore struct {
	client *redis.Client
}

func NewRedisOrderStore(addr, password string, db int) (*RedisOrderStore, error) {
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

	return &RedisOrderStore{client: client}, nil
}

func (s *RedisOrderStore) StoreCreatedOrder(idempotencyKey string, order models.Order) error {
	if idempotencyKey == "" {
		return nil
	}
	payload, err := json.Marshal(order)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return s.client.Set(ctx, idempotencyRedisKey(idempotencyKey), payload, 24*time.Hour).Err()
}

func (s *RedisOrderStore) GetCreatedOrder(idempotencyKey string) (*models.Order, error) {
	if idempotencyKey == "" {
		return nil, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	value, err := s.client.Get(ctx, idempotencyRedisKey(idempotencyKey)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var order models.Order
	if err := json.Unmarshal([]byte(value), &order); err != nil {
		return nil, err
	}
	return &order, nil
}

func (s *RedisOrderStore) CacheOrder(order models.Order) error {
	payload, err := json.Marshal(order)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return s.client.Set(ctx, orderRedisKey(order.ID), payload, 10*time.Minute).Err()
}

func (s *RedisOrderStore) GetCachedOrder(id string) (*models.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	value, err := s.client.Get(ctx, orderRedisKey(id)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var order models.Order
	if err := json.Unmarshal([]byte(value), &order); err != nil {
		return nil, err
	}
	return &order, nil
}

func idempotencyRedisKey(key string) string {
	return idempotencyKeyPrefix + key
}

func orderRedisKey(id string) string {
	return orderCacheKeyPrefix + id
}

type noopOrderCacheStore struct{}

func (noopOrderCacheStore) StoreCreatedOrder(idempotencyKey string, order models.Order) error {
	return nil
}
func (noopOrderCacheStore) GetCreatedOrder(idempotencyKey string) (*models.Order, error) {
	return nil, nil
}
func (noopOrderCacheStore) CacheOrder(order models.Order) error {
	return nil
}
func (noopOrderCacheStore) GetCachedOrder(id string) (*models.Order, error) {
	return nil, nil
}
