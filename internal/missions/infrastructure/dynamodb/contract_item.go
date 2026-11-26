package dynamodb

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// contractItem builds the DynamoDB item for a contract. Staged credits travel
// inside the item as PendingCredits until reviewer approval materializes them
// as ledger items.
func contractItem(contract *domain.MissionContract) map[string]types.AttributeValue {
	item := map[string]types.AttributeValue{
		"PK":                &types.AttributeValueMemberS{Value: "CONTRACT"},
		"SK":                &types.AttributeValueMemberS{Value: contract.ContractID},
		"OfferID":           &types.AttributeValueMemberS{Value: contract.Offer.OfferID},
		"Title":             &types.AttributeValueMemberS{Value: contract.Offer.Title},
		"Description":       &types.AttributeValueMemberS{Value: contract.Offer.Description},
		"Category":          &types.AttributeValueMemberS{Value: contract.Offer.Category},
		"Mode":              &types.AttributeValueMemberS{Value: string(contract.Offer.Mode)},
		"CompletionReward":  &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", contract.Offer.CompletionReward)},
		"Rules":             stringSliceToAttribute(contract.Offer.Rules),
		"OfferSeasonID":     &types.AttributeValueMemberS{Value: contract.Offer.SeasonID},
		"DefinitionCode":    &types.AttributeValueMemberS{Value: contract.Offer.DefinitionCode},
		"DefinitionVersion": &types.AttributeValueMemberS{Value: contract.Offer.DefinitionVersion},
		"Roster":            stringSliceToAttribute(contract.Roster()),
		"Departed":          stringSliceToAttribute(contract.Departures()),
		"TeamID":            &types.AttributeValueMemberS{Value: contract.TeamID},
		"State":             &types.AttributeValueMemberS{Value: string(contract.State)},
		"RosterFrozen":      boolToAttribute(contract.RosterFrozen()),
		"Confirmed":         stringSliceToAttribute(contract.Confirmations()),
		"SeasonID":          &types.AttributeValueMemberS{Value: contract.SeasonID},
		"Terms":             objectivesToAttribute(contract.Terms.Objectives()),
	}
	putEnglishContent(item, contract.Offer.ContentEn)
	if phases := contract.Terms.Phases(); len(phases) > 1 {
		item["Phases"] = phasesToAttribute(phases)
	}
	item["CurrentPhase"] = &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", contract.CurrentPhase())}
	if disputes := contract.Disputes(); len(disputes) > 0 {
		item["Disputes"] = disputesToAttribute(disputes)
	}
	if pending := contract.PendingCredits(); len(pending) > 0 {
		item["PendingCredits"] = pendingCreditsToAttribute(pending)
	}
	if rejected := contract.RejectedUnits(); len(rejected) > 0 {
		item["RejectedUnits"] = stringSliceToAttribute(rejected)
	}
	if invitations := contract.Invitations(); len(invitations) > 0 {
		item["Invitations"] = stringMapToAttribute(invitationStatesToStrings(invitations))
		item["InvitedAt"] = timeToAttribute(contract.InvitedAt())
	}
	return item
}

func invitationStatesToStrings(states map[string]domain.InvitationState) map[string]string {
	strings := make(map[string]string, len(states))
	for k, v := range states {
		strings[k] = string(v)
	}
	return strings
}
