//go:build integration

package http_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TEST-BND-LANG-002 (review-queue partition): a homogeneous bilingual
// pending queue serves English titles with per-item labels + header.
func TestLangReviewQueueHomogeneousEnglish(t *testing.T) {
	// Arrange
	server, err := languageReviewQueueServer(t, true, false)
	require.NoError(t, err)
	request, err := reviewQueueRequest(server.URL, "?lang=en", "Bearer any")
	require.NoError(t, err)

	// Act
	response, err := server.Client().Do(request)

	// Assert
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	require.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "en", response.Header.Get("Content-Language"))
	var raw []map[string]json.RawMessage
	require.NoError(t, json.NewDecoder(response.Body).Decode(&raw))
	require.Len(t, raw, 1)
	assert.NotContains(t, raw[0], "ContentEn")
	assert.NotContains(t, raw[0], "content_en")
	var entries []localizedReviewQueueEntry
	bytes, err := json.Marshal(raw)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(bytes, &entries))
	assert.Equal(t, "contract-rescue", entries[0].ContractID)
	assert.Equal(t, "Rescue", entries[0].Title)
	assert.Equal(t, "en", entries[0].ContentLanguage)
	assert.Equal(t, "PENDING_VALIDATION", entries[0].State)
	assert.Equal(t, []string{"player-1"}, entries[0].Roster)
	assert.Equal(t, 0, entries[0].ConfirmedCount)
	assert.Equal(t, 0, entries[0].PendingCredits)
	assert.Equal(t, "offer-rescue", entries[0].OfferID)
}

// TEST-BND-LANG-002 (review-queue partition): omitted lang and explicit es
// both serve the Spanish baseline honestly on a bilingual item.
func TestLangReviewQueueSpanishDefaultAndExplicit(t *testing.T) {
	// Arrange
	server, err := languageReviewQueueServer(t, true, false)
	require.NoError(t, err)
	cases := []struct {
		name   string
		query  string
		accept string
	}{
		{"omitted ignores Accept-Language", "", "en"},
		{"explicit es", "?lang=es", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			request, err := reviewQueueRequest(server.URL, tc.query, "Bearer any")
			require.NoError(t, err)
			if tc.accept != "" {
				request.Header.Set("Accept-Language", tc.accept)
			}

			// Act
			response, err := server.Client().Do(request)

			// Assert
			require.NoError(t, err)
			defer func() { _ = response.Body.Close() }()
			require.Equal(t, http.StatusOK, response.StatusCode)
			assert.Equal(t, "es", response.Header.Get("Content-Language"))
			var entries []localizedReviewQueueEntry
			require.NoError(t, json.NewDecoder(response.Body).Decode(&entries))
			require.Len(t, entries, 1)
			assert.Equal(t, "Rescate", entries[0].Title)
			assert.Equal(t, "es", entries[0].ContentLanguage)
		})
	}
}

// TEST-BND-LANG-002 (review-queue partition): a legacy pending contract
// honestly serves es under an English request.
func TestLangReviewQueueLegacyFallback(t *testing.T) {
	// Arrange
	server, err := languageReviewQueueServer(t, false, true)
	require.NoError(t, err)
	request, err := reviewQueueRequest(server.URL, "?lang=en", "Bearer any")
	require.NoError(t, err)

	// Act
	response, err := server.Client().Do(request)

	// Assert
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	require.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "es", response.Header.Get("Content-Language"))
	var entries []localizedReviewQueueEntry
	require.NoError(t, json.NewDecoder(response.Body).Decode(&entries))
	require.Len(t, entries, 1)
	assert.Equal(t, "contract-old", entries[0].ContractID)
	assert.Equal(t, "Rescate antiguo", entries[0].Title)
	assert.Equal(t, "es", entries[0].ContentLanguage)
}
