//go:build integration

package http_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TEST-BND-LANG-002 (review-queue partition): a mixed pending queue serves
// each entry in its own language — no header on the heterogeneous page —
// and the accepted contract stays excluded.
func TestLangReviewQueueMixedAndFiltered(t *testing.T) {
	// Arrange
	server, err := languageReviewQueueServer(t, true, true)
	require.NoError(t, err)
	request, err := reviewQueueRequest(server.URL, "?lang=en", "Bearer any")
	require.NoError(t, err)

	// Act
	response, err := server.Client().Do(request)

	// Assert
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	require.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "", response.Header.Get("Content-Language"))
	var entries []localizedReviewQueueEntry
	require.NoError(t, json.NewDecoder(response.Body).Decode(&entries))
	require.Len(t, entries, 2)
	byID := map[string]localizedReviewQueueEntry{}
	for _, entry := range entries {
		byID[entry.ContractID] = entry
	}
	assert.NotContains(t, byID, "contract-accepted")
	require.Contains(t, byID, "contract-rescue")
	require.Contains(t, byID, "contract-old")
	assert.Equal(t, "en", byID["contract-rescue"].ContentLanguage)
	assert.Equal(t, "Rescue", byID["contract-rescue"].Title)
	assert.Equal(t, "PENDING_VALIDATION", byID["contract-rescue"].State)
	assert.Equal(t, []string{"player-1"}, byID["contract-rescue"].Roster)
	assert.Equal(t, "es", byID["contract-old"].ContentLanguage)
	assert.Equal(t, "Rescate antiguo", byID["contract-old"].Title)
	assert.Equal(t, "PENDING_VALIDATION", byID["contract-old"].State)
}
