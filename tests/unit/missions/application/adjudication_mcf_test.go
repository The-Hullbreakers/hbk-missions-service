package application_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// TEST-APP-MCF-001 — characterization only: a rejected write propagates
// its error and leaves the persisted contract snapshot untouched; the
// adjudication lock itself lands at the T-12 boundary.
func TestAdjudication_ApproveCommitsEligibleCreditsBeforeLock(t *testing.T) {
	// Arrange
	seed := mcfContractStore{contract: pendingContract(t, 150)}
	writer := &reviewStore{}
	handler := application.NewApproveContractHandler(seed, writer)

	// Act
	credits, err := handler.HandleWithVerdict(reviewer(t, "moderator"), "c-1", nil)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, writer.saved)
	assert.Equal(t, domain.ContractCompleted, writer.saved.State)
	completion := 0
	for _, c := range credits {
		if c.ObjectiveID == domain.CompletionObjectiveID {
			completion = c.Amount
		}
	}
	assert.Equal(t, 150, completion, "eligible completion reward commits pre-lock")
}

func TestAdjudication_ApproveConflictLeavesPersistedContractUntouched(t *testing.T) {
	// Arrange
	persisted := pendingContract(t, 150)
	seed := mcfContractStore{contract: persisted}
	writer := &adjudicationConflictWriter{err: errAdjudicationBlocked}
	handler := application.NewApproveContractHandler(seed, writer)

	// Act
	_, err := handler.HandleWithVerdict(reviewer(t, "admin"), "c-1", nil)

	// Assert
	require.ErrorIs(t, err, errAdjudicationBlocked)
	assert.Equal(t, 1, writer.attempts)
	assert.Nil(t, writer.saved, "conflict must not persist a contract transition")
	assert.Empty(t, writer.credits, "conflict must not persist unit or completion credits")
	assert.Equal(t, domain.ContractPendingValidation, persisted.State,
		"persisted contract stays pending — domain mutation was in-memory only")
	assert.Len(t, persisted.PendingCredits(), 1, "staged unit credit survives untouched")
}

func TestAdjudication_ConfirmCreditsFreezingInProgressBeforeLock(t *testing.T) {
	// Arrange
	contract := inProgressContract(t)
	seed := mcfContractStore{contract: contract}
	writer := &reviewStore{}
	seasons := mcfSeasonStore{"season-1": domain.NewSeason("season-1", domain.SeasonStateFreezing)}
	evidence, uris := storedEvidence(contract.ContractID, "e1")
	handler := application.NewConfirmObjectiveHandlerWithEvidence(seed, writer, seasons, evidence)

	// Act
	credit, err := handler.HandleWithMedia(lifecycleMember(t), contract.ContractID, "obj-a", "unit-1", uris[0], "season-1")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "season-1", credit.SeasonID)
	assert.Positive(t, credit.Amount)
	require.NotNil(t, writer.saved)
	assert.Len(t, writer.saved.PendingCredits(), 1,
		"staged credit persists frozen under FREEZING pre-lock")
}

func TestAdjudication_ConfirmConflictLeavesContractUnstaged(t *testing.T) {
	// Arrange
	contract := inProgressContract(t)
	seed := mcfContractStore{contract: contract}
	writer := &adjudicationConflictWriter{err: errAdjudicationBlocked}
	seasons := mcfSeasonStore{"season-1": domain.NewSeason("season-1", domain.SeasonStateFreezing)}
	evidence, uris := storedEvidence(contract.ContractID, "e1")
	handler := application.NewConfirmObjectiveHandlerWithEvidence(seed, writer, seasons, evidence)

	// Act
	_, err := handler.HandleWithMedia(lifecycleMember(t), contract.ContractID, "obj-a", "unit-1", uris[0], "season-1")

	// Assert
	require.ErrorIs(t, err, errAdjudicationBlocked)
	assert.Equal(t, 1, writer.attempts)
	assert.Nil(t, writer.saved)
	assert.Equal(t, domain.ContractInProgress, contract.State,
		"persisted contract keeps its in-progress state")
	assert.Empty(t, contract.PendingCredits(), "persisted contract gains no staged credit")
}

func TestAdjudication_RejectWritesZeroCreditsAndStaysUnfenced(t *testing.T) {
	// Arrange
	persisted := pendingContract(t, 150)
	seed := mcfContractStore{contract: persisted}
	writer := &reviewStore{}
	handler := application.NewRejectContractHandler(seed, writer)

	// Act
	err := handler.Handle(reviewer(t, "moderator"), "c-1")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, writer.saved)
	assert.Equal(t, domain.ContractDisqualified, writer.saved.State)
	assert.Empty(t, writer.credits, "reject persists a state transition only — no credit channel")
	assert.Equal(t, domain.ContractPendingValidation, persisted.State,
		"persisted snapshot is only superseded by a successful save")
}
