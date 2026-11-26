package domain

// Confirm validates the season/unit and stages a new merit credit scored
// from the contract's frozen ContractTerms. The credit is returned as a
// projection for the response but only enters the ledger once a reviewer
// approves the contract (REQ-MCF-003, CHG-0031). Objectives outside the
// snapshot are rejected; unit limits and bonuses come from the terms.
func (c *MissionContract) Confirm(objectiveID, unitID string, season Season) (MeritCredit, error) {
	return c.confirm(objectiveID, unitID, "", season)
}

// ConfirmWithMedia behaves like Confirm and additionally records the already
// validated media reference on the resulting credit.
func (c *MissionContract) ConfirmWithMedia(objectiveID, unitID string, media MediaURI, season Season) (MeritCredit, error) {
	return c.confirm(objectiveID, unitID, media, season)
}

func (c *MissionContract) confirm(objectiveID, unitID string, media MediaURI, season Season) (MeritCredit, error) {
	if c == nil {
		return MeritCredit{}, ErrContractNil
	}
	if c.State.IsTerminal() {
		return MeritCredit{}, ErrContractTerminal
	}
	if c.State != ContractInProgress {
		return MeritCredit{}, ErrContractNotInProgress
	}
	if !season.CanAcceptMerits(c) {
		return MeritCredit{}, ErrSeasonNotActive
	}
	if season.SeasonID != c.SeasonID {
		return MeritCredit{}, ErrSeasonMismatch
	}
	objective, ok := c.Terms.Lookup(objectiveID)
	if !ok {
		return MeritCredit{}, ErrObjectiveNotInContract
	}
	if phase := c.Terms.ObjectivePhase(objectiveID); phase >= 0 && phase != c.currentPhase {
		return MeritCredit{}, ErrPhaseNotUnlocked
	}
	if c.Offer.Category != objective.Category {
		return MeritCredit{}, ErrCategoryMismatch
	}
	roster := c.Roster()
	if len(roster) == 0 {
		return MeritCredit{}, ErrRosterEmpty
	}

	recipient := c.recipient(roster)
	track := c.track()
	confirmationID, err := NewConfirmationID(c.ContractID, objective.ObjectiveID, unitID, recipient)
	if err != nil {
		return MeritCredit{}, err
	}
	if c.confirmedUnits[confirmationID.String()] {
		return MeritCredit{}, ErrDuplicateConfirmation
	}
	if objective.UnitLimit > 0 && c.confirmedCountFor(objective.ObjectiveID) >= objective.UnitLimit {
		return MeritCredit{}, ErrObjectiveLimitReached
	}
	c.confirmedUnits[confirmationID.String()] = true

	credit := MeritCredit{
		SeasonID:       season.SeasonID,
		Category:       objective.Category,
		Track:          track,
		RecipientID:    recipient,
		Amount:         objective.MeritPerUnit + objective.Bonus,
		ContractID:     c.ContractID,
		ObjectiveID:    objective.ObjectiveID,
		UnitID:         unitID,
		ConfirmationID: confirmationID.String(),
		MediaURI:       media,
	}
	c.pendingCredits[confirmationID.String()] = credit
	c.advancePhases()
	return credit, nil
}

// advancePhases moves the progression marker forward while the current phase's
// mandatory objectives are all confirmed — optional objectives never gate. The
// marker clamps at the last phase: a completed final phase keeps its objectives
// eligible (remaining unit credits may still confirm).
func (c *MissionContract) advancePhases() {
	for c.currentPhase < c.Terms.PhaseCount()-1 && c.Terms.PhaseComplete(c.currentPhase, c.objectiveConfirmed) {
		c.currentPhase++
	}
}

func (c *MissionContract) recipient(roster []string) string {
	if c.TeamID != "" {
		return c.TeamID
	}
	return roster[0]
}

func (c *MissionContract) track() Track {
	if c.TeamID != "" {
		return TrackTeam
	}
	return TrackIndividual
}

func (c *MissionContract) confirmedCountFor(objectiveID string) int {
	count := 0
	for raw := range c.confirmedUnits {
		id, err := ParseConfirmationID(raw)
		if err != nil {
			continue
		}
		if id.ObjectiveID() == objectiveID {
			count++
		}
	}
	return count
}
