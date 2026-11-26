package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// ListBuckets returns the distinct category/track buckets that carry credits
// for the season — the enumeration behind the standings-history read.
func (s *MeritLedgerStore) ListBuckets(seasonID string) ([]domain.StandingBucket, error) {
	ctx := context.Background()

	var exclusiveStartKey map[string]types.AttributeValue
	seen := make(map[string]domain.StandingBucket)

	for {
		result, err := s.client.Client().Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(s.tableName),
			KeyConditionExpression: aws.String("PK = :pk"),
			FilterExpression:       aws.String("SeasonID = :sid"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk":  &types.AttributeValueMemberS{Value: "MERIT_CREDIT"},
				":sid": &types.AttributeValueMemberS{Value: seasonID},
			},
			ExclusiveStartKey: exclusiveStartKey,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list buckets: %w", err)
		}

		for _, item := range result.Items {
			bucket := domain.StandingBucket{
				Category: getString(item, "Category"),
				Track:    domain.Track(getString(item, "Track")),
			}
			seen[bucket.Category+"/"+string(bucket.Track)] = bucket
		}

		if result.LastEvaluatedKey == nil {
			break
		}
		exclusiveStartKey = result.LastEvaluatedKey
	}

	buckets := make([]domain.StandingBucket, 0, len(seen))
	for _, bucket := range seen {
		buckets = append(buckets, bucket)
	}
	return buckets, nil
}
