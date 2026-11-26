package dynamodb

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
)

// outboxPartition is the single-scope partition for pending event intents.
const outboxPartition = "OUTBOX"

const (
	outboxStatusPending   = "PENDING"
	outboxStatusPublished = "PUBLISHED"
)

var (
	// ErrOutboxIntentNotFound is returned when marking an intent that was
	// never persisted; the conditional update never creates a phantom row.
	ErrOutboxIntentNotFound = errors.New("outbox intent not found")
	// ErrOutboxBatchTooLarge is returned when the fact batch cannot fit in a
	// single DynamoDB transaction; no partial finalization is written.
	ErrOutboxBatchTooLarge = errors.New("placement facts exceed the transaction item limit")
)

// outboxSortKey scopes one intent under the outbox partition.
func outboxSortKey(eventID string) string {
	return "PLACEMENT#" + eventID
}

// outboxCorrelationID derives the stable correlation shared by every fact a
// single finalization emits, so republication keeps the same identity.
func outboxCorrelationID(seasonID string) string {
	sum := sha256.Sum256([]byte("season-finalize|" + seasonID))
	return hex.EncodeToString(sum[:16])
}

// ListPending returns persisted publication intents still awaiting publish.
// DynamoDB applies Limit before the status filter, so reads keep paginating
// until `limit` pending intents are collected or the partition is exhausted;
// a non-positive limit drains every pending intent.
func (s *SeasonStore) ListPending(ctx context.Context, limit int) ([]application.PendingIntent, error) {
	pending := []application.PendingIntent{}
	var startKey map[string]types.AttributeValue
	for limit <= 0 || len(pending) < limit {
		input := &dynamodb.QueryInput{
			TableName:              aws.String(s.tableName),
			KeyConditionExpression: aws.String("PK = :pk"),
			FilterExpression:       aws.String("#status = :pending"),
			ExpressionAttributeNames: map[string]string{
				"#status": "Status",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk":      &types.AttributeValueMemberS{Value: outboxPartition},
				":pending": &types.AttributeValueMemberS{Value: outboxStatusPending},
			},
			ExclusiveStartKey: startKey,
		}
		if limit > 0 {
			input.Limit = aws.Int32(int32(limit - len(pending)))
		}
		result, err := s.client.Client().Query(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to list pending outbox intents: %w", err)
		}
		for _, item := range result.Items {
			pending = append(pending, pendingIntentFromItem(item))
		}
		if result.LastEvaluatedKey == nil {
			break
		}
		startKey = result.LastEvaluatedKey
	}
	return pending, nil
}

// MarkPublished flips one intent to PUBLISHED. The conditional update keeps
// the operation idempotent: repeats succeed, unknown ids fail without
// creating a phantom row.
func (s *SeasonStore) MarkPublished(ctx context.Context, eventID string) error {
	_, err := s.client.Client().UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: outboxPartition},
			"SK": &types.AttributeValueMemberS{Value: outboxSortKey(eventID)},
		},
		UpdateExpression:    aws.String("SET #status = :published"),
		ConditionExpression: aws.String("attribute_exists(PK)"),
		ExpressionAttributeNames: map[string]string{
			"#status": "Status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":published": &types.AttributeValueMemberS{Value: outboxStatusPublished},
		},
	})
	if err != nil {
		var cond *types.ConditionalCheckFailedException
		if errors.As(err, &cond) {
			return ErrOutboxIntentNotFound
		}
		return fmt.Errorf("failed to mark outbox intent %s published: %w", eventID, err)
	}
	return nil
}

// pendingIntentFromItem maps an outbox row back to its application port type.
func pendingIntentFromItem(item map[string]types.AttributeValue) application.PendingIntent {
	return application.PendingIntent{
		EventID:   strings.TrimPrefix(getString(item, "SK"), "PLACEMENT#"),
		EventType: getString(item, "EventType"),
		Envelope:  getString(item, "Envelope"),
	}
}
