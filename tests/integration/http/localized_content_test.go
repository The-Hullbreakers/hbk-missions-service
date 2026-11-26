//go:build integration

package http_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TEST-BND-LANG-001 (definitions partition): malformed or ambiguous lang
// query values are rejected with INVALID_QUERY_PARAMETER before catalog
// content is served; a repeated identical value is accepted.
func TestLangDefinitionsRejectsInvalidQuery(t *testing.T) {
	cases := []struct {
		name       string
		query      string
		wantStatus int
	}{
		{"unsupported language", "?lang=fr", http.StatusBadRequest},
		{"empty language", "?lang=", http.StatusBadRequest},
		{"conflicting es then en", "?lang=es&lang=en", http.StatusBadRequest},
		{"conflicting en then es", "?lang=en&lang=es", http.StatusBadRequest},
		{"repeated es accepted", "?lang=es&lang=es", http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			server, err := languageDefinitionsServer(t, true)
			require.NoError(t, err)
			request, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/missions/definitions"+tc.query, nil)
			require.NoError(t, err)
			request.Header.Set("Authorization", "Bearer any")

			// Act
			response, err := server.Client().Do(request)

			// Assert
			require.NoError(t, err)
			defer func() { _ = response.Body.Close() }()
			assert.Equal(t, tc.wantStatus, response.StatusCode)
			if tc.wantStatus == http.StatusBadRequest {
				assert.Equal(t, "", response.Header.Get("Content-Language"))
				var body struct {
					Code string `json:"code"`
				}
				require.NoError(t, json.NewDecoder(response.Body).Decode(&body))
				assert.Equal(t, "INVALID_QUERY_PARAMETER", body.Code)
			} else {
				var items []struct {
					ContentLanguage string `json:"ContentLanguage"`
				}
				require.NoError(t, json.NewDecoder(response.Body).Decode(&items))
				require.Len(t, items, 1)
				assert.Equal(t, "es", items[0].ContentLanguage)
			}
		})
	}
}

// TEST-BND-LANG-001 (definitions partition): a Spanish-only record served
// under lang=en falls back to Spanish and labels itself honestly with
// Content-Language: es, and an empty catalog stays empty.
func TestLangDefinitionsLegacyFallbackAndEmpty(t *testing.T) {
	cases := []struct {
		name  string
		query string
	}{
		{"spanish only fallback under en", "?lang=en"},
		{"omitted lang ignores accept language", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			server, err := languageDefinitionsServer(t, true)
			require.NoError(t, err)
			request, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/missions/definitions"+tc.query, nil)
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
			var items []map[string]json.RawMessage
			require.NoError(t, json.NewDecoder(response.Body).Decode(&items))
			require.Len(t, items, 1)
			assert.NotContains(t, items[0], "ContentEn")
			var title, servedLanguage string
			require.NoError(t, json.Unmarshal(items[0]["Title"], &title))
			require.NoError(t, json.Unmarshal(items[0]["ContentLanguage"], &servedLanguage))
			assert.Equal(t, "Rescate", title)
			assert.Equal(t, "es", servedLanguage)
		})
	}

	t.Run("empty catalog under en", func(t *testing.T) {
		// Arrange
		server, err := languageDefinitionsServer(t, false)
		require.NoError(t, err)
		request, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/missions/definitions?lang=en", nil)
		require.NoError(t, err)
		request.Header.Set("Authorization", "Bearer any")

		// Act
		response, err := server.Client().Do(request)

		// Assert
		require.NoError(t, err)
		defer func() { _ = response.Body.Close() }()
		require.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, "", response.Header.Get("Content-Language"))
		var items []struct {
			ContentLanguage string `json:"ContentLanguage"`
		}
		require.NoError(t, json.NewDecoder(response.Body).Decode(&items))
		assert.Empty(t, items)
	})
}
