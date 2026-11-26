package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// ErrStagedFactConflict reports that a row is already staged under the
// fact's deterministic identity with a different digest. Staged facts are
// immutable: a divergent replay is surfaced, never overwritten.
var ErrStagedFactConflict = errors.New("staged fact conflicts with previously staged content")

// PlacementFactStager persists each placement fact as an independent item.
// Writes are sequential conditional puts — one fact per write, no
// transaction — so the staged set size is not bounded by the 100-item
// transaction limit. An identical replay of the same staged identity is a
// no-op; any digest divergence conflicts.
type PlacementFactStager struct {
	client    *DynamoDBClient
	tableName string
}

// NewPlacementFactStager creates the adapter implementing
// application.PlacementFactStager.
func NewPlacementFactStager(client *DynamoDBClient, tableName string) *PlacementFactStager {
	return &PlacementFactStager{client: client, tableName: tableName}
}

// Stage writes all facts for the attempt, stopping at the first failure.
func (s *PlacementFactStager) Stage(attempt domain.AdjudicationAttempt, facts []domain.PlacementFact) error {
	ctx := context.Background()
	for i := range facts {
		if err := s.stageOne(ctx, attempt, facts[i]); err != nil {
			return fmt.Errorf("failed to stage placement fact %s: %w", facts[i].EventID, err)
		}
	}
	return nil
}

// stageOne issues the single conditional put for one fact. On a
// conditional failure it falls through to the replay path instead of
// overwriting.
func (s *PlacementFactStager) stageOne(ctx context.Context, attempt domain.AdjudicationAttempt, fact domain.PlacementFact) error {
	if fact.SeasonID != attempt.SeasonID || fact.EventID == "" {
		return domain.ErrStagedFactSetInvalid
	}
	members := make([]types.AttributeValue, len(fact.MemberIDs))
	for i, member := range fact.MemberIDs {
		members[i] = &types.AttributeValueMemberS{Value: member}
	}
	item := map[string]types.AttributeValue{
		"PK":          &types.AttributeValueMemberS{Value: "FACT#" + fact.SeasonID},
		"SK":          &types.AttributeValueMemberS{Value: attempt.ID + "#" + fact.Category + "#" + string(fact.Track) + "#" + fact.RecipientID},
		"Digest":      &types.AttributeValueMemberS{Value: fact.Digest()},
		"AttemptID":   &types.AttributeValueMemberS{Value: attempt.ID},
		"EventID":     &types.AttributeValueMemberS{Value: fact.EventID},
		"SeasonID":    &types.AttributeValueMemberS{Value: fact.SeasonID},
		"Category":    &types.AttributeValueMemberS{Value: fact.Category},
		"Track":       &types.AttributeValueMemberS{Value: string(fact.Track)},
		"RecipientID": &types.AttributeValueMemberS{Value: fact.RecipientID},
		"Rank":        &types.AttributeValueMemberN{Value: strconv.Itoa(fact.Rank)},
		"FinalMerits": &types.AttributeValueMemberN{Value: strconv.Itoa(fact.FinalMerits)},
		"MemberIDs":   &types.AttributeValueMemberL{Value: members},
	}
	_, err := s.client.Client().PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(s.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	})
	if err == nil {
		return nil
	}
	var cond *types.ConditionalCheckFailedException
	if !errors.As(err, &cond) {
		return fmt.Errorf("failed to stage fact: %w", err)
	}
	return s.replay(ctx, item, fact)
}

// replay reconciles a conditional failure: an identical stored digest is a
// no-op; a missing or divergent one is a conflict, never an overwrite.
func (s *PlacementFactStager) replay(ctx context.Context, item map[string]types.AttributeValue, fact domain.PlacementFact) error {
	result, readErr := s.client.Client().GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"PK": item["PK"],
			"SK": item["SK"],
		},
		ConsistentRead: aws.Bool(true),
	})
	if readErr != nil {
		return fmt.Errorf("failed to read staged fact: %w", readErr)
	}
	if result.Item == nil {
		return ErrStagedFactConflict
	}
	digest, ok := result.Item["Digest"].(*types.AttributeValueMemberS)
	if !ok || digest.Value != fact.Digest() {
		return ErrStagedFactConflict
	}
	return nil
}
