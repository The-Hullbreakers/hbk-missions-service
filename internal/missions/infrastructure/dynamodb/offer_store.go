package dynamodb

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// OfferStore loads and persists mission offers.
type OfferStore struct {
	client    *DynamoDBClient
	tableName string
}

// NewOfferStore creates a DynamoDB offer store.
func NewOfferStore(client *DynamoDBClient, tableName string) *OfferStore {
	return &OfferStore{client: client, tableName: tableName}
}

// Save persists an offer.
func (s *OfferStore) Save(offer domain.MissionOffer) error {
	ctx := context.Background()

	item := map[string]types.AttributeValue{
		"PK":                &types.AttributeValueMemberS{Value: "OFFER"},
		"SK":                &types.AttributeValueMemberS{Value: offer.OfferID},
		"Category":          &types.AttributeValueMemberS{Value: offer.Category},
		"Capacity":          &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", offer.Capacity)},
		"Mode":              &types.AttributeValueMemberS{Value: string(offer.Mode)},
		"SeasonID":          &types.AttributeValueMemberS{Value: offer.SeasonID},
		"DefinitionID":      &types.AttributeValueMemberS{Value: offer.DefinitionID},
		"DefinitionVersion": &types.AttributeValueMemberS{Value: offer.DefinitionVersion},
	}

	_, err := s.client.Client().PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to save offer: %w", err)
	}

	return nil
}

// Get loads an offer by identity.
func (s *OfferStore) Get(tenantID, offerID string) (domain.MissionOffer, error) {
	_ = tenantID
	ctx := context.Background()

	result, err := s.client.Client().GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "OFFER"},
			"SK": &types.AttributeValueMemberS{Value: offerID},
		},
	})
	if err != nil {
		return domain.MissionOffer{}, fmt.Errorf("failed to get offer: %w", err)
	}
	if result.Item == nil {
		return domain.MissionOffer{}, errors.New("offer not found")
	}

	id := getString(result.Item, "SK")
	definitionID := getString(result.Item, "DefinitionID")
	definitionVersion := getString(result.Item, "DefinitionVersion")
	if definitionID == "" {
		definitionID = id
		definitionVersion = "legacy"
	}

	return domain.MissionOffer{
		OfferID:           id,
		Category:          getString(result.Item, "Category"),
		Capacity:          getInt(result.Item, "Capacity"),
		Mode:              domain.MissionMode(getString(result.Item, "Mode")),
		SeasonID:          getString(result.Item, "SeasonID"),
		DefinitionID:      definitionID,
		DefinitionVersion: definitionVersion,
	}, nil
}
