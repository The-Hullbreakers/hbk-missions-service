package domain_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// A static policy answers itself through the reader port.
func TestActiveLimitPolicy_SelfReader(t *testing.T) {
	// Arrange
	policy := domain.NewActiveLimitPolicy(3, map[string]int{"BOUNTY": 2})

	// Act
	got, err := policy.GetActiveLimit()

	// Assert
	require.NoError(t, err)
	require.False(t, got.Default() != 3 || got.EffectiveLimit("BOUNTY") != 2)
}

// Overrides returns a defensive copy — mutating it does not leak back.
func TestActiveLimitPolicy_OverridesAreDefensive(t *testing.T) {
	// Arrange
	policy := domain.NewActiveLimitPolicy(5, map[string]int{"BOUNTY": 2})

	// Act
	overrides := policy.Overrides()
	overrides["BOUNTY"] = 99

	// Assert
	require.Equal(t, 2, policy.EffectiveLimit("BOUNTY"))
}

// Phased terms reject empty phase lists.
func TestNewContractTermsPhased_EmptyRejected(t *testing.T) {
	// Arrange / Act
	_, err := domain.NewContractTermsPhased(nil)

	// Assert
	require.ErrorIs(t, err, domain.ErrContractTermsInvalid)
}

// Phased terms reject a duplicated objective across phases.
func TestNewContractTermsPhased_DuplicateObjectiveRejected(t *testing.T) {
	// Arrange
	obj := domain.NewObjectiveNamed("obj-a", "a", "", "BOUNTY", 10, 0, 0)

	// Act
	_, err := domain.NewContractTermsPhased([]domain.Phase{
		{Mandatory: []domain.Objective{obj}},
		{Optional: []domain.Objective{obj}},
	})

	// Assert
	require.ErrorIs(t, err, domain.ErrContractTermsInvalid)
}

// Phased terms reject a malformed objective inside a phase.
func TestNewContractTermsPhased_InvalidObjectiveRejected(t *testing.T) {
	// Arrange — missing objective id
	bad := domain.NewObjectiveNamed("", "broken", "", "BOUNTY", 10, 0, 0)

	// Act
	_, err := domain.NewContractTermsPhased([]domain.Phase{
		{Mandatory: []domain.Objective{bad}},
	})

	// Assert
	require.ErrorIs(t, err, domain.ErrContractTermsInvalid)
}

// ObjectivePhase reports -1 for unknown objectives and PhaseComplete is false
// out of range.
func TestContractTerms_PhaseHelpersEdgeCases(t *testing.T) {
	// Arrange
	obj := domain.NewObjectiveNamed("obj-a", "a", "", "BOUNTY", 10, 0, 0)
	terms, err := domain.NewContractTermsPhased([]domain.Phase{
		{Mandatory: []domain.Objective{obj}},
	})
	require.NoError(t, err)

	// Act / Assert
	require.Equal(t, -1, terms.ObjectivePhase("missing"))
	if terms.PhaseComplete(9, func(string) bool { return true }) {
		require.Failf(t, "unexpected failure", "expected out-of-range phase to be incomplete")
	}
	require.Equal(t, 1, terms.PhaseCount())
}
