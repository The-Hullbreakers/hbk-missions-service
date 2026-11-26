package application_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func pagedBoard(t *testing.T, offers []domain.MissionOffer) *application.ListOffersHandler {
	t.Helper()
	seasons := &activeSeasonsStub{seasons: []domain.Season{domain.NewSeason("season-2026-q4", domain.SeasonStateActive)}}
	return application.NewListOffersHandler(&offerListReaderStub{offers: offers}, &boardContractsStub{}, seasons)
}

func TestListOffers_ReturnsPaginatedEnvelope(t *testing.T) {
	// Arrange
	offers := []domain.MissionOffer{
		{OfferID: "o1", SeasonID: "season-2026-q4"},
		{OfferID: "o2", SeasonID: "season-2026-q4"},
		{OfferID: "o3", SeasonID: "season-2026-q4"},
	}
	handler := pagedBoard(t, offers)
	filter := domain.OfferListFilter{Page: domain.OfferPage{Limit: 2, Offset: 2}}

	// Act
	got, err := handler.Handle("player-1", filter, domain.LanguageSpanish)

	// Assert
	require.NoError(t, err)
	require.Equal(t, 3, got.Total)
	require.Equal(t, 2, got.Limit)
	require.Equal(t, 2, got.Offset)
	require.Len(t, got.Offers, 1)
	require.Equal(t, "o3", got.Offers[0].OfferID)
}

func TestListOffers_TotalCountsAfterVisibilityExclusion(t *testing.T) {
	// Arrange
	hidden := domain.NewMissionOffer("offer-hidden", "BOUNTY", "season-2026-q4")
	hidden.DefinitionCode = "def-1"
	visible := domain.NewMissionOffer("offer-visible", "BOUNTY", "season-2026-q4")
	visible.DefinitionCode = "def-2"
	contract := domain.NewMissionContractFromState("c1", hidden, []string{"player-1"}, domain.ContractInProgress, true, "", "season-2026-q4", nil, nil, nil)
	seasons := &activeSeasonsStub{seasons: []domain.Season{domain.NewSeason("season-2026-q4", domain.SeasonStateActive)}}
	handler := application.NewListOffersHandler(
		&offerListReaderStub{offers: []domain.MissionOffer{hidden, visible}},
		&boardContractsStub{contracts: []domain.MissionContract{*contract}},
		seasons,
	)
	filter := domain.OfferListFilter{Page: domain.OfferPage{Limit: 50}}

	// Act
	got, err := handler.Handle("player-1", filter, domain.LanguageSpanish)

	// Assert
	require.NoError(t, err)
	require.Equal(t, 1, got.Total, "total must count visible offers only")
	require.Len(t, got.Offers, 1)
}

func TestListOffers_OfferCarriesObjectivesAndMeritSummary(t *testing.T) {
	// Arrange
	offer := domain.NewMissionOffer("offer-1", "PVP_AIR", "season-2026-q4")
	offer.DefinitionCode = "pvp-live-air-10"
	offer.Objectives = []domain.Objective{
		domain.NewObjectiveNamed("pvp-kills", "Derribos verificados", "", "PVP_AIR", 270, 0, 1),
		domain.NewObjectiveNamed("evidence-review", "Revisión de evidencias", "", "PVP_AIR", 30, 0, 1),
	}
	handler := pagedBoard(t, []domain.MissionOffer{offer})
	filter := domain.OfferListFilter{Page: domain.OfferPage{Limit: 50}}

	// Act
	got, err := handler.Handle("player-1", filter, domain.LanguageSpanish)

	// Assert
	require.NoError(t, err)
	require.Len(t, got.Offers, 1)
	require.Len(t, got.Offers[0].Objectives, 2)
	summary := got.Offers[0].MeritSummary()
	require.Equal(t, 300, summary.ObjectiveMerits)
	require.Equal(t, 300, summary.TotalMerits)
}

func TestListOffers_SortsBeforePaginating(t *testing.T) {
	// Arrange
	offers := []domain.MissionOffer{
		{OfferID: "o1", CompletionReward: 100, SeasonID: "season-2026-q4"},
		{OfferID: "o2", CompletionReward: 500, SeasonID: "season-2026-q4"},
		{OfferID: "o3", CompletionReward: 300, SeasonID: "season-2026-q4"},
	}
	handler := pagedBoard(t, offers)
	filter := domain.OfferListFilter{
		Page: domain.OfferPage{Limit: 1},
		Sort: domain.OfferSort{Field: domain.OfferSortCompletionReward, Descending: true},
	}

	// Act
	got, err := handler.Handle("player-1", filter, domain.LanguageSpanish)

	// Assert
	require.NoError(t, err)
	require.Equal(t, "o2", got.Offers[0].OfferID)
}
