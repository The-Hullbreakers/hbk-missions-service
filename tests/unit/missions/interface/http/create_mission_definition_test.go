package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

type fakeDefinitionWriter struct{}

func (fakeDefinitionWriter) Save(definition domain.MissionDefinition) error { return nil }

func TestCreateMissionDefinitionHandler_CreatesDefinition(t *testing.T) {
	app := application.NewCreateMissionDefinitionHandler(fakeDefinitionWriter{})
	handler := missionhttp.CreateMissionDefinitionHandler(app)

	body := `{"tenant_id":"tenant-1","definition_id":"bounty-template","category":"BOUNTY_HUNTING","capacity":10,"mode":"SOLO"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/missions/definitions", bytes.NewReader([]byte(body)))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	var result map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if result["DefinitionID"] != "bounty-template" {
		t.Fatalf("unexpected response: %+v", result)
	}
}
