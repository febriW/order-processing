package auth

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

type Claims struct {
	UserID    string `json:"user_id"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func GenerateJWT(userID, role string) (string, error) {
	return GenerateToken(userID, role, TokenTypeAccess, 15*time.Minute)
}

func GenerateRefreshToken(userID, role string) (string, error) {
	return GenerateToken(userID, role, TokenTypeRefresh, 7*24*time.Hour)
}

func GenerateToken(userID, role, tokenType string, ttl time.Duration) (string, error) {
	issuedAt := time.Now()
	expirationTime := issuedAt.Add(ttl)
	claims := &Claims{
		UserID:    userID,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey())
}

func ValidateJWT(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey(), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func jwtKey() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "jwt_secret_key"
	}
	return []byte(secret)
}
