package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// FinalizationCommitter commits the constant-size finalization transaction:
// the FREEZING season transitions to FINALIZED only while it still holds the
// matching adjudication attempt, and exactly one deterministic PENDING
// manifest is conditionally created. A stale retry can never overwrite a
// committed manifest's progress, and per-fact outbox rows are never written.
type FinalizationCommitter struct {
	client    *DynamoDBClient
	tableName string
}

// NewFinalizationCommitter creates the adapter implementing
// application.FinalizationCommitter.
func NewFinalizationCommitter(client *DynamoDBClient, tableName string) *FinalizationCommitter {
	return &FinalizationCommitter{client: client, tableName: tableName}
}

// Commit writes the guarded season transition and manifest in one
// transaction. The facts argument is only summarized into the manifest's
// count/hash — the staged rows are the source the publisher later re-verifies.
func (c *FinalizationCommitter) Commit(season domain.Season, attempt domain.AdjudicationAttempt, facts []domain.PlacementFact) error {
	set, err := domain.NewStagedFactSet(facts)
	if err != nil {
		return err
	}
	seasonTx := types.TransactWriteItem{Update: &types.Update{
		TableName: aws.String(c.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: seasonPartition()},
			"SK": &types.AttributeValueMemberS{Value: season.SeasonID},
		},
		UpdateExpression: aws.String("SET #state = :finalized"),
		ConditionExpression: aws.String("attribute_exists(PK) AND #state = :freezing " +
			"AND AdjudicationLock = :lock AND AdjudicationCutoff = :cutoff"),
		ExpressionAttributeNames: map[string]string{"#state": "State"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":finalized": &types.AttributeValueMemberS{Value: string(domain.SeasonStateFinalized)},
			":freezing":  &types.AttributeValueMemberS{Value: string(domain.SeasonStateFreezing)},
			":lock":      &types.AttributeValueMemberS{Value: attempt.ID},
			":cutoff":    &types.AttributeValueMemberS{Value: attempt.Cutoff.Format(time.RFC3339Nano)},
		},
	}}
	manifestTx := types.TransactWriteItem{Put: &types.Put{
		TableName: aws.String(c.tableName),
		Item: map[string]types.AttributeValue{
			"PK":        &types.AttributeValueMemberS{Value: "OUTBOX"},
			"SK":        &types.AttributeValueMemberS{Value: "MANIFEST#" + season.SeasonID + "#" + attempt.ID},
			"SeasonID":  &types.AttributeValueMemberS{Value: season.SeasonID},
			"AttemptID": &types.AttributeValueMemberS{Value: attempt.ID},
			"Cutoff":    &types.AttributeValueMemberS{Value: attempt.Cutoff.Format(time.RFC3339Nano)},
			"FactCount": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", set.Count())},
			"FactHash":  &types.AttributeValueMemberS{Value: set.Hash()},
			"NextIndex": &types.AttributeValueMemberN{Value: "0"},
			"Status":    &types.AttributeValueMemberS{Value: "PENDING"},
		},
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	}}
	_, err = c.client.Client().TransactWriteItems(context.Background(), &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{seasonTx, manifestTx},
	})
	if err == nil {
		return nil
	}
	var canceled *types.TransactionCanceledException
	if errors.As(err, &canceled) {
		return domain.ErrInvalidSeasonStateTransition
	}
	return fmt.Errorf("failed to commit finalization manifest: %w", err)
}
