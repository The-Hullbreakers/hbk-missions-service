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

// saveCreditBearing checks the season lock in the same transaction as the
// contract write. A read followed by a separate put would admit a raced lock.
func (s *ContractStore) saveCreditBearing(ctx context.Context, seasonID string, writes []types.TransactWriteItem) error {
	items := append([]types.TransactWriteItem{seasonUnlockedCheck(s.tableName, seasonID)}, writes...)
	_, err := s.client.Client().TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: items,
	})
	if err == nil {
		return nil
	}
	if seasonLockCheckFailed(err) {
		return adjudicationFenceError(ctx, s.client, s.tableName, seasonID, true)
	}
	return fmt.Errorf("failed to save credit-bearing contract: %w", err)
}

func seasonUnlockedCheck(tableName, seasonID string) types.TransactWriteItem {
	return types.TransactWriteItem{ConditionCheck: &types.ConditionCheck{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: seasonPartition()},
			"SK": &types.AttributeValueMemberS{Value: seasonID},
		},
		ConditionExpression:      aws.String("attribute_exists(PK) AND attribute_not_exists(AdjudicationLock) AND (#state = :active OR #state = :freezing)"),
		ExpressionAttributeNames: map[string]string{"#state": "State"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":active":   &types.AttributeValueMemberS{Value: string(domain.SeasonStateActive)},
			":freezing": &types.AttributeValueMemberS{Value: string(domain.SeasonStateFreezing)},
		},
	}}
}

func seasonAcceptingCheck(tableName, seasonID string) types.TransactWriteItem {
	return types.TransactWriteItem{ConditionCheck: &types.ConditionCheck{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: seasonPartition()},
			"SK": &types.AttributeValueMemberS{Value: seasonID},
		},
		ConditionExpression:      aws.String("attribute_exists(PK) AND attribute_not_exists(AdjudicationLock) AND #state = :active"),
		ExpressionAttributeNames: map[string]string{"#state": "State"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":active": &types.AttributeValueMemberS{Value: string(domain.SeasonStateActive)},
		},
	}}
}

func seasonLockCheckFailed(err error) bool {
	var canceled *types.TransactionCanceledException
	return errors.As(err, &canceled) && len(canceled.CancellationReasons) > 0 &&
		aws.ToString(canceled.CancellationReasons[0].Code) == "ConditionalCheckFailed"
}

func adjudicationFenceError(ctx context.Context, client *DynamoDBClient, tableName, seasonID string, allowFreezing bool) error {
	result, err := client.Client().GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: seasonPartition()},
			"SK": &types.AttributeValueMemberS{Value: seasonID},
		},
		ConsistentRead: aws.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("failed to inspect rejected season lock: %w", err)
	}
	if result.Item == nil {
		return domain.ErrSeasonNotFound
	}
	if _, locked := result.Item["AdjudicationLock"]; locked {
		return domain.ErrSeasonAdjudicating
	}
	if state := getString(result.Item, "State"); state != string(domain.SeasonStateActive) &&
		(!allowFreezing || state != string(domain.SeasonStateFreezing)) {
		return domain.ErrSeasonNotActive
	}
	return fmt.Errorf("failed to reconcile rejected season lock: %w", domain.ErrInvalidSeasonStateTransition)
}

// duplicateOnConditionalFail maps only cancellation in the credit-item
// range to the exactly-once signal, never the preceding season guard.
func duplicateOnConditionalFail(err error, creditStart, creditEnd int) error {
	if err == nil {
		return nil
	}
	var canceled *types.TransactionCanceledException
	if errors.As(err, &canceled) {
		for i, reason := range canceled.CancellationReasons {
			if reason.Code != nil && *reason.Code == "ConditionalCheckFailed" &&
				i >= creditStart && i < creditEnd {
				return domain.ErrDuplicateConfirmation
			}
		}
	}
	return nil
}
