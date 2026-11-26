package application_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

type fakeDefinitionListReader struct {
	definitions []domain.MissionDefinition
	err         error
}

func (r *fakeDefinitionListReader) List() ([]domain.MissionDefinition, error) {
	return r.definitions, r.err
}

func TestListDefinitionsHandler_ReturnsAllDefinitions(t *testing.T) {
	draft, err := domain.NewMissionDefinition("def-bounty", "Defensa de puesto", "Protege la posición", "COMBAT", domain.MissionModeSolo, 0, nil, nil)
	require.NoError(t, err)
	reader := &fakeDefinitionListReader{definitions: []domain.MissionDefinition{draft}}
	handler := application.NewListDefinitionsHandler(reader)

	got, err := handler.Handle(domain.LanguageSpanish)
	require.NoError(t, err)
	require.False(t, len(got) != 1 || got[0].DefinitionCode != "def-bounty")
}

func TestListDefinitionsHandler_PropagatesReaderError(t *testing.T) {
	reader := &fakeDefinitionListReader{err: errors.New("store down")}
	handler := application.NewListDefinitionsHandler(reader)

	_, err := handler.Handle(domain.LanguageSpanish)
	require.False(t, err == nil)
}
