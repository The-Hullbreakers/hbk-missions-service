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

type reviewContractStore struct {
	contract *domain.MissionContract
	saved    *domain.MissionContract
	credits  []domain.MeritCredit
}

func (s *reviewContractStore) Get(string) (*domain.MissionContract, error) { return s.contract, nil }
func (s *reviewContractStore) Save(c *domain.MissionContract) error        { s.saved = c; return nil }
func (s *reviewContractStore) SaveApproved(c *domain.MissionContract, credits []domain.MeritCredit) error {
	s.saved = c
	s.credits = credits
	return nil
}

// reviewServer mirrors the composition wiring for the moderation endpoints.
func reviewServer(session boundarySession, store *reviewContractStore) *httptest.Server {
	submit := application.NewSubmitContractHandler(store, store)
	approve := application.NewApproveContractHandler(store, store)
	reject := application.NewRejectContractHandler(store, store)
	auth := missionhttp.NewAuthMiddleware(session)

	mux := http.NewServeMux()
	mux.Handle("POST /api/v1/missions/contracts/{contract_id}/submit", auth.Authenticate(missionhttp.SubmitContractHandler(submit)))
	mux.Handle("POST /api/v1/missions/contracts/{contract_id}/approve", auth.Authenticate(missionhttp.ApproveContractHandler(approve)))
	mux.Handle("POST /api/v1/missions/contracts/{contract_id}/reject", auth.Authenticate(missionhttp.RejectContractHandler(reject)))
	return httptest.NewServer(mux)
}

// inProgressReviewContract returns a started contract ready for submission.
func inProgressReviewContract() *domain.MissionContract {
	contract := lifecycleContractFixture()
	member, _ := domain.NewCallerIdentity("player-1", nil, nil)
	_ = contract.Start(member)
	return contract
}

// inProgressReviewContractWithEvidence returns a started contract with one
// media-backed staged credit, satisfying the submit evidence requirement.
func inProgressReviewContractWithEvidence(t *testing.T) *domain.MissionContract {
	t.Helper()
	contract := inProgressReviewContract()
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	media, err := domain.NewMediaURI("evidence/contract-1/u1", "contract-1")
	require.NoError(t, err)
	_, err = contract.ConfirmWithMedia("objective-1", "u1", media, season)
	require.NoError(t, err)
	return contract
}

// pendingReviewContract returns a submitted contract with one staged credit.
func pendingReviewContract(t *testing.T) *domain.MissionContract {
	t.Helper()
	contract := lifecycleContractFixture()
	contract.Offer.CompletionReward = 200
	member, _ := domain.NewCallerIdentity("player-1", nil, nil)
	_ = contract.Start(member)
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	media, err := domain.NewMediaURI("evidence/contract-1/unit-1", "contract-1")
	require.NoError(t, err)
	_, err = contract.ConfirmWithMedia("objective-1", "unit-1", media, season)
	require.NoError(t, err)
	err = contract.Submit(member)
	require.NoError(t, err)
	return contract
}

func TestContractReview_SubmitEndpointMovesToPending(t *testing.T) {
	store := &reviewContractStore{contract: inProgressReviewContractWithEvidence(t)}
	server := reviewServer(boundarySession{playerID: "player-1"}, store)
	defer server.Close()

	resp := postLifecycle(t, server, "submit")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.False(t, store.saved == nil || store.saved.State != domain.ContractPendingValidation)
}

// TEST-0063-HTTP-002: an in-progress contract with no media-backed staged
// credit cannot enter review — 422 EVIDENCE_REQUIRED, nothing persisted.
func TestContractReview_SubmitEndpointRejectsWithoutEvidence(t *testing.T) {
	// Arrange
	store := &reviewContractStore{contract: inProgressReviewContract()}
	server := reviewServer(boundarySession{playerID: "player-1"}, store)
	defer server.Close()

	// Act
	resp := postLifecycle(t, server, "submit")
	defer resp.Body.Close()

	// Assert
	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	var body map[string]string
	err := json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)
	require.Equal(t, "EVIDENCE_REQUIRED", body["code"])
	require.False(t, store.saved != nil || store.contract.State != domain.ContractInProgress)
}

func TestContractReview_SubmitEndpointRejectsOutsider(t *testing.T) {
	store := &reviewContractStore{contract: inProgressReviewContract()}
	server := reviewServer(boundarySession{playerID: "outsider"}, store)
	defer server.Close()

	resp := postLifecycle(t, server, "submit")
	defer resp.Body.Close()

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestContractReview_ApproveEndpointReturnsCommittedCredits(t *testing.T) {
	store := &reviewContractStore{contract: pendingReviewContract(t)}
	server := reviewServer(boundarySession{playerID: "mod-1", roles: []any{"moderator"}}, store)
	defer server.Close()

	resp := postLifecycle(t, server, "approve")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	var credits []domain.MeritCredit
	err := json.NewDecoder(resp.Body).Decode(&credits)
	require.NoError(t, err)
	require.Len(t, credits, 2)
	require.False(t, store.saved == nil || store.saved.State != domain.ContractCompleted)
}

func TestContractReview_ApproveEndpointRejectsMember(t *testing.T) {
	store := &reviewContractStore{contract: pendingReviewContract(t)}
	server := reviewServer(boundarySession{playerID: "player-1"}, store)
	defer server.Close()

	resp := postLifecycle(t, server, "approve")
	defer resp.Body.Close()

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	require.Nil(t, store.saved)
}

func TestContractReview_RejectEndpointDisqualifies(t *testing.T) {
	store := &reviewContractStore{contract: pendingReviewContract(t)}
	server := reviewServer(boundarySession{playerID: "mod-1", roles: []any{"moderator"}}, store)
	defer server.Close()

	resp := postLifecycle(t, server, "reject")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.False(t, store.saved == nil || store.saved.State != domain.ContractDisqualified)
}
