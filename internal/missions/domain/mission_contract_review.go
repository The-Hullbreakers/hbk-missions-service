package domain

// Submit transitions an in-progress contract to PENDING_VALIDATION so a
// reviewer adjudicates the staged credits. Only a roster member may submit,
// at any confirmation level — partial fulfillment is administrable as long
// as at least one staged credit carries a media evidence reference.
func (c *MissionContract) Submit(identity CallerIdentity) error {
	if !c.CanSubmitEvidence(identity) {
		return ErrCallerNotAuthorized
	}
	if c.State.IsTerminal() {
		return ErrContractTerminal
	}
	if c.State != ContractInProgress {
		return ErrContractStateInvalid
	}
	if !c.hasStagedEvidence() {
		return ErrContractEvidenceRequired
	}
	c.State = ContractPendingValidation
	return nil
}

// hasStagedEvidence reports whether any pending credit carries a stored
// media reference — the minimum evidence bar for entering review.
func (c *MissionContract) hasStagedEvidence() bool {
	for _, credit := range c.pendingCredits {
		if credit.MediaURI != "" {
			return true
		}
	}
	return false
}

// Approve transitions a pending contract to COMPLETED and materializes every
// staged merit credit, plus the completion reward when all required
// objectives are fully confirmed. Role authorization is enforced by the
// application layer; the domain guards the state machine only.
func (c *MissionContract) Approve() ([]MeritCredit, error) {
	return c.approve(nil)
}

// ApproveWithVerdict behaves like Approve and additionally drops every staged
// unit listed in rejected: those credits never reach the ledger and the units
// are recorded as rejected for audit. A reference that was never staged fails
// the whole call — nothing mutates.
func (c *MissionContract) ApproveWithVerdict(rejected []UnitRef) ([]MeritCredit, error) {
	return c.approve(rejected)
}

func (c *MissionContract) approve(rejected []UnitRef) ([]MeritCredit, error) {
	if c == nil {
		return nil, ErrContractNil
	}
	if c.State.IsTerminal() {
		return nil, ErrContractTerminal
	}
	if c.State != ContractPendingValidation {
		return nil, ErrContractNotPendingValidation
	}
	if err := c.applyVerdict(rejected); err != nil {
		return nil, err
	}
	credits := make([]MeritCredit, 0, len(c.pendingCredits)+1)
	for _, credit := range c.pendingCredits {
		credits = append(credits, credit)
	}
	if c.Terms.AllMandatoryFulfilled(c.confirmedCountFor) && c.Offer.CompletionReward > 0 {
		completion, err := c.completionCredit()
		if err != nil {
			return nil, err
		}
		credits = append(credits, completion)
	}
	c.State = ContractCompleted
	return credits, nil
}

// applyVerdict resolves each rejected unit ref to its staged confirmation and
// moves it out of the confirmed/pending sets into the rejected record.
func (c *MissionContract) applyVerdict(rejected []UnitRef) error {
	roster := c.Roster()
	keys := make([]string, 0, len(rejected))
	seen := map[string]bool{}
	for _, ref := range rejected {
		confirmationID, err := NewConfirmationID(c.ContractID, ref.ObjectiveID, ref.UnitID, c.recipient(roster))
		if err != nil {
			return err
		}
		key := confirmationID.String()
		if seen[key] {
			continue
		}
		if _, staged := c.pendingCredits[key]; !staged {
			return ErrUnitNotStaged
		}
		seen[key] = true
		keys = append(keys, key)
	}
	for _, key := range keys {
		delete(c.pendingCredits, key)
		delete(c.confirmedUnits, key)
		c.rejectedUnits[key] = true
	}
	return nil
}

// Reject transitions a pending contract to the terminal DISQUALIFIED state and
// discards every staged credit — nothing enters the ledger.
func (c *MissionContract) Reject() error {
	if c == nil {
		return ErrContractNil
	}
	if c.State.IsTerminal() {
		return ErrContractTerminal
	}
	if c.State != ContractPendingValidation {
		return ErrContractNotPendingValidation
	}
	c.State = ContractDisqualified
	c.pendingCredits = map[string]MeritCredit{}
	return nil
}

// completionCredit builds the deterministic COMPLETION credit. The
// confirmation id is derived from the contract so a retried write under
// attribute_not_exists is a no-op rather than a second credit.
func (c *MissionContract) completionCredit() (MeritCredit, error) {
	roster := c.Roster()
	recipient := c.recipient(roster)
	confirmationID, err := NewConfirmationID(c.ContractID, CompletionObjectiveID, c.ContractID, recipient)
	if err != nil {
		return MeritCredit{}, err
	}
	return MeritCredit{
		SeasonID:       c.SeasonID,
		Category:       c.Offer.Category,
		Track:          c.track(),
		RecipientID:    recipient,
		Amount:         c.Offer.CompletionReward,
		ContractID:     c.ContractID,
		ObjectiveID:    CompletionObjectiveID,
		UnitID:         c.ContractID,
		ConfirmationID: confirmationID.String(),
	}, nil
}

// CompletionObjectiveID is the reserved objective identifier for the credit a
// contract emits when it reaches COMPLETED.
const CompletionObjectiveID = "COMPLETION"
