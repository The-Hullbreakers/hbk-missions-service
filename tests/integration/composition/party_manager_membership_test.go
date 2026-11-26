package composition_test

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"

	missioncomposition "github.com/hbk/hbk-missions-service/internal/missions/composition"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
	"github.com/hbk/hbk-missions-service/internal/missions/infrastructure/session"
)

// teamSession authenticates player-1 as a member of team-1.
type teamSession struct{}

var _ session.UserSession = teamSession{}

func (teamSession) GetUserIDFromHeader(string) (string, map[string]any, error) {
	return "player-1", map[string]any{"team_ids": []any{"team-1"}}, nil
}

func (teamSession) HasRole(map[string]any, string) bool { return true }

func newMembershipFixture(t *testing.T, partyManagerURL string) *httptest.Server {
	t.Helper()
	ctx := t.Context()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	err = client.CreateMissionsTable(ctx, "hbk-missions-composition")
	require.NoError(t, err)
	clearMissionsTable(t, ctx, client, "hbk-missions-composition")

	seasonStore := missiondynamo.NewSeasonStore(client, "hbk-missions-composition")
	err = seasonStore.Save(domain.NewSeason("season-1", domain.SeasonStateActive))
	require.NoError(t, err)
	offerStore := missiondynamo.NewOfferStore(client, "hbk-missions-composition")
	err = offerStore.Save(domain.NewTeamMissionOffer("offer-team-1", "BOUNTY_HUNTING", "season-1"))
	require.NoError(t, err)

	server, err := missioncomposition.NewServer(missioncomposition.Config{
		TableName:       "hbk-missions-composition",
		DynamoDB:        client,
		UserSession:     teamSession{},
		PartyManagerURL: partyManagerURL,
	})
	require.NoError(t, err)
	return httptest.NewServer(server)
}

// SCN-MCF-MEMBER-VALID: Party Manager confirms the teammate — contract created.
func TestCompositionServer_TeamAcceptWithVerifiedTeammate(t *testing.T) {
	party := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer party.Close()

	ts := newMembershipFixture(t, party.URL)
	defer ts.Close()

	resp, _ := postJSON(t, ts.URL+"/api/v1/missions/offers/offer-team-1/accept",
		`{"team_id":"team-1","teammates":["player-2"]}`)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

// SCN-MCF-MEMBER-INVALID: definitive 404 from Party Manager → 422.
func TestCompositionServer_TeamAcceptNonMemberRejected(t *testing.T) {
	party := httptest.NewServer(http.NotFoundHandler())
	defer party.Close()

	ts := newMembershipFixture(t, party.URL)
	defer ts.Close()

	resp, payload := postJSON(t, ts.URL+"/api/v1/missions/offers/offer-team-1/accept",
		`{"team_id":"team-1","teammates":["player-9"]}`)
	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	require.Equal(t, "TEAMMATE_NOT_MEMBER", payload["code"])
}

// SCN-MCF-MEMBER-UNAVAILABLE: Party Manager down → fail closed with 503.
func TestCompositionServer_TeamAcceptMembershipUnavailable(t *testing.T) {
	party := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer party.Close()

	ts := newMembershipFixture(t, party.URL)
	defer ts.Close()

	resp, payload := postJSON(t, ts.URL+"/api/v1/missions/offers/offer-team-1/accept",
		`{"team_id":"team-1","teammates":["player-2"]}`)
	require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	require.Equal(t, "MEMBERSHIP_UNAVAILABLE", payload["code"])
}
