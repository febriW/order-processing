package main

import (
	"fmt"
	"testing"

	"github.com/febriW/order-processing/common/models"
)

type fakeProductRepository struct {
	products map[string]models.Product
}

func newFakeProductRepository() *fakeProductRepository {
	return &fakeProductRepository{
		products: make(map[string]models.Product),
	}
}

func (r *fakeProductRepository) CreateProduct(product models.Product) error {
	r.products[product.ID] = product
	return nil
}

func (r *fakeProductRepository) GetProductByID(id string) (*models.Product, error) {
	product, exists := r.products[id]
	if !exists {
		return nil, fmt.Errorf("product not found")
	}
	return &product, nil
}

func (r *fakeProductRepository) ListProducts() ([]models.Product, error) {
	products := make([]models.Product, 0, len(r.products))
	for _, product := range r.products {
		products = append(products, product)
	}
	return products, nil
}

func (r *fakeProductRepository) UpdateProduct(product models.Product) error {
	if _, exists := r.products[product.ID]; !exists {
		return fmt.Errorf("product not found")
	}
	r.products[product.ID] = product
	return nil
}

func (r *fakeProductRepository) DeleteProduct(id string) error {
	if _, exists := r.products[id]; !exists {
		return fmt.Errorf("product not found")
	}
	delete(r.products, id)
	return nil
}

func TestProductServiceCreateProduct(t *testing.T) {
	service := NewProductService(newFakeProductRepository())

	product, err := service.CreateProduct(ProductRequest{
		Name:        "Coffee",
		Description: "Arabica beans",
		Price:       10.5,
		Stock:       20,
	})
	if err != nil {
		t.Fatalf("CreateProduct returned error: %v", err)
	}
	if product.ID == "" {
		t.Fatal("expected product ID to be generated")
	}
	if product.Name != "Coffee" {
		t.Fatalf("expected product name Coffee, got %s", product.Name)
	}
}

func TestProductServiceCreateProductValidation(t *testing.T) {
	service := NewProductService(newFakeProductRepository())

	_, err := service.CreateProduct(ProductRequest{
		Name:  "",
		Price: 10,
		Stock: 1,
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestProductServiceUpdateProduct(t *testing.T) {
	repo := newFakeProductRepository()
	service := NewProductService(repo)

	product, err := service.CreateProduct(ProductRequest{
		Name:  "Coffee",
		Price: 10.5,
		Stock: 20,
	})
	if err != nil {
		t.Fatalf("CreateProduct returned error: %v", err)
	}

	updated, err := service.UpdateProduct(product.ID, ProductRequest{
		Name:        "Tea",
		Description: "Green tea",
		Price:       5.25,
		Stock:       15,
	})
	if err != nil {
		t.Fatalf("UpdateProduct returned error: %v", err)
	}
	if updated.Name != "Tea" {
		t.Fatalf("expected product name Tea, got %s", updated.Name)
	}
	if updated.Stock != 15 {
		t.Fatalf("expected stock 15, got %d", updated.Stock)
	}
}

func TestProductServiceDeleteProduct(t *testing.T) {
	repo := newFakeProductRepository()
	service := NewProductService(repo)

	product, err := service.CreateProduct(ProductRequest{
		Name:  "Coffee",
		Price: 10.5,
		Stock: 20,
	})
	if err != nil {
		t.Fatalf("CreateProduct returned error: %v", err)
	}

	if err := service.DeleteProduct(product.ID); err != nil {
		t.Fatalf("DeleteProduct returned error: %v", err)
	}

	if _, err := service.GetProduct(product.ID); err == nil {
		t.Fatal("expected deleted product to be missing")
	}
}
