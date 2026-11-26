package domain

// Start transitions an accepted contract to IN_PROGRESS and freezes its roster.
// Only a roster member may start it. Once frozen, the
// roster copy returned by Roster() cannot affect the internal roster.
func (c *MissionContract) Start(identity CallerIdentity) error {
	if !c.CanSubmitEvidence(identity) {
		return ErrCallerNotAuthorized
	}
	if c.State != ContractAccepted {
		if c.State.IsTerminal() {
			return ErrContractTerminal
		}
		return ErrContractStateInvalid
	}
	c.State = ContractInProgress
	c.rosterFrozen = true
	return nil
}

// Leave removes the caller from the contract's active membership. The
// contract keeps running for the remaining members; the last active member
// leaving abandons it. Departed members lose all contract rights.
func (c *MissionContract) Leave(identity CallerIdentity) error {
	if !c.CanSubmitEvidence(identity) {
		return ErrCallerNotAuthorized
	}
	if c.State.IsTerminal() {
		return ErrContractTerminal
	}
	if c.hasOpenDispute() {
		return ErrDisputesPending
	}
	c.departed[identity.PlayerID()] = true
	if len(c.ActiveMembers()) == 0 {
		c.State = ContractAbandoned
	}
	return nil
}

// Abandon transitions an active contract to the terminal ABANDONED state.
// Only a roster member may abandon it.
func (c *MissionContract) Abandon(identity CallerIdentity) error {
	if !c.CanSubmitEvidence(identity) {
		return ErrCallerNotAuthorized
	}
	if c.State.IsTerminal() {
		return ErrContractTerminal
	}
	if c.hasOpenDispute() {
		return ErrDisputesPending
	}
	c.State = ContractAbandoned
	return nil
}

// Expire transitions a non-terminal contract to EXPIRED. System-driven: no
// actor is required because expiry is derived from the offer window.
func (c *MissionContract) Expire() error {
	if c.State.IsTerminal() {
		return ErrContractTerminal
	}
	c.State = ContractExpired
	return nil
}

// objectiveConfirmed reports whether the objective has at least one confirmed
// unit recorded under the frozen terms.
func (c *MissionContract) objectiveConfirmed(objectiveID string) bool {
	return c.confirmedCountFor(objectiveID) > 0
}
