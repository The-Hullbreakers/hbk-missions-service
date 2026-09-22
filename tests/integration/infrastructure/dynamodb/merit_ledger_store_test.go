package dynamodb_test

import (
	"context"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func TestMeritLedgerStore_PersistsCredit(t *testing.T) {
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
	credit := domain.MeritCredit{
		TenantID:       "tenant-1",
		SeasonID:       "season-1",
		Category:       "BOUNTY_HUNTING",
		Track:          domain.TrackIndividual,
		RecipientID:    "player-1",
		Amount:         100,
		ContractID:     "contract-1",
		ObjectiveID:    "objective-1",
		UnitID:         "target-1",
		ConfirmationID: "contract-1:objective-1:target-1:player-1",
	}

	if err := store.Save(credit); err != nil {
		t.Fatalf("could not save credit: %v", err)
	}

	total := store.TotalFor("tenant-1", "season-1", "BOUNTY_HUNTING", domain.TrackIndividual, "player-1")
	if total != 100 {
		t.Fatalf("expected total 100, got %d", total)
	}
}
