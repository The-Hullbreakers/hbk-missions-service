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

func TestAcceptMissionEndpoint_RejectsMalformedBody(t *testing.T) {
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

	req, err := http.NewRequest(http.MethodPost,
		server.URL+"/api/v1/missions/offers/offer-1/accept",
		strings.NewReader(`{"team_id":`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer any")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Nil(t, writer.saved)
}
