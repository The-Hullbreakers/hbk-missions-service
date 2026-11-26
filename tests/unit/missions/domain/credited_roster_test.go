package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// TEST-DOM-ADJ-002 — TEAM member lists come only from the frozen rosters of
// contracts whose credits actually fed that exact bucket/recipient; every
// member on a credited roster receives a personal grant even when not
// individually credited.
func TestCreditedRoster_OnlyMembersOfCreditedContracts(t *testing.T) {
	t.Parallel()

	// Arrange
	rosters := map[string][]string{
		"c-1": {"m-1", "m-2"},
		"c-2": {"m-3", "m-4"},
	}
	credited := []string{"c-1"}

	// Act
	members, err := domain.MembersOfCreditedContracts(credited, rosters)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []string{"m-1", "m-2"}, members,
		"uncredited contract c-2's frozen roster must not leak in")
}

func TestCreditedRoster_DedupedSortedDeterministicUnion(t *testing.T) {
	t.Parallel()

	// Arrange — unsorted input, shared member across credited contracts,
	// credited IDs supplied out of order.
	rosters := map[string][]string{
		"c-1": {"m-2", "m-1"},
		"c-3": {"m-3", "m-1"},
	}
	credited := []string{"c-3", "c-1"}

	// Act
	members, err := domain.MembersOfCreditedContracts(credited, rosters)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []string{"m-1", "m-2", "m-3"}, members,
		"union must be deduplicated and deterministically sorted")
}

func TestCreditedRoster_MissingNamedRosterRejected(t *testing.T) {
	t.Parallel()

	// Arrange — c-9 is named as credited but has no frozen roster row.
	rosters := map[string][]string{"c-1": {"m-1"}}
	credited := []string{"c-1", "c-9"}

	// Act
	_, err := domain.MembersOfCreditedContracts(credited, rosters)

	// Assert
	require.ErrorIs(t, err, domain.ErrCreditedRosterInvalid)
}

func TestCreditedRoster_NamedContractEmptyRosterRejected(t *testing.T) {
	t.Parallel()

	// Arrange — a TEAM roster exists but froze with zero members: the
	// contract cannot vouch for any grant recipient.
	rosters := map[string][]string{"c-1": {}}
	credited := []string{"c-1"}

	// Act
	_, err := domain.MembersOfCreditedContracts(credited, rosters)

	// Assert
	require.ErrorIs(t, err, domain.ErrCreditedRosterInvalid)
}

func TestCreditedRoster_EmptyCreditedSetRejected(t *testing.T) {
	t.Parallel()

	// Arrange — a TEAM fact cannot mint members from nothing.
	rosters := map[string][]string{"c-1": {"m-1"}}

	// Act
	_, err := domain.MembersOfCreditedContracts(nil, rosters)

	// Assert
	require.ErrorIs(t, err, domain.ErrCreditedRosterInvalid)
}

func TestCreditedRoster_ResultDoesNotShareBackingArray(t *testing.T) {
	t.Parallel()

	// Arrange
	roster := []string{"m-1", "m-2"}
	rosters := map[string][]string{"c-1": roster}
	credited := []string{"c-1"}

	// Act
	members, err := domain.MembersOfCreditedContracts(credited, rosters)

	// Assert
	require.NoError(t, err)
	require.Len(t, members, 2)
	assert.NotSame(t, &roster[0], &members[0],
		"returned members must be a copy, not an alias of the input roster")
}
