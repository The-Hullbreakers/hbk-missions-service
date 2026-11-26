package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// SaveDeparture persists a contract after a member left, releasing only the
// departing member's guards atomically: their PLAYER_ACTIVE slot and their
// definition reservation (when still owned by this contract). A reservation
// owned by a newer contract cancels the write — it is never deleted. The
// contract put itself is conditional: the stored row must still carry the
// state and the exact set of previously departed members the leaver saw, so
// a stale snapshot cancels the transaction instead of resurrecting an
// already-departed member (REG-0053-03). previousState is the non-terminal
// state the leave was decided from; the contract may only reach ABANDONED
// when this departure emptied the active roster (last-member exit).
func (s *ContractStore) SaveDeparture(contract *domain.MissionContract, departingPlayerID string, previousState domain.ContractState) error {
	if previousState.IsTerminal() {
		return fmt.Errorf("departure requires a non-terminal prior state, got %s", previousState)
	}
	if contract.State != previousState &&
		(contract.State != domain.ContractAbandoned || len(contract.ActiveMembers()) != 0) {
		return fmt.Errorf("departure of %s cannot move contract %s from %s to %s",
			departingPlayerID, contract.ContractID, previousState, contract.State)
	}
	if !contract.IsDeparted(departingPlayerID) {
		return fmt.Errorf("player %s has not departed contract %s", departingPlayerID, contract.ContractID)
	}
	ctx := context.Background()

	condition, names, values := departureWriteCondition(contract, departingPlayerID, previousState)
	items := []types.TransactWriteItem{
		{Put: &types.Put{
			TableName:                 aws.String(s.tableName),
			Item:                      contractItem(contract),
			ConditionExpression:       aws.String(condition),
			ExpressionAttributeNames:  names,
			ExpressionAttributeValues: values,
		}},
		{Update: &types.Update{
			TableName:        aws.String(s.tableName),
			Key:              playerActiveKey(departingPlayerID),
			UpdateExpression: aws.String("DELETE contract_ids :cid"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":cid": &types.AttributeValueMemberSS{Value: []string{contract.ContractID}},
			},
		}},
	}
	items = append(items, reservationReleaseTx(contract, s.tableName, []string{departingPlayerID})...)

	if _, err := s.client.Client().TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{TransactItems: items}); err != nil {
		return fmt.Errorf("failed to save departure: %w", err)
	}
	return nil
}

// departureWriteCondition builds the optimistic fence for the contract put:
// the row must exist, carry the lifecycle state the leaver saw, and its
// Departed list must equal exactly the members departed before this leave
// (the shared departed-set fence).
func departureWriteCondition(contract *domain.MissionContract, departingPlayerID string, previousState domain.ContractState) (string, map[string]string, map[string]types.AttributeValue) {
	prior := make([]string, 0, len(contract.Departures()))
	for _, id := range contract.Departures() {
		if id != departingPlayerID {
			prior = append(prior, id)
		}
	}
	names := map[string]string{"#state": "State"}
	values := map[string]types.AttributeValue{
		":state": &types.AttributeValueMemberS{Value: string(previousState)},
	}
	condition := "attribute_exists(PK) AND #state = :state AND " + departedSetMatch(prior, names, values)
	return condition, names, values
}
