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

type fakeDefinitionStore struct{}

func (fakeDefinitionStore) Get(_, _ string) (domain.MissionDefinition, error) {
	def, _ := domain.NewMissionDefinition("bounty-template", "BOUNTY_HUNTING", 10, domain.MissionModeSolo)
	return def, nil
}

func (fakeDefinitionStore) GetVersion(_, _, _ string) (domain.MissionDefinition, error) {
	def, _ := domain.NewMissionDefinition("bounty-template", "BOUNTY_HUNTING", 10, domain.MissionModeSolo)
	_ = def.Publish("v1")
	return def, nil
}

func (fakeDefinitionStore) Save(definition domain.MissionDefinition) error { return nil }

func TestPublishMissionDefinitionHandler_PublishesDefinition(t *testing.T) {
	app := application.NewPublishMissionDefinitionHandler(fakeDefinitionStore{}, fakeDefinitionStore{})
	handler := missionhttp.PublishMissionDefinitionHandler(app)

	body := `{"tenant_id":"tenant-1","version":"v1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/missions/definitions/bounty-template/publish", bytes.NewReader([]byte(body)))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if result["Version"] != "v1" {
		t.Fatalf("unexpected response: %+v", result)
	}
}
