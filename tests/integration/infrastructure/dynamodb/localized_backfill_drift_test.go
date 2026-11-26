//go:build integration

package dynamodb_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// TEST-INF-LANG-002: any divergence from the exact reviewed Spanish
// snapshot — description, rule text, objective name or description —
// makes the candidate unresolved rather than silently overwritten.
func TestBackfillOfferDriftTable(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*domain.MissionOffer)
	}{
		{"description", func(o *domain.MissionOffer) { o.Description = "Salvar a otro piloto" }},
		{"rule", func(o *domain.MissionOffer) { o.Rules[0] = "Otra regla" }},
		{"objective name", func(o *domain.MissionOffer) { o.Objectives[0].Name = "Otro objetivo" }},
		{"objective description", func(o *domain.MissionOffer) { o.Objectives[0].Description = "Otro piloto" }},
		{"objective id", func(o *domain.MissionOffer) { o.Objectives[0].ObjectiveID = "obj-2" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			client, table, setupErr := localizedStoreFixture(t)
			require.NoError(t, setupErr)
			store := missiondynamo.NewOfferStore(client, table)
			offer := spanishBackfillOffer("offer-drift", "Rescate", "v1")
			tc.mutate(&offer)
			require.NoError(t, store.Save(offer))
			before, err := rawOfferItem(t.Context(), client, table, "offer-drift")
			require.NoError(t, err)
			reviewed := backfillReviewed()

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
		})
	}
}

// TEST-INF-LANG-002: a candidate that already carries authored English —
// even partial — is left untouched and is neither updated nor a failure.
func TestBackfillOfferKeepsPriorAuthoredEnglish(t *testing.T) {
	// Arrange
	client, table, setupErr := localizedStoreFixture(t)
	require.NoError(t, setupErr)
	store := missiondynamo.NewOfferStore(client, table)
	offer := spanishBackfillOffer("offer-prior-en", "Rescate", "v1")
	offer.ContentEn = domain.LocalizedContent{Title: "Prior authored English", Description: "Keep it"}
	require.NoError(t, store.Save(offer))
	before, err := rawOfferItem(t.Context(), client, table, "offer-prior-en")
	require.NoError(t, err)
	reviewed := backfillReviewed()

	// Act
	report, err := missiondynamo.NewTranslationBackfiller(client, table).Rehearse(t.Context(), reviewed, true)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, report.Updated)
	assert.Empty(t, report.Unresolved)
	after, err := rawOfferItem(t.Context(), client, table, "offer-prior-en")
	require.NoError(t, err)
	assert.Equal(t, before, after)
}
