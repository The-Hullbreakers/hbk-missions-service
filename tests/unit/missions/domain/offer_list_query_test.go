package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestParseOfferListFilter_ParsesPagingBounds(t *testing.T) {
	// Arrange
	raw := domain.OfferListQuery{Limit: "10", Offset: "20"}

	// Act
	filter, err := domain.ParseOfferListFilter(raw)

	// Assert
	require.NoError(t, err)
	require.Equal(t, 10, filter.Page.Limit)
	require.Equal(t, 20, filter.Page.Offset)
}

func TestParseOfferListFilter_ClampsLimitAboveMax(t *testing.T) {
	// Arrange
	raw := domain.OfferListQuery{Limit: "9999"}

	// Act
	filter, err := domain.ParseOfferListFilter(raw)

	// Assert
	require.NoError(t, err)
	require.Equal(t, domain.OfferListMaxLimit, filter.Page.Limit)
}

func TestParseOfferListFilter_RejectsInvalidPaging(t *testing.T) {
	// Arrange + Act + Assert
	for _, raw := range []domain.OfferListQuery{
		{Limit: "abc"}, {Limit: "0"}, {Limit: "-5"}, {Offset: "-1"}, {Offset: "x"},
	} {
		_, err := domain.ParseOfferListFilter(raw)
		require.ErrorIs(t, err, domain.ErrInvalidOfferListFilter, "raw: %+v", raw)
	}
}

func TestParseOfferListFilter_ParsesSortFieldAndDirection(t *testing.T) {
	// Arrange
	raw := domain.OfferListQuery{Sort: "-completion_reward"}

	// Act
	filter, err := domain.ParseOfferListFilter(raw)

	// Assert
	require.NoError(t, err)
	require.Equal(t, domain.OfferSort{Field: domain.OfferSortCompletionReward, Descending: true}, filter.Sort)
}

func TestParseOfferListFilter_RejectsUnknownSortField(t *testing.T) {
	// Arrange
	raw := domain.OfferListQuery{Sort: "price"}

	// Act
	_, err := domain.ParseOfferListFilter(raw)

	// Assert
	require.ErrorIs(t, err, domain.ErrInvalidOfferListFilter)
}
