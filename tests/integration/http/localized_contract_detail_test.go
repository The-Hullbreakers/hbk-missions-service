//go:build integration

package http_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TEST-BND-LANG-002 (contract detail partition): malformed or ambiguous lang
// values are rejected with 400 INVALID_QUERY_PARAMETER before the contract
// lookup runs — proven by using a missing id that would otherwise reach
// the pre-existing store not-found error (BUG-0054-001);
// a repeated identical value is accepted.
func TestLangContractDetailRejectsInvalidQueryBeforeLookup(t *testing.T) {
	cases := []struct {
		name       string
		contractID string
		query      string
		wantStatus int
	}{
		{"unsupported language", "missing", "?lang=fr", http.StatusBadRequest},
		{"empty language", "missing", "?lang=", http.StatusBadRequest},
		{"conflicting es then en", "missing", "?lang=es&lang=en", http.StatusBadRequest},
		{"repeated es accepted", "contract-rescue", "?lang=es&lang=es", http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			server, err := languageContractsServer(t, true, false)
			require.NoError(t, err)

			// Act
			response, err := detailLangGet(server, tc.contractID, tc.query, "")

			// Assert
			require.NoError(t, err)
			defer func() { _ = response.Body.Close() }()
			assert.Equal(t, tc.wantStatus, response.StatusCode)
			if tc.wantStatus == http.StatusBadRequest {
				assert.Equal(t, "", response.Header.Get("Content-Language"))
				body, err := decodeDetail(response)
				require.NoError(t, err)
				assert.Equal(t, "INVALID_QUERY_PARAMETER", body["code"])
			}
		})
	}
}

// TEST-BND-LANG-002: a bilingual contract serves English title and frozen
// term text under lang=en and labels the payload + Content-Language header.
func TestLangContractDetailServesEnglish(t *testing.T) {
	// Arrange
	server, err := languageContractMemberServer(t)
	require.NoError(t, err)

	// Act
	response, err := detailLangGet(server, "contract-rescue", "?lang=en", "")

	// Assert
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	require.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "en", response.Header.Get("Content-Language"))
	body, err := decodeDetail(response)
	require.NoError(t, err)
	assert.NotContains(t, body, "ContentEn")
	assert.Equal(t, "en", body["content_language"])
	assert.Equal(t, "Rescue", body["title"])
	assert.Equal(t, "ACCEPTED", body["state"])
	assert.Equal(t, float64(50), body["completion_reward"])
	term, err := detailTerm(body)
	require.NoError(t, err)
	assert.Equal(t, "obj-1", term["objective_id"])
	assert.Equal(t, "Capture", term["name"])
	assert.Equal(t, "Capture the pilot", term["description"])
	assert.Equal(t, float64(10), term["merit_per_unit"])
	assert.Equal(t, float64(2), term["unit_limit"])
}

// TEST-BND-LANG-002: an omitted lang defaults to Spanish and ignores any
// Accept-Language header; the label and Content-Language say es.
func TestLangContractDetailSpanishDefaultIgnoresAcceptLanguage(t *testing.T) {
	// Arrange
	server, err := languageContractMemberServer(t)
	require.NoError(t, err)

	// Act
	response, err := detailLangGet(server, "contract-rescue", "", "en")

	// Assert
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	require.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "es", response.Header.Get("Content-Language"))
	body, err := decodeDetail(response)
	require.NoError(t, err)
	assert.Equal(t, "es", body["content_language"])
	assert.Equal(t, "Rescate", body["title"])
	term, err := detailTerm(body)
	require.NoError(t, err)
	assert.Equal(t, "Capturar", term["name"])
}

// TEST-BND-LANG-002: a legacy Spanish-only contract honestly declares es even
// when English is requested — label and header never claim en.
func TestLangContractDetailLegacyFallback(t *testing.T) {
	// Arrange
	server, err := languageContractMemberServer(t)
	require.NoError(t, err)

	// Act
	response, err := detailLangGet(server, "contract-old", "?lang=en", "")

	// Assert
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	require.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "es", response.Header.Get("Content-Language"))
	body, err := decodeDetail(response)
	require.NoError(t, err)
	assert.Equal(t, "es", body["content_language"])
	assert.Equal(t, "Rescate antiguo", body["title"])
	term, err := detailTerm(body)
	require.NoError(t, err)
	assert.Equal(t, "Capturar", term["name"])
}
