package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

func acceptServer(t *testing.T, offer domain.MissionOffer, session boundarySession) (*httptest.Server, *recordingContractWriter) {
	t.Helper()
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
	auth := missionhttp.NewAuthMiddleware(session)
	return httptest.NewServer(auth.Authenticate(missionhttp.AcceptMissionHandler(app))), writer
}

func postAccept(t *testing.T, url, body string) (int, map[string]string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer any")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var payload map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&payload)
	return resp.StatusCode, payload
}

// SCN-MCF-TEAM-SOLO-REJECT over HTTP: a team_id on a SOLO offer resolves as
// 422 TEAM_NOT_ALLOWED — the exploit is rejected at the boundary.
func TestAcceptMissionEndpoint_RejectsTeamOnSoloOffer(t *testing.T) {
	server, writer := acceptServer(t,
		domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1"),
		boundarySession{playerID: "player-1", teamIDs: []any{"team-1"}},
	)
	defer server.Close()

	status, payload := postAccept(t, server.URL+"/api/v1/missions/offers/offer-1/accept", `{"team_id":"team-1"}`)
	require.Equal(t, http.StatusUnprocessableEntity, status)
	require.Equal(t, "TEAM_NOT_ALLOWED", payload["code"])
	require.Nil(t, writer.saved)
}

// Declared teammates that the membership source cannot verify resolve as
// 422 TEAMMATE_NOT_MEMBER.
func TestAcceptMissionEndpoint_RejectsUnverifiableTeammate(t *testing.T) {
	server, writer := acceptServer(t,
		domain.NewTeamMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1"),
		boundarySession{playerID: "player-1", teamIDs: []any{"team-1"}},
	)
	defer server.Close()

	status, payload := postAccept(t, server.URL+"/api/v1/missions/offers/offer-1/accept",
		`{"team_id":"team-1","teammates":["player-2"]}`)
	require.Equal(t, http.StatusUnprocessableEntity, status)
	require.Equal(t, "TEAMMATE_NOT_MEMBER", payload["code"])
	require.Nil(t, writer.saved)
}
