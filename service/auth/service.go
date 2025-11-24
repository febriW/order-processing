package main

import (
	"fmt"

	"github.com/febriW/order-processing/common/auth"
	"github.com/febriW/order-processing/common/models"
	"github.com/google/uuid"
)

type AuthService struct {
	repo *AuthRepository
}

func NewAuthService(repo *AuthRepository) *AuthService {
	return &AuthService{repo: repo}
}

// Struct untuk request dan response
type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type AuthResponse struct {
	Token string `json:"token"`
}

func (s *AuthService) RegisterUser(req AuthRequest) (*models.User, error) {
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	newUser := models.User{
		ID:       uuid.NewString(),
		Email:    req.Email,
		Password: hashedPassword,
	}

	if err := s.repo.CreateUser(newUser); err != nil {
		return nil, fmt.Errorf("could not create user: %w", err)
	}

	createdUser, err := s.repo.GetUserByEmail(req.Email)
	if err != nil {
		return nil, err
	}

	return createdUser, nil
}

func (s *AuthService) LoginUser(req AuthRequest) (*AuthResponse, error) {
	user, err := s.repo.GetUserByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	if !auth.CheckPasswordHash(req.Password, user.Password) {
		return nil, fmt.Errorf("invalid credentials")
	}
	token, err := auth.GenerateJWT(user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("could not generate token: %w", err)
	}
	return &AuthResponse{Token: token}, nil
}
