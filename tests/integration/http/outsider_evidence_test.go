package http_test

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

// A caller outside the contract roster must not adjudicate evidence, even
// when the body forges a roster member's identity.
func TestConfirmEvidenceEndpoint_RejectsOutsiderDespiteForgedBody(t *testing.T) {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	terms, _ := domain.NewContractTerms([]domain.Objective{domain.NewObjectiveNamed("objective-1", "objective-1", "", "BOUNTY_HUNTING", 100, 0, 0)})
	contract.Terms = terms
	member, _ := domain.NewCallerIdentity("player-1", nil, nil)
	_ = contract.Start(member)

	app := application.NewConfirmObjectiveHandlerWithEvidence(
		fixedContractStore{contract: contract},
		&stagedContractWriter{},
		fixedSeasonStore{season: domain.NewSeason("season-1", domain.SeasonStateActive)},
		&evidenceBlobFake{exists: map[string]bool{"evidence/contract-1/stored": true}},
	)
	auth := missionhttp.NewAuthMiddleware(boundarySession{playerID: "outsider"})
	server := httptest.NewServer(auth.Authenticate(missionhttp.ConfirmEvidenceHandler(app)))
	defer server.Close()

	body := `{"objective_id":"objective-1","unit_id":"target-1","season_id":"season-1","player_id":"player-1"}`
	req, err := http.NewRequest(http.MethodPost,
		server.URL+"/api/v1/missions/contracts/contract-1/evidence", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer any")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}
