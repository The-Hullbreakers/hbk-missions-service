//go:build integration

package dynamodb_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// TEST-INF-LANG-002: rule ORDER is part of the exact Spanish snapshot —
// a frozen offer with identical rules reversed is unresolved, not
// silently overwritten.
func TestBackfillOfferReorderedRulesDrift(t *testing.T) {
	// Arrange
	client, table, setupErr := localizedStoreFixture(t)
	require.NoError(t, setupErr)
	store := missiondynamo.NewOfferStore(client, table)
	offer := spanishBackfillOffer("offer-drift", "Rescate", "v1")
	offer.Rules = []string{"Sin armas", "No ayudas"}
	require.NoError(t, store.Save(offer))
	before, err := rawOfferItem(t.Context(), client, table, "offer-drift")
	require.NoError(t, err)
	reviewed := backfillReviewed()
	reviewed.SourceSpanish.Rules = []string{"No ayudas", "Sin armas"}
	reviewed.English.Rules = []string{"No aids", "No weapons"}

	// Act
	report, err := missiondynamo.NewTranslationBackfiller(client, table).Rehearse(t.Context(), reviewed, false)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, report.Eligible)
	assert.Empty(t, report.Updated)
	assert.Contains(t, report.Unresolved, missiondynamo.BackfillRecord{Partition: "OFFER", ID: "offer-drift", Reason: "SPANISH_SNAPSHOT_MISMATCH"})
	after, err := rawOfferItem(t.Context(), client, table, "offer-drift")
	require.NoError(t, err)
	assert.Equal(t, before, after)
}
