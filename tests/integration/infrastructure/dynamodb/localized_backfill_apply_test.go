//go:build integration

package dynamodb_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// TEST-INF-LANG-002: apply enriches only the compatible candidate — English
// snapshot written, Spanish/economics/lineage untouched — and leaves
// divergent candidates alone.
func TestBackfillOfferApplyOnlyEnglish(t *testing.T) {
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
	report, err := missiondynamo.NewTranslationBackfiller(client, table).Rehearse(t.Context(), reviewed, true)

	// Assert
	require.NoError(t, err)
	assert.Contains(t, report.Eligible, missiondynamo.BackfillRecord{Partition: "OFFER", ID: "offer-match"})
	assert.Contains(t, report.Updated, missiondynamo.BackfillRecord{Partition: "OFFER", ID: "offer-match"})
	assert.Contains(t, report.Unresolved, missiondynamo.BackfillRecord{Partition: "OFFER", ID: "offer-drift", Reason: "SPANISH_SNAPSHOT_MISMATCH"})
	enriched, err := store.Get("offer-match")
	require.NoError(t, err)
	assert.Equal(t, "Rescue", enriched.ContentEn.Title)
	assert.Equal(t, "Save the pilot", enriched.ContentEn.Description)
	assert.Equal(t, []string{"No aids"}, enriched.ContentEn.Rules)
	require.Len(t, enriched.ContentEn.Objectives, 1)
	assert.Equal(t, "obj-1", enriched.ContentEn.Objectives[0].ObjectiveID)
	assert.Equal(t, "Capture", enriched.ContentEn.Objectives[0].Name)
	assert.Equal(t, "Capture the pilot", enriched.ContentEn.Objectives[0].Description)
	assert.Equal(t, "Rescate", enriched.Title)
	assert.Equal(t, 50, enriched.CompletionReward)
	assert.Equal(t, "def-rescue", enriched.DefinitionCode)
	assert.Equal(t, "v1", enriched.DefinitionVersion)
	served := application.LocalizeOffer(enriched, domain.LanguageEnglish)
	assert.Equal(t, domain.LanguageEnglish, served.ContentLanguage)
	assert.Equal(t, "Rescue", served.Title)
	drift, err := store.Get("offer-drift")
	require.NoError(t, err)
	assert.Equal(t, domain.LocalizedContent{}, drift.ContentEn)
	assert.Equal(t, domain.LanguageSpanish, application.LocalizeOffer(drift, domain.LanguageEnglish).ContentLanguage)
	matchAfter, err := rawOfferItem(t.Context(), client, table, "offer-match")
	require.NoError(t, err)
	driftAfter, err := rawOfferItem(t.Context(), client, table, "offer-drift")
	require.NoError(t, err)
	assert.Equal(t, matchBefore, itemWithoutContentEn(matchAfter))
	assert.Equal(t, driftBefore, driftAfter)
}
