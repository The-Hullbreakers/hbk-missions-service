package http_test

import (
	"net/http"
	"testing"
)

func TestLeaderboardEndpoint_RejectsUnauthorizedRequest(t *testing.T) {
	server := newLeaderboardTestServer(nil, nil, rejectingFakeSession{})
	defer server.Close()

	resp := getLeaderboard(t, server, "?track=individual")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestLeaderboardEndpoint_DeniesWhenTenantMissing(t *testing.T) {
	server := newLeaderboardTestServer(nil, nil, leaderboardFakeSession{tenantID: ""})
	defer server.Close()

	resp := getLeaderboard(t, server, "?track=individual")
	defer resp.Body.Close()

	assertCode(t, resp, http.StatusForbidden, "TENANT_BOUNDARY_DENIED")
}
