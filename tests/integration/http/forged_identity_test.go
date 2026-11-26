package http_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

// Boundary evidence for CHG-0020: identity fields in request bodies are never
// trusted; caller always comes from the authenticated session.

type boundarySession struct {
	playerID string
	roles    []any
	teamIDs  []any
}

func (s boundarySession) GetUserIDFromHeader(string) (string, map[string]any, error) {
	return s.playerID, map[string]any{"roles": s.roles, "team_ids": s.teamIDs}, nil
}

func (s boundarySession) HasRole(claims map[string]any, role string) bool {
	raw, _ := claims["roles"].([]any)
	for _, r := range raw {
		if r == role {
			return true
		}
	}
	return false
}

type fixedOfferReader struct{ offer domain.MissionOffer }

func (r fixedOfferReader) Get(string) (domain.MissionOffer, error) { return r.offer, nil }

type emptyActiveContracts struct{}

func (emptyActiveContracts) ListActiveByPlayer(string) ([]domain.MissionContract, error) {
	return nil, nil
}

type emptyContractHistory struct{}

func (emptyContractHistory) ListByPlayer(string, domain.ContractListFilter) ([]domain.MissionContract, error) {
	return nil, nil
}

type emptyDefinitions struct{}

func (emptyDefinitions) Get(string) (domain.MissionDefinition, error) {
	return domain.MissionDefinition{}, domain.ErrDefinitionNotFound
}

func (emptyDefinitions) GetVersion(string, string) (domain.MissionDefinition, error) {
	return domain.MissionDefinition{}, domain.ErrDefinitionNotFound
}

type recordingContractWriter struct{ saved *domain.MissionContract }

func (w *recordingContractWriter) CommitAcceptance(contract *domain.MissionContract, _ domain.MissionOffer, _ int) error {
	w.saved = contract
	return nil
}

func TestAcceptMissionEndpoint_DerivesIdentityFromSession(t *testing.T) {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	writer := &recordingContractWriter{}
	app := application.NewAcceptMissionHandler(
		domain.NewActiveLimitPolicy(3, nil),
		fixedOfferReader{offer: offer},
		emptyActiveContracts{},
		emptyContractHistory{},
		emptyDefinitions{},
		fixedSeasonStore{season: domain.NewSeason("season-1", domain.SeasonStateActive)},
		writer,
		nil,
	)
	auth := missionhttp.NewAuthMiddleware(boundarySession{playerID: "player-1"})
	server := httptest.NewServer(auth.Authenticate(missionhttp.AcceptMissionHandler(app)))
	defer server.Close()

	// Forged identity fields: player must come from the session.
	body := `{"player_id":"victim","team_id":""}`
	req, err := http.NewRequest(http.MethodPost,
		server.URL+"/api/v1/missions/offers/offer-1/accept", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer any")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotNil(t, writer.saved)
	require.False(t, !writer.saved.IncludesPlayer("player-1") || writer.saved.IncludesPlayer("victim"))

	var result map[string]any
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Equal(t, "offer-1:player-1:1", result["ContractID"])
}
