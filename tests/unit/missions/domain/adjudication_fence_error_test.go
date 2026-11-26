package domain_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// TEST-DOM-ADJ-004: callers distinguish an engaged adjudication lock from
// an inactive season or an invalid public state transition.
func TestSeasonAdjudicatingError_GivenLockedSeason_WhenWriteRejected_ThenIdentityIsDistinct(t *testing.T) {
	t.Parallel()
	// Arrange
	reason := domain.ErrSeasonAdjudicating

	// Act
	wrapped := fmt.Errorf("season credit rejected: %w", reason)

	// Assert
	require.ErrorIs(t, wrapped, domain.ErrSeasonAdjudicating)
	assert.True(t, errors.Is(wrapped, reason))
	assert.NotErrorIs(t, wrapped, domain.ErrSeasonNotActive)
	assert.NotErrorIs(t, wrapped, domain.ErrInvalidSeasonStateTransition)
}
