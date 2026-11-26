package dynamodb

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// OverallCreditsStore reads the facts for the individual season-wide ranking.
type OverallCreditsStore struct {
	client    *DynamoDBClient
	tableName string
}

func NewOverallCreditsReader(client *DynamoDBClient, tableName string) *OverallCreditsStore {
	return &OverallCreditsStore{client: client, tableName: tableName}
}

func (s *OverallCreditsStore) ListOverallCredits(seasonID string) ([]domain.MeritCredit, map[string]bool, error) {
	ctx := context.Background()
	credits := make([]domain.MeritCredit, 0)
	var startKey map[string]types.AttributeValue
	for {
		page, err := s.client.Client().Query(ctx, &dynamodb.QueryInput{
			TableName:                aws.String(s.tableName),
			KeyConditionExpression:   aws.String("PK = :pk"),
			FilterExpression:         aws.String("#season = :season AND #track = :track"),
			ProjectionExpression:     aws.String("#season, Category, #track, RecipientID, Amount, ContractID, ObjectiveID"),
			ExpressionAttributeNames: map[string]string{"#season": "SeasonID", "#track": "Track"},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk":     &types.AttributeValueMemberS{Value: "MERIT_CREDIT"},
				":season": &types.AttributeValueMemberS{Value: seasonID},
				":track":  &types.AttributeValueMemberS{Value: string(domain.TrackIndividual)},
			},
			ExclusiveStartKey: startKey,
			ConsistentRead:    aws.Bool(true),
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read overall credits: %w", err)
		}
		for _, item := range page.Items {
			number, ok := item["Amount"].(*types.AttributeValueMemberN)
			if !ok {
				return nil, nil, fmt.Errorf("invalid overall credit amount")
			}
			amount, err := strconv.Atoi(number.Value)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid overall credit amount: %w", err)
			}
			credit := domain.MeritCredit{
				SeasonID:    getString(item, "SeasonID"),
				Category:    getString(item, "Category"),
				Track:       domain.TrackIndividual,
				RecipientID: getString(item, "RecipientID"),
				Amount:      amount,
				ContractID:  getString(item, "ContractID"),
				ObjectiveID: getString(item, "ObjectiveID"),
			}
			if credit.SeasonID != seasonID || credit.Category == "" || credit.RecipientID == "" || credit.ContractID == "" || credit.ObjectiveID == "" {
				return nil, nil, fmt.Errorf("invalid overall credit facts")
			}
			credits = append(credits, credit)
		}
		if len(page.LastEvaluatedKey) == 0 {
			break
		}
		startKey = page.LastEvaluatedKey
	}
	completed := make(map[string]bool)
	seen := make(map[string]struct{})
	for _, credit := range credits {
		if _, ok := seen[credit.ContractID]; ok {
			continue
		}
		seen[credit.ContractID] = struct{}{}
		result, err := s.client.Client().GetItem(ctx, &dynamodb.GetItemInput{
			TableName: aws.String(s.tableName),
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{Value: "CONTRACT"},
				"SK": &types.AttributeValueMemberS{Value: credit.ContractID},
			},
			ConsistentRead: aws.Bool(true),
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read overall contract state: %w", err)
		}
		if result.Item != nil && domain.ContractState(getString(result.Item, "State")) == domain.ContractCompleted {
			completed[credit.ContractID] = true
		}
	}
	return credits, completed, nil
}
