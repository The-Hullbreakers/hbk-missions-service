package application_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// TEST-APP-MCF-002: a same-write lock conflict has one stable business
// identity and does not stage credits on the persisted contract.
func TestConfirmObjective_AdjudicationFencePropagatesNamedConflict(t *testing.T) {
	t.Parallel()
	// Arrange
	persisted := inProgressContract(t)
	writer := &adjudicationConflictWriter{err: domain.ErrSeasonAdjudicating}
	handler := application.NewConfirmObjectiveHandler(
		mcfContractStore{contract: persisted}, writer,
		mcfSeasonStore{"season-1": domain.NewSeason("season-1", domain.SeasonStateFreezing)},
	)

	// Act
	_, err := handler.Handle(lifecycleMember(t), persisted.ContractID, "obj-a", "unit-1", "season-1")

	// Assert
	require.ErrorIs(t, err, domain.ErrSeasonAdjudicating)
	assert.Equal(t, 1, writer.attempts)
	assert.Nil(t, writer.saved)
	assert.Equal(t, domain.ContractInProgress, persisted.State)
	assert.Empty(t, persisted.PendingCredits())
}
