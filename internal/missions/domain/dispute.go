package domain

// DisputeState is the lifecycle of a raised dispute.
type DisputeState string

const (
	// DisputeOpen means the dispute is awaiting reviewer resolution and blocks
	// member exit (leave/abandon) on the contract.
	DisputeOpen DisputeState = "OPEN"
	// DisputeResolved means the dispute was adjudicated; the outcome is
	// immutable and no longer blocks exit.
	DisputeResolved DisputeState = "RESOLVED"
)

// Dispute records a member's objection to a specific rejected credit —
// identified by the frozen objective the credit targeted.
type Dispute struct {
	DisputeID   string       `json:"dispute_id"`
	ContractID  string       `json:"contract_id"`
	ObjectiveID string       `json:"objective_id"`
	RaisedBy    string       `json:"raised_by"`
	Reason      string       `json:"reason"`
	State       DisputeState `json:"state"`
	ResolvedBy  string       `json:"resolved_by,omitempty"`
}

// RaiseDispute opens a dispute on the contract for the given objective. Only
// active roster members may raise one; the objective must exist in the
// contract's frozen terms.
func (c *MissionContract) RaiseDispute(identity CallerIdentity, objectiveID, reason string) (Dispute, error) {
	if !c.CanSubmitEvidence(identity) {
		return Dispute{}, ErrCallerNotAuthorized
	}
	if _, ok := c.Terms.Lookup(objectiveID); !ok {
		return Dispute{}, ErrObjectiveNotInContract
	}
	if c.disputes == nil {
		c.disputes = map[string]Dispute{}
	}
	dispute := Dispute{
		DisputeID:   c.ContractID + "~DISPUTE~" + objectiveID + "~" + identity.PlayerID(),
		ContractID:  c.ContractID,
		ObjectiveID: objectiveID,
		RaisedBy:    identity.PlayerID(),
		Reason:      reason,
		State:       DisputeOpen,
	}
	c.disputes[dispute.DisputeID] = dispute
	return dispute, nil
}

// ResolveDispute marks the dispute RESOLVED and records the moderator. A
// resolved dispute is immutable — re-resolving is rejected.
func (c *MissionContract) ResolveDispute(disputeID, resolvedBy string) error {
	dispute, ok := c.disputes[disputeID]
	if !ok {
		return ErrDisputeNotFound
	}
	if dispute.State != DisputeOpen {
		return ErrDisputeStateInvalid
	}
	dispute.State = DisputeResolved
	dispute.ResolvedBy = resolvedBy
	c.disputes[disputeID] = dispute
	return nil
}

// Disputes returns a defensive copy of the contract's dispute collection.
func (c *MissionContract) Disputes() []Dispute {
	out := make([]Dispute, 0, len(c.disputes))
	for _, d := range c.disputes {
		out = append(out, d)
	}
	return out
}

// RestoreDisputes rehydrates persisted disputes on load.
func (c *MissionContract) RestoreDisputes(disputes []Dispute) {
	c.disputes = make(map[string]Dispute, len(disputes))
	for _, d := range disputes {
		c.disputes[d.DisputeID] = d
	}
}

// hasOpenDispute reports whether any dispute still blocks member exit.
func (c *MissionContract) hasOpenDispute() bool {
	for _, d := range c.disputes {
		if d.State == DisputeOpen {
			return true
		}
	}
	return false
}
