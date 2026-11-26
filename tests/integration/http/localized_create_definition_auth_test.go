//go:build integration

package http_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// TEST-HTTP-LANG-003: an unauthenticated caller cannot create a draft.
func TestLangCreateDefinitionUnauthenticated(t *testing.T) {
	// Arrange
	server, store, err := localizedCreateFixture(t, false)
	require.NoError(t, err)
	request, err := postDefinitionRequest(server, spanishOnlyDraftBody, "")
	require.NoError(t, err)

	// Act
	response, err := server.Client().Do(request)

	// Assert
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
	_, getErr := store.Get("draft-spanish")
	assert.True(t, errors.Is(getErr, domain.ErrDefinitionNotFound))
}

// TEST-HTTP-LANG-003: a non-admin caller cannot create a draft.
func TestLangCreateDefinitionNonAdminForbidden(t *testing.T) {
	// Arrange
	server, store, err := localizedCreateFixture(t, true)
	require.NoError(t, err)
	request, err := postDefinitionRequest(server, spanishOnlyDraftBody, "Bearer any")
	require.NoError(t, err)

	// Act
	response, err := server.Client().Do(request)

	// Assert
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	assert.Equal(t, http.StatusForbidden, response.StatusCode)
	_, getErr := store.Get("draft-spanish")
	assert.True(t, errors.Is(getErr, domain.ErrDefinitionNotFound))
}
