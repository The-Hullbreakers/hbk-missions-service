package composition_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	missioncomposition "github.com/hbk/hbk-missions-service/internal/missions/composition"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

type acceptOutcome struct {
	status int
	code   string
}

// raceAccepts fires one POST accept per offer in parallel and returns each
// call's outcome plus any transport or decode error — no assertions inside.
func raceAccepts(baseURL, auth string, offerIDs []string) ([]acceptOutcome, []error) {
	var wg sync.WaitGroup
	outcomes := make([]acceptOutcome, len(offerIDs))
	errs := make([]error, len(offerIDs))
	for i, offerID := range offerIDs {
		wg.Add(1)
		go func(id string, n int) {
			defer wg.Done()
			req, err := http.NewRequest(http.MethodPost, baseURL+"/api/v1/missions/offers/"+id+"/accept", bytes.NewReader([]byte(`{}`)))
			if err != nil {
				errs[n] = err
				return
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", auth)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				errs[n] = err
				return
			}
			defer resp.Body.Close()
			var body map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				errs[n] = fmt.Errorf("decode accept response: %w", err)
				return
			}
			code, _ := body["code"].(string)
			outcomes[n] = acceptOutcome{status: resp.StatusCode, code: code}
		}(offerID, i)
	}
	wg.Wait()
	return outcomes, errs
}

// REG-0053-02 through the public API: one player racing two published
// versions of the same definition_code — exactly one accept lands 201, the
// other is rejected 422 MISSION_DEFINITION_ALREADY_ACTIVE.
func TestConcurrentAcceptance_SameDefinitionSerializes(t *testing.T) {
	// Arrange
	client, table := chg0053Table(t, "e2e-serializes")

	seasonStore := missiondynamo.NewSeasonStore(client, table)
	require.NoError(t, seasonStore.Save(domain.NewSeason("season-1", domain.SeasonStateActive)))
	definitionStore := missiondynamo.NewDefinitionStore(client, table)
	for _, version := range []string{"1", "2"} {
		definition, err := domain.NewMissionDefinition(
			"def-x", "Misión X", "Descripción de la misión", "BOUNTY_HUNTING",
			domain.MissionModeSolo, 0,
			[]domain.Objective{domain.NewObjectiveNamed("obj-x", "Objetivo", "", "BOUNTY_HUNTING", 10, 0, 1)}, nil)
		require.NoError(t, err)
		definition = englishize(definition)
		require.NoError(t, definition.Publish(version))
		require.NoError(t, definitionStore.Save(definition))
	}
	offerStore := missiondynamo.NewOfferStore(client, table)
	for i, offerID := range []string{"offer-x-v1", "offer-x-v2"} {
		offer := domain.NewMissionOffer(offerID, "BOUNTY_HUNTING", "season-1")
		offer.DefinitionCode = "def-x"
		offer.DefinitionVersion = []string{"1", "2"}[i]
		require.NoError(t, offerStore.Save(offer))
	}

	server, err := missioncomposition.NewServer(missioncomposition.Config{
		TableName:   table,
		DynamoDB:    client,
		UserSession: tokenSession{},
	})
	require.NoError(t, err)
	ts := httptest.NewServer(server)
	defer ts.Close()

	// Act
	outcomes, actErrs := raceAccepts(ts.URL, token("player-1"), []string{"offer-x-v1", "offer-x-v2"})

	// Assert
	for _, err := range actErrs {
		require.NoError(t, err)
	}
	created, conflicts := 0, 0
	for _, r := range outcomes {
		switch {
		case r.status == http.StatusCreated:
			created++
		case r.status == http.StatusUnprocessableEntity && r.code == "MISSION_DEFINITION_ALREADY_ACTIVE":
			conflicts++
		}
	}
	require.Equal(t, 1, created)
	require.Equal(t, 1, conflicts)
}
