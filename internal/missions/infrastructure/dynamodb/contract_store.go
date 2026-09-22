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

// ContractStore loads and persists mission contracts.
type ContractStore struct {
	client    *DynamoDBClient
	tableName string
}

// NewContractStore creates a DynamoDB contract store.
func NewContractStore(client *DynamoDBClient, tableName string) *ContractStore {
	return &ContractStore{client: client, tableName: tableName}
}

// Get loads a contract by identity.
func (s *ContractStore) Get(contractID string) (*domain.MissionContract, error) {
	ctx := context.Background()

	result, err := s.client.Client().GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "CONTRACT"},
			"SK": &types.AttributeValueMemberS{Value: contractID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get contract: %w", err)
	}
	if result.Item == nil {
		return nil, errors.New("contract not found")
	}

	return mapToContract(result.Item)
}

// Save persists a contract.
func (s *ContractStore) Save(contract *domain.MissionContract) error {
	ctx := context.Background()

	definitionID := contract.Offer.DefinitionID
	definitionVersion := contract.Offer.DefinitionVersion
	if definitionID == "" {
		definitionID = contract.Offer.OfferID
		definitionVersion = "legacy"
	}

	item := map[string]types.AttributeValue{
		"PK":                &types.AttributeValueMemberS{Value: "CONTRACT"},
		"SK":                &types.AttributeValueMemberS{Value: contract.ContractID},
		"OfferID":           &types.AttributeValueMemberS{Value: contract.Offer.OfferID},
		"Category":          &types.AttributeValueMemberS{Value: contract.Offer.Category},
		"Capacity":          &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", contract.Offer.Capacity)},
		"Mode":              &types.AttributeValueMemberS{Value: string(contract.Offer.Mode)},
		"OfferSeasonID":     &types.AttributeValueMemberS{Value: contract.Offer.SeasonID},
		"DefinitionID":      &types.AttributeValueMemberS{Value: definitionID},
		"DefinitionVersion": &types.AttributeValueMemberS{Value: definitionVersion},
		"Roster":            stringSliceToAttribute(contract.Roster()),
		"TeamID":            &types.AttributeValueMemberS{Value: contract.TeamID},
		"State":             &types.AttributeValueMemberS{Value: string(contract.State)},
		"RosterFrozen":      boolToAttribute(contract.RosterFrozen()),
		"Confirmed":         stringSliceToAttribute(contract.Confirmations()),
		"SeasonID":          &types.AttributeValueMemberS{Value: contract.SeasonID},
	}

	_, err := s.client.Client().PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to save contract: %w", err)
	}

	return nil
}

// ListActiveByPlayer lists non-terminal contracts where the player is in the roster and the offer category matches.
func (s *ContractStore) ListActiveByPlayer(tenantID, playerID, category string) ([]domain.MissionContract, error) {
	_ = tenantID
	ctx := context.Background()

	result, err := s.client.Client().Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(s.tableName),
		KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "CONTRACT"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list contracts: %w", err)
	}

	var active []domain.MissionContract
	for _, item := range result.Items {
		contract, err := mapToContract(item)
		if err != nil {
			return nil, err
		}
		if contract.Offer.Category != category || contract.State.IsTerminal() {
			continue
		}
		for _, p := range contract.Roster() {
			if p == playerID {
				active = append(active, *contract)
				break
			}
		}
	}

	return active, nil
}
