package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestOverallStandings_AggregatesIndividualCreditsAcrossCategories(t *testing.T) {
	// Arrange
	credits := []domain.MeritCredit{
		{SeasonID: "season-1", Track: domain.TrackIndividual, Category: "PVP_AIR", RecipientID: "mira", Amount: 120, ContractID: "contract-1", ObjectiveID: "objective-1"},
		{SeasonID: "season-1", Track: domain.TrackIndividual, Category: "RACING", RecipientID: "mira", Amount: 80, ContractID: "contract-1", ObjectiveID: "objective-1"},
		{SeasonID: "season-1", Track: domain.TrackIndividual, Category: "PVP_AIR", RecipientID: "kai", Amount: 190, ContractID: "contract-2", ObjectiveID: "objective-1"},
		{SeasonID: "season-1", Track: domain.TrackTeam, Category: "PVP_AIR", RecipientID: "mira", Amount: 400, ContractID: "contract-team", ObjectiveID: "objective-team"},
		{SeasonID: "season-old", Track: domain.TrackIndividual, Category: "RACING", RecipientID: "kai", Amount: 700, ContractID: "contract-old", ObjectiveID: "objective-old"},
	}
	completed := map[string]bool{"contract-1": true, "contract-2": true, "contract-team": true, "contract-old": true}

	// Act
	got := domain.ProjectOverallStandings(credits, completed, "season-1")

	// Assert
	require.Len(t, got, 2)
	require.Equal(t, "mira", got[0].RecipientID)
	require.Equal(t, "season-1", got[0].SeasonID)
	require.Equal(t, domain.TrackIndividual, got[0].Track)
	require.Equal(t, 200, got[0].Total)
	require.Equal(t, 1, got[0].CompletedMissions)
	require.Equal(t, 2, got[0].UniqueObjectives)
	require.Equal(t, "kai", got[1].RecipientID)
	require.Equal(t, 190, got[1].Total)
}

func TestOverallStandings_TieBreaksByCompletionsObjectivesThenRecipient(t *testing.T) {
	// Arrange
	credits := []domain.MeritCredit{
		{SeasonID: "season-1", Track: domain.TrackIndividual, Category: "PVP_AIR", RecipientID: "alice", Amount: 50, ContractID: "a-1", ObjectiveID: "one"},
		{SeasonID: "season-1", Track: domain.TrackIndividual, Category: "RACING", RecipientID: "alice", Amount: 50, ContractID: "a-1", ObjectiveID: "one"},
		{SeasonID: "season-1", Track: domain.TrackIndividual, Category: "PVP_AIR", RecipientID: "bob", Amount: 50, ContractID: "b-1", ObjectiveID: "one"},
		{SeasonID: "season-1", Track: domain.TrackIndividual, Category: "PVP_AIR", RecipientID: "bob", Amount: 50, ContractID: "b-2", ObjectiveID: "two"},
		{SeasonID: "season-1", Track: domain.TrackIndividual, Category: "PVP_AIR", RecipientID: "carol", Amount: 50, ContractID: "c-1", ObjectiveID: "one"},
		{SeasonID: "season-1", Track: domain.TrackIndividual, Category: "PVP_AIR", RecipientID: "carol", Amount: 50, ContractID: "c-2", ObjectiveID: "one"},
		{SeasonID: "season-1", Track: domain.TrackIndividual, Category: "PVP_AIR", RecipientID: "zeta", Amount: 50, ContractID: "z-1", ObjectiveID: "one"},
		{SeasonID: "season-1", Track: domain.TrackIndividual, Category: "PVP_AIR", RecipientID: "zeta", Amount: 50, ContractID: "z-2", ObjectiveID: "two"},
	}
	completed := map[string]bool{"a-1": true, "b-1": true, "b-2": true, "c-1": true, "c-2": true, "z-1": true, "z-2": true}

	// Act
	got := domain.ProjectOverallStandings(credits, completed, "season-1")

	// Assert
	require.Len(t, got, 4)
	require.Equal(t, []string{"bob", "zeta", "carol", "alice"}, []string{
		got[0].RecipientID, got[1].RecipientID, got[2].RecipientID, got[3].RecipientID,
	})
	require.Equal(t, 2, got[0].CompletedMissions)
	require.Equal(t, 2, got[0].UniqueObjectives)
	require.Equal(t, 1, got[2].UniqueObjectives)
	require.Equal(t, 1, got[3].CompletedMissions)
}

func TestOverallStandings_WithoutIndividualMeritsIsEmpty(t *testing.T) {
	// Arrange
	credits := []domain.MeritCredit{{SeasonID: "season-1", Track: domain.TrackTeam, Category: "PVP_AIR", RecipientID: "team-1", Amount: 10}}

	// Act
	got := domain.ProjectOverallStandings(credits, nil, "season-1")

	// Assert
	require.Empty(t, got)
}

func TestOverallStandings_InProgressContractDoesNotCountCompletion(t *testing.T) {
	// Arrange
	credits := []domain.MeritCredit{
		{SeasonID: "season-1", Track: domain.TrackIndividual, Category: "PVP_AIR", RecipientID: "mira", Amount: 50, ContractID: "done-1", ObjectiveID: "one"},
		{SeasonID: "season-1", Track: domain.TrackIndividual, Category: "PVP_AIR", RecipientID: "mira", Amount: 50, ContractID: "open-1", ObjectiveID: "two"},
	}
	completed := map[string]bool{"done-1": true}

	// Act
	got := domain.ProjectOverallStandings(credits, completed, "season-1")

	// Assert
	require.Len(t, got, 1)
	require.Equal(t, 1, got[0].CompletedMissions)
	require.Equal(t, 2, got[0].UniqueObjectives)
}
