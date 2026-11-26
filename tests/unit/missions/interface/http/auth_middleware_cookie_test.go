package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

type recordingSession struct {
	userID     string
	claims     map[string]any
	lastHeader string
}

func (s *recordingSession) GetUserIDFromHeader(header string) (string, map[string]any, error) {
	s.lastHeader = header
	return s.userID, s.claims, nil
}

func (s *recordingSession) HasRole(map[string]any, string) bool { return true }

func TestAuthMiddleware_AuthenticatesAccessTokenCookie(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := missionhttp.NewAuthMiddleware(fakeUserSession{})
	server := httptest.NewServer(middleware.Authenticate(next))
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/any", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "cookie-token"})
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAuthMiddleware_HeaderTakesPrecedenceOverCookie(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	s := &recordingSession{userID: "player-1", claims: map[string]any{}}
	middleware := missionhttp.NewAuthMiddleware(s)
	server := httptest.NewServer(middleware.Authenticate(next))
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/any", nil)
	req.Header.Set("Authorization", "Bearer header-token")
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "cookie-token"})
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "Bearer header-token", s.lastHeader)
}

func TestAuthMiddleware_RejectsInvalidCookieToken(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := missionhttp.NewAuthMiddleware(failingUserSession{})
	server := httptest.NewServer(middleware.Authenticate(next))
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/any", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "bad"})
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
