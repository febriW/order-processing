package main

import (
	"fmt"
	"testing"

	"github.com/febriW/order-processing/common/auth"
	"github.com/febriW/order-processing/common/models"
)

type fakeAuthRepository struct {
	users map[string]models.User
}

type fakeSessionStore struct {
	sessions map[string]*auth.Claims
}

func newFakeAuthRepository() *fakeAuthRepository {
	return &fakeAuthRepository{
		users: make(map[string]models.User),
	}
}

func newFakeSessionStore() *fakeSessionStore {
	return &fakeSessionStore{
		sessions: make(map[string]*auth.Claims),
	}
}

func (r *fakeAuthRepository) CreateUser(user models.User) error {
	if _, exists := r.users[user.Email]; exists {
		return fmt.Errorf("user already exists")
	}
	if user.Role == "" {
		user.Role = models.RoleBasicUser
	}
	r.users[user.Email] = user
	return nil
}

func (r *fakeAuthRepository) GetUserByEmail(email string) (*models.User, error) {
	user, exists := r.users[email]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return &user, nil
}

func (s *fakeSessionStore) StoreSession(claims *auth.Claims) error {
	s.sessions[claims.ID] = claims
	return nil
}

func (s *fakeSessionStore) ValidateSession(claims *auth.Claims) error {
	if _, exists := s.sessions[claims.ID]; !exists {
		return fmt.Errorf("session is not active")
	}
	return nil
}

func (s *fakeSessionStore) DeleteSession(tokenID string) error {
	delete(s.sessions, tokenID)
	return nil
}

func TestAuthServiceRegisterUser(t *testing.T) {
	repo := newFakeAuthRepository()
	sessionStore := newFakeSessionStore()
	service := NewAuthService(repo, sessionStore)

	user, err := service.RegisterUser(AuthRequest{
		Email:    "customer@example.com",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("RegisterUser returned error: %v", err)
	}

	if user.ID == "" {
		t.Fatal("expected user ID to be generated")
	}
	if user.Email != "customer@example.com" {
		t.Fatalf("expected email customer@example.com, got %s", user.Email)
	}
	if user.Password == "secret123" {
		t.Fatal("expected stored password to be hashed")
	}
	if !auth.CheckPasswordHash("secret123", user.Password) {
		t.Fatal("expected stored password hash to match original password")
	}
}

func TestAuthServiceLoginUser(t *testing.T) {
	hashedPassword, err := auth.HashPassword("secret123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	repo := newFakeAuthRepository()
	repo.users["customer@example.com"] = models.User{
		ID:       "user-1",
		Email:    "customer@example.com",
		Password: hashedPassword,
		Role:     models.RoleBasicUser,
	}
	sessionStore := newFakeSessionStore()
	service := NewAuthService(repo, sessionStore)

	response, err := service.LoginUser(AuthRequest{
		Email:    "customer@example.com",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("LoginUser returned error: %v", err)
	}
	if response.AccessToken == "" {
		t.Fatal("expected access token to be generated")
	}
	if response.RefreshToken == "" {
		t.Fatal("expected refresh token to be generated")
	}

	claims, err := auth.ValidateJWT(response.AccessToken)
	if err != nil {
		t.Fatalf("ValidateJWT returned error: %v", err)
	}
	if claims.UserID != "user-1" {
		t.Fatalf("expected user ID user-1, got %s", claims.UserID)
	}
	if claims.Role != models.RoleBasicUser {
		t.Fatalf("expected role basic, got %s", claims.Role)
	}
	if claims.TokenType != auth.TokenTypeAccess {
		t.Fatalf("expected access token type, got %s", claims.TokenType)
	}
	if _, exists := sessionStore.sessions[claims.ID]; !exists {
		t.Fatal("expected session to be stored after login")
	}
}

func TestAuthServiceValidateAndLogoutToken(t *testing.T) {
	hashedPassword, err := auth.HashPassword("secret123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	repo := newFakeAuthRepository()
	repo.users["customer@example.com"] = models.User{
		ID:       "user-1",
		Email:    "customer@example.com",
		Password: hashedPassword,
		Role:     models.RoleBasicUser,
	}
	service := NewAuthService(repo, newFakeSessionStore())

	response, err := service.LoginUser(AuthRequest{
		Email:    "customer@example.com",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("LoginUser returned error: %v", err)
	}

	validation, err := service.ValidateToken(response.AccessToken)
	if err != nil {
		t.Fatalf("ValidateToken returned error: %v", err)
	}
	if validation.UserID != "user-1" {
		t.Fatalf("expected user ID user-1, got %s", validation.UserID)
	}

	if err := service.LogoutUser(response.AccessToken); err != nil {
		t.Fatalf("LogoutUser returned error: %v", err)
	}

	if _, err := service.ValidateToken(response.AccessToken); err == nil {
		t.Fatal("expected token validation to fail after logout")
	}
}

func TestAuthServiceRefreshToken(t *testing.T) {
	hashedPassword, err := auth.HashPassword("secret123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	repo := newFakeAuthRepository()
	repo.users["customer@example.com"] = models.User{
		ID:       "user-1",
		Email:    "customer@example.com",
		Password: hashedPassword,
		Role:     models.RoleBasicUser,
	}
	service := NewAuthService(repo, newFakeSessionStore())

	loginResponse, err := service.LoginUser(AuthRequest{
		Email:    "customer@example.com",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("LoginUser returned error: %v", err)
	}

	refreshResponse, err := service.RefreshToken(RefreshTokenRequest{
		RefreshToken: loginResponse.RefreshToken,
	})
	if err != nil {
		t.Fatalf("RefreshToken returned error: %v", err)
	}
	if refreshResponse.AccessToken == "" {
		t.Fatal("expected refreshed access token")
	}
	if refreshResponse.RefreshToken == "" {
		t.Fatal("expected rotated refresh token")
	}
	if refreshResponse.RefreshToken == loginResponse.RefreshToken {
		t.Fatal("expected refresh token rotation")
	}
}

func TestAuthServiceLoginUserInvalidEmail(t *testing.T) {
	service := NewAuthService(newFakeAuthRepository())

	_, err := service.LoginUser(AuthRequest{
		Email:    "missing@example.com",
		Password: "secret123",
	})
	if err == nil {
		t.Fatal("expected invalid credentials error")
	}
	if err.Error() != "invalid credentials" {
		t.Fatalf("expected invalid credentials error, got %v", err)
	}
}

func TestAuthServiceLoginUserInvalidPassword(t *testing.T) {
	hashedPassword, err := auth.HashPassword("secret123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	repo := newFakeAuthRepository()
	repo.users["customer@example.com"] = models.User{
		ID:       "user-1",
		Email:    "customer@example.com",
		Password: hashedPassword,
		Role:     models.RoleBasicUser,
	}
	service := NewAuthService(repo)

	_, err = service.LoginUser(AuthRequest{
		Email:    "customer@example.com",
		Password: "wrong-password",
	})
	if err == nil {
		t.Fatal("expected invalid credentials error")
	}
	if err.Error() != "invalid credentials" {
		t.Fatalf("expected invalid credentials error, got %v", err)
	}
}
