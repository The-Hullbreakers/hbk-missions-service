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
	handler := missionhttp.NewAuthMiddleware(memberSession{}).Authenticate(missionhttp.CreateMissionDefinitionHandler(app))

	body := `{"definition_code":"bounty-template","title":"Bounty","description":"desc","category":"BOUNTY_HUNTING","mode":"SOLO","objectives":[{"objective_id":"obj-1","name":"obj-1","merit_per_unit":100,"unit_limit":1}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/missions/definitions", bytes.NewReader([]byte(body)))
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	var result map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if result["DefinitionCode"] != "bounty-template" {
		t.Fatalf("unexpected response: %+v", result)
	}
}
