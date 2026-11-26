package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// completedContractIDs resolves which credited contract ids reached
// COMPLETED — the tie-break only counts finished contracts, never credits
// from contracts still in progress or in another non-terminal state.
func (s *MeritLedgerStore) completedContractIDs(builders map[string]*standingBuilder) (map[string]bool, error) {
	ctx := context.Background()

	unique := make(map[string]bool)
	for _, b := range builders {
		for contractID := range b.contracts {
			unique[contractID] = true
		}
	}

	completed := make(map[string]bool)
	for contractID := range unique {
		result, err := s.client.Client().GetItem(ctx, &dynamodb.GetItemInput{
			TableName: aws.String(s.tableName),
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{Value: "CONTRACT"},
				"SK": &types.AttributeValueMemberS{Value: contractID},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to load contract state: %w", err)
		}
		if result.Item == nil {
			continue
		}
		if domain.ContractState(getString(result.Item, "State")) == domain.ContractCompleted {
			completed[contractID] = true
		}
	}
	return completed, nil
}
