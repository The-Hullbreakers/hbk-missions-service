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

func phasedReviewed() missiondynamo.ReviewedMissionTranslation {
	reviewed := backfillReviewed()
	reviewed.SourceSpanish.Objectives = append(reviewed.SourceSpanish.Objectives,
		domain.LocalizedObjective{ObjectiveID: "obj-2", Name: "Rastrear", Description: "Seguir al piloto"})
	reviewed.English.Objectives = append(reviewed.English.Objectives,
		domain.LocalizedObjective{ObjectiveID: "obj-2", Name: "Track", Description: "Follow the pilot"})
	return reviewed
}

// phasedContractFixture builds a two-phase accepted contract; the second
// phase objective name may diverge from the reviewed source.
func phasedContractFixture(phase2Name string) (*domain.MissionContract, error) {
	offer := spanishBackfillOffer("offer-snapshot", "Rescate", "v1")
	offer.Objectives = append(offer.Objectives,
		domain.NewObjectiveNamed("obj-2", "Rastrear", "Seguir al piloto", "BOUNTY_HUNTING", 5, 0, 1))
	phase2 := append([]domain.Objective(nil), offer.Objectives[1:]...)
	phase2[0].Name = phase2Name
	terms, err := domain.NewContractTermsPhased([]domain.Phase{
		{Mandatory: []domain.Objective{offer.Objectives[0]}},
		{Mandatory: phase2},
	})
	if err != nil {
		return nil, err
	}
	contract := domain.NewMissionContract("contract-phased", offer, []string{"player-1"})
	contract.Terms = terms
	return contract, nil
}

// TEST-INF-LANG-002: apply enriches a phased contract — English snapshot
// and localized phase terms — preserving phase order and economics.
func TestBackfillContractApplyPhasedServesEnglish(t *testing.T) {
	// Arrange
	client, table, setupErr := localizedStoreFixture(t)
	require.NoError(t, setupErr)
	store := missiondynamo.NewContractStore(client, table)
	contract, err := phasedContractFixture("Rastrear")
	require.NoError(t, err)
	require.NoError(t, store.Save(contract))
	before, err := rawContractItem(t.Context(), client, table, "contract-phased")
	require.NoError(t, err)
	reviewed := phasedReviewed()

	// Act
	report, err := missiondynamo.NewTranslationBackfiller(client, table).Rehearse(t.Context(), reviewed, true)

	// Assert
	require.NoError(t, err)
	assert.Contains(t, report.Updated, missiondynamo.BackfillRecord{Partition: "CONTRACT", ID: "contract-phased"})
	enriched, err := store.Get("contract-phased")
	require.NoError(t, err)
	served := application.LocalizeContract(*enriched, domain.LanguageEnglish)
	assert.Equal(t, domain.LanguageEnglish, served.ContentLanguage)
	phases := served.Terms.Phases()
	require.Len(t, phases, 2)
	require.Len(t, phases[0].Mandatory, 1)
	assert.Equal(t, "Capture", phases[0].Mandatory[0].Name)
	assert.Equal(t, 10, phases[0].Mandatory[0].MeritPerUnit)
	require.Len(t, phases[1].Mandatory, 1)
	assert.Equal(t, "Track", phases[1].Mandatory[0].Name)
	assert.Equal(t, 5, phases[1].Mandatory[0].MeritPerUnit)
	after, err := rawContractItem(t.Context(), client, table, "contract-phased")
	require.NoError(t, err)
	assert.Equal(t, before, itemWithoutContentEn(after))
}

// TEST-INF-LANG-002: phase-term divergence is unresolved on dry run.
func TestBackfillContractPhasedTermsDrift(t *testing.T) {
	// Arrange
	client, table, setupErr := localizedStoreFixture(t)
	require.NoError(t, setupErr)
	store := missiondynamo.NewContractStore(client, table)
	contract, err := phasedContractFixture("Otro rastreo")
	require.NoError(t, err)
	require.NoError(t, store.Save(contract))
	before, err := rawContractItem(t.Context(), client, table, "contract-phased")
	require.NoError(t, err)
	reviewed := phasedReviewed()

	// Act
	report, err := missiondynamo.NewTranslationBackfiller(client, table).Rehearse(t.Context(), reviewed, false)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, report.Eligible)
	assert.Empty(t, report.Updated)
	assert.Contains(t, report.Unresolved, missiondynamo.BackfillRecord{Partition: "CONTRACT", ID: "contract-phased", Reason: "SPANISH_SNAPSHOT_MISMATCH"})
	after, err := rawContractItem(t.Context(), client, table, "contract-phased")
	require.NoError(t, err)
	assert.Equal(t, before, after)
}
