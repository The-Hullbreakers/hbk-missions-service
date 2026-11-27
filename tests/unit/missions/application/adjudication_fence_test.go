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
	evidence, uris := storedEvidence(persisted.ContractID, "e1")
	handler := application.NewConfirmObjectiveHandlerWithEvidence(
		mcfContractStore{contract: persisted}, writer,
		mcfSeasonStore{"season-1": domain.NewSeason("season-1", domain.SeasonStateFreezing)},
		evidence,
	)

	// Act
	_, err := handler.HandleWithMedia(lifecycleMember(t), persisted.ContractID, "obj-a", "unit-1", uris[0], "season-1")

	// Assert
	require.ErrorIs(t, err, domain.ErrSeasonAdjudicating)
	assert.Equal(t, 1, writer.attempts)
	assert.Nil(t, writer.saved)
	assert.Equal(t, domain.ContractInProgress, persisted.State)
	assert.Empty(t, persisted.PendingCredits())
}
