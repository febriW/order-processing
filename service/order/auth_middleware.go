package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/febriW/order-processing/common/models"
	"github.com/febriW/order-processing/common/utils"
)

type contextKey string

const userIDContextKey contextKey = "user_id"

type AuthMiddleware struct {
	validateURL string
	client      *http.Client
}

type validateResponse struct {
	Status int `json:"status"`
	Data   struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	} `json:"data"`
}

func NewAuthMiddleware(validateURL string) *AuthMiddleware {
	return &AuthMiddleware{
		validateURL: validateURL,
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (m *AuthMiddleware) RequireRoles(next http.HandlerFunc, roles ...string) http.Handler {
	allowedRoles := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowedRoles[role] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			utils.RespondWithError(w, http.StatusUnauthorized, "Authorization bearer token is required")
			return
		}

		userID, userRole, err := m.validateToken(token)
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}

		if _, allowed := allowedRoles[userRole]; !allowed {
			utils.RespondWithError(w, http.StatusForbidden, "You do not have permission to access this resource")
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) validateToken(authorization string) (string, string, error) {
	req, err := http.NewRequest(http.MethodGet, m.validateURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", authorization)

	response, err := m.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("could not validate token")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("invalid token")
	}

	var payload validateResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", "", fmt.Errorf("invalid auth response")
	}
	if strings.TrimSpace(payload.Data.Role) == "" || strings.TrimSpace(payload.Data.UserID) == "" {
		return "", "", fmt.Errorf("invalid auth response")
	}

	return payload.Data.UserID, payload.Data.Role, nil
}

func userIDFromContext(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(userIDContextKey).(string)
	if !ok || strings.TrimSpace(value) == "" {
		return "", false
	}
	return value, true
}

func basicUserRoles() []string {
	return []string{models.RoleBasicUser, models.RoleAdmin, models.RoleSuperAdmin}
}
