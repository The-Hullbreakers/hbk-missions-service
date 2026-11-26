package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// TEST-DOM-ADJ-002 — the internal adjudication attempt carries a stable
// deterministic identity and the cutoff minted once by the lock.
func TestAdjudicationAttempt_DeterministicIDSeasonAndCutoff(t *testing.T) {
	t.Parallel()

	// Arrange
	cutoff := time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)
	expected, expectedErr := domain.NewAdjudicationAttempt("season-1", cutoff)

	// Act
	attempt, err := domain.NewAdjudicationAttempt("season-1", cutoff)

	// Assert
	require.NoError(t, err)
	require.NoError(t, expectedErr)
	assert.Equal(t, "ADJ#season-1", attempt.ID)
	assert.Equal(t, "season-1", attempt.SeasonID)
	assert.Equal(t, cutoff, attempt.Cutoff, "cutoff is carried verbatim from the lock")
	assert.Equal(t, expected, attempt,
		"same season and cutoff must mint an identical attempt value")
}

func TestAdjudicationAttempt_DifferentSeasonMintsDifferentID(t *testing.T) {
	t.Parallel()

	// Arrange
	cutoff := time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)
	other, otherErr := domain.NewAdjudicationAttempt("season-2", cutoff)

	// Act
	attempt, err := domain.NewAdjudicationAttempt("season-1", cutoff)

	// Assert
	require.NoError(t, err)
	require.NoError(t, otherErr)
	assert.NotEqual(t, other.ID, attempt.ID)
	assert.Equal(t, "ADJ#season-2", other.ID)
}

func TestAdjudicationAttempt_RejectsBlankSeason(t *testing.T) {
	t.Parallel()

	// Arrange
	cutoff := time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)

	// Act
	_, err := domain.NewAdjudicationAttempt(" ", cutoff)

	// Assert
	require.ErrorIs(t, err, domain.ErrAdjudicationAttemptInvalid)
}

func TestAdjudicationAttempt_RejectsZeroCutoff(t *testing.T) {
	t.Parallel()

	// Arrange — zero time; the cutoff must be supplied by the lock, never
	// defaulted inside the domain.

	// Act
	_, err := domain.NewAdjudicationAttempt("season-1", time.Time{})

	// Assert
	require.ErrorIs(t, err, domain.ErrAdjudicationAttemptInvalid)
}
