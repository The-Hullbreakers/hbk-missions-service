//go:build integration

package http_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TEST-HTTP-LANG-001 (list-contracts partition): malformed lang queries
// are rejected before the player roster is read.
func TestLangListContractsRejectsInvalidQuery(t *testing.T) {
	// Arrange
	server, err := languageHTTPServer(t, nil)
	require.NoError(t, err)
	cases := []struct {
		name  string
		extra string
	}{
		{"unsupported", "&lang=fr"},
		{"empty", "&lang="},
		{"conflicting es then en", "&lang=es&lang=en"},
		{"conflicting en then es", "&lang=en&lang=es"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			request, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/missions/contracts?season_id=season-1&status=all"+tc.extra, nil)
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
		request, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/missions/contracts?season_id=season-1&status=all&lang=es&lang=es", nil)
		require.NoError(t, err)
		request.Header.Set("Authorization", "Bearer any")

		// Act
		response, err := server.Client().Do(request)

		// Assert
		require.NoError(t, err)
		defer func() { _ = response.Body.Close() }()
		require.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, "", response.Header.Get("Content-Language"))
	})
}

// TEST-HTTP-LANG-001 (list-contracts partition): an empty roster stays
// empty with no Content-Language header.
func TestLangListContractsEmpty(t *testing.T) {
	// Arrange
	server, err := languageHTTPServer(t, nil)
	require.NoError(t, err)
	request, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/missions/contracts?season_id=season-1&status=all&lang=en", nil)
	require.NoError(t, err)
	request.Header.Set("Authorization", "Bearer any")

	// Act
	response, err := server.Client().Do(request)

	// Assert
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	require.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "", response.Header.Get("Content-Language"))
	var items []json.RawMessage
	require.NoError(t, json.NewDecoder(response.Body).Decode(&items))
	assert.Empty(t, items)
}

// TEST-HTTP-LANG-001 (list-contracts partition): an unauthenticated caller
// is rejected by the middleware before any lang handling.
func TestLangListContractsUnauthorized(t *testing.T) {
	// Arrange
	server, err := languageHTTPServer(t, nil)
	require.NoError(t, err)
	request, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/missions/contracts?season_id=season-1&status=all&lang=fr", nil)
	require.NoError(t, err)

	// Act
	response, err := server.Client().Do(request)

	// Assert
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
	assert.Equal(t, "", response.Header.Get("Content-Language"))
	var body map[string]string
	require.NoError(t, json.NewDecoder(response.Body).Decode(&body))
	assert.Equal(t, "UNAUTHORIZED", body["code"])
}
