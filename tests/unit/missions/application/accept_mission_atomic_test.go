package application_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// allActiveContractsStore returns the whole active set, ignoring category.
type allActiveContractsStore struct {
	contracts []domain.MissionContract
}

func (s allActiveContractsStore) ListActiveByPlayer(playerID string) ([]domain.MissionContract, error) {
	return s.contracts, nil
}

// recordingAcceptanceWriter captures the single atomic commit.
type recordingAcceptanceWriter struct {
	contract *domain.MissionContract
	offer    domain.MissionOffer
	limit    int
	calls    int
	failWith error
}

func (w *recordingAcceptanceWriter) CommitAcceptance(contract *domain.MissionContract, offer domain.MissionOffer, limit int) error {
	w.calls++
	w.contract = contract
	w.offer = offer
	w.limit = limit
	return w.failWith
}

func activeContractForOffer(category, definitionCode, offerID string) domain.MissionContract {
	offer := domain.NewMissionOffer(offerID, category, "season-1")
	offer.DefinitionCode = definitionCode
	contract := domain.NewMissionContract("c-"+offerID, offer, []string{"player-1"})
	return *contract
}

// REQ-MB-014: acceptance is rejected when the offer's bound season is not
// ACTIVE — no contract, no write attempt.
func TestAcceptMission_RejectsWhenBoundSeasonNotActive(t *testing.T) {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	writer := &recordingAcceptanceWriter{}
	handler := application.NewAcceptMissionHandler(
		domain.NewActiveLimitPolicy(3, nil),
		fixedOfferStore{offer: offer},
		allActiveContractsStore{},
		historyContracts{},
		versionedDefinitions{},
		fixedSeasonStore{season: domain.NewSeason("season-1", domain.SeasonStateFinalized)},
		writer,
		nil,
	)

	_, err := handler.Handle(testIdentity(t, "player-1", nil), "offer-1", "", nil)
	require.ErrorIs(t, err, domain.ErrSeasonNotActive)
	require.Equal(t, 0, writer.calls)
}

// REQ-MB-004: the whole acceptance commits through one atomic write carrying
// the contract and each roster member's active-set entry.
func TestAcceptMission_CommitsAtomicallyThroughWriter(t *testing.T) {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	writer := &recordingAcceptanceWriter{}
	handler := application.NewAcceptMissionHandler(
		domain.NewActiveLimitPolicy(3, nil),
		fixedOfferStore{offer: offer},
		allActiveContractsStore{},
		historyContracts{},
		versionedDefinitions{},
		fixedSeasonStore{season: domain.NewSeason("season-1", domain.SeasonStateActive)},
		writer,
		nil,
	)

	contract, err := handler.Handle(testIdentity(t, "player-1", nil), "offer-1", "", nil)
	require.NoError(t, err)
	require.Equal(t, 1, writer.calls)
	require.Equal(t, contract, writer.contract)
	require.Equal(t, "offer-1", writer.offer.OfferID)
	require.Equal(t, 3, writer.limit)
}

// REQ-MB-003: the limit counts the player's total active set across
// categories, not only the offer's category.
func TestAcceptMission_LimitCountsAllActiveCategories(t *testing.T) {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	active := []domain.MissionContract{
		activeContractForOffer("TRADE", "def-2", "offer-b"),
		activeContractForOffer("EXPLORATION", "def-3", "offer-c"),
		activeContractForOffer("BOUNTY_HUNTING", "def-4", "offer-d"),
	}
	for i := range active {
		member, _ := domain.NewCallerIdentity("player-1", nil, nil)
		_ = active[i].Start(member)
	}
	writer := &recordingAcceptanceWriter{}
	handler := application.NewAcceptMissionHandler(
		domain.NewActiveLimitPolicy(3, nil),
		fixedOfferStore{offer: offer},
		allActiveContractsStore{contracts: active},
		historyContracts{},
		versionedDefinitions{},
		fixedSeasonStore{season: domain.NewSeason("season-1", domain.SeasonStateActive)},
		writer,
		nil,
	)

	_, err := handler.Handle(testIdentity(t, "player-1", nil), "offer-1", "", nil)
	require.ErrorIs(t, err, domain.ErrActiveMissionLimitReached)
}
