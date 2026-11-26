package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestNewPlacementFact_DerivesDeterministicEventID(t *testing.T) {
	t.Parallel()

	// Arrange / Act
	fact, err := domain.NewPlacementFact(domain.PlacementFactParams{
		SeasonID: "2956-S3", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual,
		RecipientID: "player-1", Rank: 1, FinalMerits: 18500,
	})

	// Assert
	require.NoError(t, err)
	require.NotEmpty(t, fact.EventID)
	again, err := domain.NewPlacementFact(domain.PlacementFactParams{
		SeasonID: "2956-S3", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual,
		RecipientID: "player-1", Rank: 1, FinalMerits: 18500,
	})
	require.NoError(t, err)
	require.Equal(t, fact.EventID, again.EventID)
}

func TestNewPlacementFact_TeamRequiresFrozenMembers(t *testing.T) {
	t.Parallel()

	// Arrange / Act
	_, err := domain.NewPlacementFact(domain.PlacementFactParams{
		SeasonID: "2956-S3", Category: "BOUNTY_HUNTING", Track: domain.TrackTeam,
		RecipientID: "team-alpha", Rank: 1, FinalMerits: 300,
	})

	// Assert
	require.ErrorIs(t, err, domain.ErrTeamFactRequiresMembers)
}

func TestNewPlacementFact_RejectsInvalidRank(t *testing.T) {
	t.Parallel()

	// Arrange / Act
	_, err := domain.NewPlacementFact(domain.PlacementFactParams{
		SeasonID: "2956-S3", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual,
		RecipientID: "player-1", Rank: 0, FinalMerits: 150,
	})

	// Assert
	require.ErrorIs(t, err, domain.ErrPlacementFactInvalid)
}
