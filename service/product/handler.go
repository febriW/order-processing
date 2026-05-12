package main

import (
	"encoding/json"
	"net/http"

	"github.com/febriW/order-processing/common/models"
	"github.com/febriW/order-processing/common/utils"
	"github.com/gorilla/mux"
)

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

type ProductHandler struct {
	service *ProductService
}

func NewProductHandler(service *ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

// CreateProductHandler godoc
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
func (h *ProductHandler) CreateProductHandler(w http.ResponseWriter, r *http.Request) {
	var req ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	product, err := h.service.CreateProduct(req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, "Product created successfully", product)
}

// ListProductsHandler godoc
// @Summary List products
// @Tags product
// @Produce json
// @Security BearerAuth
// @Success 200 {object} productListDataResponse
// @Failure 401 {object} productEmptyResponse
// @Failure 500 {object} productEmptyResponse
// @Router /products [get]
func (h *ProductHandler) ListProductsHandler(w http.ResponseWriter, r *http.Request) {
	products, err := h.service.ListProducts()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, "Products retrieved successfully", products)
}

// GetProductHandler godoc
// @Summary Get product by id
// @Tags product
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Success 200 {object} productDataResponse
// @Failure 401 {object} productEmptyResponse
// @Failure 404 {object} productEmptyResponse
// @Router /products/{id} [get]
func (h *ProductHandler) GetProductHandler(w http.ResponseWriter, r *http.Request) {
	product, err := h.service.GetProduct(mux.Vars(r)["id"])
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, "Product retrieved successfully", product)
}

// UpdateProductHandler godoc
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
func (h *ProductHandler) UpdateProductHandler(w http.ResponseWriter, r *http.Request) {
	var req ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	product, err := h.service.UpdateProduct(mux.Vars(r)["id"], req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, "Product updated successfully", product)
}

// DeleteProductHandler godoc
// @Summary Delete product by id
// @Tags product
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Success 200 {object} productEmptyResponse
// @Failure 401 {object} productEmptyResponse
// @Failure 404 {object} productEmptyResponse
// @Router /products/{id} [delete]
func (h *ProductHandler) DeleteProductHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteProduct(mux.Vars(r)["id"]); err != nil {
		utils.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, "Product deleted successfully", nil)
}
