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

// SaveApproved commits the adjudicated contract state and every materialized
// merit credit in one DynamoDB transaction: the contract put, the roster
// active-set release, the still-active members' definition-reservation
// release, and one conditional credit put per staged credit.
// Exactly-once comes from the confirmation-id conditional put — a retried
// approval under a stale read surfaces the duplicate signal. The contract
// put requires the row to exist with the same departed set the adjudication
// saw, so a review started before a committed leave can neither overwrite
// the departure nor attach credits to a stale roster (REG-0053-05).
func (s *ContractStore) SaveApproved(contract *domain.MissionContract, credits []domain.MeritCredit) error {
	ctx := context.Background()

	names := map[string]string{}
	values := map[string]types.AttributeValue{}
	condition := "attribute_exists(PK) AND " + departedSetMatch(contract.Departures(), names, values)
	items := []types.TransactWriteItem{
		{Put: &types.Put{
			TableName:                 aws.String(s.tableName),
			Item:                      contractItem(contract),
			ConditionExpression:       aws.String(condition),
			ExpressionAttributeNames:  names,
			ExpressionAttributeValues: values,
		}},
	}
	items = append(items, rosterReleaseTx(contract, s.tableName)...)
	items = append(items, reservationReleaseTx(contract, s.tableName, contract.ActiveMembers())...)
	creditStart := len(items)
	for _, credit := range credits {
		items = append(items, types.TransactWriteItem{Put: &types.Put{
			TableName:           aws.String(s.tableName),
			Item:                creditItem(credit),
			ConditionExpression: aws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
		}})
	}

	if len(credits) > 0 {
		err := s.saveCreditBearing(ctx, contract.SeasonID, items)
		if dup := duplicateOnConditionalFail(err, creditStart+1, len(items)+1); dup != nil {
			return dup
		}
		return err
	}
	_, err := s.client.Client().TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{TransactItems: items})
	if dup := duplicateOnConditionalFail(err, creditStart, len(items)); dup != nil {
		return dup
	}
	if err != nil {
		return fmt.Errorf("failed to commit approval: %w", err)
	}
	return nil
}

// rosterReleaseTx removes the contract id from each roster member's active set;
// shared by every terminal-state write path.
func rosterReleaseTx(contract *domain.MissionContract, tableName string) []types.TransactWriteItem {
	items := make([]types.TransactWriteItem, 0, len(contract.Roster()))
	for _, playerID := range contract.Roster() {
		items = append(items, types.TransactWriteItem{Update: &types.Update{
			TableName:        aws.String(tableName),
			Key:              playerActiveKey(playerID),
			UpdateExpression: aws.String("DELETE contract_ids :cid"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":cid": &types.AttributeValueMemberSS{Value: []string{contract.ContractID}},
			},
		}})
	}
	return items
}

// pendingCreditsToAttribute serializes staged credits as a list of maps so the
// contract item carries the full adjudication context.
func pendingCreditsToAttribute(credits []domain.MeritCredit) types.AttributeValue {
	items := make([]types.AttributeValue, 0, len(credits))
	for _, c := range credits {
		m := map[string]types.AttributeValue{
			"ConfirmationID": &types.AttributeValueMemberS{Value: c.ConfirmationID},
			"ObjectiveID":    &types.AttributeValueMemberS{Value: c.ObjectiveID},
			"UnitID":         &types.AttributeValueMemberS{Value: c.UnitID},
			"RecipientID":    &types.AttributeValueMemberS{Value: c.RecipientID},
			"SeasonID":       &types.AttributeValueMemberS{Value: c.SeasonID},
			"Category":       &types.AttributeValueMemberS{Value: c.Category},
			"Track":          &types.AttributeValueMemberS{Value: string(c.Track)},
			"ContractID":     &types.AttributeValueMemberS{Value: c.ContractID},
			"Amount":         &types.AttributeValueMemberN{Value: strconv.Itoa(c.Amount)},
		}
		if c.MediaURI != "" {
			m["MediaURI"] = &types.AttributeValueMemberS{Value: string(c.MediaURI)}
		}
		items = append(items, &types.AttributeValueMemberM{Value: m})
	}
	return &types.AttributeValueMemberL{Value: items}
}

// pendingCreditsFromItem restores staged credits from a persisted contract item.
func pendingCreditsFromItem(item map[string]types.AttributeValue) []domain.MeritCredit {
	list, ok := item["PendingCredits"].(*types.AttributeValueMemberL)
	if !ok || len(list.Value) == 0 {
		return nil
	}
	credits := make([]domain.MeritCredit, 0, len(list.Value))
	for _, v := range list.Value {
		m, ok := v.(*types.AttributeValueMemberM)
		if !ok {
			continue
		}
		credits = append(credits, domain.MeritCredit{
			ConfirmationID: getString(m.Value, "ConfirmationID"),
			ObjectiveID:    getString(m.Value, "ObjectiveID"),
			UnitID:         getString(m.Value, "UnitID"),
			RecipientID:    getString(m.Value, "RecipientID"),
			SeasonID:       getString(m.Value, "SeasonID"),
			Category:       getString(m.Value, "Category"),
			Track:          domain.Track(getString(m.Value, "Track")),
			ContractID:     getString(m.Value, "ContractID"),
			Amount:         getInt(m.Value, "Amount"),
			MediaURI:       domain.MediaURI(getString(m.Value, "MediaURI")),
		})
	}
	return credits
}
