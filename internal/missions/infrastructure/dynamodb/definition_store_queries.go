package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func definitionFromItem(item map[string]types.AttributeValue) domain.MissionDefinition {
	return domain.MissionDefinition{
		DefinitionCode:   getString(item, "DefinitionCode"),
		Title:            getString(item, "Title"),
		Description:      getString(item, "Description"),
		Category:         getString(item, "Category"),
		Mode:             domain.MissionMode(getString(item, "Mode")),
		CompletionReward: getInt(item, "CompletionReward"),
		State:            domain.DefinitionState(getString(item, "State")),
		Version:          getString(item, "Version"),
		Objectives:       objectivesFromItem(item, "Objectives"),
		Rules:            getStringSlice(item, "Rules"),
		ContentEn:        englishContentFromItem(item),
	}
}

// List returns all definition items, including published version entries.
func (s *DefinitionStore) List() ([]domain.MissionDefinition, error) {
	ctx := context.Background()

	result, err := s.client.Client().Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(s.tableName),
		KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: definitionPartition()},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list definitions: %w", err)
	}

	definitions := make([]domain.MissionDefinition, 0, len(result.Items))
	for _, item := range result.Items {
		definitions = append(definitions, definitionFromItem(item))
	}
	return definitions, nil
}

// Get loads the current definition by code.
func (s *DefinitionStore) Get(definitionCode string) (domain.MissionDefinition, error) {
	ctx := context.Background()

	result, err := s.client.Client().GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: definitionPartition()},
			"SK": &types.AttributeValueMemberS{Value: definitionCode},
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
func (s *DefinitionStore) GetVersion(definitionCode, version string) (domain.MissionDefinition, error) {
	ctx := context.Background()

	result, err := s.client.Client().GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: definitionPartition()},
			"SK": &types.AttributeValueMemberS{Value: definitionCode + "#" + version},
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
