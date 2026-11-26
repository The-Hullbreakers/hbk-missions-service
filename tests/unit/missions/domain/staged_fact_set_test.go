package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// TEST-DOM-ADJ-003 — the staged fact set owns an immutable, order-independent
// aggregate hash over per-fact digests; assertions cover semantics, never
// opaque hash bytes.
func TestStagedFactSet_CountAndStableHash(t *testing.T) {
	t.Parallel()

	// Arrange
	p := adjFactParams()
	f1 := mustAdjFact(t, p)
	p.Rank, p.RecipientID, p.FinalMerits = 2, "team-b", 200
	f2 := mustAdjFact(t, p)
	facts := []domain.PlacementFact{f1, f2}
	earlier, earlierErr := domain.NewStagedFactSet(facts)

	// Act
	set, err := domain.NewStagedFactSet(facts)

	// Assert
	require.NoError(t, err)
	require.NoError(t, earlierErr)
	assert.Equal(t, 2, set.Count())
	assert.NotEmpty(t, set.Hash())
	assert.Equal(t, earlier.Hash(), set.Hash(),
		"the same fact set must aggregate to the same hash")
}

func TestStagedFactSet_OrderIndependent(t *testing.T) {
	t.Parallel()

	// Arrange
	p := adjFactParams()
	f1 := mustAdjFact(t, p)
	p.Rank, p.RecipientID = 2, "team-b"
	f2 := mustAdjFact(t, p)
	forward, forwardErr := domain.NewStagedFactSet([]domain.PlacementFact{f1, f2})

	// Act
	reverse, err := domain.NewStagedFactSet([]domain.PlacementFact{f2, f1})

	// Assert
	require.NoError(t, err)
	require.NoError(t, forwardErr)
	assert.Equal(t, forward.Hash(), reverse.Hash(),
		"staging order must not affect the aggregate hash")
}

func TestStagedFactSet_EmptySetHasStableNonemptyHash(t *testing.T) {
	t.Parallel()

	// Arrange
	again, againErr := domain.NewStagedFactSet(nil)

	// Act
	set, err := domain.NewStagedFactSet(nil)

	// Assert
	require.NoError(t, err)
	require.NoError(t, againErr)
	assert.Equal(t, 0, set.Count())
	assert.NotEmpty(t, set.Hash(), "an empty manifest still needs a real hash")
	assert.Equal(t, again.Hash(), set.Hash())
}

func TestStagedFactSet_DuplicateEventIDRejected(t *testing.T) {
	t.Parallel()

	// Arrange
	f1 := mustAdjFact(t, adjFactParams())
	duplicate := f1

	// Act
	_, err := domain.NewStagedFactSet([]domain.PlacementFact{f1, duplicate})

	// Assert
	require.ErrorIs(t, err, domain.ErrStagedFactSetInvalid)
}

func TestStagedFactSet_SourceMutationCannotChangeHash(t *testing.T) {
	t.Parallel()

	// Arrange
	members := []string{"m-1", "m-2"}
	p := adjFactParams()
	p.MemberIDs = members
	f1 := mustAdjFact(t, p)
	facts := []domain.PlacementFact{f1}
	set, setErr := domain.NewStagedFactSet(facts)
	require.NoError(t, setErr)
	before := set.Hash()
	members[0] = "m-9"
	facts[0].FinalMerits = 9999

	// Act
	after := set.Hash()

	// Assert
	assert.Equal(t, before, after,
		"aliasing the source members or fact slice must not rewrite the staged hash")
}
