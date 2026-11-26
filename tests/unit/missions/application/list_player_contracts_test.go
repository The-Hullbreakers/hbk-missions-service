package application_test

import (
	"errors"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestListPlayerContracts_ReturnsPlayerContracts(t *testing.T) {
	// Arrange
	offer := domain.NewMissionOffer("offer-1", "BOUNTY", "season-2026-q4")
	mine := domain.NewMissionContract("c1", offer, []string{"player-1"})
	other := domain.NewMissionContract("c2", offer, []string{"player-2"})
	reader := &playerContractsReaderStub{contracts: []domain.MissionContract{*mine, *other}}
	handler := application.NewListPlayerContractsHandler(reader)
	filter := domain.ContractListFilter{Status: domain.ContractStatusAll}

	// Act
	got, err := handler.Handle("player-1", filter, domain.LanguageSpanish)

	// Assert
	require.NoError(t, err)
	require.False(t, len(got) != 1 || got[0].ContractID != "c1")
	require.Equal(t, "player-1", reader.gotPlayer)
}

func TestListPlayerContracts_PropagatesReaderError(t *testing.T) {
	// Arrange
	boom := errors.New("store unavailable")
	handler := application.NewListPlayerContractsHandler(&playerContractsReaderStub{err: boom})

	// Act
	_, err := handler.Handle("player-1", domain.ContractListFilter{}, domain.LanguageSpanish)

	// Assert
	require.ErrorIs(t, err, boom)
}
