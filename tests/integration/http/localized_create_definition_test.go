//go:build integration

package http_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// TEST-HTTP-LANG-003 / SCN-MB-CONTENT-DRAFT-EN: bilingual draft persists authored English; response stays Spanish shape.
func TestLangCreateDefinitionBilingualPersistsEnglish(t *testing.T) {
	// Arrange
	server, store, err := localizedCreateFixture(t, false)
	require.NoError(t, err)
	request, err := postDefinitionRequest(server, bilingualDraftBody, "Bearer any")
	require.NoError(t, err)

	// Act
	response, err := server.Client().Do(request)

	// Assert
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	require.Equal(t, http.StatusCreated, response.StatusCode)
	var raw map[string]json.RawMessage
	require.NoError(t, json.NewDecoder(response.Body).Decode(&raw))
	assert.NotContains(t, raw, "ContentEn")
	assert.NotContains(t, raw, "translations")
	var draft struct {
		DefinitionCode   string `json:"DefinitionCode"`
		Title            string `json:"Title"`
		State            string `json:"State"`
		CompletionReward int    `json:"CompletionReward"`
		Objectives       []struct {
			ObjectiveID  string `json:"ObjectiveID"`
			MeritPerUnit int    `json:"MeritPerUnit"`
		} `json:"Objectives"`
	}
	rawBytes, err := json.Marshal(raw)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(rawBytes, &draft))
	assert.Equal(t, "draft-bilingual", draft.DefinitionCode)
	assert.Equal(t, "Rescate", draft.Title)
	assert.Equal(t, "DRAFT", draft.State)
	assert.Equal(t, 50, draft.CompletionReward)
	require.Len(t, draft.Objectives, 1)
	assert.Equal(t, "obj-1", draft.Objectives[0].ObjectiveID)
	assert.Equal(t, 10, draft.Objectives[0].MeritPerUnit)
	stored, err := store.Get("draft-bilingual")
	require.NoError(t, err)
	assert.Equal(t, "Rescate", stored.Title)
	assert.Equal(t, 50, stored.CompletionReward)
	require.Len(t, stored.Objectives, 1)
	assert.Equal(t, "Capturar", stored.Objectives[0].Name)
	assert.Equal(t, "Sin daños", stored.Objectives[0].Description)
	assert.Equal(t, []string{"Captura al piloto"}, stored.Rules)
	assert.Equal(t, "Rescue", stored.ContentEn.Title)
	assert.Equal(t, "Save the pilot", stored.ContentEn.Description)
	assert.Equal(t, []string{"Capture the pilot"}, stored.ContentEn.Rules)
	require.Len(t, stored.ContentEn.Objectives, 1)
	assert.Equal(t, "obj-1", stored.ContentEn.Objectives[0].ObjectiveID)
	assert.Equal(t, "Capture", stored.ContentEn.Objectives[0].Name)
	assert.Equal(t, "No damage", stored.ContentEn.Objectives[0].Description)
	require.Len(t, stored.Objectives, 1)
	assert.Equal(t, 10, stored.Objectives[0].MeritPerUnit)
}

// TEST-HTTP-LANG-003: Spanish-only draft stays valid with zero English.
func TestLangCreateDefinitionSpanishOnlyValid(t *testing.T) {
	// Arrange
	server, store, err := localizedCreateFixture(t, false)
	require.NoError(t, err)
	request, err := postDefinitionRequest(server, spanishOnlyDraftBody, "Bearer any")
	require.NoError(t, err)

	// Act
	response, err := server.Client().Do(request)

	// Assert
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	require.Equal(t, http.StatusCreated, response.StatusCode)
	stored, err := store.Get("draft-spanish")
	require.NoError(t, err)
	assert.Equal(t, "Rescate", stored.Title)
	assert.Equal(t, domain.LocalizedContent{}, stored.ContentEn)
}

// TEST-HTTP-LANG-003: missing required Spanish fields reject the draft —
// translations cannot compensate — and nothing persists.
func TestLangCreateDefinitionMissingSpanishRejected(t *testing.T) {
	// Arrange
	server, store, err := localizedCreateFixture(t, false)
	require.NoError(t, err)
	request, err := postDefinitionRequest(server, missingTitleDraftBody, "Bearer any")
	require.NoError(t, err)

	// Act
	response, err := server.Client().Do(request)

	// Assert
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	assert.Equal(t, http.StatusUnprocessableEntity, response.StatusCode)
	_, getErr := store.Get("draft-invalid")
	assert.True(t, errors.Is(getErr, domain.ErrDefinitionNotFound))
}
