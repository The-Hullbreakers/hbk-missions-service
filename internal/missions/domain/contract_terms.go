package domain

import "encoding/json"

// ContractTerms is the frozen objective and reward snapshot a contract
// inherited from its definition version at acceptance. Confirmation and
// completion are evaluated against it, so later definition edits never
// change what a live contract pays or requires.
type ContractTerms struct {
	objectives []Objective
	phases     []Phase
}

// Phase groups the mandatory and optional objectives of one contract stage.
type Phase struct {
	Mandatory []Objective
	Optional  []Objective
}

// NewContractTerms validates and freezes the objective list.
func NewContractTerms(objectives []Objective) (ContractTerms, error) {
	if len(objectives) == 0 {
		return ContractTerms{}, ErrContractTermsInvalid
	}
	copied := make([]Objective, len(objectives))
	copy(copied, objectives)
	for _, o := range copied {
		if o.ObjectiveID == "" || o.Category == "" || o.MeritPerUnit < 0 || o.UnitLimit < 0 || o.Bonus < 0 {
			return ContractTerms{}, ErrContractTermsInvalid
		}
	}
	return ContractTerms{objectives: copied, phases: []Phase{{Mandatory: copied}}}, nil
}

// NewContractTermsPhased validates and freezes an ordered phase sequence;
// every objective must satisfy the same invariants as flat terms and belong
// to exactly one phase entry.
func NewContractTermsPhased(phases []Phase) (ContractTerms, error) {
	if len(phases) == 0 {
		return ContractTerms{}, ErrContractTermsInvalid
	}
	seen := map[string]bool{}
	all := []Objective{}
	copied := make([]Phase, len(phases))
	for i, phase := range phases {
		copied[i] = Phase{
			Mandatory: append([]Objective(nil), phase.Mandatory...),
			Optional:  append([]Objective(nil), phase.Optional...),
		}
		for _, o := range append(phase.Mandatory, phase.Optional...) {
			if o.ObjectiveID == "" || o.Category == "" || o.MeritPerUnit < 0 || o.UnitLimit < 0 || o.Bonus < 0 || seen[o.ObjectiveID] {
				return ContractTerms{}, ErrContractTermsInvalid
			}
			seen[o.ObjectiveID] = true
			all = append(all, o)
		}
	}
	if len(all) == 0 {
		return ContractTerms{}, ErrContractTermsInvalid
	}
	return ContractTerms{objectives: all, phases: copied}, nil
}

// Lookup resolves an objective from the frozen terms.
func (t ContractTerms) Lookup(objectiveID string) (Objective, bool) {
	for _, o := range t.objectives {
		if o.ObjectiveID == objectiveID {
			return o, true
		}
	}
	return Objective{}, false
}

// Objectives returns a defensive copy of the frozen objective list.
func (t ContractTerms) Objectives() []Objective {
	copied := make([]Objective, len(t.objectives))
	copy(copied, t.objectives)
	return copied
}

// IsZero reports whether the terms were never set (legacy contracts).
func (t ContractTerms) IsZero() bool {
	return len(t.objectives) == 0
}

// MarshalJSON exposes the frozen objective list in serialized contract views.
func (t ContractTerms) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Objectives())
}

// AllConfirmed reports whether every frozen objective has at least one
// confirmed unit in the given confirmation set.
func (t ContractTerms) AllConfirmed(confirmed func(objectiveID string) bool) bool {
	if t.IsZero() {
		return false
	}
	for _, o := range t.objectives {
		if !confirmed(o.ObjectiveID) {
			return false
		}
	}
	return true
}
