package dynamodb

import (
	"strings"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func validateReviewedTranslation(reviewed ReviewedMissionTranslation) error {
	es, en := reviewed.SourceSpanish, reviewed.English
	if strings.TrimSpace(reviewed.DefinitionCode) == "" ||
		strings.TrimSpace(reviewed.DefinitionVersion) == "" ||
		strings.TrimSpace(es.Title) == "" || strings.TrimSpace(es.Description) == "" ||
		strings.TrimSpace(en.Title) == "" || strings.TrimSpace(en.Description) == "" ||
		len(es.Rules) != len(en.Rules) || len(es.Objectives) != len(en.Objectives) {
		return domain.ErrDefinitionInvalid
	}
	for i := range es.Rules {
		if strings.TrimSpace(es.Rules[i]) == "" || strings.TrimSpace(en.Rules[i]) == "" {
			return domain.ErrDefinitionInvalid
		}
	}
	ids := make(map[string]domain.LocalizedObjective, len(en.Objectives))
	for _, objective := range en.Objectives {
		if objective.ObjectiveID == "" || strings.TrimSpace(objective.Name) == "" {
			return domain.ErrDefinitionInvalid
		}
		if _, exists := ids[objective.ObjectiveID]; exists {
			return domain.ErrDefinitionInvalid
		}
		ids[objective.ObjectiveID] = objective
	}
	seen := make(map[string]bool, len(es.Objectives))
	for _, objective := range es.Objectives {
		translated, ok := ids[objective.ObjectiveID]
		if objective.ObjectiveID == "" || seen[objective.ObjectiveID] ||
			strings.TrimSpace(objective.Name) == "" || !ok ||
			(objective.Description != "" && strings.TrimSpace(translated.Description) == "") {
			return domain.ErrDefinitionInvalid
		}
		seen[objective.ObjectiveID] = true
	}
	return nil
}
