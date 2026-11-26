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
