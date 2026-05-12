package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/febriW/order-processing/common/models"
	"github.com/febriW/order-processing/common/utils"
)

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

func (m *AuthMiddleware) RequireRoles(roles ...string) func(http.Handler) http.Handler {
	allowedRoles := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowedRoles[role] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				utils.RespondWithError(w, http.StatusUnauthorized, "Authorization bearer token is required")
				return
			}

			userRole, err := m.validateToken(token)
			if err != nil {
				utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
				return
			}

			if _, allowed := allowedRoles[userRole]; !allowed {
				utils.RespondWithError(w, http.StatusForbidden, "You do not have permission to access this resource")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (m *AuthMiddleware) validateToken(authorization string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, m.validateURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", authorization)

	response, err := m.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not validate token")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("invalid token")
	}

	var payload validateResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("invalid auth response")
	}
	if strings.TrimSpace(payload.Data.Role) == "" {
		return "", fmt.Errorf("invalid auth response")
	}

	return payload.Data.Role, nil
}

func basicUserRoles() []string {
	return []string{models.RoleBasicUser, models.RoleAdmin, models.RoleSuperAdmin}
}

func adminRoles() []string {
	return []string{models.RoleAdmin, models.RoleSuperAdmin}
}
