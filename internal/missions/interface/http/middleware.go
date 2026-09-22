package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/infrastructure/session"
)

type contextKey string

const (
	callerIDKey contextKey = "callerID"
	claimsKey   contextKey = "claims"
)

// AuthMiddleware wraps HTTP handlers with OIDC token validation.
type AuthMiddleware struct {
	userSession session.UserSession
}

// NewAuthMiddleware creates an auth middleware with the given session validator.
func NewAuthMiddleware(userSession session.UserSession) *AuthMiddleware {
	return &AuthMiddleware{userSession: userSession}
}

// Authenticate ensures the request carries a valid token before calling next.
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeAuthError(w)
			return
		}

		userID, claims, err := m.userSession.GetUserIDFromHeader(authHeader)
		if err != nil {
			writeAuthError(w)
			return
		}

		ctx := context.WithValue(r.Context(), callerIDKey, userID)
		ctx = context.WithValue(ctx, claimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func writeAuthError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"code": "UNAUTHORIZED"})
}
