package http_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

type fixedDefinitionLister struct{ definitions []domain.MissionDefinition }

func (r fixedDefinitionLister) List() ([]domain.MissionDefinition, error) {
	return r.definitions, nil
}

func newDefinitionsTestServer(definitions []domain.MissionDefinition, roles []interface{}) *httptest.Server {
	app := application.NewListDefinitionsHandler(fixedDefinitionLister{definitions: definitions})
	guarded := missionhttp.RequireRole(definitionsSession{roles: roles}, "admin", "moderator")(missionhttp.ListDefinitionsHandler(app))
	auth := missionhttp.NewAuthMiddleware(definitionsSession{roles: roles})
	return httptest.NewServer(auth.Authenticate(guarded))
}

// definitionsSession maps claim roles through a real membership check.
type definitionsSession struct{ roles []interface{} }

func (s definitionsSession) GetUserIDFromHeader(string) (string, map[string]any, error) {
	return "admin-1", map[string]any{"roles": s.roles}, nil
}

func (s definitionsSession) HasRole(claims map[string]any, role string) bool {
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

func getDefinitions(t *testing.T, server *httptest.Server) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/missions/definitions", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer any")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func TestListDefinitionsEndpoint_ReturnsDefinitionsForAdmin(t *testing.T) {
	definition, err := domain.NewMissionDefinition("def-bounty", "Defensa de puesto", "Protege la posición", "BOUNTY_HUNTING", domain.MissionModeSolo, 0, nil, nil)
	require.NoError(t, err)
	server := newDefinitionsTestServer([]domain.MissionDefinition{definition}, []interface{}{"admin"})
	defer server.Close()

	resp := getDefinitions(t, server)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	var result []map[string]any
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.False(t, len(result) != 1 || result[0]["DefinitionCode"] != "def-bounty")
}

func TestListDefinitionsEndpoint_RejectsMemberRole(t *testing.T) {
	server := newDefinitionsTestServer(nil, []interface{}{"member"})
	defer server.Close()

	resp := getDefinitions(t, server)
	defer resp.Body.Close()

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}
