package domain

import "strings"

// WithObjectiveText returns an independent terms snapshot whose objective
// Name/Description carry the given translation, matched by stable
// ObjectiveID across both the flat objective list and every phase's
// mandatory and optional slots. Ids, categories, scoring, phase membership
// and ordering never change, and neither the receiver nor the input
// translations are mutated.
func (t ContractTerms) WithObjectiveText(translations []LocalizedObjective) (ContractTerms, error) {
	if t.IsZero() {
		return ContractTerms{}, ErrContractTermsInvalid
	}
	byID, err := objectiveTranslationsByID(translations, len(t.objectives))
	if err != nil {
		return ContractTerms{}, err
	}
	for _, objective := range t.objectives {
		translation, ok := byID[objective.ObjectiveID]
		if !ok || strings.TrimSpace(translation.Name) == "" {
			return ContractTerms{}, ErrContractTermsInvalid
		}
		if objective.Description != "" && strings.TrimSpace(translation.Description) == "" {
			return ContractTerms{}, ErrContractTermsInvalid
		}
	}
	translated := ContractTerms{
		objectives: make([]Objective, len(t.objectives)),
		phases:     make([]Phase, len(t.phases)),
	}
	for i, objective := range t.objectives {
		translated.objectives[i] = translateObjective(objective, byID)
	}
	for i, phase := range t.phases {
		translated.phases[i] = Phase{
			Mandatory: translateObjectives(phase.Mandatory, byID),
			Optional:  translateObjectives(phase.Optional, byID),
		}
	}
	return translated, nil
}

// objectiveTranslationsByID indexes translations by ObjectiveID, requiring
// exactly one entry per stored objective: count mismatch, an unknown id or
// a duplicate invalidates the snapshot.
func objectiveTranslationsByID(translations []LocalizedObjective, expected int) (map[string]LocalizedObjective, error) {
	if len(translations) != expected {
		return nil, ErrContractTermsInvalid
	}
	byID := make(map[string]LocalizedObjective, len(translations))
	for _, translation := range translations {
		if _, exists := byID[translation.ObjectiveID]; exists {
			return nil, ErrContractTermsInvalid
		}
		byID[translation.ObjectiveID] = translation
	}
	return byID, nil
}

// translateObjectives returns a fresh slice with each objective's text
// replaced by its ID-matched translation; economics stay untouched.
func translateObjectives(objectives []Objective, byID map[string]LocalizedObjective) []Objective {
	out := make([]Objective, len(objectives))
	for i, objective := range objectives {
		out[i] = translateObjective(objective, byID)
	}
	return out
}

func translateObjective(objective Objective, byID map[string]LocalizedObjective) Objective {
	translation := byID[objective.ObjectiveID]
	objective.Name = translation.Name
	objective.Description = translation.Description
	return objective
}
