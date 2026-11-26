package domain

import (
	"sort"
	"time"
)

// MissionContract is a player or team commitment to a published mission offer.
type MissionContract struct {
	ContractID     string
	Offer          MissionOffer
	Terms          ContractTerms
	State          ContractState
	SeasonID       string
	roster         []string
	TeamID         string
	rosterFrozen   bool
	departed       map[string]bool
	confirmedUnits map[string]bool
	rejectedUnits  map[string]bool
	pendingCredits map[string]MeritCredit
	invitations    map[string]InvitationState
	invitedAt      time.Time
	currentPhase   int
	disputes       map[string]Dispute
}

// CurrentPhase returns the index of the phase whose objectives are eligible.
func (c *MissionContract) CurrentPhase() int { return c.currentPhase }

// RestoreCurrentPhase rehydrates persisted phase progression.
func (c *MissionContract) RestoreCurrentPhase(index int) { c.currentPhase = index }

// NewMissionContract creates an accepted solo contract with the given identity, offer and roster.
func NewMissionContract(contractID string, offer MissionOffer, roster []string) *MissionContract {
	copied := make([]string, len(roster))
	copy(copied, roster)
	c := NewMissionContractFromState(contractID, offer, copied, ContractAccepted, false, "", offer.SeasonID, nil, nil, nil)
	c.updateRosterFrozen()
	return c
}

// NewTeamMissionContract creates an accepted team contract with the given identity, offer, roster and team.
// The requester (roster[0]) is confirmed; every other member is invited pending.
func NewTeamMissionContract(contractID string, offer MissionOffer, roster []string, teamID string) *MissionContract {
	copied := make([]string, len(roster))
	copy(copied, roster)
	invitations, invitedAt := newPendingInvitations(copied, time.Now())
	c := NewMissionContractFromState(contractID, offer, copied, ContractAccepted, false, teamID, offer.SeasonID, nil, nil, nil)
	c.invitations = invitations
	c.invitedAt = invitedAt
	c.updateRosterFrozen()
	return c
}

// NewMissionContractFromState reconstructs a contract from persisted state.
func NewMissionContractFromState(contractID string, offer MissionOffer, roster []string, state ContractState, rosterFrozen bool, teamID string, seasonID string, confirmed []string, departed []string, pending []MeritCredit) *MissionContract {
	confirmedUnits := make(map[string]bool, len(confirmed))
	for _, id := range confirmed {
		confirmedUnits[id] = true
	}
	departedSet := make(map[string]bool, len(departed))
	for _, id := range departed {
		departedSet[id] = true
	}
	pendingCredits := make(map[string]MeritCredit, len(pending))
	for _, credit := range pending {
		pendingCredits[credit.ConfirmationID] = credit
	}
	return &MissionContract{
		ContractID:     contractID,
		Offer:          offer,
		State:          state,
		SeasonID:       seasonID,
		roster:         roster,
		TeamID:         teamID,
		rosterFrozen:   rosterFrozen,
		departed:       departedSet,
		confirmedUnits: confirmedUnits,
		rejectedUnits:  map[string]bool{},
		pendingCredits: pendingCredits,
	}
}

// CanSubmitEvidence reports whether the caller may submit evidence for this
// contract: the caller must be a roster member. Privileged roles are evaluated
// by the application layer on top of this rule.
func (c *MissionContract) CanSubmitEvidence(identity CallerIdentity) bool {
	if c == nil {
		return false
	}
	return c.IncludesPlayer(identity.PlayerID()) && !c.IsDeparted(identity.PlayerID())
}

// Summary returns a lightweight, policy-ready value object for acceptance decisions.
func (c *MissionContract) Summary() ActiveContractSummary {
	definitionCode := c.Offer.DefinitionCode
	if definitionCode == "" {
		definitionCode = c.Offer.OfferID
	}
	return NewActiveContractSummary(c.Offer.OfferID, c.Offer.Category, c.State, definitionCode)
}

// Confirmations returns the already confirmed unit identifiers.
func (c *MissionContract) Confirmations() []string {
	confirmed := make([]string, 0, len(c.confirmedUnits))
	for id := range c.confirmedUnits {
		confirmed = append(confirmed, id)
	}
	return confirmed
}

// RejectedUnits returns the confirmation ids the approving reviewer marked
// as failed — the staged credits were discarded and never reached the ledger.
func (c *MissionContract) RejectedUnits() []string {
	rejected := make([]string, 0, len(c.rejectedUnits))
	for id := range c.rejectedUnits {
		rejected = append(rejected, id)
	}
	sort.Strings(rejected)
	return rejected
}

// RestoreRejectedUnits rehydrates the persisted adjudication record.
func (c *MissionContract) RestoreRejectedUnits(ids []string) {
	c.rejectedUnits = make(map[string]bool, len(ids))
	for _, id := range ids {
		c.rejectedUnits[id] = true
	}
}

// PendingCredits returns the credits staged by evidence confirmations that a
// reviewer has not yet adjudicated. Approval releases them to the ledger;
// rejection discards them.
func (c *MissionContract) PendingCredits() []MeritCredit {
	credits := make([]MeritCredit, 0, len(c.pendingCredits))
	for _, credit := range c.pendingCredits {
		credits = append(credits, credit)
	}
	return credits
}
