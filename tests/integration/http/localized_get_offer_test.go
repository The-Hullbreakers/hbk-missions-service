//go:build integration

package http_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TEST-BND-LANG-001 (get-offer partition): malformed lang queries are
// rejected before the offer lookup — a missing id would otherwise be 404.
func TestLangGetOfferRejectsInvalidQueryBeforeLookup(t *testing.T) {
	// Arrange
	server, err := languageOfferServer(t)
	require.NoError(t, err)
	cases := []struct {
		name  string
		query string
	}{
		{"unsupported", "?lang=fr"},
		{"empty", "?lang="},
		{"conflicting", "?lang=es&lang=en"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			request, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/missions/offers/missing"+tc.query, nil)
			require.NoError(t, err)
			request.Header.Set("Authorization", "Bearer any")

			// Act
			response, err := server.Client().Do(request)

			// Assert
			require.NoError(t, err)
			defer func() { _ = response.Body.Close() }()
			require.Equal(t, http.StatusBadRequest, response.StatusCode)
			assert.Equal(t, "", response.Header.Get("Content-Language"))
			var body map[string]string
			require.NoError(t, json.NewDecoder(response.Body).Decode(&body))
			assert.Equal(t, "INVALID_QUERY_PARAMETER", body["code"])
		})
	}
}

// TEST-BND-LANG-001 (get-offer partition): a Spanish-only legacy offer is
// served in Spanish under lang=en, an omitted lang, or a repeated same
// value — labeled honestly with Content-Language: es.
func TestLangGetOfferServesLegacySpanish(t *testing.T) {
	// Arrange
	server, err := languageOfferServer(t)
	require.NoError(t, err)
	cases := []struct {
		name  string
		query string
	}{
		{"english falls back", "?lang=en"},
		{"omitted ignores accept language", ""},
		{"repeated en accepted", "?lang=en&lang=en"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			request, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/missions/offers/offer-legacy"+tc.query, nil)
			require.NoError(t, err)
			request.Header.Set("Authorization", "Bearer any")
			request.Header.Set("Accept-Language", "en")

			// Act
			response, err := server.Client().Do(request)

			// Assert
			require.NoError(t, err)
			defer func() { _ = response.Body.Close() }()
			require.Equal(t, http.StatusOK, response.StatusCode)
			assert.Equal(t, "es", response.Header.Get("Content-Language"))
			var offer map[string]json.RawMessage
			require.NoError(t, json.NewDecoder(response.Body).Decode(&offer))
			assert.NotContains(t, offer, "ContentEn")
			var offerID, title, category string
			require.NoError(t, json.Unmarshal(offer["OfferID"], &offerID))
			require.NoError(t, json.Unmarshal(offer["Title"], &title))
			require.NoError(t, json.Unmarshal(offer["Category"], &category))
			assert.Equal(t, "offer-legacy", offerID)
			assert.Equal(t, "Rescate", title)
			assert.Equal(t, "BOUNTY_HUNTING", category)
			var summary struct {
				ObjectiveMerits  int `json:"ObjectiveMerits"`
				CompletionReward int `json:"CompletionReward"`
				TotalMerits      int `json:"TotalMerits"`
			}
			require.NoError(t, json.Unmarshal(offer["MeritSummary"], &summary))
			assert.Equal(t, 20, summary.ObjectiveMerits)
			assert.Equal(t, 50, summary.CompletionReward)
			assert.Equal(t, 70, summary.TotalMerits)
			var objectives []struct {
				ObjectiveID  string `json:"ObjectiveID"`
				MeritPerUnit int    `json:"MeritPerUnit"`
				UnitLimit    int    `json:"UnitLimit"`
			}
			require.NoError(t, json.Unmarshal(offer["Objectives"], &objectives))
			require.Len(t, objectives, 1)
			assert.Equal(t, "obj-1", objectives[0].ObjectiveID)
			assert.Equal(t, 10, objectives[0].MeritPerUnit)
			assert.Equal(t, 2, objectives[0].UnitLimit)
			var servedLanguage string
			require.NoError(t, json.Unmarshal(offer["ContentLanguage"], &servedLanguage))
			assert.Equal(t, "es", servedLanguage)
		})
	}
}

// TEST-BND-LANG-001 (get-offer partition): an unknown offer id keeps the
// existing 404 contract regardless of lang — no Content-Language.
func TestLangGetOfferUnknownRemains404(t *testing.T) {
	// Arrange
	server, err := languageOfferServer(t)
	require.NoError(t, err)
	request, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/missions/offers/missing?lang=en", nil)
	require.NoError(t, err)
	request.Header.Set("Authorization", "Bearer any")

	// Act
	response, err := server.Client().Do(request)

	// Assert
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	require.Equal(t, http.StatusNotFound, response.StatusCode)
	assert.Equal(t, "", response.Header.Get("Content-Language"))
	var body map[string]string
	require.NoError(t, json.NewDecoder(response.Body).Decode(&body))
	assert.Equal(t, "OFFER_NOT_AVAILABLE", body["code"])
}
