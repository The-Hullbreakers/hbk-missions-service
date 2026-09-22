package domain

// ConfirmObjective validates a unit and creates a merit credit for an active contract and season.
func ConfirmObjective(contract *MissionContract, objective Objective, unitID string, season Season, tenantID string) (MeritCredit, error) {
	return contract.Confirm(objective, unitID, season, tenantID)
}
