package dynamodb

import (
	"context"
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
		"Title":             &types.AttributeValueMemberS{Value: offer.Title},
		"Description":       &types.AttributeValueMemberS{Value: offer.Description},
		"Category":          &types.AttributeValueMemberS{Value: offer.Category},
		"Mode":              &types.AttributeValueMemberS{Value: string(offer.Mode)},
		"CompletionReward":  &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", offer.CompletionReward)},
		"Objectives":        objectivesToAttribute(offer.Objectives),
		"Rules":             stringSliceToAttribute(offer.Rules),
		"SeasonID":          &types.AttributeValueMemberS{Value: offer.SeasonID},
		"DefinitionCode":    &types.AttributeValueMemberS{Value: offer.DefinitionCode},
		"DefinitionVersion": &types.AttributeValueMemberS{Value: offer.DefinitionVersion},
	}
	putEnglishContent(item, offer.ContentEn)

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
func (s *OfferStore) Get(offerID string) (domain.MissionOffer, error) {
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
		return domain.MissionOffer{}, domain.ErrOfferNotAvailable
	}

	return mapOfferItem(result.Item), nil
}

// List returns the offers matching the board filter. Structural filters are
// pushed into a FilterExpression. The free-text query stays in memory.
func (s *OfferStore) List(filter domain.OfferListFilter) ([]domain.MissionOffer, error) {
	ctx := context.Background()

	input := &dynamodb.QueryInput{
		TableName:              aws.String(s.tableName),
		KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "OFFER"},
		},
	}
	applyOfferFilter(input, filter)

	result, err := s.client.Client().Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list offers: %w", err)
	}

	offers := make([]domain.MissionOffer, 0, len(result.Items))
	for _, item := range result.Items {
		offer := mapOfferItem(item)
		if !filter.Match(offer) {
			continue
		}
		offers = append(offers, offer)
	}
	return offers, nil
}
