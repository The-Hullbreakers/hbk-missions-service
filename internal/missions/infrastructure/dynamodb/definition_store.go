package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// DefinitionStore loads and persists mission definitions.
type DefinitionStore struct {
	client    *DynamoDBClient
	tableName string
}

// NewDefinitionStore creates a DynamoDB definition store.
func NewDefinitionStore(client *DynamoDBClient, tableName string) *DefinitionStore {
	return &DefinitionStore{client: client, tableName: tableName}
}

// definitionPartition returns the single-scope partition key for definitions.
func definitionPartition() string {
	return "DEFINITION"
}

func definitionItem(definition domain.MissionDefinition, sortKey string) map[string]types.AttributeValue {
	item := map[string]types.AttributeValue{
		"PK":               &types.AttributeValueMemberS{Value: definitionPartition()},
		"SK":               &types.AttributeValueMemberS{Value: sortKey},
		"DefinitionCode":   &types.AttributeValueMemberS{Value: definition.DefinitionCode},
		"Title":            &types.AttributeValueMemberS{Value: definition.Title},
		"Description":      &types.AttributeValueMemberS{Value: definition.Description},
		"Category":         &types.AttributeValueMemberS{Value: definition.Category},
		"Mode":             &types.AttributeValueMemberS{Value: string(definition.Mode)},
		"CompletionReward": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", definition.CompletionReward)},
		"State":            &types.AttributeValueMemberS{Value: string(definition.State)},
		"Version":          &types.AttributeValueMemberS{Value: definition.Version},
		"Objectives":       objectivesToAttribute(definition.Objectives),
		"Rules":            stringSliceToAttribute(definition.Rules),
	}
	putEnglishContent(item, definition.ContentEn)
	return item
}

// Save persists a definition.
func (s *DefinitionStore) Save(definition domain.MissionDefinition) error {
	ctx := context.Background()

	if _, err := s.client.Client().PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      definitionItem(definition, definition.DefinitionCode),
	}); err != nil {
		return fmt.Errorf("failed to save definition: %w", err)
	}

	if definition.State == domain.DefinitionStatePublished && definition.Version != "" {
		if _, err := s.client.Client().PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(s.tableName),
			Item:      definitionItem(definition, definition.DefinitionCode+"#"+definition.Version),
		}); err != nil {
			return fmt.Errorf("failed to save definition version: %w", err)
		}
	}

	return nil
}
