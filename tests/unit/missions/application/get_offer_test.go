package application_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

type offerReaderStub struct {
	offer domain.MissionOffer
	err   error
	gotID string
}

func (s *offerReaderStub) Get(offerID string) (domain.MissionOffer, error) {
	s.gotID = offerID
	return s.offer, s.err
}

func TestGetOffer_ReturnsOffer(t *testing.T) {
	// Arrange
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	store := &offerReaderStub{offer: offer}
	handler := application.NewGetOfferHandler(store)

	// Act
	got, err := handler.Handle("offer-1", domain.LanguageSpanish)

	// Assert
	require.NoError(t, err)
	require.Equal(t, "offer-1", got.OfferID)
	require.Equal(t, "offer-1", store.gotID)
}

func TestGetOffer_MissingPropagatesErrOfferNotAvailable(t *testing.T) {
	// Arrange
	store := &offerReaderStub{err: domain.ErrOfferNotAvailable}
	handler := application.NewGetOfferHandler(store)

	// Act
	_, err := handler.Handle("offer-x", domain.LanguageSpanish)

	// Assert
	require.ErrorIs(t, err, domain.ErrOfferNotAvailable)
}
