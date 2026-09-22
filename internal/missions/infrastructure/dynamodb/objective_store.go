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

// ObjectiveStore loads objectives from DynamoDB.
type ObjectiveStore struct {
	client    *DynamoDBClient
	tableName string
}

// NewObjectiveStore creates a DynamoDB objective store.
func NewObjectiveStore(client *DynamoDBClient, tableName string) *ObjectiveStore {
	return &ObjectiveStore{client: client, tableName: tableName}
}

// Save persists an objective.
func (s *ObjectiveStore) Save(objective domain.Objective) error {
	ctx := context.Background()

	item := map[string]types.AttributeValue{
		"PK":           &types.AttributeValueMemberS{Value: "OBJECTIVE"},
		"SK":           &types.AttributeValueMemberS{Value: objective.ObjectiveID},
		"Category":     &types.AttributeValueMemberS{Value: objective.Category},
		"MeritPerUnit": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", objective.MeritPerUnit)},
	}

	_, err := s.client.Client().PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to save objective: %w", err)
	}

	return nil
}

// Get loads an objective by identity.
func (s *ObjectiveStore) Get(objectiveID string) (domain.Objective, error) {
	ctx := context.Background()

	result, err := s.client.Client().GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "OBJECTIVE"},
			"SK": &types.AttributeValueMemberS{Value: objectiveID},
		},
	})
	if err != nil {
		return domain.Objective{}, fmt.Errorf("failed to get objective: %w", err)
	}
	if result.Item == nil {
		return domain.Objective{}, errors.New("objective not found")
	}

	return domain.NewObjective(
		getString(result.Item, "SK"),
		getString(result.Item, "Category"),
		getInt(result.Item, "MeritPerUnit"),
	), nil
}
