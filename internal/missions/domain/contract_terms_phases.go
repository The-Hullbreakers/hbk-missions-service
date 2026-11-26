package domain

// AllMandatoryConfirmed reports whether every mandatory objective across all
// phases has a confirmed unit — the completion bar for phased contracts;
// optional objectives never gate it.
func (t ContractTerms) AllMandatoryConfirmed(confirmed func(objectiveID string) bool) bool {
	if t.IsZero() {
		return false
	}
	for i := range t.phases {
		if !t.PhaseComplete(i, confirmed) {
			return false
		}
	}
	return true
}

// Phases returns a defensive copy of the ordered phase sequence.
func (t ContractTerms) Phases() []Phase {
	copied := make([]Phase, len(t.phases))
	for i, phase := range t.phases {
		copied[i] = Phase{
			Mandatory: append([]Objective(nil), phase.Mandatory...),
			Optional:  append([]Objective(nil), phase.Optional...),
		}
	}
	return copied
}

// PhaseCount returns the number of phases in the terms.
func (t ContractTerms) PhaseCount() int { return len(t.phases) }

// ObjectivePhase returns the index of the phase containing the objective.
func (t ContractTerms) ObjectivePhase(objectiveID string) int {
	for i, phase := range t.phases {
		for _, o := range append(phase.Mandatory, phase.Optional...) {
			if o.ObjectiveID == objectiveID {
				return i
			}
		}
	}
	return -1
}

// AllMandatoryFulfilled reports whether every mandatory objective across all
// phases reached its unit limit in confirmed (non-rejected) units — the bar
// for the completion reward. An unlimited objective (`UnitLimit` 0) is
// fulfilled with at least one confirmed unit.
func (t ContractTerms) AllMandatoryFulfilled(count func(objectiveID string) int) bool {
	if t.IsZero() {
		return false
	}
	for i := range t.phases {
		for _, o := range t.phases[i].Mandatory {
			required := o.UnitLimit
			if required == 0 {
				required = 1
			}
			if count(o.ObjectiveID) < required {
				return false
			}
		}
	}
	return true
}

// PhaseComplete reports whether every mandatory objective of the phase has a
// confirmed unit.
func (t ContractTerms) PhaseComplete(index int, confirmed func(objectiveID string) bool) bool {
	if index < 0 || index >= len(t.phases) {
		return false
	}
	for _, o := range t.phases[index].Mandatory {
		if !confirmed(o.ObjectiveID) {
			return false
		}
	}
	return true
}
