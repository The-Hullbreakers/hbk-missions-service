package session

// UserSession validates OIDC tokens and extracts caller identity and roles.
type UserSession interface {
	GetUserIDFromHeader(authHeader string) (string, map[string]any, error)
	HasRole(claims map[string]any, role string) bool
}
