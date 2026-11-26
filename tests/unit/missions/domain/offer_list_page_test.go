package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func offerSet() []domain.MissionOffer {
	return []domain.MissionOffer{
		{OfferID: "o2", Category: "RACING", CompletionReward: 500, Mode: domain.MissionModeSolo},
		{OfferID: "o1", Category: "BOUNTY", CompletionReward: 100, Mode: domain.MissionModeTeam},
		{OfferID: "o3", Category: "BOUNTY", CompletionReward: 300, Mode: domain.MissionModeTeam},
	}
}

func TestPaginateOffers_SlicesAndReportsTotal(t *testing.T) {
	// Arrange
	offers := offerSet()
	page := domain.OfferPage{Limit: 2, Offset: 1}

	// Act
	got := domain.PaginateOffers(offers, page)

	// Assert
	require.Equal(t, 3, got.Total)
	require.Equal(t, 2, got.Limit)
	require.Equal(t, 1, got.Offset)
	require.Len(t, got.Offers, 2)
	require.Equal(t, "o1", got.Offers[0].OfferID)
}

func TestPaginateOffers_OffsetBeyondTotalReturnsEmpty(t *testing.T) {
	// Arrange
	offers := offerSet()

	// Act
	got := domain.PaginateOffers(offers, domain.OfferPage{Limit: 10, Offset: 50})

	// Assert
	require.Equal(t, 3, got.Total)
	require.Empty(t, got.Offers)
}

func TestSortOffers_OrdersByCompletionRewardDescending(t *testing.T) {
	// Arrange
	offers := offerSet()

	// Act
	got := domain.SortOffers(offers, domain.OfferSort{Field: domain.OfferSortCompletionReward, Descending: true})

	// Assert
	require.Equal(t, []string{"o2", "o3", "o1"}, []string{got[0].OfferID, got[1].OfferID, got[2].OfferID})
}

func TestSortOffers_OrdersByCategoryWithStableTieBreak(t *testing.T) {
	// Arrange
	offers := offerSet()

	// Act
	got := domain.SortOffers(offers, domain.OfferSort{Field: domain.OfferSortCategory})

	// Assert
	require.Equal(t, []string{"o1", "o3", "o2"}, []string{got[0].OfferID, got[1].OfferID, got[2].OfferID})
}

func TestSortOffers_EmptySortKeepsOfferIDOrder(t *testing.T) {
	// Arrange
	offers := offerSet()

	// Act
	got := domain.SortOffers(offers, domain.OfferSort{})

	// Assert
	require.Equal(t, []string{"o1", "o2", "o3"}, []string{got[0].OfferID, got[1].OfferID, got[2].OfferID})
}

func TestCollectFacets_ReturnsDistinctSortedValues(t *testing.T) {
	// Arrange
	offers := offerSet()

	// Act
	facets := domain.CollectFacets(offers)

	// Assert
	require.Equal(t, []string{"BOUNTY", "RACING"}, facets.Categories)
	require.Equal(t, []domain.MissionMode{domain.MissionModeSolo, domain.MissionModeTeam}, facets.Modes)
}
