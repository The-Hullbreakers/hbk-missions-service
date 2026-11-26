package domain_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestParseContractListFilter_DefaultsToAllSeasonsAndStatuses(t *testing.T) {
	// Arrange
	raw := domain.ContractListQuery{}

	// Act
	filter, err := domain.ParseContractListFilter(raw)

	// Assert
	require.NoError(t, err)
	require.Equal(t, domain.ContractStatusAll, filter.Status)
}

func TestParseContractListFilter_AcceptsKnownValues(t *testing.T) {
	// Arrange
	raw := domain.ContractListQuery{SeasonID: "season-2026-q4", Status: "ACTIVE"}

	// Act
	filter, err := domain.ParseContractListFilter(raw)

	// Assert
	require.NoError(t, err)
	require.False(t, filter.SeasonID != "season-2026-q4" || filter.Status != domain.ContractStatusActive)
}

func TestParseContractListFilter_AcceptsExactState(t *testing.T) {
	// Arrange
	raw := domain.ContractListQuery{Status: "COMPLETED"}

	// Act
	filter, err := domain.ParseContractListFilter(raw)

	// Assert
	require.False(t, err != nil || filter.Status != domain.ContractStatusFilter(domain.ContractCompleted))
}

func TestParseContractListFilter_RejectsUnknownStatus(t *testing.T) {
	// Arrange
	raw := domain.ContractListQuery{Status: "LOST"}

	// Act
	_, err := domain.ParseContractListFilter(raw)

	// Assert
	require.ErrorIs(t, err, domain.ErrInvalidContractStatusFilter)
}

func TestContractListFilter_Match(t *testing.T) {
	// Arrange
	offer := domain.NewMissionOffer("offer-1", "BOUNTY", "season-2026-q4")
	active := domain.NewMissionContractFromState("c1", offer, []string{"p1"}, domain.ContractInProgress, true, "", "season-2026-q4", nil, nil, nil)
	completed := domain.NewMissionContractFromState("c2", offer, []string{"p1"}, domain.ContractCompleted, true, "", "season-2026-q3", nil, nil, nil)

	// Act + Assert
	activeFilter := domain.ContractListFilter{Status: domain.ContractStatusActive}
	require.False(t, !activeFilter.Match(*active))
	require.False(t, activeFilter.Match(*completed))

	seasonFilter := domain.ContractListFilter{SeasonID: "season-2026-q4", Status: domain.ContractStatusAll}
	require.False(t, !seasonFilter.Match(*active))
	require.False(t, seasonFilter.Match(*completed))

	exactFilter := domain.ContractListFilter{Status: domain.ContractStatusFilter(domain.ContractCompleted)}
	require.False(t, !exactFilter.Match(*completed))
	require.False(t, exactFilter.Match(*active))
}
