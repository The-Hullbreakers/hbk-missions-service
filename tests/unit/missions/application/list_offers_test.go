package application_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestListOffers_ReturnsOffersForActiveSeason(t *testing.T) {
	// Arrange
	offer := domain.NewMissionOffer("offer-1", "BOUNTY", "season-2026-q4")
	offers := &offerListReaderStub{offers: []domain.MissionOffer{offer}}
	contracts := &boardContractsStub{}
	seasons := &activeSeasonsStub{seasons: []domain.Season{domain.NewSeason("season-2026-q4", domain.SeasonStateActive)}}
	handler := application.NewListOffersHandler(offers, contracts, seasons)
	filter := domain.OfferListFilter{Status: domain.OfferStatusAvailable}

	// Act
	got, err := handler.Handle("player-1", filter, domain.LanguageSpanish)

	// Assert
	require.NoError(t, err)
	require.False(t, len(got.Offers) != 1 || got.Offers[0].OfferID != "offer-1")
	require.Equal(t, "season-2026-q4", offers.gotFilter.SeasonID)
}

func TestListOffers_KeepsExplicitSeason(t *testing.T) {
	// Arrange
	offers := &offerListReaderStub{}
	contracts := &boardContractsStub{}
	seasons := &activeSeasonsStub{}
	handler := application.NewListOffersHandler(offers, contracts, seasons)
	filter := domain.OfferListFilter{SeasonID: "season-2026-q3", Status: domain.OfferStatusAll}

	// Act
	_, err := handler.Handle("player-1", filter, domain.LanguageSpanish)

	// Assert
	require.NoError(t, err)
	require.Equal(t, "season-2026-q3", offers.gotFilter.SeasonID)
}

func TestListOffers_FailsWithoutActiveSeasonWhenSeasonOmitted(t *testing.T) {
	// Arrange
	handler := application.NewListOffersHandler(&offerListReaderStub{}, &boardContractsStub{}, &activeSeasonsStub{})
	filter := domain.OfferListFilter{Status: domain.OfferStatusAvailable}

	// Act
	_, err := handler.Handle("player-1", filter, domain.LanguageSpanish)

	// Assert
	require.ErrorIs(t, err, domain.ErrNoActiveSeason)
}

func TestListOffers_PropagatesReaderErrors(t *testing.T) {
	// Arrange
	boom := errors.New("store unavailable")
	seasons := &activeSeasonsStub{seasons: []domain.Season{domain.NewSeason("season-2026-q4", domain.SeasonStateActive)}}

	// Act + Assert: offer reader error
	offersErr := application.NewListOffersHandler(&offerListReaderStub{err: boom}, &boardContractsStub{}, seasons)
	_, err := offersErr.Handle("player-1", domain.OfferListFilter{}, domain.LanguageSpanish)
	require.ErrorIs(t, err, boom)

	// Act + Assert: contracts reader error
	contractsErr := application.NewListOffersHandler(&offerListReaderStub{}, &boardContractsStub{err: boom}, seasons)
	_, err = contractsErr.Handle("player-1", domain.OfferListFilter{}, domain.LanguageSpanish)
	require.ErrorIs(t, err, boom)

	// Act + Assert: seasons reader error
	seasonsErr := application.NewListOffersHandler(&offerListReaderStub{}, &boardContractsStub{}, &activeSeasonsStub{err: boom})
	_, err = seasonsErr.Handle("player-1", domain.OfferListFilter{}, domain.LanguageSpanish)
	require.ErrorIs(t, err, boom)
}

func TestListOffers_RejectsMultipleActiveSeasons(t *testing.T) {
	// Arrange
	seasons := &activeSeasonsStub{seasons: []domain.Season{
		domain.NewSeason("s1", domain.SeasonStateActive),
		domain.NewSeason("s2", domain.SeasonStateActive),
	}}
	handler := application.NewListOffersHandler(&offerListReaderStub{}, &boardContractsStub{}, seasons)

	// Act
	_, err := handler.Handle("player-1", domain.OfferListFilter{}, domain.LanguageSpanish)

	// Assert
	require.ErrorIs(t, err, domain.ErrActiveSeasonAlreadyExists)
}
