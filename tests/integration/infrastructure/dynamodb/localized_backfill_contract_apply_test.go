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

// TEST-INF-LANG-002: apply enriches only the matching contract — English
// snapshot on Offer + localized frozen terms — leaving Spanish text,
// economics, roster, state, and lineage untouched.
func TestBackfillContractApplyOnlyEnglish(t *testing.T) {
	// Arrange
	client, table, setupErr := localizedStoreFixture(t)
	require.NoError(t, setupErr)
	store := missiondynamo.NewContractStore(client, table)
	match, err := contractBackfillFixture("contract-match", "Capturar")
	require.NoError(t, err)
	drift, err := contractBackfillFixture("contract-drift", "Otro objetivo")
	require.NoError(t, err)
	require.NoError(t, store.Save(match))
	require.NoError(t, store.Save(drift))
	matchBefore, err := rawContractItem(t.Context(), client, table, "contract-match")
	require.NoError(t, err)
	driftBefore, err := rawContractItem(t.Context(), client, table, "contract-drift")
	require.NoError(t, err)
	reviewed := backfillReviewed()

	// Act
	report, err := missiondynamo.NewTranslationBackfiller(client, table).Rehearse(t.Context(), reviewed, true)

	// Assert
	require.NoError(t, err)
	assert.Contains(t, report.Updated, missiondynamo.BackfillRecord{Partition: "CONTRACT", ID: "contract-match"})
	assert.Contains(t, report.Unresolved, missiondynamo.BackfillRecord{Partition: "CONTRACT", ID: "contract-drift", Reason: "SPANISH_SNAPSHOT_MISMATCH"})
	enriched, err := store.Get("contract-match")
	require.NoError(t, err)
	assert.Equal(t, "Rescue", enriched.Offer.ContentEn.Title)
	assert.Equal(t, []string{"No aids"}, enriched.Offer.ContentEn.Rules)
	served := application.LocalizeContract(*enriched, domain.LanguageEnglish)
	assert.Equal(t, domain.LanguageEnglish, served.ContentLanguage)
	assert.Equal(t, "Rescue", served.Offer.Title)
	terms := served.Terms.Objectives()
	require.Len(t, terms, 1)
	assert.Equal(t, "Capture", terms[0].Name)
	assert.Equal(t, 10, terms[0].MeritPerUnit)
	assert.Equal(t, "Rescate", enriched.Offer.Title)
	assert.Equal(t, []string{"player-1"}, enriched.Roster())
	assert.Equal(t, "offer-snapshot", enriched.Offer.OfferID)
	assert.Equal(t, "v1", enriched.Offer.DefinitionVersion)
	driftRead, err := store.Get("contract-drift")
	require.NoError(t, err)
	assert.Equal(t, domain.LocalizedContent{}, driftRead.Offer.ContentEn)
	assert.Equal(t, domain.LanguageSpanish, application.LocalizeContract(*driftRead, domain.LanguageEnglish).ContentLanguage)
	matchAfter, err := rawContractItem(t.Context(), client, table, "contract-match")
	require.NoError(t, err)
	driftAfter, err := rawContractItem(t.Context(), client, table, "contract-drift")
	require.NoError(t, err)
	assert.Equal(t, matchBefore, itemWithoutContentEn(matchAfter))
	assert.Equal(t, driftBefore, driftAfter)
}
