package dynamodb_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

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

func TestMeritLedgerStore_ListStandingsByBucket(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	if err != nil {
		t.Fatalf("could not create client: %v", err)
	}

	if err := client.CreateMissionsTable(ctx, "hbk-missions"); err != nil {
		t.Fatalf("could not create table: %v", err)
	}
	clearMissionsTable(t, ctx, client, "hbk-missions")

	store := missiondynamo.NewMeritLedgerStore(client, "hbk-missions")
	seedActiveSeason(t, client, "hbk-missions", "season-1")
	for _, credit := range []domain.MeritCredit{
		{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, RecipientID: "player-1", Amount: 100, ContractID: "contract-1", ObjectiveID: "objective-1", UnitID: "target-1", ConfirmationID: "c1"},
		{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, RecipientID: "player-1", Amount: 50, ContractID: "contract-1", ObjectiveID: "objective-1", UnitID: "target-2", ConfirmationID: "c2"},
		{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, RecipientID: "player-2", Amount: 80, ContractID: "contract-2", ObjectiveID: "objective-1", UnitID: "target-3", ConfirmationID: "c3"},
	} {
		if err := store.Save(credit); err != nil {
			t.Fatalf("could not save credit: %v", err)
		}
	}

	standings, err := store.ListStandings("season-1", "BOUNTY_HUNTING", domain.TrackIndividual, 0)
	if err != nil {
		t.Fatalf("could not list standings: %v", err)
	}
	if len(standings) != 2 {
		t.Fatalf("expected 2 standings, got %d", len(standings))
	}
	if standings[0].Total != 150 || standings[0].RecipientID != "player-1" {
		t.Fatalf("expected player-1 with 150, got %s with %d", standings[0].RecipientID, standings[0].Total)
	}
	if standings[1].Total != 80 || standings[1].RecipientID != "player-2" {
		t.Fatalf("expected player-2 with 80, got %s with %d", standings[1].RecipientID, standings[1].Total)
	}
}
