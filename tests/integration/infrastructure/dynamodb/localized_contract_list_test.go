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

// TEST-INF-LANG-002: a player's mixed contract list serves each entry in
// its own language — bilingual English, legacy Spanish — while an
// outsider's contract stays invisible.
func TestLocalizedContracts_MixedSnapshotsRoundTrip(t *testing.T) {
	// Arrange
	client, table, setupErr := localizedStoreFixture(t)
	require.NoError(t, setupErr)
	store := missiondynamo.NewContractStore(client, table)
	bilingual, bilingualErr := localizedBilingualContract()
	require.NoError(t, bilingualErr)
	legacy, legacyErr := localizedLegacyContract()
	require.NoError(t, legacyErr)
	outsider := domain.NewMissionContract("contract-other", legacy.Offer, []string{"player-2"})
	outsider.Terms = legacy.Terms
	require.NoError(t, store.Save(bilingual))
	require.NoError(t, store.Save(legacy))
	require.NoError(t, store.Save(outsider))

	// Act
	served, err := application.NewListPlayerContractsHandler(store).Handle("player-1", domain.ContractListFilter{Status: domain.ContractStatusAll, SeasonID: "season-1"}, domain.LanguageEnglish)

	// Assert
	require.NoError(t, err)
	require.Len(t, served, 2)
	byID := map[string]application.LocalizedContract{}
	for _, entry := range served {
		byID[entry.ContractID] = entry
	}
	require.NotContains(t, byID, "contract-other")
	require.Contains(t, byID, "contract-rescue")
	require.Contains(t, byID, "contract-old")
	assert.Equal(t, domain.LanguageEnglish, byID["contract-rescue"].ContentLanguage)
	assert.Equal(t, "Rescue", byID["contract-rescue"].Offer.Title)
	bilingualTerms := byID["contract-rescue"].Terms.Objectives()
	require.Len(t, bilingualTerms, 1)
	assert.Equal(t, "Capture", bilingualTerms[0].Name)
	assert.Equal(t, 10, bilingualTerms[0].MeritPerUnit)
	assert.Equal(t, domain.LanguageSpanish, byID["contract-old"].ContentLanguage)
	assert.Equal(t, "Rescate antiguo", byID["contract-old"].Offer.Title)
	legacyTerms := byID["contract-old"].Terms.Objectives()
	require.Len(t, legacyTerms, 1)
	assert.Equal(t, "Capturar", legacyTerms[0].Name)
	assert.Equal(t, 10, legacyTerms[0].MeritPerUnit)
}
