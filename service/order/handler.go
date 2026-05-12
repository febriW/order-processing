package main

import (
	"encoding/json"
	"net/http"

	"github.com/febriW/order-processing/common/models"
	"github.com/febriW/order-processing/common/middleware"
	"github.com/febriW/order-processing/common/utils"
	"github.com/gorilla/mux"
)

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

type OrderHandler struct {
	service *OrderService
}

func NewOrderHandler(service *OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

// CreateOrderHandler godoc
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
func (h *OrderHandler) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "invalid user context")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	var req CreateOrderRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	order, err := h.service.CreateOrder(userID, r.Header.Get("Idempotency-Key"), req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, "Order created successfully", order)
}

// GetOrderHandler godoc
// @Summary Get order by id
// @Tags order
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} orderDataResponse
// @Failure 401 {object} orderEmptyResponse
// @Failure 404 {object} orderEmptyResponse
// @Router /orders/{id} [get]
func (h *OrderHandler) GetOrderHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "invalid user context")
		return
	}

	order, err := h.service.GetOrderByID(userID, mux.Vars(r)["id"])
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, "Order retrieved successfully", order)
}

// ListOrdersHandler godoc
// @Summary List orders
// @Tags order
// @Produce json
// @Security BearerAuth
// @Success 200 {object} orderListDataResponse
// @Failure 401 {object} orderEmptyResponse
// @Failure 500 {object} orderEmptyResponse
// @Router /orders [get]
func (h *OrderHandler) ListOrdersHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "invalid user context")
		return
	}

	orders, err := h.service.ListOrders(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, "Orders retrieved successfully", orders)
}
