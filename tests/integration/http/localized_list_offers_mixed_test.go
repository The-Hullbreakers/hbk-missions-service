//go:build integration

package http_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TEST-HTTP-LANG-003 (list-offers partition): a mixed board under lang=en
// serves each offer in its own language — bilingual English, legacy
// Spanish — and the heterogeneous page carries no Content-Language.
// Currently red for TWO distinct causes: the boundary drops ContentLanguage
// (MarshalJSON) AND persistence drops ContentEn (T-0054-007 dependency).
func TestLangListOffersMixedActualLanguages(t *testing.T) {
	// Arrange
	server, err := languageMixedOffersServer(t)
	require.NoError(t, err)
	request, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/missions/offers?season_id=season-1&status=all&lang=en", nil)
	require.NoError(t, err)
	request.Header.Set("Authorization", "Bearer any")

	// Act
	response, err := server.Client().Do(request)

	// Assert
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	require.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "", response.Header.Get("Content-Language"))
	var page struct {
		Offers []map[string]json.RawMessage `json:"offers"`
		Total  int                          `json:"total"`
	}
	require.NoError(t, json.NewDecoder(response.Body).Decode(&page))
	require.Equal(t, 2, page.Total)
	require.Len(t, page.Offers, 2)
	assert.NotContains(t, page.Offers[0], "ContentEn")
	assert.NotContains(t, page.Offers[1], "ContentEn")
	var firstID, firstTitle, firstLanguage string
	require.NoError(t, json.Unmarshal(page.Offers[0]["OfferID"], &firstID))
	require.NoError(t, json.Unmarshal(page.Offers[0]["Title"], &firstTitle))
	assert.Equal(t, "offer-bilingual", firstID)
	assert.Equal(t, "Hunt", firstTitle)
	require.NoError(t, json.Unmarshal(page.Offers[0]["ContentLanguage"], &firstLanguage))
	assert.Equal(t, "en", firstLanguage)
	var secondID, secondTitle, secondLanguage string
	require.NoError(t, json.Unmarshal(page.Offers[1]["OfferID"], &secondID))
	require.NoError(t, json.Unmarshal(page.Offers[1]["Title"], &secondTitle))
	assert.Equal(t, "offer-legacy", secondID)
	assert.Equal(t, "Rescate", secondTitle)
	require.NoError(t, json.Unmarshal(page.Offers[1]["ContentLanguage"], &secondLanguage))
	assert.Equal(t, "es", secondLanguage)
}
