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

// phasedContractFixture builds a two-phase contract: phase 1 mandatory obj-a +
// optional obj-x; phase 2 mandatory obj-b.
func phasedContractFixture() *domain.MissionContract {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY", "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})

	objA := domain.NewObjectiveNamed("obj-a", "mandatory a", "", "BOUNTY", 10, 0, 0)
	objX := domain.NewObjectiveNamed("obj-x", "optional x", "", "BOUNTY", 5, 0, 0)
	objB := domain.NewObjectiveNamed("obj-b", "mandatory b", "", "BOUNTY", 20, 0, 0)
	terms, _ := domain.NewContractTermsPhased([]domain.Phase{
		{Mandatory: []domain.Objective{objA}, Optional: []domain.Objective{objX}},
		{Mandatory: []domain.Objective{objB}},
	})
	contract.Terms = terms
	member, _ := domain.NewCallerIdentity("player-1", nil, nil)
	_ = contract.Start(member)
	return contract
}

// The detail projection exposes the phase structure and progression.
func TestContractDetail_ExposesPhases(t *testing.T) {
	reader := &detailReader{contract: phasedContractFixture()}
	server := detailServer(boundarySession{playerID: "player-1"}, reader)
	defer server.Close()

	resp := getContractDetail(t, server, "contract-1")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	var detail struct {
		CurrentPhase int `json:"current_phase"`
		Phases       []struct {
			Mandatory []struct {
				ObjectiveID string `json:"objective_id"`
			} `json:"mandatory"`
			Optional []struct {
				ObjectiveID string `json:"objective_id"`
			} `json:"optional"`
		} `json:"phases"`
	}
	err := json.NewDecoder(resp.Body).Decode(&detail)
	require.NoError(t, err)
	require.Equal(t, 0, detail.CurrentPhase)
	require.Len(t, detail.Phases, 2)
	require.False(t, detail.Phases[0].Mandatory[0].ObjectiveID != "obj-a" || detail.Phases[0].Optional[0].ObjectiveID != "obj-x")
	require.Equal(t, "obj-b", detail.Phases[1].Mandatory[0].ObjectiveID)
}

// An out-of-phase confirmation returns 422 PHASE_NOT_UNLOCKED.
func TestConfirmEvidence_Phase2ObjectiveRejected(t *testing.T) {
	// Arrange — phased contract wired into the confirm handler.
	contract := phasedContractFixture()
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	app := application.NewConfirmObjectiveHandlerWithEvidence(
		fixedContractStore{contract: contract},
		&stagedContractWriter{},
		fixedSeasonStore{season: season},
		&evidenceBlobFake{exists: map[string]bool{"evidence/contract-1/stored": true}})
	auth := missionhttp.NewAuthMiddleware(roleSessionFake{playerID: "player-1"})
	server := httptest.NewServer(auth.Authenticate(missionhttp.ConfirmEvidenceHandler(app)))
	defer server.Close()

	// Act — obj-b belongs to phase 2 while phase 1 is incomplete
	resp := postEvidence(t, server, `{"objective_id":"obj-b","unit_id":"u1","season_id":"season-1","media_uri":"evidence/contract-1/stored"}`)

	// Assert
	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	var body map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&body)
	require.Equal(t, "PHASE_NOT_UNLOCKED", body["code"])
}
