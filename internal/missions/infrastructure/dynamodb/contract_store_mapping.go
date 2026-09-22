package dynamodb

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func mapToContract(item map[string]types.AttributeValue) (*domain.MissionContract, error) {
	contractID := getString(item, "SK")
	offerID := getString(item, "OfferID")
	definitionID := getString(item, "DefinitionID")
	definitionVersion := getString(item, "DefinitionVersion")
	if definitionID == "" {
		definitionID = offerID
		definitionVersion = "legacy"
	}
	offer := domain.MissionOffer{
		OfferID:           offerID,
		Category:          getString(item, "Category"),
		Capacity:          getInt(item, "Capacity"),
		Mode:              domain.MissionMode(getString(item, "Mode")),
		SeasonID:          getString(item, "OfferSeasonID"),
		DefinitionID:      definitionID,
		DefinitionVersion: definitionVersion,
	}
	state := domain.ContractState(getString(item, "State"))
	roster := getStringSlice(item, "Roster")
	teamID := getString(item, "TeamID")
	seasonID := getString(item, "SeasonID")
	rosterFrozen := getBool(item, "RosterFrozen")
	confirmed := getStringSlice(item, "Confirmed")

	contract := domain.NewMissionContractFromState(contractID, offer, roster, state, rosterFrozen, teamID, seasonID, confirmed)
	return contract, nil
}
