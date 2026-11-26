package application_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// localizedOfferListReader models the store: it applies the filter's
// structural Match — including Q whenever the handler forwards one — so the
// tests can prove the application layer keeps text search on the effective
// served language instead of the store's Spanish prefilter.
type localizedOfferListReader struct {
	offers    []domain.MissionOffer
	gotFilter domain.OfferListFilter
}

func (r *localizedOfferListReader) List(filter domain.OfferListFilter) ([]domain.MissionOffer, error) {
	r.gotFilter = filter
	var out []domain.MissionOffer
	for _, offer := range r.offers {
		if filter.Match(offer) {
			out = append(out, offer)
		}
	}
	return out, nil
}

func bilingualListOffer(offerID, esTitle, enTitle string) domain.MissionOffer {
	offer := domain.NewMissionOfferWithMode(offerID, "BOUNTY_HUNTING", domain.MissionModeSolo, "season-1")
	offer.Title = esTitle
	offer.Description = "Descripción"
	offer.CompletionReward = 100
	offer.Rules = []string{"Regla uno"}
	offer.Objectives = []domain.Objective{
		domain.NewObjectiveNamed("obj-1", "Capturar", "Objetivo en español", "BOUNTY", 10, 5, 3),
	}
	offer.ContentEn = domain.LocalizedContent{
		Title:       enTitle,
		Description: "English briefing",
		Rules:       []string{"First rule"},
		Objectives:  []domain.LocalizedObjective{{ObjectiveID: "obj-1", Name: "Capture", Description: "English objective"}},
	}
	return offer
}

func legacySpanishOffer(offerID string) domain.MissionOffer {
	offer := domain.NewMissionOfferWithMode(offerID, "BOUNTY_HUNTING", domain.MissionModeSolo, "season-1")
	offer.Title = "Misión legada"
	offer.Description = "Contenido solo en español"
	return offer
}

// TEST-APP-LANG-001 (ListOffers partition): a mixed catalog under lang=en
// serves each item in its effective language with an honest label while
// ids, economics, ordering and page shape stay unchanged.
func TestListOffersLocalized_ServesPerItemLanguage(t *testing.T) {
	// Arrange
	reader := &localizedOfferListReader{offers: []domain.MissionOffer{
		bilingualListOffer("offer-1", "Cacería", "Bounty hunt"),
		legacySpanishOffer("offer-2"),
	}}
	handler := application.NewListOffersHandler(reader, &boardContractsStub{}, &activeSeasonsStub{})
	filter := domain.OfferListFilter{SeasonID: "season-1", Page: domain.OfferPage{Limit: 50, Offset: 0}}

	// Act
	page, err := handler.Handle("player-1", filter, domain.LanguageEnglish)

	// Assert
	require.NoError(t, err)
	require.Len(t, page.Offers, 2)
	assert.Equal(t, "offer-1", page.Offers[0].OfferID)
	assert.Equal(t, "Bounty hunt", page.Offers[0].Title)
	assert.Equal(t, "English briefing", page.Offers[0].Description)
	assert.Equal(t, []string{"First rule"}, page.Offers[0].Rules)
	require.Len(t, page.Offers[0].Objectives, 1)
	assert.Equal(t, "obj-1", page.Offers[0].Objectives[0].ObjectiveID)
	assert.Equal(t, "Capture", page.Offers[0].Objectives[0].Name)
	assert.Equal(t, "English objective", page.Offers[0].Objectives[0].Description)
	assert.Equal(t, 10, page.Offers[0].Objectives[0].MeritPerUnit)
	assert.Equal(t, 5, page.Offers[0].Objectives[0].Bonus)
	assert.Equal(t, 3, page.Offers[0].Objectives[0].UnitLimit)
	assert.Equal(t, 100, page.Offers[0].CompletionReward)
	assert.Equal(t, domain.LanguageEnglish, page.Offers[0].ContentLanguage)
	assert.Equal(t, "offer-2", page.Offers[1].OfferID)
	assert.Equal(t, "Misión legada", page.Offers[1].Title)
	assert.Equal(t, domain.LanguageSpanish, page.Offers[1].ContentLanguage)
	assert.Equal(t, 2, page.Total)
	assert.Equal(t, 50, page.Limit)
	assert.Equal(t, 0, page.Offset)
}

// TEST-APP-LANG-001 (ListOffers partition): `q` matches the effective
// served language — after localization, before ordering and pagination —
// and is never forwarded to the store's Spanish prefilter.
func TestListOffersLocalized_QueryMatchesEffectiveLanguage(t *testing.T) {
	tests := []struct {
		name       string
		q          string
		expectedID string
	}{
		{"english term matches bilingual en title though es base does not", "Bounty", "offer-1"},
		{"spanish term matches es-only legacy fallback under en", "legada", "offer-2"},
		{"spanish base term does not match when en is served", "Cacería", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			reader := &localizedOfferListReader{offers: []domain.MissionOffer{
				bilingualListOffer("offer-1", "Cacería", "Bounty hunt"),
				legacySpanishOffer("offer-2"),
			}}
			handler := application.NewListOffersHandler(reader, &boardContractsStub{}, &activeSeasonsStub{})
			filter := domain.OfferListFilter{SeasonID: "season-1", Q: tc.q, Page: domain.OfferPage{Limit: 50, Offset: 0}}

			// Act
			page, err := handler.Handle("player-1", filter, domain.LanguageEnglish)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, "", reader.gotFilter.Q)
			if tc.expectedID == "" {
				assert.Equal(t, 0, page.Total)
				assert.Empty(t, page.Offers)
				return
			}
			assert.Equal(t, 1, page.Total)
			require.Len(t, page.Offers, 1)
			assert.Equal(t, tc.expectedID, page.Offers[0].OfferID)
		})
	}
}
