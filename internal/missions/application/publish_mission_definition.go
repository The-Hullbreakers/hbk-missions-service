package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// DefinitionReader loads a mission definition by identity.
type DefinitionReader interface {
	Get(tenantID, definitionID string) (domain.MissionDefinition, error)
	GetVersion(tenantID, definitionID, version string) (domain.MissionDefinition, error)
}

// PublishMissionDefinitionHandler publishes a draft definition version.
type PublishMissionDefinitionHandler struct {
	definitions DefinitionReader
	writer      DefinitionWriter
}

// NewPublishMissionDefinitionHandler creates a handler with the required ports.
func NewPublishMissionDefinitionHandler(definitions DefinitionReader, writer DefinitionWriter) *PublishMissionDefinitionHandler {
	return &PublishMissionDefinitionHandler{definitions: definitions, writer: writer}
}

// Handle loads a draft definition, publishes it with the given version, and persists it.
func (h *PublishMissionDefinitionHandler) Handle(tenantID, definitionID, version string) (domain.MissionDefinition, error) {
	definition, err := h.definitions.Get(tenantID, definitionID)
	if err != nil {
		return domain.MissionDefinition{}, err
	}
	if err := definition.Publish(version); err != nil {
		return domain.MissionDefinition{}, err
	}
	if err := h.writer.Save(definition); err != nil {
		return domain.MissionDefinition{}, err
	}
	return definition, nil
}
