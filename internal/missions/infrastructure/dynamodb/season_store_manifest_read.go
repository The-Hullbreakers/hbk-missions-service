package dynamodb

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// ListPendingManifests never exposes staged facts without a committed manifest.
func (s *SeasonStore) ListPendingManifests(ctx context.Context, limit int) ([]application.PlacementManifest, error) {
	pending := []application.PlacementManifest{}
	var start map[string]types.AttributeValue
	for {
		page, err := s.client.Client().Query(ctx, &dynamodb.QueryInput{
			TableName:                aws.String(s.tableName),
			KeyConditionExpression:   aws.String("PK = :pk AND begins_with(SK, :prefix)"),
			FilterExpression:         aws.String("#status = :pending"),
			ExpressionAttributeNames: map[string]string{"#status": "Status"},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk":      &types.AttributeValueMemberS{Value: outboxPartition},
				":prefix":  &types.AttributeValueMemberS{Value: "MANIFEST#"},
				":pending": &types.AttributeValueMemberS{Value: outboxStatusPending},
			},
			ConsistentRead: aws.Bool(true), ExclusiveStartKey: start,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list pending placement manifests: %w", err)
		}
		for _, item := range page.Items {
			manifest, mapErr := manifestFromItem(item)
			if mapErr != nil {
				return nil, mapErr
			}
			pending = append(pending, manifest)
			if limit > 0 && len(pending) >= limit {
				return pending, nil
			}
		}
		if len(page.LastEvaluatedKey) == 0 {
			return pending, nil
		}
		start = page.LastEvaluatedKey
	}
}

// ListRemaining validates the entire staged set before exposing any fact
// for publication; the progress cursor only determines which ones to return.
func (s *SeasonStore) ListRemaining(ctx context.Context, manifest application.PlacementManifest) ([]application.PendingIntent, error) {
	facts := []domain.PlacementFact{}
	var start map[string]types.AttributeValue
	for {
		page, err := s.client.Client().Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(s.tableName),
			KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :prefix)"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk":     &types.AttributeValueMemberS{Value: "FACT#" + manifest.SeasonID},
				":prefix": &types.AttributeValueMemberS{Value: manifest.AttemptID + "#"},
			},
			ConsistentRead: aws.Bool(true), ExclusiveStartKey: start,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to read staged placement facts: %w", err)
		}
		for _, item := range page.Items {
			fact, mapErr := stagedPlacementFact(item, manifest)
			if mapErr != nil {
				return nil, mapErr
			}
			facts = append(facts, fact)
		}
		if len(page.LastEvaluatedKey) == 0 {
			break
		}
		start = page.LastEvaluatedKey
	}
	set, err := domain.NewStagedFactSet(facts)
	if err != nil {
		return nil, fmt.Errorf("invalid staged placement facts: %w", err)
	}
	if err := set.Verify(manifest.FactCount, manifest.FactHash); err != nil {
		return nil, fmt.Errorf("staged placement facts differ from committed manifest: %w", err)
	}
	if manifest.NextIndex < 0 || manifest.NextIndex > len(facts) {
		return nil, fmt.Errorf("invalid placement manifest progress")
	}
	remaining := make([]application.PendingIntent, 0, len(facts)-manifest.NextIndex)
	for _, fact := range facts[manifest.NextIndex:] {
		item, err := outboxPlacementItem(fact, manifest.Cutoff.Format(time.RFC3339Nano))
		if err != nil {
			return nil, err
		}
		remaining = append(remaining, pendingIntentFromItem(item))
	}
	return remaining, nil
}

func stagedPlacementFact(item map[string]types.AttributeValue, manifest application.PlacementManifest) (domain.PlacementFact, error) {
	rank, rankErr := manifestNumber(item, "Rank")
	merits, meritsErr := manifestNumber(item, "FinalMerits")
	members, ok := item["MemberIDs"].(*types.AttributeValueMemberL)
	if rankErr != nil || meritsErr != nil || !ok || getString(item, "SeasonID") != manifest.SeasonID ||
		getString(item, "AttemptID") != manifest.AttemptID {
		return domain.PlacementFact{}, fmt.Errorf("invalid staged placement fact fields")
	}
	memberIDs := make([]string, 0, len(members.Value))
	for _, member := range members.Value {
		value, valid := member.(*types.AttributeValueMemberS)
		if !valid || value.Value == "" {
			return domain.PlacementFact{}, fmt.Errorf("invalid staged placement member")
		}
		memberIDs = append(memberIDs, value.Value)
	}
	track := domain.Track(getString(item, "Track"))
	if track != domain.TrackIndividual && track != domain.TrackTeam {
		return domain.PlacementFact{}, fmt.Errorf("invalid staged placement track")
	}
	fact, err := domain.NewPlacementFact(domain.PlacementFactParams{
		SeasonID: manifest.SeasonID, Category: getString(item, "Category"), Track: track,
		RecipientID: getString(item, "RecipientID"), Rank: rank, FinalMerits: merits, MemberIDs: memberIDs,
	})
	expectedSK := manifest.AttemptID + "#" + fact.Category + "#" + string(fact.Track) + "#" + fact.RecipientID
	if err != nil || fact.EventID != getString(item, "EventID") ||
		fact.Digest() != getString(item, "Digest") || getString(item, "SK") != expectedSK {
		return domain.PlacementFact{}, fmt.Errorf("staged placement fact identity or digest mismatch")
	}
	return fact, nil
}

func manifestNumber(item map[string]types.AttributeValue, field string) (int, error) {
	number, ok := item[field].(*types.AttributeValueMemberN)
	if !ok {
		return 0, fmt.Errorf("missing manifest number %s", field)
	}
	return strconv.Atoi(number.Value)
}
