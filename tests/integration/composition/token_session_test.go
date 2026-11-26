package composition_test

import (
	"strings"

	"github.com/hbk/hbk-missions-service/internal/missions/infrastructure/session"
)

// tokenSession derives player/role claims from the bearer token
// "<player>[:role,...]". Test-only stand-in for the real OIDC validator.
type tokenSession struct{}

var _ session.UserSession = tokenSession{}

func (tokenSession) GetUserIDFromHeader(header string) (string, map[string]any, error) {
	token := strings.TrimPrefix(header, "Bearer ")
	parts := strings.Split(token, ":")
	claims := map[string]any{}
	if len(parts) > 1 && parts[1] != "" {
		roles := make([]any, 0)
		for _, role := range strings.Split(parts[1], ",") {
			roles = append(roles, role)
		}
		claims["roles"] = roles
	}
	return parts[0], claims, nil
}

func (tokenSession) HasRole(claims map[string]any, role string) bool {
	raw, _ := claims["roles"].([]any)
	for _, r := range raw {
		if r == role {
			return true
		}
	}
	return false
}

func token(player string, roles ...string) string {
	return "Bearer " + player + ":" + strings.Join(roles, ",")
}
