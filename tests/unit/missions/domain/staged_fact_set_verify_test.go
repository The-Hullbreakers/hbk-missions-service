package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// TEST-DOM-ADJ-003 — Verify is the staging gatekeeper: only the exact
// (count, hash) pair captured at staging passes.
func TestStagedFactSet_VerifyAcceptsMatchingCountAndHash(t *testing.T) {
	t.Parallel()

	// Arrange
	p := adjFactParams()
	f1 := mustAdjFact(t, p)
	p.Rank, p.RecipientID = 2, "team-b"
	set, setErr := domain.NewStagedFactSet([]domain.PlacementFact{f1, mustAdjFact(t, p)})
	require.NoError(t, setErr)
	count, hash := set.Count(), set.Hash()

	// Act
	err := set.Verify(count, hash)

	// Assert
	require.NoError(t, err)
}

func TestStagedFactSet_VerifyRejectsCountMismatch(t *testing.T) {
	t.Parallel()

	// Arrange
	set, setErr := domain.NewStagedFactSet([]domain.PlacementFact{mustAdjFact(t, adjFactParams())})
	require.NoError(t, setErr)
	hash := set.Hash()

	// Act
	err := set.Verify(set.Count()+1, hash)

	// Assert
	require.ErrorIs(t, err, domain.ErrStagedFactSetMismatch)
}

func TestStagedFactSet_VerifyRejectsHashMismatch(t *testing.T) {
	t.Parallel()

	// Arrange
	set, setErr := domain.NewStagedFactSet([]domain.PlacementFact{mustAdjFact(t, adjFactParams())})
	require.NoError(t, setErr)
	count := set.Count()

	// Act
	err := set.Verify(count, "tampered-hash")

	// Assert
	require.ErrorIs(t, err, domain.ErrStagedFactSetMismatch)
}

func TestStagedFactSet_VerifyEmptySetAcceptsOwnShape(t *testing.T) {
	t.Parallel()

	// Arrange — an empty committed manifest verifies against its own
	// declared zero-count shape and must complete.
	set, setErr := domain.NewStagedFactSet(nil)
	require.NoError(t, setErr)
	count, hash := set.Count(), set.Hash()

	// Act
	err := set.Verify(count, hash)

	// Assert
	require.NoError(t, err)
}

func TestStagedFactSet_VerifyRejectsEmptyAgainstPopulatedExpectation(t *testing.T) {
	t.Parallel()

	// Arrange — a populated staged set must never verify against the empty
	// manifest's declared shape.
	populated, popErr := domain.NewStagedFactSet([]domain.PlacementFact{mustAdjFact(t, adjFactParams())})
	require.NoError(t, popErr)
	empty, emptyErr := domain.NewStagedFactSet(nil)
	require.NoError(t, emptyErr)
	count, hash := empty.Count(), empty.Hash()

	// Act
	err := populated.Verify(count, hash)

	// Assert
	require.ErrorIs(t, err, domain.ErrStagedFactSetMismatch)
}
