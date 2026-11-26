package application_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// policyWriterStub is a hand-written ActiveLimitPolicyWriter double.
type policyWriterStub struct {
	saved domain.ActiveLimitPolicy
	err   error
}

func (s *policyWriterStub) SaveActiveLimit(policy domain.ActiveLimitPolicy) error {
	s.saved = policy
	return s.err
}

func TestUpdateActiveLimit_PersistsPolicy(t *testing.T) {
	// Arrange
	writer := &policyWriterStub{}
	handler := application.NewUpdateActiveLimitHandler(writer)

	// Act
	err := handler.Handle(3, map[string]int{"BOUNTY_HUNTING": 2})

	// Assert
	require.NoError(t, err)
	require.False(t, writer.saved.Default() != 3 || writer.saved.EffectiveLimit("BOUNTY_HUNTING") != 2)
}

func TestUpdateActiveLimit_NoOverridesUsesDefaultEverywhere(t *testing.T) {
	// Arrange
	writer := &policyWriterStub{}
	handler := application.NewUpdateActiveLimitHandler(writer)

	// Act
	err := handler.Handle(4, nil)

	// Assert
	require.NoError(t, err)
	require.Equal(t, 4, writer.saved.EffectiveLimit("EXPLORATION"))
}

func TestUpdateActiveLimit_NonPositiveRejected(t *testing.T) {
	// Arrange
	handler := application.NewUpdateActiveLimitHandler(&policyWriterStub{})

	// Act
	err := handler.Handle(0, nil)

	// Assert
	require.ErrorIs(t, err, domain.ErrInvalidPolicy)
}

func TestUpdateActiveLimit_NonPositiveOverrideRejected(t *testing.T) {
	// Arrange
	handler := application.NewUpdateActiveLimitHandler(&policyWriterStub{})

	// Act
	err := handler.Handle(3, map[string]int{"BOUNTY_HUNTING": -1})

	// Assert
	require.ErrorIs(t, err, domain.ErrInvalidPolicy)
}
