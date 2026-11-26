package domain_test

import (
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	"github.com/stretchr/testify/require"
)

// REQ-MB-013: a published definition declares its objectives and reward
// rules so acceptance can freeze them into contract terms.
func TestMissionDefinition_CarriesObjectives(t *testing.T) {
	definition, err := domain.NewMissionDefinition(
		"def-1", "Defensa de puesto", "Protege la posición",
		"BOUNTY", domain.MissionModeSolo, 0,
		[]domain.Objective{
			domain.NewObjectiveNamed("obj-kills", "Elimina atacantes", "", "BOUNTY", 10, 5, 100),
		}, nil,
	)
	require.NoError(t, err)

	require.Equal(t, 1, len(definition.Objectives))
	objective := definition.Objectives[0]
	require.Equal(t, "obj-kills", objective.ObjectiveID)
	require.Equal(t, 10, objective.MeritPerUnit)
	require.Equal(t, 5, objective.Bonus)
	require.Equal(t, 100, objective.UnitLimit)
}
