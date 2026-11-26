//go:build integration

package http_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TEST-BND-LANG-002: the reviewer detail keeps its full wire shape while the
// content localizes — pending_credits/rejected_units/phases keys survive
// (invitations stays omitempty) alongside the en label.
func TestLangContractDetailReviewerKeepsFullView(t *testing.T) {
	// Arrange
	server, err := languageContractsServer(t, true, false)
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
	assert.Equal(t, "en", body["content_language"])
	assert.Equal(t, "Rescue", body["title"])
	for _, key := range []string{"pending_credits", "rejected_units", "phases", "current_phase"} {
		assert.Contains(t, body, key)
	}
}

// TEST-BND-LANG-002: a roster member still gets the minimized projection —
// the same translated content plus the member-filtered credit fields.
func TestLangContractDetailMemberKeepsMinimizedView(t *testing.T) {
	// Arrange
	server, err := languageContractMemberServer(t)
	require.NoError(t, err)

	// Act
	response, err := detailLangGet(server, "contract-rescue", "?lang=en", "")

	// Assert
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	require.Equal(t, http.StatusOK, response.StatusCode)
	body, err := decodeDetail(response)
	require.NoError(t, err)
	assert.Equal(t, "en", body["content_language"])
	assert.Equal(t, "Rescue", body["title"])
	term, err := detailTerm(body)
	require.NoError(t, err)
	assert.Equal(t, "Capture", term["name"])
	assert.Contains(t, body, "pending_credits")
	assert.Empty(t, body["pending_credits"])
	assert.Empty(t, body["rejected_units"])
}

// TEST-BND-LANG-002: an outsider is still denied before any localization —
// 403 with the unchanged authorization code and no Content-Language.
func TestLangContractDetailOutsiderDenied(t *testing.T) {
	// Arrange
	server, err := languageContractMemberServer(t)
	require.NoError(t, err)

	// Act
	response, err := detailLangGet(server, "contract-foreign", "?lang=en", "")

	// Assert
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	assert.Equal(t, http.StatusForbidden, response.StatusCode)
	assert.Equal(t, "", response.Header.Get("Content-Language"))
	body, err := decodeDetail(response)
	require.NoError(t, err)
	assert.Equal(t, "FORBIDDEN", body["code"])
}

// TEST-BND-LANG-002: an unknown contract id still returns the unchanged
// 404 CONTRACT_NOT_FOUND with no Content-Language header.
func TestLangContractDetailUnknownReturns404(t *testing.T) {
	// Arrange
	server, err := languageContractMemberServer(t)
	require.NoError(t, err)

	// Act
	response, err := detailLangGet(server, "contract-missing", "?lang=en", "")

	// Assert
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	assert.Equal(t, http.StatusNotFound, response.StatusCode)
	assert.Equal(t, "", response.Header.Get("Content-Language"))
	body, err := decodeDetail(response)
	require.NoError(t, err)
	assert.Equal(t, "CONTRACT_NOT_FOUND", body["code"])
}
