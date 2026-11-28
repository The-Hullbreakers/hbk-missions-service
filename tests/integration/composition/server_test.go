package composition_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	missioncomposition "github.com/hbk/hbk-missions-service/internal/missions/composition"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
	"github.com/hbk/hbk-missions-service/internal/missions/infrastructure/session"
)

type fakeUserSession struct{}

var _ session.UserSession = fakeUserSession{}

func (s fakeUserSession) GetUserIDFromHeader(string) (string, map[string]any, error) {
	return "player-1", map[string]any{"tenant_id": "tenant-1"}, nil
}

func (s fakeUserSession) HasRole(map[string]any, string) bool { return true }

// clearMissionsTable deletes all items from the given table. It is used to
// isolate integration tests that share the same local DynamoDB instance.
func clearMissionsTable(t *testing.T, ctx context.Context, client *missiondynamo.DynamoDBClient, tableName string) {
	t.Helper()

	var exclusiveStartKey map[string]types.AttributeValue
	for {
		result, err := client.Client().Scan(ctx, &dynamodb.ScanInput{
			TableName:         aws.String(tableName),
			ExclusiveStartKey: exclusiveStartKey,
		})
		if err != nil {
			t.Fatalf("could not scan table: %v", err)
		}

		for _, item := range result.Items {
			_, err := client.Client().DeleteItem(ctx, &dynamodb.DeleteItemInput{
				TableName: aws.String(tableName),
				Key: map[string]types.AttributeValue{
					"PK": item["PK"],
					"SK": item["SK"],
				},
			})
			if err != nil {
				t.Fatalf("could not delete item: %v", err)
			}
		}

		if result.LastEvaluatedKey == nil {
			break
		}
		exclusiveStartKey = result.LastEvaluatedKey
	}
}

func TestCompositionServer_HealthAndConfirm(t *testing.T) {
	ctx := t.Context()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	if err != nil {
		t.Fatalf("could not create client: %v", err)
	}

	if err := client.CreateMissionsTable(ctx, "hbk-missions-composition"); err != nil {
		t.Fatalf("could not create table: %v", err)
	}

	contractStore := missiondynamo.NewContractStore(client, "hbk-missions-composition")
	seasonStore := missiondynamo.NewSeasonStore(client, "hbk-missions-composition")

	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	terms, _ := domain.NewContractTerms([]domain.Objective{
		domain.NewObjectiveNamed("objective-1", "objective-1", "", "BOUNTY_HUNTING", 100, 0, 0),
	})
	contract.Terms = terms
	member, _ := domain.NewCallerIdentity("player-1", nil, nil)

	contract.Start(member)
	if err := contractStore.Save(contract); err != nil {
		t.Fatalf("could not seed contract: %v", err)
	}

	if err := seasonStore.Save(domain.NewSeason("season-1", domain.SeasonStateActive)); err != nil {
		t.Fatalf("could not seed season: %v", err)
	}

	server, err := missioncomposition.NewServer(missioncomposition.Config{
		TableName:   "hbk-missions-composition",
		DynamoDB:    client,
		UserSession: fakeUserSession{},
		Evidence:    &blobStoreFake{stored: map[string]bool{"evidence/contract-1/e1": true}},
	})
	if err != nil {
		t.Fatalf("could not create server: %v", err)
	}

	ts := httptest.NewServer(server)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatalf("health request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from healthz, got %d", resp.StatusCode)
	}

	body := `{"objective_id":"objective-1","unit_id":"target-1","season_id":"season-1","media_uri":"evidence/contract-1/e1","tenant_id":"tenant-1"}`
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/missions/contracts/contract-1/evidence", strings.NewReader(body))
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer any")

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("confirm request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
}
