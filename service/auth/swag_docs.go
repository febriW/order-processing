package main

import "github.com/febriW/order-processing/common/models"

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

// registerUser godoc
// @Summary Register user
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body AuthRequest true "Register payload"
// @Success 201 {object} authUserDataResponse
// @Failure 400 {object} authEmptyResponse
// @Failure 500 {object} authEmptyResponse
// @Router /auth/register [post]
func registerUser() {}

// loginUser godoc
// @Summary Login user
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body AuthRequest true "Login payload"
// @Success 200 {object} authTokenDataResponse
// @Failure 400 {object} authEmptyResponse
// @Failure 401 {object} authEmptyResponse
// @Router /auth/login [post]
func loginUser() {}

// refreshToken godoc
// @Summary Refresh token pair
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body RefreshTokenRequest true "Refresh payload"
// @Success 200 {object} authTokenDataResponse
// @Failure 400 {object} authEmptyResponse
// @Failure 401 {object} authEmptyResponse
// @Router /auth/refresh [post]
func refreshToken() {}

// validateToken godoc
// @Summary Validate bearer token
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} authValidationDataResponse
// @Failure 401 {object} authEmptyResponse
// @Router /auth/validate [get]
func validateToken() {}

// logoutUser godoc
// @Summary Logout user
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} authEmptyResponse
// @Failure 401 {object} authEmptyResponse
// @Router /auth/logout [post]
func logoutUser() {}
