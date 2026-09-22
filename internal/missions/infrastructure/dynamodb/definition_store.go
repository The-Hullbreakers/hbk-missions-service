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

func definitionFromItem(item map[string]types.AttributeValue) domain.MissionDefinition {
	definitionID := getString(item, "DefinitionID")
	if definitionID == "" {
		definitionID = getString(item, "SK")
	}
	return domain.MissionDefinition{
		DefinitionID: definitionID,
		Category:     getString(item, "Category"),
		Mode:         domain.MissionMode(getString(item, "Mode")),
		Capacity:     getInt(item, "Capacity"),
		State:        domain.DefinitionState(getString(item, "State")),
		Version:      getString(item, "Version"),
	}
}

// Save persists a definition.
func (s *DefinitionStore) Save(definition domain.MissionDefinition) error {
	ctx := context.Background()

	item := map[string]types.AttributeValue{
		"PK":           &types.AttributeValueMemberS{Value: "DEFINITION"},
		"SK":           &types.AttributeValueMemberS{Value: definition.DefinitionID},
		"DefinitionID": &types.AttributeValueMemberS{Value: definition.DefinitionID},
		"Category":     &types.AttributeValueMemberS{Value: definition.Category},
		"Capacity":     &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", definition.Capacity)},
		"Mode":         &types.AttributeValueMemberS{Value: string(definition.Mode)},
		"State":        &types.AttributeValueMemberS{Value: string(definition.State)},
		"Version":      &types.AttributeValueMemberS{Value: definition.Version},
	}

	_, err := s.client.Client().PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to save definition: %w", err)
	}

	if definition.State == domain.DefinitionStatePublished && definition.Version != "" {
		versioned := map[string]types.AttributeValue{
			"PK":           &types.AttributeValueMemberS{Value: "DEFINITION"},
			"SK":           &types.AttributeValueMemberS{Value: definition.DefinitionID + "#" + definition.Version},
			"DefinitionID": &types.AttributeValueMemberS{Value: definition.DefinitionID},
			"Category":     &types.AttributeValueMemberS{Value: definition.Category},
			"Capacity":     &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", definition.Capacity)},
			"Mode":         &types.AttributeValueMemberS{Value: string(definition.Mode)},
			"State":        &types.AttributeValueMemberS{Value: string(definition.State)},
			"Version":      &types.AttributeValueMemberS{Value: definition.Version},
		}
		_, err = s.client.Client().PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(s.tableName),
			Item:      versioned,
		})
		if err != nil {
			return fmt.Errorf("failed to save definition version: %w", err)
		}
	}

	return nil
}

// Get loads the current definition by identity.
func (s *DefinitionStore) Get(tenantID, definitionID string) (domain.MissionDefinition, error) {
	_ = tenantID
	ctx := context.Background()

	result, err := s.client.Client().GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "DEFINITION"},
			"SK": &types.AttributeValueMemberS{Value: definitionID},
		},
	})
	if err != nil {
		return domain.MissionDefinition{}, fmt.Errorf("failed to get definition: %w", err)
	}
	if result.Item == nil {
		return domain.MissionDefinition{}, domain.ErrDefinitionNotFound
	}

	return definitionFromItem(result.Item), nil
}

// GetVersion loads a specific published definition version.
func (s *DefinitionStore) GetVersion(tenantID, definitionID, version string) (domain.MissionDefinition, error) {
	_ = tenantID
	ctx := context.Background()

	result, err := s.client.Client().GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "DEFINITION"},
			"SK": &types.AttributeValueMemberS{Value: definitionID + "#" + version},
		},
	})
	if err != nil {
		return domain.MissionDefinition{}, fmt.Errorf("failed to get definition version: %w", err)
	}
	if result.Item == nil {
		return domain.MissionDefinition{}, domain.ErrDefinitionNotFound
	}

	def := definitionFromItem(result.Item)
	if !def.State.IsPublished() {
		return domain.MissionDefinition{}, domain.ErrDefinitionNotPublished
	}
	return def, nil
}
