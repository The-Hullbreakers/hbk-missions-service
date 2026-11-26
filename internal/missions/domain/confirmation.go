package domain

// ConfirmObjective validates a unit and creates a merit credit for an active contract and season.
func ConfirmObjective(contract *MissionContract, objectiveID, unitID string, season Season) (MeritCredit, error) {
	return contract.Confirm(objectiveID, unitID, season)
}
