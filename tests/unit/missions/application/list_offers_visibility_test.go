package application_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestListOffers_HidesLegacyOffersWithoutDefinitionCode(t *testing.T) {
	// Arrange: legacy data uses OfferID as the definition identity.
	hidden := domain.NewMissionOffer("offer-hidden", "BOUNTY", "season-2026-q4")
	visible := domain.NewMissionOffer("offer-visible", "BOUNTY", "season-2026-q4")
	offers := &offerListReaderStub{offers: []domain.MissionOffer{hidden, visible}}
	activeContract := domain.NewMissionContractFromState("c1", hidden, []string{"player-1"}, domain.ContractInProgress, true, "", "season-2026-q4", nil, nil, nil)
	contracts := &boardContractsStub{contracts: []domain.MissionContract{*activeContract}}
	seasons := &activeSeasonsStub{seasons: []domain.Season{domain.NewSeason("season-2026-q4", domain.SeasonStateActive)}}
	handler := application.NewListOffersHandler(offers, contracts, seasons)

	// Act
	got, err := handler.Handle("player-1", domain.OfferListFilter{}, domain.LanguageSpanish)

	// Assert
	require.False(t, err != nil || len(got.Offers) != 1 || got.Offers[0].OfferID != "offer-visible")
}

// departedContract builds a real in-progress team contract and departs the
// leaver through the domain Leave API. It returns errors instead of
// asserting so a setup failure can never masquerade as the intended RED.
func departedContract(offer domain.MissionOffer, contractID string, roster []string, leaver string) (*domain.MissionContract, error) {
	contract := domain.NewMissionContractFromState(contractID, offer, roster, domain.ContractInProgress, true, "team-1", offer.SeasonID, nil, nil, nil)
	identity, err := domain.NewCallerIdentity(leaver, nil, nil)
	if err != nil {
		return nil, err
	}
	if err := contract.Leave(identity); err != nil {
		return nil, err
	}
	return contract, nil
}

func TestListOffers_DepartedMemberSeesTemplateAgain(t *testing.T) {
	// Arrange: player-1 left the team contract bound to def-x; roster history
	// keeps them, active membership does not. REG-0053-01.
	boardOffer := domain.NewMissionOffer("offer-x", "BOUNTY", "season-2026-q4")
	boardOffer.DefinitionCode = "def-x"
	otherOffer := domain.NewMissionOffer("offer-y", "BOUNTY", "season-2026-q4")
	otherOffer.DefinitionCode = "def-y"
	contract, err := departedContract(boardOffer, "contract-x", []string{"player-1", "player-2"}, "player-1")
	if err != nil {
		t.Fatalf("arrange departed contract: %v", err)
	}
	offers := &offerListReaderStub{offers: []domain.MissionOffer{boardOffer, otherOffer}}
	contracts := &boardContractsStub{contracts: []domain.MissionContract{*contract}}
	seasons := &activeSeasonsStub{seasons: []domain.Season{domain.NewSeason("season-2026-q4", domain.SeasonStateActive)}}
	handler := application.NewListOffersHandler(offers, contracts, seasons)

	// Act
	got, err := handler.Handle("player-1", domain.OfferListFilter{}, domain.LanguageSpanish)

	// Assert
	require.NoError(t, err)
	require.Len(t, got.Offers, 2)
	require.Equal(t, 1, contracts.activeCalls)
}

func TestListOffers_ActiveRosterMemberKeepsTemplateHidden(t *testing.T) {
	// Arrange: same departed contract, but player-2 is still an active member.
	boardOffer := domain.NewMissionOffer("offer-x", "BOUNTY", "season-2026-q4")
	boardOffer.DefinitionCode = "def-x"
	otherOffer := domain.NewMissionOffer("offer-y", "BOUNTY", "season-2026-q4")
	otherOffer.DefinitionCode = "def-y"
	contract, err := departedContract(boardOffer, "contract-x", []string{"player-1", "player-2"}, "player-1")
	if err != nil {
		t.Fatalf("arrange departed contract: %v", err)
	}
	offers := &offerListReaderStub{offers: []domain.MissionOffer{boardOffer, otherOffer}}
	contracts := &boardContractsStub{contracts: []domain.MissionContract{*contract}}
	seasons := &activeSeasonsStub{seasons: []domain.Season{domain.NewSeason("season-2026-q4", domain.SeasonStateActive)}}
	handler := application.NewListOffersHandler(offers, contracts, seasons)

	// Act
	got, err := handler.Handle("player-2", domain.OfferListFilter{}, domain.LanguageSpanish)

	// Assert
	require.NoError(t, err)
	require.Len(t, got.Offers, 1)
	require.Equal(t, "offer-y", got.Offers[0].OfferID)
}

func TestListOffers_HidesDefinitionsAlreadyActive(t *testing.T) {
	// Arrange
	hidden := domain.NewMissionOffer("offer-hidden", "BOUNTY", "season-2026-q4")
	hidden.DefinitionCode = "def-1"
	visible := domain.NewMissionOffer("offer-visible", "BOUNTY", "season-2026-q4")
	visible.DefinitionCode = "def-2"
	offers := &offerListReaderStub{offers: []domain.MissionOffer{hidden, visible}}
	activeContract := domain.NewMissionContractFromState("c1", hidden, []string{"player-1"}, domain.ContractInProgress, true, "", "season-2026-q4", nil, nil, nil)
	contracts := &boardContractsStub{contracts: []domain.MissionContract{*activeContract}}
	seasons := &activeSeasonsStub{seasons: []domain.Season{domain.NewSeason("season-2026-q4", domain.SeasonStateActive)}}
	handler := application.NewListOffersHandler(offers, contracts, seasons)
	filter := domain.OfferListFilter{Status: domain.OfferStatusAvailable}

	// Act
	got, err := handler.Handle("player-1", filter, domain.LanguageSpanish)

	// Assert
	require.NoError(t, err)
	require.False(t, len(got.Offers) != 1 || got.Offers[0].OfferID != "offer-visible")
	require.Equal(t, "player-1", contracts.gotActivePlayer)
}
