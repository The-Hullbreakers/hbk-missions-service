package domain

// MissionContract is a player or team commitment to a published mission offer.
type MissionContract struct {
	ContractID     string
	Offer          MissionOffer
	State          ContractState
	SeasonID       string
	roster         []string
	TeamID         string
	rosterFrozen   bool
	confirmedUnits map[string]bool
}

// Roster returns a defensive copy of the frozen roster.
func (c *MissionContract) Roster() []string {
	copied := make([]string, len(c.roster))
	copy(copied, c.roster)
	return copied
}

// NewMissionContract creates an accepted solo contract with the given identity, offer and roster.
func NewMissionContract(contractID string, offer MissionOffer, roster []string) *MissionContract {
	copied := make([]string, len(roster))
	copy(copied, roster)
	return NewMissionContractFromState(contractID, offer, copied, ContractAccepted, false, "", offer.SeasonID, nil)
}

// NewTeamMissionContract creates an accepted team contract with the given identity, offer, roster and team.
func NewTeamMissionContract(contractID string, offer MissionOffer, roster []string, teamID string) *MissionContract {
	copied := make([]string, len(roster))
	copy(copied, roster)
	return NewMissionContractFromState(contractID, offer, copied, ContractAccepted, false, teamID, offer.SeasonID, nil)
}

// NewMissionContractFromState reconstructs a contract from persisted state.
func NewMissionContractFromState(contractID string, offer MissionOffer, roster []string, state ContractState, rosterFrozen bool, teamID string, seasonID string, confirmed []string) *MissionContract {
	confirmedUnits := make(map[string]bool, len(confirmed))
	for _, id := range confirmed {
		confirmedUnits[id] = true
	}
	return &MissionContract{
		ContractID:     contractID,
		Offer:          offer,
		State:          state,
		SeasonID:       seasonID,
		roster:         roster,
		TeamID:         teamID,
		rosterFrozen:   rosterFrozen,
		confirmedUnits: confirmedUnits,
	}
}

// Start transitions an accepted contract to IN_PROGRESS and freezes its roster.
// Once frozen, the roster copy returned by Roster() cannot affect the internal roster.
func (c *MissionContract) Start() {
	if c.State == ContractAccepted {
		c.State = ContractInProgress
		c.rosterFrozen = true
	}
}

// Abandon transitions an active contract to the terminal ABANDONED state.
func (c *MissionContract) Abandon() {
	if !c.State.IsTerminal() {
		c.State = ContractAbandoned
	}
}

// RosterFrozen returns true once the roster has been frozen at the start of the contract.
func (c *MissionContract) RosterFrozen() bool {
	return c.rosterFrozen
}

// Summary returns a lightweight, policy-ready value object for acceptance decisions.
func (c *MissionContract) Summary() ActiveContractSummary {
	definitionID := c.Offer.DefinitionID
	if definitionID == "" {
		definitionID = c.Offer.OfferID
	}
	return NewActiveContractSummary(c.Offer.OfferID, c.Offer.Category, c.State, definitionID)
}

// Confirmations returns the already confirmed unit identifiers.
func (c *MissionContract) Confirmations() []string {
	confirmed := make([]string, 0, len(c.confirmedUnits))
	for id := range c.confirmedUnits {
		confirmed = append(confirmed, id)
	}
	return confirmed
}
