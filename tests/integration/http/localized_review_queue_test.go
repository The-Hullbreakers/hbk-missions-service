//go:build integration

package http_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type localizedReviewQueueEntry struct {
	ContractID      string   `json:"contract_id"`
	OfferID         string   `json:"offer_id"`
	DefinitionCode  string   `json:"definition_code"`
	Title           string   `json:"title"`
	ContentLanguage string   `json:"content_language"`
	State           string   `json:"state"`
	Roster          []string `json:"roster"`
	ConfirmedCount  int      `json:"confirmed_count"`
	PendingCredits  int      `json:"pending_credits"`
}

func reviewQueueRequest(serverURL, query, auth string) (*http.Request, error) {
	request, err := http.NewRequest(http.MethodGet, serverURL+"/api/v1/missions/review-queue"+query, nil)
	if err != nil {
		return nil, err
	}
	if auth != "" {
		request.Header.Set("Authorization", auth)
	}
	return request, nil
}

// TEST-BND-LANG-002 (review-queue partition): malformed lang is rejected
// before any queue read.
func TestLangReviewQueueRejectsInvalidQuery(t *testing.T) {
	// Arrange
	server, err := languageReviewQueueServer(t, true, false)
	require.NoError(t, err)
	cases := []struct {
		name  string
		query string
		want  int
	}{
		{"unsupported", "?lang=fr", 400},
		{"empty", "?lang=", 400},
		{"conflicting", "?lang=es&lang=en", 400},
		{"reverse conflict", "?lang=en&lang=es", 400},
		{"repeated en accepted", "?lang=en&lang=en", 200},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			request, err := reviewQueueRequest(server.URL, tc.query, "Bearer any")
			require.NoError(t, err)

			// Act
			response, err := server.Client().Do(request)

			// Assert
			require.NoError(t, err)
			defer func() { _ = response.Body.Close() }()
			assert.Equal(t, tc.want, response.StatusCode)
			if tc.want == 400 {
				assert.Equal(t, "", response.Header.Get("Content-Language"))
				var body map[string]string
				require.NoError(t, json.NewDecoder(response.Body).Decode(&body))
				assert.Equal(t, "INVALID_QUERY_PARAMETER", body["code"])
			} else {
				assert.Equal(t, "en", response.Header.Get("Content-Language"))
			}
		})
	}
}

// TEST-BND-LANG-002 (review-queue partition): an empty pending queue stays
// an empty JSON array with no Content-Language.
func TestLangReviewQueueEmpty(t *testing.T) {
	// Arrange
	server, err := languageReviewQueueServer(t, false, false)
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
	var entries []json.RawMessage
	require.NoError(t, json.NewDecoder(response.Body).Decode(&entries))
	assert.Empty(t, entries)
}
