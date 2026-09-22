package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// DefinitionWriter persists a mission definition.
type DefinitionWriter interface {
	Save(definition domain.MissionDefinition) error
}

// CreateMissionDefinitionHandler creates a new mission definition in DRAFT.
type CreateMissionDefinitionHandler struct {
	definitions DefinitionWriter
}

// NewCreateMissionDefinitionHandler creates a handler with the required ports.
func NewCreateMissionDefinitionHandler(definitions DefinitionWriter) *CreateMissionDefinitionHandler {
	return &CreateMissionDefinitionHandler{definitions: definitions}
}

// Handle validates and persists a new draft definition.
func (h *CreateMissionDefinitionHandler) Handle(tenantID, definitionID, category string, capacity int, mode domain.MissionMode) (domain.MissionDefinition, error) {
	_ = tenantID
	definition, err := domain.NewMissionDefinition(definitionID, category, capacity, mode)
	if err != nil {
		return domain.MissionDefinition{}, err
	}
	if err := h.definitions.Save(definition); err != nil {
		return domain.MissionDefinition{}, err
	}
	return definition, nil
}
