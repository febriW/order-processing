package main

import "github.com/febriW/order-processing/common/models"

// @title Order Service API
// @version 1.0
// @description Order endpoints for order-processing.
// @BasePath /
// @schemes http
// @accept json
// @produce json
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

type orderDataResponse struct {
	Status  int          `json:"status"`
	Message string       `json:"message"`
	Data    models.Order `json:"data"`
}

type orderListDataResponse struct {
	Status  int            `json:"status"`
	Message string         `json:"message"`
	Data    []models.Order `json:"data"`
}

type orderEmptyResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// createOrder godoc
// @Summary Create order
// @Tags order
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Idempotency-Key header string true "Idempotency key"
// @Param payload body CreateOrderRequest true "Order payload"
// @Success 201 {object} orderDataResponse
// @Failure 400 {object} orderEmptyResponse
// @Failure 401 {object} orderEmptyResponse
// @Router /orders [post]
func createOrder() {}

// listOrders godoc
// @Summary List orders
// @Tags order
// @Produce json
// @Security BearerAuth
// @Success 200 {object} orderListDataResponse
// @Failure 401 {object} orderEmptyResponse
// @Failure 500 {object} orderEmptyResponse
// @Router /orders [get]
func listOrders() {}

// getOrder godoc
// @Summary Get order by id
// @Tags order
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} orderDataResponse
// @Failure 401 {object} orderEmptyResponse
// @Failure 404 {object} orderEmptyResponse
// @Router /orders/{id} [get]
func getOrder() {}
