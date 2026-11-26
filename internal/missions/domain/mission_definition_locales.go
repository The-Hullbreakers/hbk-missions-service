package domain

import "strings"

// englishSnapshotComplete reports whether the definition's optional
// English snapshot structurally covers every Spanish player-facing
// field: a nonblank title and markdown description, one rule per
// Spanish rule in order with no blank entries, and exactly one
// objective translation per Spanish ObjectiveID with a nonblank name
// and a description whenever the Spanish objective carries one.
// A new draft version may publish only when this holds; historical
// Spanish-only records already published are never revalidated.
func (d MissionDefinition) englishSnapshotComplete() bool {
	en := d.ContentEn
	if strings.TrimSpace(en.Title) == "" || strings.TrimSpace(en.Description) == "" {
		return false
	}
	if len(en.Rules) != len(d.Rules) {
		return false
	}
	for _, rule := range en.Rules {
		if strings.TrimSpace(rule) == "" {
			return false
		}
	}
	if len(en.Objectives) != len(d.Objectives) {
		return false
	}
	byID := make(map[string]LocalizedObjective, len(en.Objectives))
	for _, translated := range en.Objectives {
		if _, duplicate := byID[translated.ObjectiveID]; duplicate {
			return false
		}
		byID[translated.ObjectiveID] = translated
	}
	for _, objective := range d.Objectives {
		translated, ok := byID[objective.ObjectiveID]
		if !ok || strings.TrimSpace(translated.Name) == "" {
			return false
		}
		if objective.Description != "" && strings.TrimSpace(translated.Description) == "" {
			return false
		}
	}
	return true
}
