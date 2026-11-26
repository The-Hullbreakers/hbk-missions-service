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

type queueStore struct {
	contracts []domain.MissionContract
	calls     int
}

func (s *queueStore) ListPendingReview() ([]domain.MissionContract, error) {
	s.calls++
	return s.contracts, nil
}

// queueServer mirrors the composition wiring for the review queue route.
func queueServer(session boundarySession, store *queueStore) *httptest.Server {
	queue := application.NewListPendingReviewHandler(store)
	auth := missionhttp.NewAuthMiddleware(session)

	mux := http.NewServeMux()
	mux.Handle("GET /api/v1/missions/review-queue", auth.Authenticate(missionhttp.ReviewQueueHandler(queue)))
	return httptest.NewServer(mux)
}

// getReviewQueue performs the authenticated queue request.
func getReviewQueue(t *testing.T, server *httptest.Server) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/missions/review-queue", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer any")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

// reviewQueueEntry is the response projection a reviewer consumes.
type reviewQueueEntry struct {
	ContractID     string   `json:"contract_id"`
	OfferID        string   `json:"offer_id"`
	DefinitionCode string   `json:"definition_code"`
	Title          string   `json:"title"`
	State          string   `json:"state"`
	Roster         []string `json:"roster"`
	SeasonID       string   `json:"season_id"`
	TeamID         string   `json:"team_id"`
	ConfirmedCount int      `json:"confirmed_count"`
	PendingCredits int      `json:"pending_credits"`
}

func TestReviewQueue_ReviewerListsPendingContracts(t *testing.T) {
	pending := pendingReviewContract(t)
	store := &queueStore{contracts: []domain.MissionContract{*pending}}
	server := queueServer(boundarySession{playerID: "mod-1", roles: []any{"moderator"}}, store)
	defer server.Close()

	resp := getReviewQueue(t, server)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	var entries []reviewQueueEntry
	err := json.NewDecoder(resp.Body).Decode(&entries)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	entry := entries[0]
	require.False(t, entry.ContractID != pending.ContractID || entry.State != string(domain.ContractPendingValidation))
	require.False(t, entry.PendingCredits != 1 || entry.ConfirmedCount != 1)
	require.False(t, len(entry.Roster) != 1 || entry.Roster[0] != "player-1")
}

func TestReviewQueue_MemberDenied(t *testing.T) {
	store := &queueStore{contracts: []domain.MissionContract{*pendingReviewContract(t)}}
	server := queueServer(boundarySession{playerID: "player-1"}, store)
	defer server.Close()

	resp := getReviewQueue(t, server)
	defer resp.Body.Close()

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	require.Equal(t, 0, store.calls)
}

func TestReviewQueue_EmptyQueueReturnsEmptyList(t *testing.T) {
	store := &queueStore{contracts: []domain.MissionContract{}}
	server := queueServer(boundarySession{playerID: "mod-1", roles: []any{"admin"}}, store)
	defer server.Close()

	resp := getReviewQueue(t, server)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	var entries []reviewQueueEntry
	err := json.NewDecoder(resp.Body).Decode(&entries)
	require.NoError(t, err)
	require.Len(t, entries, 0)
}
