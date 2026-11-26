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

// MeritLedgerStore writes and reads merit credits from DynamoDB.
type MeritLedgerStore struct {
	client    *DynamoDBClient
	tableName string
}

// NewMeritLedgerStore creates a store and ensures the table exists.
func NewMeritLedgerStore(client *DynamoDBClient, tableName string) *MeritLedgerStore {
	return &MeritLedgerStore{client: client, tableName: tableName}
}

// creditItem builds the DynamoDB item for a merit credit.
func creditItem(credit domain.MeritCredit) map[string]types.AttributeValue {
	item := map[string]types.AttributeValue{
		"PK":             &types.AttributeValueMemberS{Value: "MERIT_CREDIT"},
		"SK":             &types.AttributeValueMemberS{Value: credit.ConfirmationID},
		"GSI1PK":         &types.AttributeValueMemberS{Value: creditBucketKey(credit)},
		"GSI1SK":         &types.AttributeValueMemberS{Value: credit.ConfirmationID},
		"GSI2PK":         &types.AttributeValueMemberS{Value: leaderboardBucketKey(credit)},
		"GSI2SK":         &types.AttributeValueMemberS{Value: leaderboardSortKey(credit)},
		"SeasonID":       &types.AttributeValueMemberS{Value: credit.SeasonID},
		"Category":       &types.AttributeValueMemberS{Value: credit.Category},
		"Track":          &types.AttributeValueMemberS{Value: string(credit.Track)},
		"RecipientID":    &types.AttributeValueMemberS{Value: credit.RecipientID},
		"Amount":         &types.AttributeValueMemberN{Value: strconv.Itoa(credit.Amount)},
		"ContractID":     &types.AttributeValueMemberS{Value: credit.ContractID},
		"ObjectiveID":    &types.AttributeValueMemberS{Value: credit.ObjectiveID},
		"UnitID":         &types.AttributeValueMemberS{Value: credit.UnitID},
		"ConfirmationID": &types.AttributeValueMemberS{Value: credit.ConfirmationID},
	}
	if credit.MediaURI != "" {
		item["MediaURI"] = &types.AttributeValueMemberS{Value: string(credit.MediaURI)}
	}
	return item
}

// Save records a credit in the ledger.
func (s *MeritLedgerStore) Save(credit domain.MeritCredit) error {
	ctx := context.Background()

	_, err := s.client.Client().TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			seasonUnlockedCheck(s.tableName, credit.SeasonID),
			{Put: &types.Put{TableName: aws.String(s.tableName), Item: creditItem(credit)}},
		},
	})
	if seasonLockCheckFailed(err) {
		return adjudicationFenceError(ctx, s.client, s.tableName, credit.SeasonID, true)
	}
	if err != nil {
		return fmt.Errorf("failed to save merit credit: %w", err)
	}

	return nil
}
