package composition_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	missioncomposition "github.com/hbk/hbk-missions-service/internal/missions/composition"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func TestCompositionServer_ListSeasonsByStatus(t *testing.T) {
	ctx := t.Context()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	if err != nil {
		t.Fatalf("could not create client: %v", err)
	}

	if err := client.CreateMissionsTable(ctx, "hbk-missions-composition"); err != nil {
		t.Fatalf("could not create table: %v", err)
	}
	clearMissionsTable(t, ctx, client, "hbk-missions-composition")

	seasonStore := missiondynamo.NewSeasonStore(client, "hbk-missions-composition")
	if err := seasonStore.Save(domain.NewSeason("season-active", domain.SeasonStateActive)); err != nil {
		t.Fatalf("could not save active season: %v", err)
	}
	if err := seasonStore.Save(domain.NewSeason("season-archived", domain.SeasonStateArchived)); err != nil {
		t.Fatalf("could not save archived season: %v", err)
	}

	server, err := missioncomposition.NewServer(missioncomposition.Config{
		TableName:   "hbk-missions-composition",
		DynamoDB:    client,
		UserSession: fakeUserSession{},
	})
	if err != nil {
		t.Fatalf("could not create server: %v", err)
	}

	ts := httptest.NewServer(server)
	defer ts.Close()

	cases := []struct {
		query        string
		expectedCode int
		expectedIDs  []string
	}{
		{"", http.StatusOK, []string{"season-active"}},
		{"?status=active", http.StatusOK, []string{"season-active"}},
		{"?status=inactive", http.StatusOK, []string{"season-archived"}},
		{"?status=all", http.StatusOK, []string{"season-active", "season-archived"}},
		{"?status=closed", http.StatusBadRequest, nil},
	}

	for _, c := range cases {
		req, err := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/seasons"+c.query, nil)
		if err != nil {
			t.Fatalf("could not create request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer any")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("list seasons request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != c.expectedCode {
			t.Fatalf("expected %d for %q, got %d", c.expectedCode, c.query, resp.StatusCode)
		}

		if c.expectedCode != http.StatusOK {
			continue
		}

		var result []map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("could not decode response for %q: %v", c.query, err)
		}
		if len(result) != len(c.expectedIDs) {
			t.Fatalf("expected %d seasons for %q, got %d", len(c.expectedIDs), c.query, len(result))
		}
		for i, id := range c.expectedIDs {
			if result[i]["SeasonID"] != id {
				t.Fatalf("expected %q at index %d for %q, got %v", id, i, c.query, result[i]["SeasonID"])
			}
		}
	}
}
