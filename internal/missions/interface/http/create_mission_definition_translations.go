package http

import (
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// createMissionDefinitionTranslations carries optional authored
// translations on draft creation; Spanish fields stay required.
type createMissionDefinitionTranslations struct {
	En *createMissionDefinitionEnglish `json:"en"`
}

type createMissionDefinitionEnglish struct {
	Title       string                                    `json:"title"`
	Description string                                    `json:"description"`
	Rules       []string                                  `json:"rules"`
	Objectives  []createMissionDefinitionEnglishObjective `json:"objectives"`
}

type createMissionDefinitionEnglishObjective struct {
	ObjectiveID string `json:"objective_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (t createMissionDefinitionTranslations) englishContent() *domain.LocalizedContent {
	if t.En == nil {
		return nil
	}
	content := &domain.LocalizedContent{Title: t.En.Title, Description: t.En.Description, Rules: t.En.Rules}
	for _, o := range t.En.Objectives {
		content.Objectives = append(content.Objectives, domain.LocalizedObjective{
			ObjectiveID: o.ObjectiveID, Name: o.Name, Description: o.Description,
		})
	}
	return content
}
