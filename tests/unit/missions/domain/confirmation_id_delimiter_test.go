package domain_test

import (
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	"github.com/stretchr/testify/require"
)

// BUG-0021-05: contract ids contain colons, so the serialized identity must
// use a delimiter that cannot collide.
func TestConfirmationID_RoundTripsColonContractID(t *testing.T) {
	id, err := domain.NewConfirmationID("off-1:player-1:1", "obj-a", "unit-1", "player-1")
	require.NoError(t, err)

	parsed, err := domain.ParseConfirmationID(id.String())

	require.NoError(t, err)
	require.Equal(t, "off-1:player-1:1", parsed.ContractID())
	require.Equal(t, "obj-a", parsed.ObjectiveID())
	require.Equal(t, "unit-1", parsed.UnitID())
	require.Equal(t, "player-1", parsed.RecipientID())
}
