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

// playerActiveKey identifies the per-player active-contract set item used by
// the atomic acceptance guard. Membership is a StringSet of contract ids so a
// terminal release is an idempotent DELETE and the set can never drift.
func playerActiveKey(playerID string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: "PLAYER_ACTIVE"},
		"SK": &types.AttributeValueMemberS{Value: playerID},
	}
}

// CommitAcceptance writes the contract, each roster member's per-definition
// reservation, and each member's active-set guard in a single transaction.
// A ConditionalCheckFailed on a reservation maps to
// ErrMissionDefinitionAlreadyActive (BUG-0053-02) — checked first to mirror
// DecideAcceptance priority; one on a player's active set maps to
// ErrActiveMissionLimitReached (BUG-0022-02).
func (s *ContractStore) CommitAcceptance(contract *domain.MissionContract, offer domain.MissionOffer, effectiveLimit int) error {
	ctx := context.Background()

	items := []types.TransactWriteItem{
		seasonAcceptingCheck(s.tableName, contract.SeasonID),
		{Put: &types.Put{TableName: aws.String(s.tableName), Item: contractItem(contract)}},
	}

	// One definition reservation per roster member: conditional put.
	reservationIdx := len(items)
	for _, playerID := range contract.Roster() {
		items = append(items, reservationPutTx(contract, playerID, s.tableName))
	}

	// One active-set guard per roster member: ADD the contract id iff the set
	// is still below the effective limit.
	counterIdx := len(items)
	for _, playerID := range contract.Roster() {
		items = append(items, types.TransactWriteItem{Update: &types.Update{
			TableName:           aws.String(s.tableName),
			Key:                 playerActiveKey(playerID),
			UpdateExpression:    aws.String("ADD contract_ids :cid"),
			ConditionExpression: aws.String("attribute_not_exists(contract_ids) OR size(contract_ids) < :limit"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":cid":   &types.AttributeValueMemberSS{Value: []string{contract.ContractID}},
				":limit": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", effectiveLimit)},
			},
		}})
	}

	_, err := s.client.Client().TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{TransactItems: items})
	if err != nil {
		if seasonLockCheckFailed(err) {
			return adjudicationFenceError(ctx, s.client, s.tableName, contract.SeasonID, false)
		}
		var txCanceled *types.TransactionCanceledException
		if errors.As(err, &txCanceled) {
			definitionConflict, limitConflict := false, false
			for i, reason := range txCanceled.CancellationReasons {
				if reason.Code == nil || *reason.Code != "ConditionalCheckFailed" {
					continue
				}
				switch {
				case i >= reservationIdx && i < counterIdx:
					definitionConflict = true
				case i >= counterIdx:
					limitConflict = true
				}
			}
			if definitionConflict {
				return domain.ErrMissionDefinitionAlreadyActive
			}
			if limitConflict {
				return domain.ErrActiveMissionLimitReached
			}
		}
		return fmt.Errorf("failed to commit acceptance: %w", err)
	}
	return nil
}
