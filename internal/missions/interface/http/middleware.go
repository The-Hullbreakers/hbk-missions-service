package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
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
			cookie, err := r.Cookie("access_token")
			if err != nil || cookie.Value == "" {
				writeAuthError(w)
				return
			}
			authHeader = "Bearer " + cookie.Value
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

// CallerIdentityFromContext rebuilds the domain caller identity from the
// authenticated session values stored by AuthMiddleware. The returned error is
// non-nil when required claims are missing.
func CallerIdentityFromContext(ctx context.Context) (domain.CallerIdentity, error) {
	userID, _ := ctx.Value(callerIDKey).(string)
	claims, _ := ctx.Value(claimsKey).(map[string]any)
	return domain.NewCallerIdentity(
		userID,
		stringSliceClaim(claims, "roles"),
		stringSliceClaim(claims, "team_ids"),
	)
}

func stringSliceClaim(claims map[string]any, key string) []string {
	raw, ok := claims[key].([]interface{})
	if !ok {
		if single, ok := claims[key].(string); ok && single != "" {
			return []string{single}
		}
		return nil
	}
	values := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok {
			values = append(values, s)
		}
	}
	return values
}
