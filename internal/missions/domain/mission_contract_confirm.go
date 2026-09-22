package domain

// Confirm validates the objective/season/unit and records a new merit credit.
func (c *MissionContract) Confirm(objective Objective, unitID string, season Season, tenantID string) (MeritCredit, error) {
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

	return MeritCredit{
		TenantID:       tenantID,
		SeasonID:       season.SeasonID,
		Category:       objective.Category,
		Track:          track,
		RecipientID:    recipient,
		Amount:         objective.MeritPerUnit + objective.Bonus,
		ContractID:     c.ContractID,
		ObjectiveID:    objective.ObjectiveID,
		UnitID:         unitID,
		ConfirmationID: confirmationID.String(),
	}, nil
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
