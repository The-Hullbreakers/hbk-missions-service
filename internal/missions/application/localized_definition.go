package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// LocalizedDefinition is the definition read projection in the language
// the caller selected: the definition fields keep their shape while
// ContentLanguage honestly reports the locale actually served.
type LocalizedDefinition struct {
	domain.MissionDefinition
	ContentLanguage domain.Language
}

// LocalizeDefinition projects a definition into the requested language via
// the shared atomic text resolution; the stored definition is never
// mutated.
func LocalizeDefinition(definition domain.MissionDefinition, language domain.Language) LocalizedDefinition {
	text := localizeMissionText(definition.Title, definition.Description, definition.Rules, definition.Objectives, definition.ContentEn, language)
	served := LocalizedDefinition{MissionDefinition: definition, ContentLanguage: text.Language}
	served.Title = text.Title
	served.Description = text.Description
	served.Rules = text.Rules
	served.Objectives = text.Objectives
	return served
}
