package main

import (
	"fmt"

	"github.com/febriW/order-processing/common/auth"
	"github.com/febriW/order-processing/common/models"
	"github.com/google/uuid"
)

type AuthService struct {
	repo     AuthUserRepository
	sessions AuthSessionStore
}

type AuthUserRepository interface {
	CreateUser(user models.User) error
	GetUserByEmail(email string) (*models.User, error)
}

func NewAuthService(repo AuthUserRepository, sessions ...AuthSessionStore) *AuthService {
	sessionStore := AuthSessionStore(noopSessionStore{})
	if len(sessions) > 0 && sessions[0] != nil {
		sessionStore = sessions[0]
	}
	return &AuthService{repo: repo, sessions: sessionStore}
}

// Struct untuk request dan response
type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Token        string `json:"token,omitempty"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AuthValidationResponse struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
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
		return nil, fmt.Errorf("failed to retrieve created user: %w", err)
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
	accessToken, err := auth.GenerateJWT(user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("could not generate token: %w", err)
	}
	refreshToken, err := auth.GenerateRefreshToken(user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("could not generate refresh token: %w", err)
	}

	accessClaims, err := auth.ValidateJWT(accessToken)
	if err != nil {
		return nil, fmt.Errorf("could not validate generated token: %w", err)
	}
	refreshClaims, err := auth.ValidateJWT(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("could not validate generated refresh token: %w", err)
	}
	if err := s.sessions.StoreSession(accessClaims); err != nil {
		return nil, fmt.Errorf("could not store access session: %w", err)
	}
	if err := s.sessions.StoreSession(refreshClaims); err != nil {
		return nil, fmt.Errorf("could not store refresh session: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Token:        accessToken,
	}, nil
}

func (s *AuthService) ValidateToken(token string) (*AuthValidationResponse, error) {
	claims, err := auth.ValidateJWT(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}
	if claims.TokenType != auth.TokenTypeAccess {
		return nil, fmt.Errorf("invalid token")
	}
	if err := s.sessions.ValidateSession(claims); err != nil {
		return nil, fmt.Errorf("invalid session")
	}
	return &AuthValidationResponse{
		UserID: claims.UserID,
		Role:   claims.Role,
	}, nil
}

func (s *AuthService) RefreshToken(req RefreshTokenRequest) (*AuthResponse, error) {
	claims, err := auth.ValidateJWT(req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}
	if claims.TokenType != auth.TokenTypeRefresh {
		return nil, fmt.Errorf("invalid refresh token")
	}
	if err := s.sessions.ValidateSession(claims); err != nil {
		return nil, fmt.Errorf("invalid refresh session")
	}

	accessToken, err := auth.GenerateJWT(claims.UserID, claims.Role)
	if err != nil {
		return nil, fmt.Errorf("could not generate access token: %w", err)
	}
	newRefreshToken, err := auth.GenerateRefreshToken(claims.UserID, claims.Role)
	if err != nil {
		return nil, fmt.Errorf("could not generate refresh token: %w", err)
	}

	accessClaims, err := auth.ValidateJWT(accessToken)
	if err != nil {
		return nil, fmt.Errorf("could not validate generated access token: %w", err)
	}
	refreshClaims, err := auth.ValidateJWT(newRefreshToken)
	if err != nil {
		return nil, fmt.Errorf("could not validate generated refresh token: %w", err)
	}

	if err := s.sessions.DeleteSession(claims.ID); err != nil {
		return nil, fmt.Errorf("could not rotate refresh session: %w", err)
	}
	if err := s.sessions.StoreSession(accessClaims); err != nil {
		return nil, fmt.Errorf("could not store access session: %w", err)
	}
	if err := s.sessions.StoreSession(refreshClaims); err != nil {
		return nil, fmt.Errorf("could not store refresh session: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		Token:        accessToken,
	}, nil
}

func (s *AuthService) LogoutUser(token string) error {
	claims, err := auth.ValidateJWT(token)
	if err != nil {
		return fmt.Errorf("invalid token")
	}
	if claims.TokenType != auth.TokenTypeAccess {
		return fmt.Errorf("invalid token")
	}
	if err := s.sessions.DeleteSession(claims.ID); err != nil {
		return fmt.Errorf("could not delete session: %w", err)
	}
	return nil
}
