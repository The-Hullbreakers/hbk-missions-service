//go:build integration

package http_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// TEST-HTTP-LANG-003 / SCN-MB-CONTENT-INCOMPLETE: publishing a draft with
// missing or incomplete English rejects 422 and leaves draft + store
// untouched — no version row written.
func TestLangPublishDefinitionIncompleteEnglishRejected(t *testing.T) {
	// Arrange
	cases := []struct {
		name string
		en   *domain.LocalizedContent
	}{
		{"missing english", nil},
		{"incomplete rule", &domain.LocalizedContent{
			Title: "Rescue", Description: "Save the pilot",
			Objectives: []domain.LocalizedObjective{
				{ObjectiveID: "obj-1", Name: "Capture", Description: "No damage"},
			},
		}},
		{"incomplete objective", &domain.LocalizedContent{
			Title: "Rescue", Description: "Save the pilot",
			Rules: []string{"Capture the pilot"},
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			server, store, err := localizedPublishFixture(t, tc.en)
			require.NoError(t, err)
			request, err := publishRequest(server)
			require.NoError(t, err)

			// Act
			response, err := server.Client().Do(request)

			// Assert
			require.NoError(t, err)
			defer func() { _ = response.Body.Close() }()
			assert.Equal(t, http.StatusUnprocessableEntity, response.StatusCode)
			var body map[string]string
			require.NoError(t, json.NewDecoder(response.Body).Decode(&body))
			assert.Equal(t, "INVALID_MISSION_DEFINITION", body["code"])
			stored, err := store.Get("draft-publish")
			require.NoError(t, err)
			assert.Equal(t, domain.DefinitionStateDraft, stored.State)
			assert.Equal(t, "", stored.Version)
			_, getErr := store.GetVersion("draft-publish", "v1")
			assert.ErrorIs(t, getErr, domain.ErrDefinitionNotFound)
		})
	}
}

// TEST-HTTP-LANG-003 / SCN-MB-CONTENT-INCOMPLETE: a complete bilingual
// draft publishes — persisted version carries the frozen English snapshot
// while the response stays the Spanish shape.
func TestLangPublishDefinitionCompleteEnglishPublishes(t *testing.T) {
	// Arrange
	server, store, err := localizedPublishFixture(t, publishableEnglish())
	require.NoError(t, err)
	request, err := publishRequest(server)
	require.NoError(t, err)

	// Act
	response, err := server.Client().Do(request)

	// Assert
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	require.Equal(t, http.StatusOK, response.StatusCode)
	var raw map[string]json.RawMessage
	require.NoError(t, json.NewDecoder(response.Body).Decode(&raw))
	assert.NotContains(t, raw, "ContentEn")
	var title string
	require.NoError(t, json.Unmarshal(raw["Title"], &title))
	assert.Equal(t, "Rescate", title)
	var state string
	require.NoError(t, json.Unmarshal(raw["State"], &state))
	assert.Equal(t, "PUBLISHED", state)
	versioned, err := store.GetVersion("draft-publish", "v1")
	require.NoError(t, err)
	assert.Equal(t, domain.DefinitionStatePublished, versioned.State)
	assert.Equal(t, "v1", versioned.Version)
	assert.Equal(t, "Rescue", versioned.ContentEn.Title)
	assert.Equal(t, "Save the pilot", versioned.ContentEn.Description)
	assert.Equal(t, []string{"Capture the pilot"}, versioned.ContentEn.Rules)
	require.Len(t, versioned.ContentEn.Objectives, 1)
	assert.Equal(t, "obj-1", versioned.ContentEn.Objectives[0].ObjectiveID)
	assert.Equal(t, "Capture", versioned.ContentEn.Objectives[0].Name)
	assert.Equal(t, "No damage", versioned.ContentEn.Objectives[0].Description)
	assert.Equal(t, "Rescate", versioned.Title)
	assert.Equal(t, 50, versioned.CompletionReward)
}
