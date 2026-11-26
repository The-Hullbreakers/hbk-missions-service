package dynamodb

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// definitionGuardKey identifies one player's reservation for a stable mission
// definition: one item per (player, definition), owned by the contract that
// reserved it. The sort key uses the legacy OfferID fallback, matching
// Summary() and board exclusion.
func definitionGuardKey(playerID, definitionCode string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: "PLAYER_DEF_ACTIVE#" + playerID},
		"SK": &types.AttributeValueMemberS{Value: definitionCode},
	}
}

// contractDefinitionCode resolves the reservation identity for a contract.
func contractDefinitionCode(contract *domain.MissionContract) string {
	if contract.Offer.DefinitionCode != "" {
		return contract.Offer.DefinitionCode
	}
	return contract.Offer.OfferID
}

// reservationPutTx reserves the definition for the roster member inside the
// acceptance transaction; any existing reservation cancels the write.
func reservationPutTx(contract *domain.MissionContract, playerID, tableName string) types.TransactWriteItem {
	item := definitionGuardKey(playerID, contractDefinitionCode(contract))
	item["contract_id"] = &types.AttributeValueMemberS{Value: contract.ContractID}
	return types.TransactWriteItem{Put: &types.Put{
		TableName:           aws.String(tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	}}
}

// reservationReleaseTx deletes each member's reservation only when it is
// absent (legacy contracts) or still owned by this contract — a stale save
// can never destroy a newer reservation owned by another contract.
func reservationReleaseTx(contract *domain.MissionContract, tableName string, members []string) []types.TransactWriteItem {
	items := make([]types.TransactWriteItem, 0, len(members))
	for _, playerID := range members {
		items = append(items, types.TransactWriteItem{Delete: &types.Delete{
			TableName:           aws.String(tableName),
			Key:                 definitionGuardKey(playerID, contractDefinitionCode(contract)),
			ConditionExpression: aws.String("attribute_not_exists(PK) OR contract_id = :cid"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":cid": &types.AttributeValueMemberS{Value: contract.ContractID},
			},
		}})
	}
	return items
}
