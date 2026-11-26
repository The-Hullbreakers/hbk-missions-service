//go:build integration

package dynamodb_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// TEST-INF-LANG-002: a dry run reports the compatible offer as eligible,
// reports Spanish-drift candidates as unresolved, and mutates nothing.
func TestBackfillOfferDryRunReportsSpanishDrift(t *testing.T) {
	// Arrange
	client, table, setupErr := localizedStoreFixture(t)
	require.NoError(t, setupErr)
	store := missiondynamo.NewOfferStore(client, table)
	require.NoError(t, store.Save(spanishBackfillOffer("offer-match", "Rescate", "v1")))
	require.NoError(t, store.Save(spanishBackfillOffer("offer-drift", "Rescate anterior", "v1")))
	reviewed := backfillReviewed()
	matchBefore, err := rawOfferItem(t.Context(), client, table, "offer-match")
	require.NoError(t, err)
	driftBefore, err := rawOfferItem(t.Context(), client, table, "offer-drift")
	require.NoError(t, err)

	// Act
	report, err := missiondynamo.NewTranslationBackfiller(client, table).Rehearse(t.Context(), reviewed, false)

	// Assert
	require.NoError(t, err)
	assert.Contains(t, report.Eligible, missiondynamo.BackfillRecord{Partition: "OFFER", ID: "offer-match"})
	assert.Empty(t, report.Updated)
	assert.Contains(t, report.Unresolved, missiondynamo.BackfillRecord{Partition: "OFFER", ID: "offer-drift", Reason: "SPANISH_SNAPSHOT_MISMATCH"})
	matchAfter, err := rawOfferItem(t.Context(), client, table, "offer-match")
	require.NoError(t, err)
	driftAfter, err := rawOfferItem(t.Context(), client, table, "offer-drift")
	require.NoError(t, err)
	assert.Equal(t, matchBefore, matchAfter)
	assert.Equal(t, driftBefore, driftAfter)
	_, matchHasEn := matchAfter["ContentEn"]
	_, driftHasEn := driftAfter["ContentEn"]
	assert.False(t, matchHasEn)
	assert.False(t, driftHasEn)
}

// TEST-INF-LANG-002: a wrong-version candidate is skipped with
// VERSION_MISMATCH and never appears in Eligible/Updated.
func TestBackfillOfferRejectsWrongVersion(t *testing.T) {
	// Arrange
	client, table, setupErr := localizedStoreFixture(t)
	require.NoError(t, setupErr)
	store := missiondynamo.NewOfferStore(client, table)
	require.NoError(t, store.Save(spanishBackfillOffer("offer-wrong-version", "Rescate", "v2")))
	reviewed := backfillReviewed()
	before, err := rawOfferItem(t.Context(), client, table, "offer-wrong-version")
	require.NoError(t, err)

	// Act
	report, err := missiondynamo.NewTranslationBackfiller(client, table).Rehearse(t.Context(), reviewed, false)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, report.Eligible)
	assert.Empty(t, report.Updated)
	assert.Contains(t, report.Unresolved, missiondynamo.BackfillRecord{Partition: "OFFER", ID: "offer-wrong-version", Reason: "VERSION_MISMATCH"})
	after, err := rawOfferItem(t.Context(), client, table, "offer-wrong-version")
	require.NoError(t, err)
	assert.Equal(t, before, after)
}
