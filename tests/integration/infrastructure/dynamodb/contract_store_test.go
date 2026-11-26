package dynamodb_test

import (
	"context"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func TestContractStore_RoundTrip(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	if err != nil {
		t.Fatalf("could not create client: %v", err)
	}

	if err := client.CreateMissionsTable(ctx, "hbk-missions"); err != nil {
		t.Fatalf("could not create table: %v", err)
	}
	clearMissionsTable(t, ctx, client, "hbk-missions")

	store := missiondynamo.NewContractStore(client, "hbk-missions")
	seedActiveSeason(t, client, "hbk-missions", "season-1")
	definition, _ := domain.NewMissionDefinition("bounty-template", "T", "D", "BOUNTY_HUNTING", domain.MissionModeSolo, 0, nil, nil)
	definition = englishize(definition)
	_ = definition.Publish("v1")
	offer, _ := domain.NewMissionOfferFromDefinition("offer-1", "season-1", definition)
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	member, _ := domain.NewCallerIdentity("player-1", nil, nil)

	contract.Start(member)

	if err := store.Save(contract); err != nil {
		t.Fatalf("could not save contract: %v", err)
	}

	loaded, err := store.Get("contract-1")
	if err != nil {
		t.Fatalf("could not get contract: %v", err)
	}

	if loaded.ContractID != "contract-1" {
		t.Fatalf("expected contract-1, got %s", loaded.ContractID)
	}
	if loaded.State != domain.ContractInProgress {
		t.Fatalf("expected IN_PROGRESS, got %s", loaded.State)
	}
	if loaded.Offer.OfferID != "offer-1" || loaded.Offer.DefinitionCode != "bounty-template" {
		t.Fatalf("unexpected offer in contract: %+v", loaded.Offer)
	}
	roster := loaded.Roster()
	if len(roster) != 1 || roster[0] != "player-1" {
		t.Fatalf("expected roster [player-1], got %v", roster)
	}

	active, err := store.ListActiveByPlayer("player-1")
	if err != nil {
		t.Fatalf("could not list active contracts: %v", err)
	}
	if len(active) != 1 || active[0].ContractID != "contract-1" {
		t.Fatalf("unexpected active contracts: %+v", active)
	}

	// A contract for a different player or category should not be active for player-1.
	other, _ := domain.NewMissionOfferFromDefinition("offer-2", "season-1", definition)
	otherContract := domain.NewMissionContract("contract-2", other, []string{"player-2"})
	otherContract.Start(member)
	if err := store.Save(otherContract); err != nil {
		t.Fatalf("could not seed second contract: %v", err)
	}

	active2, err := store.ListActiveByPlayer("player-1")
	if err != nil {
		t.Fatalf("could not list active contracts: %v", err)
	}
	if len(active2) != 1 {
		t.Fatalf("expected one active contract, got %d", len(active2))
	}

	_, err = store.Get("missing")
	if err == nil {
		t.Fatalf("expected an error for missing contract")
	}
}
