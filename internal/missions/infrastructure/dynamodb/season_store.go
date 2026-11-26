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

// SeasonStore loads seasons from DynamoDB.
type SeasonStore struct {
	client    *DynamoDBClient
	tableName string
}

// NewSeasonStore creates a DynamoDB season store.
func NewSeasonStore(client *DynamoDBClient, tableName string) *SeasonStore {
	return &SeasonStore{client: client, tableName: tableName}
}

// seasonPartition returns the single-scope partition key for seasons.
func seasonPartition() string {
	return "SEASON"
}

// seasonItem builds the persistence attributes for a season.
func seasonItem(season domain.Season) map[string]types.AttributeValue {
	item := map[string]types.AttributeValue{
		"PK":       &types.AttributeValueMemberS{Value: seasonPartition()},
		"SK":       &types.AttributeValueMemberS{Value: season.SeasonID},
		"State":    &types.AttributeValueMemberS{Value: string(season.State)},
		"Official": boolToAttribute(season.Official),
	}
	if !season.StartsAt.IsZero() {
		item["StartsAt"] = timeToAttribute(season.StartsAt)
	}
	if season.Schedule != nil {
		item["ScheduleActiveDays"] = &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", season.Schedule.ActiveDays)}
		item["ScheduleOffDays"] = &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", season.Schedule.OffDays)}
		if season.Schedule.ActiveWeeks > 0 || season.Schedule.OffWeeks > 0 {
			item["ScheduleActiveWeeks"] = &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", season.Schedule.ActiveWeeks)}
			item["ScheduleOffWeeks"] = &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", season.Schedule.OffWeeks)}
		}
		item["ScheduleSeriesPrefix"] = &types.AttributeValueMemberS{Value: season.Schedule.Series.Prefix}
		item["ScheduleSeriesIteration"] = &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", season.Schedule.Series.Iteration)}
	}
	return item
}

// Create persists a new season; an existing id can never be overwritten —
// the conditional put fails regardless of caller bugs above the store.
func (s *SeasonStore) Create(season domain.Season) error {
	ctx := context.Background()

	_, err := s.client.Client().PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(s.tableName),
		Item:                seasonItem(season),
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	})
	if err != nil {
		var cond *types.ConditionalCheckFailedException
		if errors.As(err, &cond) {
			return domain.ErrSeasonAlreadyExists
		}
		return fmt.Errorf("failed to create season: %w", err)
	}
	return nil
}

// Save upserts a season; lifecycle transitions reuse it on existing ids.
func (s *SeasonStore) Save(season domain.Season) error {
	ctx := context.Background()

	_, err := s.client.Client().PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      seasonItem(season),
	})
	if err != nil {
		return fmt.Errorf("failed to save season: %w", err)
	}

	return nil
}

// ListActive returns the seasons with State = ACTIVE.
func (s *SeasonStore) ListActive() ([]domain.Season, error) {
	ctx := context.Background()

	result, err := s.client.Client().Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(s.tableName),
		KeyConditionExpression: aws.String("PK = :pk"),
		FilterExpression:       aws.String("#state = :state"),
		ExpressionAttributeNames: map[string]string{
			"#state": "State",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":    &types.AttributeValueMemberS{Value: seasonPartition()},
			":state": &types.AttributeValueMemberS{Value: string(domain.SeasonStateActive)},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list active seasons: %w", err)
	}

	seasons := make([]domain.Season, 0, len(result.Items))
	for _, item := range result.Items {
		seasons = append(seasons, seasonFromItem(item))
	}

	return seasons, nil
}

// Get loads a season by identity.
func (s *SeasonStore) Get(seasonID string) (domain.Season, error) {
	ctx := context.Background()

	result, err := s.client.Client().GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: seasonPartition()},
			"SK": &types.AttributeValueMemberS{Value: seasonID},
		},
	})
	if err != nil {
		return domain.Season{}, fmt.Errorf("failed to get season: %w", err)
	}
	if result.Item == nil {
		return domain.Season{}, domain.ErrSeasonNotFound
	}

	return seasonFromItem(result.Item), nil
}
