package dynamodb

import (
	"context"
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

// Save persists a season.
func (s *SeasonStore) Save(season domain.Season) error {
	ctx := context.Background()

	item := map[string]types.AttributeValue{
		"PK":       &types.AttributeValueMemberS{Value: "SEASON"},
		"SK":       &types.AttributeValueMemberS{Value: season.SeasonID},
		"State":    &types.AttributeValueMemberS{Value: string(season.State)},
		"Official": boolToAttribute(season.Official),
	}

	if !season.StartsAt.IsZero() {
		item["StartsAt"] = timeToAttribute(season.StartsAt)
	}
	if season.Schedule != nil {
		item["ScheduleActiveWeeks"] = &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", season.Schedule.ActiveWeeks)}
		item["ScheduleOffWeeks"] = &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", season.Schedule.OffWeeks)}
		item["ScheduleSeriesPrefix"] = &types.AttributeValueMemberS{Value: season.Schedule.Series.Prefix}
		item["ScheduleSeriesIteration"] = &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", season.Schedule.Series.Iteration)}
	}

	_, err := s.client.Client().PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to save season: %w", err)
	}

	return nil
}

// ListActive returns all seasons with State = ACTIVE.
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
			":pk":    &types.AttributeValueMemberS{Value: "SEASON"},
			":state": &types.AttributeValueMemberS{Value: string(domain.SeasonStateActive)},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list active seasons: %w", err)
	}

	var seasons []domain.Season
	for _, item := range result.Items {
		seasons = append(seasons, seasonFromItem(item))
	}

	return seasons, nil
}

// List returns seasons that match the given status filter.
func (s *SeasonStore) List(filter domain.SeasonStatusFilter) ([]domain.Season, error) {
	ctx := context.Background()

	result, err := s.client.Client().Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(s.tableName),
		KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "SEASON"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list seasons: %w", err)
	}

	var seasons []domain.Season
	for _, item := range result.Items {
		season := seasonFromItem(item)
		if filter.Match(season) {
			seasons = append(seasons, season)
		}
	}

	return seasons, nil
}

// Get loads a season by identity.
func (s *SeasonStore) Get(seasonID string) (domain.Season, error) {
	ctx := context.Background()

	result, err := s.client.Client().GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "SEASON"},
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
