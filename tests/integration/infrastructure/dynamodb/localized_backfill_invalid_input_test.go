//go:build integration

package dynamodb_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// TEST-INF-LANG-002: a structurally incomplete reviewed English
// translation is rejected before any enrichment — the candidate item is
// left byte-identical.
func TestBackfillOfferRejectsIncompleteEnglishInput(t *testing.T) {
	// Arrange
	client, table, setupErr := localizedStoreFixture(t)
	require.NoError(t, setupErr)
	store := missiondynamo.NewOfferStore(client, table)
	require.NoError(t, store.Save(spanishBackfillOffer("offer-match", "Rescate", "v1")))
	before, err := rawOfferItem(t.Context(), client, table, "offer-match")
	require.NoError(t, err)
	reviewed := backfillReviewed()
	reviewed.English.Objectives = nil

	// Act
	report, err := missiondynamo.NewTranslationBackfiller(client, table).Rehearse(t.Context(), reviewed, true)

	// Assert
	assert.ErrorIs(t, err, domain.ErrDefinitionInvalid)
	assert.Empty(t, report.Updated)
	after, err := rawOfferItem(t.Context(), client, table, "offer-match")
	require.NoError(t, err)
	assert.Equal(t, before, after)
	_, hasEn := after["ContentEn"]
	assert.False(t, hasEn)
}
