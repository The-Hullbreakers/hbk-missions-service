package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// List returns the seasons that match the given status filter.
func (s *SeasonStore) List(filter domain.SeasonStatusFilter) ([]domain.Season, error) {
	ctx := context.Background()

	result, err := s.client.Client().Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(s.tableName),
		KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: seasonPartition()},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list seasons: %w", err)
	}

	seasons := make([]domain.Season, 0, len(result.Items))
	for _, item := range result.Items {
		season := seasonFromItem(item)
		if filter.Match(season) {
			seasons = append(seasons, season)
		}
	}

	return seasons, nil
}
