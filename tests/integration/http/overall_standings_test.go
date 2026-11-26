//go:build integration

package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func overallRouteFixture(t *testing.T) (*httptest.Server, error) {
	return languageSessionServer(t, previewJourneySession{userID: "kai"}, func(db *missiondynamo.DynamoDBClient, table string) error {
		if err := missiondynamo.NewSeasonStore(db, table).Save(domain.NewSeason("season-1", domain.SeasonStateActive)); err != nil {
			return err
		}
		ledger := missiondynamo.NewMeritLedgerStore(db, table)
		for _, credit := range []domain.MeritCredit{
			{SeasonID: "season-1", Track: domain.TrackIndividual, Category: "PVP_AIR", RecipientID: "mira", Amount: 120, ContractID: "c1", ObjectiveID: "o1", ConfirmationID: "confirm-1"},
			{SeasonID: "season-1", Track: domain.TrackIndividual, Category: "RACING", RecipientID: "mira", Amount: 80, ContractID: "c1", ObjectiveID: "o2", ConfirmationID: "confirm-2"},
			{SeasonID: "season-1", Track: domain.TrackIndividual, Category: "PVP_AIR", RecipientID: "kai", Amount: 190, ContractID: "c2", ObjectiveID: "o1", ConfirmationID: "confirm-3"},
		} {
			if err := ledger.Save(credit); err != nil {
				return err
			}
		}
		return nil
	})
}

func TestOverallStandingsRoute_ReturnsSeasonAndCallerRankThroughComposition(t *testing.T) {
	// Arrange
	server, setupErr := overallRouteFixture(t)
	require.NoError(t, setupErr)
	request, requestErr := http.NewRequest(http.MethodGet, server.URL+"/api/v1/missions/seasons/season-1/standings/overall?limit=1", nil)
	require.NoError(t, requestErr)
	request.Header.Set("Authorization", "Bearer preview")
	// Act
	response, err := http.DefaultClient.Do(request)
	// Assert
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, http.StatusOK, response.StatusCode)
	var body struct {
		SeasonID     string `json:"season_id"`
		Track        string `json:"track"`
		Participants int    `json:"participants"`
		Items        []struct {
			Rank        int    `json:"rank"`
			RecipientID string `json:"recipient_id"`
			Total       int    `json:"total"`
		} `json:"items"`
		NextCursor string `json:"next_cursor"`
		MyStanding *struct {
			Rank        int    `json:"rank"`
			RecipientID string `json:"recipient_id"`
			Total       int    `json:"total"`
		} `json:"my_standing"`
	}
	require.NoError(t, json.NewDecoder(response.Body).Decode(&body))
	require.Equal(t, "season-1", body.SeasonID)
	require.Equal(t, "INDIVIDUAL", body.Track)
	require.Equal(t, 2, body.Participants)
	require.Len(t, body.Items, 1)
	require.Equal(t, "mira", body.Items[0].RecipientID)
	require.Equal(t, 200, body.Items[0].Total)
	require.NotEmpty(t, body.NextCursor)
	require.NotNil(t, body.MyStanding)
	require.Equal(t, "kai", body.MyStanding.RecipientID)
	require.Equal(t, 2, body.MyStanding.Rank)
}

func TestOverallStandingsRoute_RequiresSession(t *testing.T) {
	// Arrange
	server, setupErr := overallRouteFixture(t)
	require.NoError(t, setupErr)
	request, requestErr := http.NewRequest(http.MethodGet, server.URL+"/api/v1/missions/seasons/season-1/standings/overall", nil)
	require.NoError(t, requestErr)
	// Act
	response, err := http.DefaultClient.Do(request)
	// Assert
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, http.StatusUnauthorized, response.StatusCode)
}

func TestOverallStandingsRoute_RejectsBadLimitAndCursor(t *testing.T) {
	// Arrange
	server, setupErr := overallRouteFixture(t)
	require.NoError(t, setupErr)
	cases := []struct {
		name, query string
	}{
		{"zero-limit", "?limit=0"},
		{"oversized-limit", "?limit=51"},
		{"invalid-cursor", "?cursor=not-base64"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			request, requestErr := http.NewRequest(http.MethodGet, server.URL+"/api/v1/missions/seasons/season-1/standings/overall"+tc.query, nil)
			require.NoError(t, requestErr)
			request.Header.Set("Authorization", "Bearer preview")
			// Act
			response, err := http.DefaultClient.Do(request)
			// Assert
			require.NoError(t, err)
			defer response.Body.Close()
			require.Equal(t, http.StatusBadRequest, response.StatusCode)
		})
	}
}

func TestOverallStandingsRoute_UnknownSeasonReturns404(t *testing.T) {
	// Arrange
	server, setupErr := overallRouteFixture(t)
	require.NoError(t, setupErr)
	request, requestErr := http.NewRequest(http.MethodGet, server.URL+"/api/v1/missions/seasons/season-missing/standings/overall", nil)
	require.NoError(t, requestErr)
	request.Header.Set("Authorization", "Bearer preview")
	// Act
	response, err := http.DefaultClient.Do(request)
	// Assert
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, http.StatusNotFound, response.StatusCode)
}
