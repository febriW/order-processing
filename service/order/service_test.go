package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/febriW/order-processing/common/models"
)

type fakeOrderRepository struct {
	orders map[string]models.Order
}

type fakeOrderCache struct {
	idempotency map[string]models.Order
	byID        map[string]models.Order
}

type fakeOrderPublisher struct {
	published []models.Order
}

func newFakeOrderRepository() *fakeOrderRepository {
	return &fakeOrderRepository{orders: make(map[string]models.Order)}
}

func newFakeOrderCache() *fakeOrderCache {
	return &fakeOrderCache{
		idempotency: make(map[string]models.Order),
		byID:        make(map[string]models.Order),
	}
}

func (r *fakeOrderRepository) CreateOrder(order models.Order) error {
	r.orders[order.ID] = order
	return nil
}

func (r *fakeOrderRepository) GetOrderByID(id string) (*models.Order, error) {
	order, ok := r.orders[id]
	if !ok {
		return nil, fmt.Errorf("order not found")
	}
	return &order, nil
}

func (r *fakeOrderRepository) ListOrdersByUser(userID string) ([]models.Order, error) {
	result := make([]models.Order, 0)
	for _, order := range r.orders {
		if order.UserID == userID {
			result = append(result, order)
		}
	}
	return result, nil
}

func (c *fakeOrderCache) StoreCreatedOrder(idempotencyKey string, order models.Order) error {
	if idempotencyKey != "" {
		c.idempotency[idempotencyKey] = order
	}
	return nil
}

func (c *fakeOrderCache) GetCreatedOrder(idempotencyKey string) (*models.Order, error) {
	order, ok := c.idempotency[idempotencyKey]
	if !ok {
		return nil, nil
	}
	return &order, nil
}

func (c *fakeOrderCache) CacheOrder(order models.Order) error {
	c.byID[order.ID] = order
	return nil
}

func (c *fakeOrderCache) GetCachedOrder(id string) (*models.Order, error) {
	order, ok := c.byID[id]
	if !ok {
		return nil, nil
	}
	return &order, nil
}

func (p *fakeOrderPublisher) PublishOrderCreated(order models.Order) error {
	p.published = append(p.published, order)
	return nil
}

func TestCreateOrder(t *testing.T) {
	repo := newFakeOrderRepository()
	cache := newFakeOrderCache()
	publisher := &fakeOrderPublisher{}
	service := NewOrderService(repo, cache, publisher)

	order, err := service.CreateOrder("user-1", "idem-1", CreateOrderRequest{
		ProductID: "product-1",
		Quantity:  2,
		Amount:    120000,
	})
	if err != nil {
		t.Fatalf("CreateOrder returned error: %v", err)
	}

	if order.ID == "" || order.Status != models.OrderStatusPending {
		t.Fatalf("unexpected order response: %+v", order)
	}
	if order.UserID != "user-1" {
		t.Fatalf("expected user-1, got %s", order.UserID)
	}
	if len(publisher.published) != 1 {
		t.Fatalf("expected one order event, got %d", len(publisher.published))
	}
}

func TestCreateOrderIdempotency(t *testing.T) {
	repo := newFakeOrderRepository()
	cache := newFakeOrderCache()
	publisher := &fakeOrderPublisher{}
	service := NewOrderService(repo, cache, publisher)

	cached := models.Order{
		ID:        "existing-order",
		UserID:    "user-1",
		ProductID: "product-1",
		Quantity:  1,
		Amount:    1000,
		Status:    models.OrderStatusPending,
		CreatedAt: time.Now().UTC(),
	}
	cache.idempotency["idem-1"] = cached

	order, err := service.CreateOrder("user-1", "idem-1", CreateOrderRequest{
		ProductID: "product-1",
		Quantity:  1,
		Amount:    1000,
	})
	if err != nil {
		t.Fatalf("CreateOrder returned error: %v", err)
	}
	if order.ID != "existing-order" {
		t.Fatalf("expected existing-order, got %s", order.ID)
	}
	if len(repo.orders) != 0 {
		t.Fatalf("expected repository write to be skipped, got %d", len(repo.orders))
	}
}
