//go:build integration

package dynamodb_test

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func backfillReviewed() missiondynamo.ReviewedMissionTranslation {
	return missiondynamo.ReviewedMissionTranslation{
		DefinitionCode: "def-rescue", DefinitionVersion: "v1",
		SourceSpanish: domain.LocalizedContent{
			Title: "Rescate", Description: "Salvar al piloto", Rules: []string{"No ayudas"},
			Objectives: []domain.LocalizedObjective{
				{ObjectiveID: "obj-1", Name: "Capturar", Description: "Capturar al piloto"},
			},
		},
		English: domain.LocalizedContent{
			Title: "Rescue", Description: "Save the pilot", Rules: []string{"No aids"},
			Objectives: []domain.LocalizedObjective{
				{ObjectiveID: "obj-1", Name: "Capture", Description: "Capture the pilot"},
			},
		},
	}
}

// spanishBackfillOffer builds a legacy es-only offer whose text matches the
// reviewed source snapshot exactly.
func spanishBackfillOffer(offerID, title, version string) domain.MissionOffer {
	offer := domain.NewMissionOffer(offerID, "BOUNTY_HUNTING", "season-1")
	offer.Title = title
	offer.Description = "Salvar al piloto"
	offer.Rules = []string{"No ayudas"}
	offer.Objectives = []domain.Objective{
		domain.NewObjectiveNamed("obj-1", "Capturar", "Capturar al piloto", "BOUNTY_HUNTING", 10, 0, 2),
	}
	offer.CompletionReward = 50
	offer.DefinitionCode = "def-rescue"
	offer.DefinitionVersion = version
	return offer
}

func rawOfferItem(ctx context.Context, client *missiondynamo.DynamoDBClient, table, offerID string) (map[string]types.AttributeValue, error) {
	item, err := client.Client().GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "OFFER"},
			"SK": &types.AttributeValueMemberS{Value: offerID},
		},
	})
	if err != nil {
		return nil, err
	}
	return item.Item, nil
}

// itemWithoutContentEn returns a copy with only the ContentEn key removed.
func itemWithoutContentEn(item map[string]types.AttributeValue) map[string]types.AttributeValue {
	out := make(map[string]types.AttributeValue, len(item))
	for k, v := range item {
		if k != "ContentEn" {
			out[k] = v
		}
	}
	return out
}
