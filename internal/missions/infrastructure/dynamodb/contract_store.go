package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// ContractStore loads and persists mission contracts.
type ContractStore struct {
	client    *DynamoDBClient
	tableName string
}

// NewContractStore creates a DynamoDB contract store.
func NewContractStore(client *DynamoDBClient, tableName string) *ContractStore {
	return &ContractStore{client: client, tableName: tableName}
}

// Get loads a contract by identity.
func (s *ContractStore) Get(contractID string) (*domain.MissionContract, error) {
	ctx := context.Background()

	result, err := s.client.Client().GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "CONTRACT"},
			"SK": &types.AttributeValueMemberS{Value: contractID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get contract: %w", err)
	}
	if result.Item == nil {
		return nil, domain.ErrContractNil
	}

	return mapToContract(result.Item)
}

// Save persists a contract. Terminal states also release the contract id from
// each roster member's active set so the acceptance limit frees the slot, and
// release each still-active member's definition reservation — a departed
// member's reservation is never touched because it may belong to a newer
// attempt. Every contract put is fenced on the stored departed set matching
// the candidate's: a snapshot written before a committed departure conflicts
// instead of resurrecting the leaver (REG-0053-04).
func (s *ContractStore) Save(contract *domain.MissionContract) error {
	ctx := context.Background()

	names := map[string]string{}
	values := map[string]types.AttributeValue{}
	condition := departedSetMatch(contract.Departures(), names, values)

	if !contract.State.IsTerminal() {
		if len(contract.PendingCredits()) > 0 {
			return s.saveCreditBearing(ctx, contract.SeasonID, []types.TransactWriteItem{{Put: &types.Put{
				TableName: aws.String(s.tableName), Item: contractItem(contract),
				ConditionExpression:      aws.String(condition),
				ExpressionAttributeNames: names, ExpressionAttributeValues: values,
			}}})
		}
		_, err := s.client.Client().PutItem(ctx, &dynamodb.PutItemInput{
			TableName:                 aws.String(s.tableName),
			Item:                      contractItem(contract),
			ConditionExpression:       aws.String(condition),
			ExpressionAttributeNames:  names,
			ExpressionAttributeValues: values,
		})
		if err != nil {
			return fmt.Errorf("failed to save contract: %w", err)
		}
		return nil
	}

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
	if len(contract.PendingCredits()) > 0 {
		return s.saveCreditBearing(ctx, contract.SeasonID, items)
	}
	if _, err := s.client.Client().TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{TransactItems: items}); err != nil {
		return fmt.Errorf("failed to save contract: %w", err)
	}
	return nil
}

// departedSetMatch returns the shared write fence fragment: the stored
// Departed list must equal exactly expected (order-independent size plus a
// contains per member). A missing Departed attribute is tolerated only when
// expected is empty — which also lets an absent row pass for the legacy
// create path, while a candidate with departed members can never create one.
// Placeholders are added to names/values in the same step they are referenced.
func departedSetMatch(expected []string, names map[string]string, values map[string]types.AttributeValue) string {
	names["#departed"] = "Departed"
	values[":departedCount"] = &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", len(expected))}
	if len(expected) == 0 {
		return "(attribute_not_exists(#departed) OR size(#departed) = :departedCount)"
	}
	condition := "size(#departed) = :departedCount"
	for i, member := range expected {
		key := fmt.Sprintf(":departed%d", i)
		condition += fmt.Sprintf(" AND contains(#departed, %s)", key)
		values[key] = &types.AttributeValueMemberS{Value: member}
	}
	return "(" + condition + ")"
}

// ListPendingReview lists every contract awaiting reviewer adjudication —
// the PENDING_VALIDATION set the review queue surfaces.
func (s *ContractStore) ListPendingReview() ([]domain.MissionContract, error) {
	contracts, err := s.listAll()
	if err != nil {
		return nil, err
	}

	pending := make([]domain.MissionContract, 0, len(contracts))
	for _, contract := range contracts {
		if contract.State == domain.ContractPendingValidation {
			pending = append(pending, contract)
		}
	}
	return pending, nil
}
