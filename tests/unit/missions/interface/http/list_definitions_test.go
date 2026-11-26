package http_test

import (
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

type fakeDefinitionLister struct {
	definitions []domain.MissionDefinition
	err         error
}

func (r *fakeDefinitionLister) List() ([]domain.MissionDefinition, error) {
	return r.definitions, r.err
}

func TestListDefinitionsHandler_ReturnsDefinitions(t *testing.T) {
	definition, err := domain.NewMissionDefinition("def-bounty", "Defensa de puesto", "Protege la posición", "COMBAT", domain.MissionModeSolo, 0, nil, nil)
	require.NoError(t, err)
	app := application.NewListDefinitionsHandler(&fakeDefinitionLister{definitions: []domain.MissionDefinition{definition}})
	handler := missionhttp.ListDefinitionsHandler(app)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/missions/definitions", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var result []map[string]any
	err = json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err)
	require.False(t, len(result) != 1 || result[0]["DefinitionCode"] != "def-bounty")
}

func TestListDefinitionsHandler_Returns500OnReaderError(t *testing.T) {
	app := application.NewListDefinitionsHandler(&fakeDefinitionLister{err: errors.New("store down")})
	handler := missionhttp.ListDefinitionsHandler(app)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/missions/definitions", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}
