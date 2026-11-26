package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func draftOfferDefinition(t *testing.T) domain.MissionDefinition {
	t.Helper()
	objective := domain.NewObjectiveNamed("obj-1", "Eliminate target", "", "BOUNTY_HUNTING", 10, 0, 0)
	definition, err := domain.NewMissionDefinition("def-x", "Bounty", "desc", "BOUNTY_HUNTING", domain.MissionModeSolo, 100, []domain.Objective{objective}, nil)
	require.NoError(t, err)
	return definition
}

func publishedOfferDefinition(t *testing.T) domain.MissionDefinition {
	t.Helper()
	definition := draftOfferDefinition(t)
	definition.ContentEn = domain.LocalizedContent{
		Title: "Bounty", Description: "desc",
		Objectives: []domain.LocalizedObjective{
			{ObjectiveID: "obj-1", Name: "Eliminate target"},
		},
	}
	require.NoError(t, definition.Publish("v1"))
	return definition
}

func TestNewMissionOfferFromDefinition_RejectsEmptyOfferID(t *testing.T) {
	// Arrange
	definition := publishedOfferDefinition(t)

	// Act
	_, err := domain.NewMissionOfferFromDefinition("", "season-1", definition)

	// Assert
	require.ErrorIs(t, err, domain.ErrOfferInvalid)
}

func TestNewMissionOfferFromDefinition_RejectsEmptySeasonID(t *testing.T) {
	// Arrange
	definition := publishedOfferDefinition(t)

	// Act
	_, err := domain.NewMissionOfferFromDefinition("offer-1", "", definition)

	// Assert
	require.ErrorIs(t, err, domain.ErrOfferInvalid)
}

func TestNewMissionOfferFromDefinition_RejectsEmptyDefinitionCode(t *testing.T) {
	// Arrange
	definition := publishedOfferDefinition(t)
	definition.DefinitionCode = ""

	// Act
	_, err := domain.NewMissionOfferFromDefinition("offer-1", "season-1", definition)

	// Assert
	require.ErrorIs(t, err, domain.ErrOfferInvalid)
}

func TestNewMissionOfferFromDefinition_RejectsDraftDefinition(t *testing.T) {
	// Arrange
	definition := draftOfferDefinition(t)

	// Act
	_, err := domain.NewMissionOfferFromDefinition("offer-1", "season-1", definition)

	// Assert
	require.ErrorIs(t, err, domain.ErrDefinitionNotPublished)
}

func TestNewMissionOfferFromDefinition_RejectsRetiredDefinition(t *testing.T) {
	// Arrange
	definition := publishedOfferDefinition(t)
	require.NoError(t, definition.Retire())

	// Act
	_, err := domain.NewMissionOfferFromDefinition("offer-1", "season-1", definition)

	// Assert
	require.ErrorIs(t, err, domain.ErrDefinitionNotPublished)
}

func TestNewMissionOfferFromDefinition_FreezesPublishedContent(t *testing.T) {
	// Arrange
	definition := publishedOfferDefinition(t)

	// Act
	offer, err := domain.NewMissionOfferFromDefinition("offer-1", "season-1", definition)

	// Assert
	require.NoError(t, err)
	require.Equal(t, "offer-1", offer.OfferID)
	require.Equal(t, "season-1", offer.SeasonID)
	require.Equal(t, "def-x", offer.DefinitionCode)
	require.Equal(t, "v1", offer.DefinitionVersion)
	require.Equal(t, definition.Title, offer.Title)
	require.Equal(t, definition.Mode, offer.Mode)
	require.Equal(t, definition.CompletionReward, offer.CompletionReward)
}
