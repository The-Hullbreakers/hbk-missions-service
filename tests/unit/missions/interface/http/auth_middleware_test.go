package http_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/infrastructure/session"
	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

type fakeUserSession struct{}

var _ session.UserSession = fakeUserSession{}

func (s fakeUserSession) GetUserIDFromHeader(string) (string, map[string]any, error) {
	return "player-1", map[string]any{}, nil
}

func (s fakeUserSession) HasRole(map[string]any, string) bool { return true }

type failingUserSession struct{}

var _ session.UserSession = failingUserSession{}

func (failingUserSession) GetUserIDFromHeader(string) (string, map[string]any, error) {
	return "", nil, errors.New("invalid token")
}
func (failingUserSession) HasRole(map[string]any, string) bool { return false }

func TestAuthMiddleware_RequiresAuthorization(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := missionhttp.NewAuthMiddleware(fakeUserSession{})
	server := httptest.NewServer(middleware.Authenticate(next))
	defer server.Close()

	resp, err := http.Post(server.URL+"/any", "application/json", nil)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestAuthMiddleware_AllowsAuthorizedRequest(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := missionhttp.NewAuthMiddleware(fakeUserSession{})
	server := httptest.NewServer(middleware.Authenticate(next))
	defer server.Close()

	req, err := http.NewRequest(http.MethodPost, server.URL+"/any", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer valid-token")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAuthMiddleware_RejectsInvalidToken(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := missionhttp.NewAuthMiddleware(failingUserSession{})
	server := httptest.NewServer(middleware.Authenticate(next))
	defer server.Close()

	req, _ := http.NewRequest(http.MethodPost, server.URL+"/any", nil)
	req.Header.Set("Authorization", "Bearer bad")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid token, got %d", resp.StatusCode)
	}
}
