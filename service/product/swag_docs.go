package main

import "github.com/febriW/order-processing/common/models"

// @title Product Service API
// @version 1.0
// @description Product endpoints for order-processing.
// @BasePath /
// @schemes http
// @accept json
// @produce json
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

type productDataResponse struct {
	Status  int            `json:"status"`
	Message string         `json:"message"`
	Data    models.Product `json:"data"`
}

type productListDataResponse struct {
	Status  int              `json:"status"`
	Message string           `json:"message"`
	Data    []models.Product `json:"data"`
}

type productEmptyResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// createProduct godoc
// @Summary Create product
// @Tags product
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body ProductRequest true "Product payload"
// @Success 201 {object} productDataResponse
// @Failure 400 {object} productEmptyResponse
// @Failure 401 {object} productEmptyResponse
// @Router /products [post]
func createProduct() {}

// listProducts godoc
// @Summary List products
// @Tags product
// @Produce json
// @Security BearerAuth
// @Success 200 {object} productListDataResponse
// @Failure 401 {object} productEmptyResponse
// @Failure 500 {object} productEmptyResponse
// @Router /products [get]
func listProducts() {}

// getProduct godoc
// @Summary Get product by id
// @Tags product
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Success 200 {object} productDataResponse
// @Failure 401 {object} productEmptyResponse
// @Failure 404 {object} productEmptyResponse
// @Router /products/{id} [get]
func getProduct() {}

// updateProduct godoc
// @Summary Update product by id
// @Tags product
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Param payload body ProductRequest true "Product payload"
// @Success 200 {object} productDataResponse
// @Failure 400 {object} productEmptyResponse
// @Failure 401 {object} productEmptyResponse
// @Router /products/{id} [put]
func updateProduct() {}

// deleteProduct godoc
// @Summary Delete product by id
// @Tags product
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Success 200 {object} productEmptyResponse
// @Failure 401 {object} productEmptyResponse
// @Failure 404 {object} productEmptyResponse
// @Router /products/{id} [delete]
func deleteProduct() {}
