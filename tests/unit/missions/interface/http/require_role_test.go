package http_test

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"

	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

// rolesSession reports the real claim membership so RequireRole is exercised
// through the session port, not a fixed boolean.
type rolesSession struct {
	claims map[string]any
}

func (s rolesSession) GetUserIDFromHeader(string) (string, map[string]any, error) {
	return "player-1", s.claims, nil
}

func (s rolesSession) HasRole(claims map[string]any, role string) bool {
	roles, ok := claims["roles"].([]interface{})
	if !ok {
		return false
	}
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

func requireRoleServer(claims map[string]any, roles ...string) *httptest.Server {
	session := rolesSession{claims: claims}
	auth := missionhttp.NewAuthMiddleware(session)
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return httptest.NewServer(auth.Authenticate(missionhttp.RequireRole(session, roles...)(next)))
}

func getWithAuth(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer test-token")
	return http.DefaultClient.Do(req)
}

func TestRequireRole_AllowsCallerWithRole(t *testing.T) {
	server := requireRoleServer(map[string]any{"roles": []interface{}{"admin"}}, "admin")
	defer server.Close()

	resp, err := getWithAuth(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRequireRole_AllowsAnyListedRole(t *testing.T) {
	server := requireRoleServer(map[string]any{"roles": []interface{}{"moderator"}}, "admin", "moderator")
	defer server.Close()

	resp, err := getWithAuth(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRequireRole_RejectsCallerWithoutRole(t *testing.T) {
	server := requireRoleServer(map[string]any{"roles": []interface{}{"member"}}, "admin")
	defer server.Close()

	resp, err := getWithAuth(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}
