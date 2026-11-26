package dynamodb

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func mapToContract(item map[string]types.AttributeValue) (*domain.MissionContract, error) {
	contractID := getString(item, "SK")
	offerID := getString(item, "OfferID")
	offer := domain.MissionOffer{
		OfferID:           offerID,
		Title:             getString(item, "Title"),
		Description:       getString(item, "Description"),
		Category:          getString(item, "Category"),
		Mode:              domain.MissionMode(getString(item, "Mode")),
		CompletionReward:  getInt(item, "CompletionReward"),
		Rules:             getStringSlice(item, "Rules"),
		SeasonID:          getString(item, "OfferSeasonID"),
		DefinitionCode:    getString(item, "DefinitionCode"),
		DefinitionVersion: getString(item, "DefinitionVersion"),
		ContentEn:         englishContentFromItem(item),
	}
	state := domain.ContractState(getString(item, "State"))
	roster := getStringSlice(item, "Roster")
	departed := getStringSlice(item, "Departed")
	teamID := getString(item, "TeamID")
	seasonID := getString(item, "SeasonID")
	rosterFrozen := getBool(item, "RosterFrozen")
	confirmed := getStringSlice(item, "Confirmed")

	contract := domain.NewMissionContractFromState(contractID, offer, roster, state, rosterFrozen, teamID, seasonID, confirmed, departed, pendingCreditsFromItem(item))
	contract.RestoreRejectedUnits(getStringSlice(item, "RejectedUnits"))
	contract.RestoreInvitationState(invitationsFromItem(item), getTime(item, "InvitedAt"))
	contract.RestoreCurrentPhase(getInt(item, "CurrentPhase"))
	contract.RestoreDisputes(disputesFromItem(item, "Disputes"))
	if phases := phasesFromItem(item, "Phases"); len(phases) > 0 {
		if terms, err := domain.NewContractTermsPhased(phases); err == nil {
			contract.Terms = terms
		}
	} else if terms, err := domain.NewContractTerms(objectivesFromItem(item, "Terms")); err == nil {
		contract.Terms = terms
	}
	return contract, nil
}

func invitationsFromItem(item map[string]types.AttributeValue) map[string]domain.InvitationState {
	raw := getStringMap(item, "Invitations")
	if len(raw) == 0 {
		return nil
	}
	invitations := make(map[string]domain.InvitationState, len(raw))
	for member, state := range raw {
		invitations[member] = domain.InvitationState(state)
	}
	return invitations
}
