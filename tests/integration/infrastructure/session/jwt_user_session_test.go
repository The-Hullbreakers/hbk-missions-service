package session_test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"

	"github.com/hbk/hbk-missions-service/internal/missions/infrastructure/session"
)

func TestJWTUserSession_GetUserIDFromHeader_ValidToken(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("could not generate key: %v", err)
	}

	issuer, shutdown := startFakeOIDCProvider(t, &privateKey.PublicKey)
	defer shutdown()

	claims := jwt.MapClaims{
		"iss": issuer,
		"sub": "player-1",
		"aud": "missions-service",
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Add(-time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-key"
	signed, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("could not sign token: %v", err)
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("NewJWTUserSession panicked: %v", r)
		}
	}()

	s := session.NewJWTUserSession(issuer, "missions-service", []string{"missions-service"})
	userID, parsedClaims, err := s.GetUserIDFromHeader("Bearer " + signed)
	if err != nil {
		t.Fatalf("expected valid token to pass: %v", err)
	}
	if userID != "player-1" {
		t.Fatalf("expected userID player-1, got %s", userID)
	}
	if parsedClaims["sub"] != "player-1" {
		t.Fatalf("expected subject claim in parsed claims")
	}
}

func startFakeOIDCProvider(t *testing.T, publicKey *rsa.PublicKey) (string, func()) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"issuer":   serverURL(r),
				"jwks_uri": serverURL(r) + "/keys",
			})
		case "/keys":
			jwk := jose.JSONWebKey{Key: publicKey, KeyID: "test-key", Algorithm: "RS256", Use: "sig"}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{jwk}})
		default:
			http.NotFound(w, r)
		}
	}))

	return server.URL, server.Close
}

func serverURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, r.Host)
}
