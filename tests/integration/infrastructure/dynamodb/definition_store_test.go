package dynamodb_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func TestDefinitionStore_RoundTrip(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	if err != nil {
		t.Fatalf("could not create client: %v", err)
	}

	if err := client.CreateMissionsTable(ctx, "hbk-missions"); err != nil {
		t.Fatalf("could not create table: %v", err)
	}

	store := missiondynamo.NewDefinitionStore(client, "hbk-missions")
	definition, _ := domain.NewMissionDefinition("bounty-template", "BOUNTY_HUNTING", 10, domain.MissionModeSolo)

	if err := store.Save(definition); err != nil {
		t.Fatalf("could not save definition: %v", err)
	}

	loaded, err := store.Get("tenant-1", "bounty-template")
	if err != nil {
		t.Fatalf("could not get definition: %v", err)
	}
	if loaded.DefinitionID != "bounty-template" || loaded.State != domain.DefinitionStateDraft {
		t.Fatalf("unexpected definition: %+v", loaded)
	}
}

func TestDefinitionStore_GetVersion(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	if err != nil {
		t.Fatalf("could not create client: %v", err)
	}

	if err := client.CreateMissionsTable(ctx, "hbk-missions"); err != nil {
		t.Fatalf("could not create table: %v", err)
	}

	store := missiondynamo.NewDefinitionStore(client, "hbk-missions")
	definition, _ := domain.NewMissionDefinition("bounty-template", "BOUNTY_HUNTING", 10, domain.MissionModeSolo)
	_ = definition.Publish("v1")

	if err := store.Save(definition); err != nil {
		t.Fatalf("could not save definition: %v", err)
	}

	loaded, err := store.GetVersion("tenant-1", "bounty-template", "v1")
	if err != nil {
		t.Fatalf("could not get definition version: %v", err)
	}
	if loaded.DefinitionID != "bounty-template" || loaded.Version != "v1" {
		t.Fatalf("unexpected versioned definition: %+v", loaded)
	}
}

func TestDefinitionStore_GetVersionNotPublished(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	if err != nil {
		t.Fatalf("could not create client: %v", err)
	}

	if err := client.CreateMissionsTable(ctx, "hbk-missions"); err != nil {
		t.Fatalf("could not create table: %v", err)
	}

	_, err = client.Client().PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("hbk-missions"),
		Item: map[string]types.AttributeValue{
			"PK":           &types.AttributeValueMemberS{Value: "DEFINITION"},
			"SK":           &types.AttributeValueMemberS{Value: "bounty-template#v1"},
			"DefinitionID": &types.AttributeValueMemberS{Value: "bounty-template"},
			"Category":     &types.AttributeValueMemberS{Value: "BOUNTY_HUNTING"},
			"Capacity":     &types.AttributeValueMemberN{Value: "10"},
			"Mode":         &types.AttributeValueMemberS{Value: string(domain.MissionModeSolo)},
			"State":        &types.AttributeValueMemberS{Value: string(domain.DefinitionStateDraft)},
			"Version":      &types.AttributeValueMemberS{Value: "v1"},
		},
	})
	if err != nil {
		t.Fatalf("could not seed versioned draft: %v", err)
	}

	store := missiondynamo.NewDefinitionStore(client, "hbk-missions")
	_, err = store.GetVersion("tenant-1", "bounty-template", "v1")
	if !errors.Is(err, domain.ErrDefinitionNotPublished) {
		t.Fatalf("expected ErrDefinitionNotPublished, got %v", err)
	}
}

func TestDefinitionStore_GetVersionUnknown(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	if err != nil {
		t.Fatalf("could not create client: %v", err)
	}

	if err := client.CreateMissionsTable(ctx, "hbk-missions"); err != nil {
		t.Fatalf("could not create table: %v", err)
	}

	store := missiondynamo.NewDefinitionStore(client, "hbk-missions")
	_, err = store.GetVersion("tenant-1", "missing", "v1")
	if !errors.Is(err, domain.ErrDefinitionNotFound) {
		t.Fatalf("expected ErrDefinitionNotFound, got %v", err)
	}
}
