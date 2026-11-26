package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// DefinitionListReader lists mission definitions and versions.
type DefinitionListReader interface {
	List() ([]domain.MissionDefinition, error)
}

// ListDefinitionsHandler orchestrates the definition catalog listing use case.
type ListDefinitionsHandler struct {
	definitions DefinitionListReader
}

// NewListDefinitionsHandler creates a handler with the required reader port.
func NewListDefinitionsHandler(definitions DefinitionListReader) *ListDefinitionsHandler {
	return &ListDefinitionsHandler{definitions: definitions}
}

// Handle returns all definitions in the catalog projected into the
// requested language; each item reports the locale actually served.
func (h *ListDefinitionsHandler) Handle(language domain.Language) ([]LocalizedDefinition, error) {
	definitions, err := h.definitions.List()
	if err != nil {
		return nil, err
	}
	served := make([]LocalizedDefinition, 0, len(definitions))
	for _, definition := range definitions {
		served = append(served, LocalizeDefinition(definition, language))
	}
	return served, nil
}
