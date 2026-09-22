package http_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestLeaderboardEndpoint_ReturnsStandings(t *testing.T) {
	server := newLeaderboardTestServer([]domain.Standing{
		{RecipientID: "player-1", Total: 150, CompletedMissions: 3, UniqueObjectives: 4, Track: domain.TrackIndividual},
		{RecipientID: "player-2", Total: 100, Track: domain.TrackIndividual},
	}, nil, leaderboardFakeSession{tenantID: "tenant-1"})
	defer server.Close()

	resp := getLeaderboard(t, server, "?track=individual&limit=5&min_merits=10")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("could not decode body: %v", err)
	}
	items, ok := body["items"].([]any)
	if !ok || len(items) != 2 {
		t.Fatalf("expected 2 items, got %v", body)
	}
	first := items[0].(map[string]any)
	if first["completed_missions"].(float64) != 3 {
		t.Fatalf("expected completed_missions 3, got %v", first["completed_missions"])
	}
}

func TestLeaderboardEndpoint_RejectsMissingTrack(t *testing.T) {
	server := newLeaderboardTestServer(nil, nil, leaderboardFakeSession{tenantID: "tenant-1"})
	defer server.Close()

	resp := getLeaderboard(t, server, "")
	defer resp.Body.Close()

	assertCode(t, resp, http.StatusBadRequest, "INVALID_QUERY_PARAMETER")
}

func TestLeaderboardEndpoint_RejectsInvalidTrack(t *testing.T) {
	server := newLeaderboardTestServer(nil, nil, leaderboardFakeSession{tenantID: "tenant-1"})
	defer server.Close()

	resp := getLeaderboard(t, server, "?track=invalid")
	defer resp.Body.Close()

	assertCode(t, resp, http.StatusBadRequest, "INVALID_QUERY_PARAMETER")
}

func TestLeaderboardEndpoint_ReturnsNotFoundForMissingSeason(t *testing.T) {
	server := newLeaderboardTestServer(nil, fmt.Errorf("%w: season-2", domain.ErrSeasonNotFound), leaderboardFakeSession{tenantID: "tenant-1"})
	defer server.Close()

	resp := getLeaderboard(t, server, "?track=individual")
	defer resp.Body.Close()

	assertCode(t, resp, http.StatusNotFound, "SEASON_NOT_FOUND")
}

func TestLeaderboardEndpoint_RejectsInvalidCursor(t *testing.T) {
	server := newLeaderboardTestServer(nil, nil, leaderboardFakeSession{tenantID: "tenant-1"})
	defer server.Close()

	resp := getLeaderboard(t, server, "?track=individual&cursor=not-base64")
	defer resp.Body.Close()

	assertCode(t, resp, http.StatusBadRequest, "INVALID_QUERY_PARAMETER")
}

func assertCode(t *testing.T, resp *http.Response, status int, code string) {
	t.Helper()
	if resp.StatusCode != status {
		t.Fatalf("expected %d, got %d", status, resp.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("could not decode error body: %v", err)
	}
	if body["code"] != code {
		t.Fatalf("expected code %s, got %s", code, body["code"])
	}
}
