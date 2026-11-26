package http_test

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

// policyStoreDouble records saved policies.
type policyStoreDouble struct {
	saved domain.ActiveLimitPolicy
	err   error
}

func (s *policyStoreDouble) SaveActiveLimit(policy domain.ActiveLimitPolicy) error {
	s.saved = policy
	return s.err
}

// policyServer mirrors the composition wiring for the policy route.
func policyServer(session boundarySession, writer *policyStoreDouble) *httptest.Server {
	app := application.NewUpdateActiveLimitHandler(writer)
	auth := missionhttp.NewAuthMiddleware(session)
	requireAdminOrMod := missionhttp.RequireRole(session, "admin", "moderator")

	mux := http.NewServeMux()
	mux.Handle("PUT /api/v1/missions/policies/active-contract-limit",
		auth.Authenticate(requireAdminOrMod(missionhttp.UpdateActiveLimitHandler(app))))
	return httptest.NewServer(mux)
}

func putPolicy(t *testing.T, server *httptest.Server, payload string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPut,
		server.URL+"/api/v1/missions/policies/active-contract-limit",
		bytes.NewBufferString(payload))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer any")
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

// An admin tunes the global limit; it is persisted for the next acceptance.
func TestUpdateActiveLimit_AdminPersists(t *testing.T) {
	// Arrange
	writer := &policyStoreDouble{}
	server := policyServer(boundarySession{playerID: "admin-1", roles: []any{"admin"}}, writer)
	defer server.Close()

	// Act
	resp := putPolicy(t, server, `{"default":3,"overrides":{"BOUNTY_HUNTING":2}}`)

	// Assert
	require.False(t, resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent)
	require.False(t, writer.saved.Default() != 3 || writer.saved.EffectiveLimit("BOUNTY_HUNTING") != 2)
}

// A non-admin cannot change the policy.
func TestUpdateActiveLimit_MemberDenied(t *testing.T) {
	// Arrange
	server := policyServer(boundarySession{playerID: "player-1"}, &policyStoreDouble{})
	defer server.Close()

	// Act
	resp := putPolicy(t, server, `{"default":3}`)

	// Assert
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// A non-positive limit is rejected with 422 INVALID_POLICY.
func TestUpdateActiveLimit_NonPositiveRejected(t *testing.T) {
	// Arrange
	server := policyServer(boundarySession{playerID: "admin-1", roles: []any{"admin"}}, &policyStoreDouble{})
	defer server.Close()

	// Act
	resp := putPolicy(t, server, `{"default":0}`)

	// Assert
	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	var body map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&body)
	require.Equal(t, "INVALID_POLICY", body["code"])
}
