package main

import (
	"fmt"
	"strings"

	"github.com/febriW/order-processing/common/models"
	"github.com/google/uuid"
)

type ProductService struct {
	repo ProductRepository
}

type ProductRepository interface {
	CreateProduct(product models.Product) error
	GetProductByID(id string) (*models.Product, error)
	ListProducts() ([]models.Product, error)
	UpdateProduct(product models.Product) error
	DeleteProduct(id string) error
}

type ProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
}

func NewProductService(repo ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) CreateProduct(req ProductRequest) (*models.Product, error) {
	if err := validateProductRequest(req); err != nil {
		return nil, err
	}

	product := models.Product{
		ID:          uuid.NewString(),
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Price:       req.Price,
		Stock:       req.Stock,
	}

	if err := s.repo.CreateProduct(product); err != nil {
		return nil, fmt.Errorf("could not create product: %w", err)
	}

	return s.repo.GetProductByID(product.ID)
}

func (s *ProductService) GetProduct(id string) (*models.Product, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	return s.repo.GetProductByID(id)
}

func (s *ProductService) ListProducts() ([]models.Product, error) {
	return s.repo.ListProducts()
}

func (s *ProductService) UpdateProduct(id string, req ProductRequest) (*models.Product, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	if err := validateProductRequest(req); err != nil {
		return nil, err
	}

	product := models.Product{
		ID:          id,
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Price:       req.Price,
		Stock:       req.Stock,
	}

	if err := s.repo.UpdateProduct(product); err != nil {
		return nil, fmt.Errorf("could not update product: %w", err)
	}

	return s.repo.GetProductByID(id)
}

func (s *ProductService) DeleteProduct(id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("product id is required")
	}
	return s.repo.DeleteProduct(id)
}

func validateProductRequest(req ProductRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("product name is required")
	}
	if req.Price < 0 {
		return fmt.Errorf("product price cannot be negative")
	}
	if req.Stock < 0 {
		return fmt.Errorf("product stock cannot be negative")
	}
	return nil
}
