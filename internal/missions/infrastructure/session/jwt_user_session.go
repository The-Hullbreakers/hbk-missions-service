package session

import (
	"context"
	"fmt"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
)

// JWTUserSession validates OIDC ID tokens from HBK Identity.
type JWTUserSession struct {
	verifier         *oidc.IDTokenVerifier
	clientID         string
	allowedAudiences []string
}

// NewJWTUserSession creates a verifier from the issuer or returns an unconfigured session.
func NewJWTUserSession(issuer, clientID string, allowedAudiences []string) *JWTUserSession {
	if issuer == "" || clientID == "" {
		return &JWTUserSession{verifier: nil, clientID: clientID, allowedAudiences: allowedAudiences}
	}

	provider, err := oidc.NewProvider(context.Background(), issuer)
	if err != nil {
		panic(fmt.Sprintf("failed to create OIDC provider: %v", err))
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID:          clientID,
		SkipClientIDCheck: true,
	})

	return &JWTUserSession{verifier: verifier, clientID: clientID, allowedAudiences: allowedAudiences}
}

// GetUserIDFromHeader extracts and verifies a Bearer token.
func (s *JWTUserSession) GetUserIDFromHeader(authHeader string) (string, map[string]any, error) {
	if s.verifier == nil {
		return "", nil, fmt.Errorf("OIDC not configured - cannot verify token")
	}

	if authHeader == "" {
		return "", nil, fmt.Errorf("authorization header missing")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", nil, fmt.Errorf("invalid authorization header format")
	}

	token := parts[1]

	idToken, err := s.verifier.Verify(context.Background(), token)
	if err != nil {
		return "", nil, fmt.Errorf("failed to verify token: %w", err)
	}

	var claims jwt.MapClaims
	if err := idToken.Claims(&claims); err != nil {
		return "", nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	if err := s.ValidateAudience(claims); err != nil {
		return "", nil, err
	}

	return idToken.Subject, claims, nil
}

// ValidateAudience checks the token audience against allowed audiences.
func (s *JWTUserSession) ValidateAudience(claims map[string]any) error {
	if len(s.allowedAudiences) == 0 {
		return nil
	}

	aud, ok := claims["aud"]
	if !ok {
		return fmt.Errorf("audience claim missing")
	}

	var audList []string
	switch v := aud.(type) {
	case string:
		audList = []string{v}
	case []string:
		audList = v
	case []interface{}:
		audList = make([]string, 0, len(v))
		for _, a := range v {
			if str, ok := a.(string); ok {
				audList = append(audList, str)
			}
		}
	default:
		return fmt.Errorf("invalid audience claim format")
	}

	for _, allowedAud := range s.allowedAudiences {
		for _, tokenAud := range audList {
			if tokenAud == allowedAud {
				return nil
			}
		}
	}

	return fmt.Errorf("token audience %v not in allowed audiences %v", audList, s.allowedAudiences)
}
