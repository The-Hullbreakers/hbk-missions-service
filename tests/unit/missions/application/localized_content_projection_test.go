package application_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// TEST-APP-LANG-001 (GetOffer partition, whole-offer integrity): an English
// objective translation without a name is incomplete and falls back
// atomically to Spanish.
func TestGetOfferLocalized_MissingObjectiveNameFallsBack(t *testing.T) {
	// Arrange
	offer := integrityOfferFixture()
	contentEn := completeEnglish()
	contentEn.Objectives[0].Name = ""
	offer.ContentEn = contentEn
	handler := application.NewGetOfferHandler(&localizedOfferReaderStub{offer: offer})

	// Act
	served, err := handler.Handle("offer-int", domain.LanguageEnglish)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "Cacería", served.Title)
	assert.Equal(t, "Capturar", served.Objectives[0].Name)
	assert.Equal(t, domain.LanguageSpanish, served.ContentLanguage)
}

// TEST-APP-LANG-001 (GetOffer partition, whole-offer integrity): duplicate
// English ObjectiveIDs cannot be matched uniquely and fall back to Spanish.
func TestGetOfferLocalized_DuplicateEnglishObjectiveIDsFallBack(t *testing.T) {
	// Arrange
	offer := integrityOfferFixture()
	offer.Objectives = append(offer.Objectives,
		domain.NewObjectiveNamed("obj-2", "Segundo", "Segundo objetivo", "BOUNTY", 20, 0, 1))
	contentEn := completeEnglish()
	contentEn.Objectives = append(contentEn.Objectives,
		domain.LocalizedObjective{ObjectiveID: "obj-1", Name: "Again", Description: "Duplicate ID"})
	offer.ContentEn = contentEn
	handler := application.NewGetOfferHandler(&localizedOfferReaderStub{offer: offer})

	// Act
	served, err := handler.Handle("offer-int", domain.LanguageEnglish)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "Cacería", served.Title)
	assert.Equal(t, "Capturar", served.Objectives[0].Name)
	assert.Equal(t, "Segundo", served.Objectives[1].Name)
	assert.Equal(t, domain.LanguageSpanish, served.ContentLanguage)
}

// TEST-APP-LANG-001 (GetOffer partition, whole-offer integrity): a
// whitespace-only English title or description counts as missing and falls
// back atomically to Spanish.
func TestGetOfferLocalized_WhitespaceEnglishFallsBack(t *testing.T) {
	tests := []struct {
		name     string
		mutateEn func(domain.LocalizedContent) domain.LocalizedContent
	}{
		{
			name: "whitespace title",
			mutateEn: func(en domain.LocalizedContent) domain.LocalizedContent {
				en.Title = "   "
				return en
			},
		},
		{
			name: "whitespace description",
			mutateEn: func(en domain.LocalizedContent) domain.LocalizedContent {
				en.Description = "\t\n "
				return en
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			offer := integrityOfferFixture()
			offer.ContentEn = tc.mutateEn(completeEnglish())
			handler := application.NewGetOfferHandler(&localizedOfferReaderStub{offer: offer})

			// Act
			served, err := handler.Handle("offer-int", domain.LanguageEnglish)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, "Cacería", served.Title)
			assert.Equal(t, "Descripción", served.Description)
			assert.Equal(t, domain.LanguageSpanish, served.ContentLanguage)
		})
	}
}

// TEST-APP-LANG-001 (GetOffer partition, whole-offer integrity): the served
// projection copies rules and objectives — mutating it must not rewrite the
// stored offer's Spanish or English content.
func TestGetOfferLocalized_ProjectionDoesNotAliasSource(t *testing.T) {
	// Arrange
	offer := integrityOfferFixture()
	offer.ContentEn = completeEnglish()
	handler := application.NewGetOfferHandler(&localizedOfferReaderStub{offer: offer})

	// Act
	served, err := handler.Handle("offer-int", domain.LanguageEnglish)

	// Assert
	require.NoError(t, err)
	served.Rules[0] = "Mutated served rule"
	served.Objectives[0].Name = "Mutated served name"
	assert.Equal(t, []string{"Regla uno"}, offer.Rules)
	assert.Equal(t, "Capturar", offer.Objectives[0].Name)
	assert.Equal(t, []string{"First rule"}, offer.ContentEn.Rules)
	assert.Equal(t, "Capture", offer.ContentEn.Objectives[0].Name)
}
