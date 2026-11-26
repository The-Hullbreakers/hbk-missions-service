package dynamodb

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// ListStandings returns the standings for the given bucket, excluding those below minMerits.
func (s *MeritLedgerStore) ListStandings(seasonID, category string, track domain.Track, minMerits int) ([]domain.Standing, error) {
	ctx := context.Background()
	key := leaderboardBucketKeyFromParts(seasonID, category, track)

	var exclusiveStartKey map[string]types.AttributeValue
	builder := make(map[string]*standingBuilder)

	for {
		result, err := s.client.Client().Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(s.tableName),
			IndexName:              aws.String("GSI2"),
			KeyConditionExpression: aws.String("GSI2PK = :pk"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: key},
			},
			ExclusiveStartKey: exclusiveStartKey,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list standings: %w", err)
		}

		for _, item := range result.Items {
			recipientID := getString(item, "RecipientID")
			b, ok := builder[recipientID]
			if !ok {
				b = newStandingBuilder()
				builder[recipientID] = b
			}
			if v, ok := item["Amount"].(*types.AttributeValueMemberN); ok {
				amount, _ := strconv.Atoi(v.Value)
				b.total += amount
			}
			b.contracts[getString(item, "ContractID")] = true
			b.objectives[getString(item, "ObjectiveID")] = true
		}

		if result.LastEvaluatedKey == nil {
			break
		}
		exclusiveStartKey = result.LastEvaluatedKey
	}

	completed, err := s.completedContractIDs(builder)
	if err != nil {
		return nil, err
	}

	var standings []domain.Standing
	for recipientID, b := range builder {
		if b.total < minMerits {
			continue
		}
		completedCount := 0
		for contractID := range b.contracts {
			if completed[contractID] {
				completedCount++
			}
		}
		standings = append(standings, domain.Standing{
			SeasonID:          seasonID,
			Category:          category,
			Track:             track,
			RecipientID:       recipientID,
			Total:             b.total,
			CompletedMissions: completedCount,
			UniqueObjectives:  len(b.objectives),
		})
	}

	sort.Slice(standings, func(i, j int) bool {
		return domain.CompareStanding(standings[i], standings[j]) < 0
	})

	return standings, nil
}

// TotalFor returns the sum of credits for the given bucket.
func (s *MeritLedgerStore) TotalFor(seasonID, category string, track domain.Track, recipientID string) int {
	ctx := context.Background()
	key := bucketKey(seasonID, category, track, recipientID)

	var total int
	var exclusiveStartKey map[string]types.AttributeValue

	for {
		result, err := s.client.Client().Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(s.tableName),
			IndexName:              aws.String("GSI1"),
			KeyConditionExpression: aws.String("GSI1PK = :pk"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: key},
			},
			ExclusiveStartKey: exclusiveStartKey,
		})
		if err != nil {
			return 0
		}

		for _, item := range result.Items {
			if v, ok := item["Amount"].(*types.AttributeValueMemberN); ok {
				amount, _ := strconv.Atoi(v.Value)
				total += amount
			}
		}

		if result.LastEvaluatedKey == nil {
			break
		}
		exclusiveStartKey = result.LastEvaluatedKey
	}

	return total
}
