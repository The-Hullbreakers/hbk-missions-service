package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestParseOfferListFilter_DefaultsToAvailableAndActiveSeason(t *testing.T) {
	// Arrange
	raw := domain.OfferListQuery{}

	// Act
	filter, err := domain.ParseOfferListFilter(raw)

	// Assert
	require.NoError(t, err)
	require.Equal(t, domain.OfferStatusAvailable, filter.Status)
	require.Empty(t, filter.SeasonID)
	require.Equal(t, domain.OfferPage{Limit: domain.OfferListDefaultLimit, Offset: 0}, filter.Page)
	require.Empty(t, filter.Sort.Field)
}

func TestParseOfferListFilter_AcceptsKnownValues(t *testing.T) {
	// Arrange
	raw := domain.OfferListQuery{
		SeasonID: "season-2026-q4",
		Status:   "ALL",
		Category: "BOUNTY",
		Mode:     "TEAM",
	}

	// Act
	filter, err := domain.ParseOfferListFilter(raw)

	// Assert
	require.NoError(t, err)
	require.Equal(t, "season-2026-q4", filter.SeasonID)
	require.Equal(t, []string{"BOUNTY"}, filter.Categories)
	require.Equal(t, []domain.MissionMode{domain.MissionModeTeam}, filter.Modes)
}

func TestParseOfferListFilter_AcceptsMultiValueCategoryAndMode(t *testing.T) {
	// Arrange
	raw := domain.OfferListQuery{
		Category: "bounty, RACING ,",
		Mode:     "solo,team",
	}

	// Act
	filter, err := domain.ParseOfferListFilter(raw)

	// Assert
	require.NoError(t, err)
	require.Equal(t, []string{"BOUNTY", "RACING"}, filter.Categories)
	require.Equal(t, []domain.MissionMode{domain.MissionModeSolo, domain.MissionModeTeam}, filter.Modes)
}

func TestParseOfferListFilter_RejectsUnknownStatus(t *testing.T) {
	// Arrange
	raw := domain.OfferListQuery{Status: "CLOSED"}

	// Act
	_, err := domain.ParseOfferListFilter(raw)

	// Assert
	require.ErrorIs(t, err, domain.ErrInvalidOfferStatusFilter)
}

func TestParseOfferListFilter_RejectsUnknownModeInList(t *testing.T) {
	// Arrange
	raw := domain.OfferListQuery{Mode: "solo,arcade"}

	// Act
	_, err := domain.ParseOfferListFilter(raw)

	// Assert
	require.ErrorIs(t, err, domain.ErrInvalidOfferModeFilter)
}

func TestOfferListFilter_Match(t *testing.T) {
	// Arrange
	bounty := domain.NewMissionOffer("offer-1", "BOUNTY", "season-2026-q4")
	exploration := domain.NewMissionOfferWithMode("offer-2", "EXPLORATION", domain.MissionModeTeam, "season-2026-q3")
	filter := domain.OfferListFilter{
		SeasonID:   "season-2026-q4",
		Categories: []string{"BOUNTY"},
		Status:     domain.OfferStatusAvailable,
	}

	// Act + Assert
	require.True(t, filter.Match(bounty), "expected filter to match bounty offer of the same season")
	require.False(t, filter.Match(exploration), "expected filter not to match different category/season")

	modeFilter := domain.OfferListFilter{Modes: []domain.MissionMode{domain.MissionModeTeam}}
	require.True(t, modeFilter.Match(exploration), "expected mode filter to match team offer")
	require.False(t, modeFilter.Match(bounty), "expected mode filter not to match solo offer")
}

func TestOfferListFilter_Match_MultiValueIsOr(t *testing.T) {
	// Arrange
	bounty := domain.NewMissionOffer("offer-1", "BOUNTY", "s1")
	racing := domain.NewMissionOffer("offer-2", "RACING", "s1")
	escort := domain.NewMissionOffer("offer-3", "ESCORT", "s1")
	filter := domain.OfferListFilter{Categories: []string{"BOUNTY", "RACING"}}

	// Act + Assert
	require.True(t, filter.Match(bounty))
	require.True(t, filter.Match(racing))
	require.False(t, filter.Match(escort))
}

func TestOfferListFilter_Match_QueryIsCaseInsensitiveSubstring(t *testing.T) {
	// Arrange
	offer := domain.MissionOffer{OfferID: "o1", Title: "Night Drift", Description: "Street race"}
	other := domain.MissionOffer{OfferID: "o2", Title: "Harbor Escort", Description: "Guard ships"}
	byTitle := domain.OfferListFilter{Q: "drift"}
	byDescription := domain.OfferListFilter{Q: "SHIPS"}

	// Act + Assert
	require.True(t, byTitle.Match(offer))
	require.False(t, byTitle.Match(other))
	require.True(t, byDescription.Match(other))
}
