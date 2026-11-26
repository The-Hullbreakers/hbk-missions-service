package application

import (
	"strings"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// localizedMissionText is the resolved player-facing content of one mission
// record: the effective language's title, description, ordered rules and
// objectives — always copies, never aliasing the stored record.
type localizedMissionText struct {
	Title       string
	Description string
	Rules       []string
	Objectives  []domain.Objective
	Language    domain.Language
}

// localizeMissionText resolves the served locale atomically. English is
// served only when the frozen translation is complete — title, description,
// positional rules and every objective's name and description — and is then
// substituted as a whole in base objective order; anything incomplete falls
// back to the Spanish scalars. Objective translation descriptions are
// required only when the Spanish objective carries one.
func localizeMissionText(title, description string, rules []string, objectives []domain.Objective, en domain.LocalizedContent, language domain.Language) localizedMissionText {
	served := localizedMissionText{
		Title:       title,
		Description: description,
		Rules:       append([]string(nil), rules...),
		Objectives:  append([]domain.Objective(nil), objectives...),
		Language:    domain.LanguageSpanish,
	}
	if language != domain.LanguageEnglish || !englishTextComplete(en, rules, objectives) {
		return served
	}
	served.Title = en.Title
	served.Description = en.Description
	served.Rules = append([]string(nil), en.Rules...)
	translations := objectivesByID(en.Objectives)
	for i, objective := range served.Objectives {
		translation := translations[objective.ObjectiveID]
		served.Objectives[i].Name = translation.Name
		served.Objectives[i].Description = translation.Description
	}
	served.Language = domain.LanguageEnglish
	return served
}

// englishTextComplete reports whether the English snapshot covers every
// player-facing field. Rules stay positional; objective translations are
// matched by stable ObjectiveID and must be unique.
func englishTextComplete(en domain.LocalizedContent, rules []string, objectives []domain.Objective) bool {
	if strings.TrimSpace(en.Title) == "" || strings.TrimSpace(en.Description) == "" {
		return false
	}
	if len(en.Rules) != len(rules) {
		return false
	}
	for _, rule := range en.Rules {
		if strings.TrimSpace(rule) == "" {
			return false
		}
	}
	return englishObjectivesComplete(en.Objectives, objectives)
}

// englishObjectivesComplete requires one uniquely-IDed, non-blank English
// name per base objective; the description is required only when the
// Spanish objective carries one.
func englishObjectivesComplete(translations []domain.LocalizedObjective, objectives []domain.Objective) bool {
	if len(translations) != len(objectives) {
		return false
	}
	byID := objectivesByID(translations)
	if len(byID) != len(translations) {
		return false
	}
	for _, objective := range objectives {
		translation, ok := byID[objective.ObjectiveID]
		if !ok || strings.TrimSpace(translation.Name) == "" {
			return false
		}
		if objective.Description != "" && strings.TrimSpace(translation.Description) == "" {
			return false
		}
	}
	return true
}

// objectivesByID indexes objective translations by their stable ID;
// duplicates collapse so a size mismatch exposes them.
func objectivesByID(translations []domain.LocalizedObjective) map[string]domain.LocalizedObjective {
	byID := make(map[string]domain.LocalizedObjective, len(translations))
	for _, translation := range translations {
		byID[translation.ObjectiveID] = translation
	}
	return byID
}
