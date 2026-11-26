package application_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestListOfferFacets_ReturnsDistinctValuesForActiveSeason(t *testing.T) {
	// Arrange
	offers := &offerListReaderStub{offers: []domain.MissionOffer{
		{OfferID: "o1", Category: "RACING", Mode: domain.MissionModeTeam},
		{OfferID: "o2", Category: "BOUNTY", Mode: domain.MissionModeSolo},
		{OfferID: "o3", Category: "RACING", Mode: domain.MissionModeSolo},
	}}
	seasons := &activeSeasonsStub{seasons: []domain.Season{domain.NewSeason("season-2026-q4", domain.SeasonStateActive)}}
	handler := application.NewListOfferFacetsHandler(offers, seasons)

	// Act
	facets, err := handler.Handle(domain.OfferListFilter{})

	// Assert
	require.NoError(t, err)
	require.Equal(t, []string{"BOUNTY", "RACING"}, facets.Categories)
	require.Equal(t, []domain.MissionMode{domain.MissionModeSolo, domain.MissionModeTeam}, facets.Modes)
	require.Equal(t, "season-2026-q4", offers.gotFilter.SeasonID, "facets must resolve the active season like the list")
}

func TestListOfferFacets_FailsWithoutActiveSeason(t *testing.T) {
	// Arrange
	handler := application.NewListOfferFacetsHandler(&offerListReaderStub{}, &activeSeasonsStub{})

	// Act
	_, err := handler.Handle(domain.OfferListFilter{})

	// Assert
	require.ErrorIs(t, err, domain.ErrNoActiveSeason)
}
