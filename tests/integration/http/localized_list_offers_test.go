//go:build integration

package http_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TEST-HTTP-LANG-001 (list-offers partition): malformed lang queries are
// rejected with INVALID_QUERY_PARAMETER before the board is read.
func TestLangListOffersRejectsInvalidQuery(t *testing.T) {
	// Arrange
	server, err := languageOfferServer(t)
	require.NoError(t, err)
	cases := []struct {
		name  string
		extra string
	}{
		{"unsupported", "&lang=fr"},
		{"empty", "&lang="},
		{"conflicting", "&lang=es&lang=en"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			request, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/missions/offers?season_id=season-1&status=all"+tc.extra, nil)
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

	t.Run("repeated es accepted", func(t *testing.T) {
		// Arrange
		request, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/missions/offers?season_id=season-1&status=all&lang=es&lang=es", nil)
		require.NoError(t, err)
		request.Header.Set("Authorization", "Bearer any")

		// Act
		response, err := server.Client().Do(request)

		// Assert
		require.NoError(t, err)
		defer func() { _ = response.Body.Close() }()
		require.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, "es", response.Header.Get("Content-Language"))
	})
}

// TEST-HTTP-LANG-003 (list-offers partition): a Spanish-only board serves
// Spanish under lang=en or an omitted lang — ContentEn is never exposed.
func TestLangListOffersServesLegacySpanish(t *testing.T) {
	// Arrange
	server, err := languageOfferServer(t)
	require.NoError(t, err)
	cases := []struct {
		name  string
		query string
	}{
		{"english falls back", "?season_id=season-1&status=all&lang=en"},
		{"omitted ignores accept language", "?season_id=season-1&status=all"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			request, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/missions/offers"+tc.query, nil)
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
			var page struct {
				Offers []map[string]json.RawMessage `json:"offers"`
				Total  int                          `json:"total"`
			}
			require.NoError(t, json.NewDecoder(response.Body).Decode(&page))
			require.Equal(t, 1, page.Total)
			require.Len(t, page.Offers, 1)
			assert.NotContains(t, page.Offers[0], "ContentEn")
			var offerID, title, servedLanguage string
			require.NoError(t, json.Unmarshal(page.Offers[0]["OfferID"], &offerID))
			require.NoError(t, json.Unmarshal(page.Offers[0]["Title"], &title))
			assert.Equal(t, "offer-legacy", offerID)
			assert.Equal(t, "Rescate", title)
			var summary struct {
				TotalMerits int `json:"TotalMerits"`
			}
			require.NoError(t, json.Unmarshal(page.Offers[0]["MeritSummary"], &summary))
			assert.Equal(t, 70, summary.TotalMerits)
			require.NoError(t, json.Unmarshal(page.Offers[0]["ContentLanguage"], &servedLanguage))
			assert.Equal(t, "es", servedLanguage)
		})
	}
}

// TEST-HTTP-LANG-003 (list-offers partition): an empty board stays empty
// and carries no Content-Language header.
func TestLangListOffersEmptyHasNoContentLanguage(t *testing.T) {
	// Arrange
	server, err := languageOfferServer(t)
	require.NoError(t, err)
	request, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/missions/offers?season_id=other-season&status=all&lang=en", nil)
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
	assert.Equal(t, 0, page.Total)
	assert.Empty(t, page.Offers)
}
