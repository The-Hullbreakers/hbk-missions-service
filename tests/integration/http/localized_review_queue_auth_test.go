//go:build integration

package http_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TEST-BND-LANG-002 (review-queue partition): an unauthenticated request is
// rejected by the middleware before any lang parsing.
func TestLangReviewQueueUnauthenticated(t *testing.T) {
	// Arrange
	server, err := languageReviewQueueServer(t, false, false)
	require.NoError(t, err)
	request, err := reviewQueueRequest(server.URL, "?lang=fr", "")
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

// TEST-BND-LANG-002 (review-queue partition): a non-reviewer member is
// forbidden regardless of language.
func TestLangReviewQueueNonReviewerForbidden(t *testing.T) {
	// Arrange
	server, err := languageSessionServer(t, langMemberSession{}, nil)
	require.NoError(t, err)
	request, err := reviewQueueRequest(server.URL, "?lang=en", "Bearer any")
	require.NoError(t, err)

	// Act
	response, err := server.Client().Do(request)

	// Assert
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	assert.Equal(t, http.StatusForbidden, response.StatusCode)
	assert.Equal(t, "", response.Header.Get("Content-Language"))
}
