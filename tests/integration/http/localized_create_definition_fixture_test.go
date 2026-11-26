//go:build integration

package http_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

const bilingualDraftBody = `{"definition_code":"draft-bilingual","title":"Rescate","description":"Salvar al piloto","category":"BOUNTY_HUNTING","mode":"SOLO","completion_reward":50,"rules":["Captura al piloto"],"objectives":[{"objective_id":"obj-1","name":"Capturar","description":"Sin daños","category":"BOUNTY_HUNTING","merit_per_unit":10,"unit_limit":2}],"translations":{"en":{"title":"Rescue","description":"Save the pilot","rules":["Capture the pilot"],"objectives":[{"objective_id":"obj-1","name":"Capture","description":"No damage"}]}}}`

const spanishOnlyDraftBody = `{"definition_code":"draft-spanish","title":"Rescate","description":"Salvar al piloto","category":"BOUNTY_HUNTING","mode":"SOLO","completion_reward":50,"rules":["Captura al piloto"],"objectives":[{"objective_id":"obj-1","name":"Capturar","description":"Sin daños","category":"BOUNTY_HUNTING","merit_per_unit":10,"unit_limit":2}]}`

const missingTitleDraftBody = `{"definition_code":"draft-invalid","description":"Salvar al piloto","category":"BOUNTY_HUNTING","mode":"SOLO","completion_reward":50,"translations":{"en":{"title":"Rescue"}}}`

// localizedCreateFixture boots the real server and captures the definition
// store for post-POST persistence observation.
func localizedCreateFixture(t *testing.T, member bool) (*httptest.Server, *missiondynamo.DefinitionStore, error) {
	var store *missiondynamo.DefinitionStore
	seed := func(client *missiondynamo.DynamoDBClient, table string) error {
		store = missiondynamo.NewDefinitionStore(client, table)
		return nil
	}
	var server *httptest.Server
	var err error
	if member {
		server, err = languageSessionServer(t, definitionsSession{roles: []interface{}{"member"}}, seed)
	} else {
		server, err = languageHTTPServer(t, seed)
	}
	return server, store, err
}

func postDefinitionRequest(server *httptest.Server, body, auth string) (*http.Request, error) {
	request, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/missions/definitions", strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	if auth != "" {
		request.Header.Set("Authorization", auth)
	}
	return request, nil
}
