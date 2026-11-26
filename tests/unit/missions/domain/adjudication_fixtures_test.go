package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// adjFactParams returns a valid TEAM placement-fact parameter set so that
// member identity participates in every digest/staging assertion.
func adjFactParams() domain.PlacementFactParams {
	return domain.PlacementFactParams{
		SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackTeam,
		RecipientID: "team-a", Rank: 1, FinalMerits: 300,
		MemberIDs: []string{"m-1", "m-2"},
	}
}

func mustAdjFact(t *testing.T, p domain.PlacementFactParams) domain.PlacementFact {
	t.Helper()
	fact, err := domain.NewPlacementFact(p)
	require.NoError(t, err)
	return fact
}
