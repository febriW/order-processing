package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/febriW/order-processing/common/models"
	"github.com/google/uuid"
)

type OrderService struct {
	repo      OrderRepository
	cache     OrderCacheStore
	publisher OrderEventPublisher
}

type OrderRepository interface {
	CreateOrder(order models.Order) error
	GetOrderByID(id string) (*models.Order, error)
	ListOrdersByUser(userID string) ([]models.Order, error)
}

type CreateOrderRequest struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Amount    float64 `json:"amount"`
}

func NewOrderService(repo OrderRepository, cache OrderCacheStore, publisher OrderEventPublisher) *OrderService {
	if cache == nil {
		cache = noopOrderCacheStore{}
	}
	if publisher == nil {
		publisher = noopOrderEventPublisher{}
	}
	return &OrderService{repo: repo, cache: cache, publisher: publisher}
}

func (s *OrderService) CreateOrder(userID, idempotencyKey string, req CreateOrderRequest) (*models.Order, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("user id is required")
	}
	if err := validateCreateOrderRequest(req); err != nil {
		return nil, err
	}

	if existing, err := s.cache.GetCreatedOrder(strings.TrimSpace(idempotencyKey)); err == nil && existing != nil {
		return existing, nil
	}

	order := models.Order{
		ID:        uuid.NewString(),
		UserID:    userID,
		ProductID: strings.TrimSpace(req.ProductID),
		Quantity:  req.Quantity,
		Amount:    req.Amount,
		Status:    models.OrderStatusPending,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.repo.CreateOrder(order); err != nil {
		return nil, fmt.Errorf("could not create order: %w", err)
	}

	created, err := s.repo.GetOrderByID(order.ID)
	if err != nil {
		return nil, err
	}

	_ = s.cache.StoreCreatedOrder(strings.TrimSpace(idempotencyKey), *created)
	_ = s.cache.CacheOrder(*created)
	_ = s.publisher.PublishOrderCreated(*created)

	return created, nil
}

func (s *OrderService) GetOrderByID(userID, id string) (*models.Order, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("user id is required")
	}

	if cached, err := s.cache.GetCachedOrder(id); err == nil && cached != nil {
		if cached.UserID != userID {
			return nil, fmt.Errorf("order not found")
		}
		return cached, nil
	}

	order, err := s.repo.GetOrderByID(id)
	if err != nil {
		return nil, err
	}
	if order.UserID != userID {
		return nil, fmt.Errorf("order not found")
	}

	_ = s.cache.CacheOrder(*order)
	return order, nil
}

func (s *OrderService) ListOrders(userID string) ([]models.Order, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("user id is required")
	}
	return s.repo.ListOrdersByUser(userID)
}

func validateCreateOrderRequest(req CreateOrderRequest) error {
	if strings.TrimSpace(req.ProductID) == "" {
		return fmt.Errorf("product id is required")
	}
	if req.Quantity <= 0 {
		return fmt.Errorf("quantity must be greater than zero")
	}
	if req.Amount <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}
	if req.Amount > 1000000000 {
		return fmt.Errorf("amount is too large")
	}
	return nil
}
