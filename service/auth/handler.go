package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/febriW/order-processing/common/models"
	"github.com/febriW/order-processing/common/utils"
)

// @title Auth Service API
// @version 1.0
// @description Authentication endpoints for order-processing.
// @BasePath /
// @schemes http
// @accept json
// @produce json
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

type authUserDataResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    models.User `json:"data"`
}

type authTokenDataResponse struct {
	Status  int          `json:"status"`
	Message string       `json:"message"`
	Data    AuthResponse `json:"data"`
}

type authValidationDataResponse struct {
	Status  int                    `json:"status"`
	Message string                 `json:"message"`
	Data    AuthValidationResponse `json:"data"`
}

type authEmptyResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type AuthHandler struct {
	service *AuthService
}

func NewAuthHandler(service *AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// RegisterHandler godoc
// @Summary Register user
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body AuthRequest true "Register payload"
// @Success 201 {object} authUserDataResponse
// @Failure 400 {object} authEmptyResponse
// @Failure 500 {object} authEmptyResponse
// @Router /auth/register [post]
func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	user, err := h.service.RegisterUser(req)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	user.Password = ""
	utils.RespondWithJSON(w, http.StatusCreated, "User registered successfully", user)
}

// LoginHandler godoc
// @Summary Login user
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body AuthRequest true "Login payload"
// @Success 200 {object} authTokenDataResponse
// @Failure 400 {object} authEmptyResponse
// @Failure 401 {object} authEmptyResponse
// @Router /auth/login [post]
func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	response, err := h.service.LoginUser(req)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, "Login successful", response)
}

// RefreshTokenHandler godoc
// @Summary Refresh token pair
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body RefreshTokenRequest true "Refresh payload"
// @Success 200 {object} authTokenDataResponse
// @Failure 400 {object} authEmptyResponse
// @Failure 401 {object} authEmptyResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	response, err := h.service.RefreshToken(req)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, "Token refreshed successfully", response)
}

// ValidateTokenHandler godoc
// @Summary Validate bearer token
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} authValidationDataResponse
// @Failure 401 {object} authEmptyResponse
// @Router /auth/validate [get]
func (h *AuthHandler) ValidateTokenHandler(w http.ResponseWriter, r *http.Request) {
	token, err := bearerToken(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	response, err := h.service.ValidateToken(token)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, "Token is valid", response)
}

// LogoutHandler godoc
// @Summary Logout user
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} authEmptyResponse
// @Failure 401 {object} authEmptyResponse
// @Router /auth/logout [post]
func (h *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	token, err := bearerToken(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if err := h.service.LogoutUser(token); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, "Logout successful", nil)
}

func bearerToken(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", fmt.Errorf("authorization bearer token is required")
	}

	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", fmt.Errorf("authorization header must use bearer token")
	}

	return parts[1], nil
}
