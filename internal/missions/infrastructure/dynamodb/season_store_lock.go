package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// AdjudicationLock persists the internal adjudication attempt on the season
// item. One conditional write claims the lock only while the season is
// FREEZING and no lock exists; a raced or retried acquire resumes the
// already-persisted stable attempt (same ID and cutoff) instead of minting
// a parallel one. The public season state is never altered here.
type AdjudicationLock struct {
	client    *DynamoDBClient
	tableName string
}

// NewAdjudicationLock creates the adapter implementing
// application.AdjudicationLock.
func NewAdjudicationLock(client *DynamoDBClient, tableName string) *AdjudicationLock {
	return &AdjudicationLock{client: client, tableName: tableName}
}

// Acquire claims or resumes the adjudication attempt for the season. The
// cutoff is minted once per lock — retries return the persisted value.
func (a *AdjudicationLock) Acquire(seasonID string) (domain.AdjudicationAttempt, error) {
	ctx := context.Background()
	attempt, err := domain.NewAdjudicationAttempt(seasonID, time.Now().UTC())
	if err != nil {
		return domain.AdjudicationAttempt{}, err
	}
	_, err = a.client.Client().UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(a.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "SEASON"},
			"SK": &types.AttributeValueMemberS{Value: seasonID},
		},
		UpdateExpression:    aws.String("SET AdjudicationLock = :lock, AdjudicationCutoff = :cutoff"),
		ConditionExpression: aws.String("attribute_exists(PK) AND #state = :freezing AND attribute_not_exists(AdjudicationLock)"),
		ExpressionAttributeNames: map[string]string{
			"#state": "State",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":lock":     &types.AttributeValueMemberS{Value: attempt.ID},
			":cutoff":   &types.AttributeValueMemberS{Value: attempt.Cutoff.Format(time.RFC3339Nano)},
			":freezing": &types.AttributeValueMemberS{Value: string(domain.SeasonStateFreezing)},
		},
	})
	if err == nil {
		return attempt, nil
	}
	var cond *types.ConditionalCheckFailedException
	if !errors.As(err, &cond) {
		return domain.AdjudicationAttempt{}, fmt.Errorf("failed to acquire adjudication lock: %w", err)
	}
	return a.resume(ctx, seasonID)
}

// resume returns the attempt persisted by the lock holder. Any anomaly —
// missing season, non-FREEZING state, or a lock/cutoff that cannot re-mint
// the deterministic identity — is surfaced, never silently accepted.
func (a *AdjudicationLock) resume(ctx context.Context, seasonID string) (domain.AdjudicationAttempt, error) {
	result, err := a.client.Client().GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(a.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "SEASON"},
			"SK": &types.AttributeValueMemberS{Value: seasonID},
		},
		ConsistentRead: aws.Bool(true),
	})
	if err != nil {
		return domain.AdjudicationAttempt{}, fmt.Errorf("failed to read adjudication lock: %w", err)
	}
	if result.Item == nil {
		return domain.AdjudicationAttempt{}, domain.ErrSeasonNotFound
	}
	if getString(result.Item, "State") != string(domain.SeasonStateFreezing) {
		return domain.AdjudicationAttempt{}, domain.ErrInvalidSeasonStateTransition
	}
	storedLock := getString(result.Item, "AdjudicationLock")
	attempt, attemptErr := domain.NewAdjudicationAttempt(seasonID, getTime(result.Item, "AdjudicationCutoff"))
	if attemptErr != nil || storedLock != attempt.ID {
		return domain.AdjudicationAttempt{}, domain.ErrAdjudicationAttemptInvalid
	}
	return attempt, nil
}

// ListLocked returns every season row currently holding an adjudication
// lock, read consistently, for the internal orphaned-lock sweep. Rows whose
// lock metadata cannot be reminted are surfaced, never silently accepted.
func (a *AdjudicationLock) ListLocked(ctx context.Context) ([]application.LockedSeason, error) {
	locked := []application.LockedSeason{}
	var start map[string]types.AttributeValue
	for {
		page, err := a.client.Client().Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(a.tableName),
			KeyConditionExpression: aws.String("PK = :pk"),
			FilterExpression:       aws.String("attribute_exists(AdjudicationLock)"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: "SEASON"},
			},
			ConsistentRead: aws.Bool(true), ExclusiveStartKey: start,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list adjudication locks: %w", err)
		}
		for _, item := range page.Items {
			seasonID := getString(item, "SK")
			attempt, attemptErr := domain.NewAdjudicationAttempt(seasonID, getTime(item, "AdjudicationCutoff"))
			if attemptErr != nil || getString(item, "AdjudicationLock") != attempt.ID {
				return nil, domain.ErrAdjudicationAttemptInvalid
			}
			locked = append(locked, application.LockedSeason{
				SeasonID: seasonID, AttemptID: attempt.ID, Cutoff: attempt.Cutoff,
			})
		}
		if len(page.LastEvaluatedKey) == 0 {
			return locked, nil
		}
		start = page.LastEvaluatedKey
	}
}
