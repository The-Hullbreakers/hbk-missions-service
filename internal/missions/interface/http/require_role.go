package http

import (
	"encoding/json"
	nethttp "net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/infrastructure/session"
)

// RequireRole returns middleware that allows the request only when the caller
// claims contain at least one of the given roles. It must run after
// AuthMiddleware so claims are already in the request context.
func RequireRole(userSession session.UserSession, roles ...string) func(nethttp.Handler) nethttp.Handler {
	return func(next nethttp.Handler) nethttp.Handler {
		return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
			claims, ok := r.Context().Value(claimsKey).(map[string]any)
			if !ok {
				writeForbidden(w)
				return
			}

			for _, role := range roles {
				if userSession.HasRole(claims, role) {
					next.ServeHTTP(w, r)
					return
				}
			}

			writeForbidden(w)
		})
	}
}

func writeForbidden(w nethttp.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(nethttp.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]string{"code": "FORBIDDEN"})
}
