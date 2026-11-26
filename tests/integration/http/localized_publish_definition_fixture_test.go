//go:build integration

package http_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// localizedPublishFixture boots the real server seeded with one Spanish
// draft and captures its store for state observation.
func localizedPublishFixture(t *testing.T, en *domain.LocalizedContent) (*httptest.Server, *missiondynamo.DefinitionStore, error) {
	var store *missiondynamo.DefinitionStore
	seed := func(client *missiondynamo.DynamoDBClient, table string) error {
		store = missiondynamo.NewDefinitionStore(client, table)
		draft, err := domain.NewMissionDefinition("draft-publish", "Rescate", "Salvar al piloto",
			"BOUNTY_HUNTING", domain.MissionModeSolo, 50,
			[]domain.Objective{
				domain.NewObjectiveNamed("obj-1", "Capturar", "Sin danos", "BOUNTY_HUNTING", 10, 0, 2),
			}, []string{"Captura al piloto"})
		if err != nil {
			return err
		}
		if en != nil {
			draft.ContentEn = en.Clone()
		}
		return store.Save(draft)
	}
	server, err := languageHTTPServer(t, seed)
	return server, store, err
}

// publishableEnglish is a complete English snapshot that passes domain
// publication completeness validation.
func publishableEnglish() *domain.LocalizedContent {
	return &domain.LocalizedContent{
		Title: "Rescue", Description: "Save the pilot",
		Rules: []string{"Capture the pilot"},
		Objectives: []domain.LocalizedObjective{
			{ObjectiveID: "obj-1", Name: "Capture", Description: "No damage"},
		},
	}
}

func publishRequest(server *httptest.Server) (*http.Request, error) {
	request, err := http.NewRequest(http.MethodPost,
		server.URL+"/api/v1/missions/definitions/draft-publish/publish",
		strings.NewReader(`{"version":"v1"}`))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer any")
	return request, nil
}
