package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// TEST-DOM-ADJ-002/003 — regression branches: identity-less or
// payload-divergent restaging and invalid credited-roster inputs.
func TestStagedFactSet_GivenEmptyEventID_WhenStaged_ThenRejected(t *testing.T) {
	t.Parallel()

	// Arrange — a fact that never persisted its event identity cannot be
	// staged: replay could restage divergent content under a blank key.
	fact := domain.PlacementFact{
		SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackTeam,
		RecipientID: "team-a", Rank: 1, FinalMerits: 300,
		MemberIDs: []string{"m-1", "m-2"},
	}

	// Act
	_, err := domain.NewStagedFactSet([]domain.PlacementFact{fact})

	// Assert
	require.ErrorIs(t, err, domain.ErrStagedFactSetInvalid)
}

func TestStagedFactSet_GivenSameEventIDDivergentPayload_WhenStaged_ThenRejected(t *testing.T) {
	t.Parallel()

	// Arrange — two facts sharing one EventID but disagreeing on content:
	// the collision must surface, never silently restage either payload.
	f1 := domain.PlacementFact{
		EventID: "evt-a", SeasonID: "season-1", Category: "BOUNTY_HUNTING",
		Track: domain.TrackTeam, RecipientID: "team-a", Rank: 1,
		FinalMerits: 300, MemberIDs: []string{"m-1", "m-2"},
	}
	divergent := f1
	divergent.FinalMerits = 9999

	// Act
	_, err := domain.NewStagedFactSet([]domain.PlacementFact{f1, divergent})

	// Assert
	require.ErrorIs(t, err, domain.ErrStagedFactSetInvalid)
}

func TestCreditedRoster_GivenBlankCreditedID_WhenResolved_ThenRejected(t *testing.T) {
	t.Parallel()

	for name, blank := range map[string]string{"empty": "", "whitespace": "   "} {
		t.Run("id_"+name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			rosters := map[string][]string{"c-1": {"m-1"}}
			credited := []string{"c-1", blank}

			// Act
			_, err := domain.MembersOfCreditedContracts(credited, rosters)

			// Assert
			require.ErrorIs(t, err, domain.ErrCreditedRosterInvalid)
		})
	}
}

func TestCreditedRoster_GivenBlankMember_WhenResolved_ThenRejected(t *testing.T) {
	t.Parallel()

	for name, blank := range map[string]string{"empty": "", "whitespace": "   "} {
		t.Run("member_"+name, func(t *testing.T) {
			t.Parallel()

			// Arrange — a frozen roster row carrying a blank member cannot
			// mint a personal grant recipient.
			rosters := map[string][]string{"c-1": {"m-1", blank}}
			credited := []string{"c-1"}

			// Act
			_, err := domain.MembersOfCreditedContracts(credited, rosters)

			// Assert
			require.ErrorIs(t, err, domain.ErrCreditedRosterInvalid)
		})
	}
}

func TestNewPlacementFact_GivenMemberSlice_WhenConstructed_ThenDoesNotAlias(t *testing.T) {
	t.Parallel()

	// Arrange — caller-owned member slice must not share its backing
	// array with the constructed fact.
	members := []string{"m-1", "m-2"}
	params := adjFactParams()
	params.MemberIDs = members

	// Act
	fact, err := domain.NewPlacementFact(params)

	// Assert
	require.NoError(t, err)
	assert.NotSame(t, &members[0], &fact.MemberIDs[0],
		"fact must clone the member slice, not alias the caller's array")
}
