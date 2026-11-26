//go:build integration

package http_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/composition"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

type previewJourneySession struct{ userID string }

func (s previewJourneySession) GetUserIDFromHeader(header string) (string, map[string]any, error) {
	if header != "Bearer preview" {
		return "", nil, errors.New("unauthorized")
	}
	return s.userID, map[string]any{"roles": []any{"member"}}, nil
}

func (previewJourneySession) HasRole(map[string]any, string) bool { return false }

// TEST-0055-006: public routes render the preview through real composition.
func TestSeasonPreviewJourney(t *testing.T) {
	// Arrange
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", "http://localhost:8000")
	require.NoError(t, err)
	result, err := db.Client().Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String("hbk-missions-preview-mid-season"),
		ConsistentRead:         aws.Bool(true),
		KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "CONTRACT"},
		},
		Limit: aws.Int32(1),
	})
	require.NoError(t, err)
	require.NotEmpty(t, result.Items)
	roster := result.Items[0]["Roster"].(*types.AttributeValueMemberL).Value
	require.NotEmpty(t, roster)
	member := roster[0].(*types.AttributeValueMemberS).Value
	handler, err := composition.NewServer(composition.Config{
		TableName: "hbk-missions-preview-mid-season", DynamoDB: db,
		UserSession:     previewJourneySession{userID: member},
		ShutdownContext: ctx, SchedulerTick: time.Hour,
	})
	require.NoError(t, err)
	server := httptest.NewServer(handler)
	defer server.Close()
	client := &http.Client{Timeout: 5 * time.Second}
	cases := []struct {
		name, path, field string
		minimum           int
		expected          any
	}{
		{"current-season", "/api/v1/missions/seasons/current", "SeasonID", 0, "preview-mid-season"},
		{"offers", "/api/v1/missions/offers", "offers", 1, nil},
		{"statistics", "/api/v1/missions/seasons/current/statistics", "Buckets", 4, nil},
		{"leaderboard", "/api/v1/missions/seasons/preview-mid-season/leaderboards/PVP_AIR?track=INDIVIDUAL", "items", 1, nil},
		{"my-contracts", "/api/v1/missions/contracts?season_id=preview-mid-season", "", 1, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			request, err := http.NewRequest(http.MethodGet, server.URL+tc.path, nil)
			require.NoError(t, err)
			request.Header.Set("Authorization", "Bearer preview")
			// Act
			response, err := client.Do(request)
			// Assert
			require.NoError(t, err)
			defer response.Body.Close()
			require.Equal(t, http.StatusOK, response.StatusCode)
			var body any
			require.NoError(t, json.NewDecoder(response.Body).Decode(&body))
			if tc.field == "" {
				assert.GreaterOrEqual(t, len(body.([]any)), tc.minimum)
			} else if tc.expected != nil {
				assert.Equal(t, tc.expected, body.(map[string]any)[tc.field])
			} else {
				assert.GreaterOrEqual(t, len(body.(map[string]any)[tc.field].([]any)), tc.minimum)
			}
		})
	}
	t.Run("missing-auth", func(t *testing.T) {
		// Arrange
		request, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/missions/seasons/current", nil)
		require.NoError(t, err)
		// Act
		response, err := client.Do(request)
		// Assert
		require.NoError(t, err)
		defer response.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
	})
}
