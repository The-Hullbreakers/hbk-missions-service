package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// DefinitionWriter persists a mission definition.
type DefinitionWriter interface {
	Save(definition domain.MissionDefinition) error
}

// CreateMissionDefinitionRequest carries every field needed to author a
// draft mission definition. ContentEn is optional: nil creates a
// Spanish-only draft, non-nil freezes a deep copy of the English snapshot.
type CreateMissionDefinitionRequest struct {
	DefinitionCode   string
	Title            string
	Description      string
	Category         string
	Mode             domain.MissionMode
	CompletionReward int
	Objectives       []domain.Objective
	Rules            []string
	ContentEn        *domain.LocalizedContent
}

// CreateMissionDefinitionHandler creates a new mission definition in DRAFT.
type CreateMissionDefinitionHandler struct {
	definitions DefinitionWriter
}

// NewCreateMissionDefinitionHandler creates a handler with the required ports.
func NewCreateMissionDefinitionHandler(definitions DefinitionWriter) *CreateMissionDefinitionHandler {
	return &CreateMissionDefinitionHandler{definitions: definitions}
}

// Handle validates and persists a new draft definition with its player-facing
// content, completion reward, named objective rules, readable rules, and an
// optional frozen English snapshot.
func (h *CreateMissionDefinitionHandler) Handle(request CreateMissionDefinitionRequest) (domain.MissionDefinition, error) {
	definition, err := domain.NewMissionDefinition(request.DefinitionCode, request.Title, request.Description, request.Category, request.Mode, request.CompletionReward, request.Objectives, request.Rules)
	if err != nil {
		return domain.MissionDefinition{}, err
	}
	if request.ContentEn != nil {
		definition.ContentEn = request.ContentEn.Clone()
	}
	if err := h.definitions.Save(definition); err != nil {
		return domain.MissionDefinition{}, err
	}
	return definition, nil
}
