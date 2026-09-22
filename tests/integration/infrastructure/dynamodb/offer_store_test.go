package dynamodb_test

import (
	"context"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func TestOfferStore_RoundTrip(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	if err != nil {
		t.Fatalf("could not create client: %v", err)
	}

	if err := client.CreateMissionsTable(ctx, "hbk-missions"); err != nil {
		t.Fatalf("could not create table: %v", err)
	}

	store := missiondynamo.NewOfferStore(client, "hbk-missions")
	definition, _ := domain.NewMissionDefinition("bounty-template", "BOUNTY_HUNTING", 10, domain.MissionModeSolo)
	_ = definition.Publish("v1")
	offer, _ := domain.NewMissionOfferFromDefinition("offer-1", "season-1", definition, 0)

	if err := store.Save(offer); err != nil {
		t.Fatalf("could not save offer: %v", err)
	}

	loaded, err := store.Get("tenant-1", "offer-1")
	if err != nil {
		t.Fatalf("could not get offer: %v", err)
	}
	if loaded.OfferID != "offer-1" || loaded.DefinitionID != "bounty-template" || loaded.DefinitionVersion != "v1" {
		t.Fatalf("unexpected offer: %+v", loaded)
	}

	_, err = store.Get("tenant-1", "missing")
	if err == nil {
		t.Fatalf("expected an error for missing offer")
	}
}
