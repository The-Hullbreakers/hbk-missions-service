package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func namedObjective() domain.Objective {
	return domain.NewObjectiveNamed("obj-a", "Objetivo", "", "BOUNTY_HUNTING", 50, 0, 0)
}

func TestMissionDefinition_StoresRules(t *testing.T) {
	rules := []string{"Solo PU/LIVE", "Video continuo sin cortes"}

	def, err := domain.NewMissionDefinition(
		"rescue-siege", "Rescate en el asedio", "Extrae al piloto",
		"BOUNTY_HUNTING", domain.MissionModeSolo, 200,
		[]domain.Objective{namedObjective()}, rules,
	)

	require.NoError(t, err)
	require.Equal(t, rules, def.Rules)
}

func TestMissionDefinition_AllowsMissingRules(t *testing.T) {
	def, err := domain.NewMissionDefinition(
		"rescue-siege", "Rescate en el asedio", "Extrae al piloto",
		"BOUNTY_HUNTING", domain.MissionModeSolo, 200,
		[]domain.Objective{namedObjective()}, nil,
	)

	require.NoError(t, err)
	require.Empty(t, def.Rules)
}

func TestMissionDefinition_RejectsBlankRule(t *testing.T) {
	_, err := domain.NewMissionDefinition(
		"rescue-siege", "Rescate en el asedio", "Extrae al piloto",
		"BOUNTY_HUNTING", domain.MissionModeSolo, 200,
		[]domain.Objective{namedObjective()}, []string{"Solo PU/LIVE", "  "},
	)

	require.ErrorIs(t, err, domain.ErrDefinitionInvalid)
}

func TestMissionDefinition_StoresMarkdownDescriptionVerbatim(t *testing.T) {
	description := "## Requisitos\n\n- Solo PU/LIVE\n- Video continuo"

	def, err := domain.NewMissionDefinition(
		"rescue-siege", "Rescate en el asedio", description,
		"BOUNTY_HUNTING", domain.MissionModeSolo, 200,
		[]domain.Objective{namedObjective()}, nil,
	)

	require.NoError(t, err)
	require.Equal(t, description, def.Description)
}

func TestMissionOfferFromDefinition_FreezesRules(t *testing.T) {
	// Arrange
	rules := []string{"Solo PU/LIVE", "Video continuo sin cortes"}
	definition, err := domain.NewMissionDefinition(
		"pvp-live-air-10", "Cazador de escuadrones", "10 derribos",
		"PVP_AIR", domain.MissionModeSolo, 0,
		[]domain.Objective{domain.NewObjectiveNamed("pvp-kills", "Derribos", "", "PVP_AIR", 270, 0, 1)},
		rules,
	)
	require.NoError(t, err)
	definition.ContentEn = domain.LocalizedContent{
		Title: "Squad hunter", Description: "10 takedowns",
		Rules: []string{"PU/LIVE only", "Continuous uncut video"},
		Objectives: []domain.LocalizedObjective{
			{ObjectiveID: "pvp-kills", Name: "Takedowns"},
		},
	}
	_ = definition.Publish("v1")

	// Act
	offer, err := domain.NewMissionOfferFromDefinition("offer-1", "season-1", definition)

	// Assert
	require.NoError(t, err)
	require.Equal(t, rules, offer.Rules)
}
