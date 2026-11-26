package application_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// TEST-APP-LANG-001 (ListOffers partition): the effective-language `q`
// filter runs before pagination — Total counts the filtered population and
// the page slices inside it. Offer-1 carries the term only in its Spanish
// base so a Spanish prefilter would wrongly include it.
func TestListOffersLocalized_QueryFiltersBeforePagination(t *testing.T) {
	// Arrange
	reader := &localizedOfferListReader{offers: []domain.MissionOffer{
		bilingualListOffer("offer-1", "Bounty vieja", "Old hunt"),
		bilingualListOffer("offer-2", "Cacería dos", "Bounty hunt two"),
		bilingualListOffer("offer-3", "Cacería tres", "Bounty hunt three"),
	}}
	handler := application.NewListOffersHandler(reader, &boardContractsStub{}, &activeSeasonsStub{})
	filter := domain.OfferListFilter{SeasonID: "season-1", Q: "Bounty", Page: domain.OfferPage{Limit: 1, Offset: 1}}

	// Act
	page, err := handler.Handle("player-1", filter, domain.LanguageEnglish)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "", reader.gotFilter.Q)
	assert.Equal(t, 2, page.Total)
	require.Len(t, page.Offers, 1)
	assert.Equal(t, "offer-3", page.Offers[0].OfferID)
	assert.Equal(t, "Bounty hunt three", page.Offers[0].Title)
	assert.Equal(t, domain.LanguageEnglish, page.Offers[0].ContentLanguage)
}
