package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// ListActiveByPlayer lists the player's non-terminal contracts across all
// categories; the active limit counts the total set and departed members are
// excluded from active membership.
func (s *ContractStore) ListActiveByPlayer(playerID string) ([]domain.MissionContract, error) {
	contracts, err := s.listAll()
	if err != nil {
		return nil, err
	}

	var active []domain.MissionContract
	for _, contract := range contracts {
		if contract.State.IsTerminal() {
			continue
		}
		if contract.IncludesPlayer(playerID) && !contract.IsDeparted(playerID) {
			active = append(active, contract)
		}
	}
	return active, nil
}

// ListByPlayer lists the contracts where the player is in the roster, matching the filter.
func (s *ContractStore) ListByPlayer(playerID string, filter domain.ContractListFilter) ([]domain.MissionContract, error) {
	contracts, err := s.listAll()
	if err != nil {
		return nil, err
	}

	mine := make([]domain.MissionContract, 0, len(contracts))
	for _, contract := range contracts {
		if contract.IncludesPlayer(playerID) && filter.Match(contract) {
			mine = append(mine, contract)
		}
	}
	return mine, nil
}

func (s *ContractStore) listAll() ([]domain.MissionContract, error) {
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

	contracts := make([]domain.MissionContract, 0, len(result.Items))
	for _, item := range result.Items {
		contract, err := mapToContract(item)
		if err != nil {
			return nil, err
		}
		contracts = append(contracts, *contract)
	}
	return contracts, nil
}
