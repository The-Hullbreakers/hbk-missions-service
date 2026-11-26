package application_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

type lifecycleStore struct {
	contract *domain.MissionContract
	saved    *domain.MissionContract
}

func (s *lifecycleStore) Get(string) (*domain.MissionContract, error) { return s.contract, nil }
func (s *lifecycleStore) Save(c *domain.MissionContract) error        { s.saved = c; return nil }

func (s *lifecycleStore) SaveDeparture(c *domain.MissionContract, _ string, _ domain.ContractState) error {
	return s.Save(c)
}

func rewardedLifecycleFixture(t *testing.T, reward int) *domain.MissionContract {
	t.Helper()
	definition, err := domain.NewMissionDefinition(
		"rescue-siege", "Rescate en el asedio", "Extrae al piloto",
		"BOUNTY", domain.MissionModeSolo, reward,
		[]domain.Objective{domain.NewObjectiveNamed("obj-a", "Llega al piloto", "", "BOUNTY", 10, 0, 0)}, nil,
	)
	require.NoError(t, err)
	definition.ContentEn = domain.LocalizedContent{
		Title: "Rescue", Description: "Extract the pilot",
		Objectives: []domain.LocalizedObjective{
			{ObjectiveID: "obj-a", Name: "Reach the pilot"},
		},
	}
	_ = definition.Publish("v1")
	offer, err := domain.NewMissionOfferFromDefinition("off-1", "season-1", definition)
	require.NoError(t, err)
	contract := domain.NewMissionContract("c-1", offer, []string{"player-1"})
	snapshot, err := domain.NewContractTerms([]domain.Objective{
		domain.NewObjectiveNamed("obj-a", "Llega al piloto", "", "BOUNTY", 10, 0, 0),
	})
	require.NoError(t, err)
	contract.Terms = snapshot
	return contract
}

func lifecycleFixture(t *testing.T, terms []domain.Objective) *domain.MissionContract {
	t.Helper()
	offer := domain.NewMissionOffer("off-1", "BOUNTY", "season-1")
	contract := domain.NewMissionContract("c-1", offer, []string{"player-1"})
	if terms != nil {
		snapshot, err := domain.NewContractTerms(terms)
		require.NoError(t, err)
		contract.Terms = snapshot
	}
	return contract
}

func lifecycleMember(t *testing.T) domain.CallerIdentity {
	t.Helper()
	id, err := domain.NewCallerIdentity("player-1", nil, nil)
	require.NoError(t, err)
	return id
}

func lifecycleOutsider(t *testing.T) domain.CallerIdentity {
	t.Helper()
	id, err := domain.NewCallerIdentity("player-9", nil, nil)
	require.NoError(t, err)
	return id
}

func TestStartContract_StartsAndPersists(t *testing.T) {
	contract := lifecycleFixture(t, nil)
	store := &lifecycleStore{contract: contract}
	handler := application.NewStartContractHandler(store, store)

	err := handler.Handle(lifecycleMember(t), "c-1")

	require.NoError(t, err)
	require.False(t, store.saved == nil || store.saved.State != domain.ContractInProgress)
}

func TestStartContract_RejectsOutsider(t *testing.T) {
	contract := lifecycleFixture(t, nil)
	store := &lifecycleStore{contract: contract}
	handler := application.NewStartContractHandler(store, store)

	err := handler.Handle(lifecycleOutsider(t), "c-1")

	require.ErrorIs(t, err, domain.ErrCallerNotAuthorized)
	require.Nil(t, store.saved)
}

func TestAbandonContract_ReleasesAndPersists(t *testing.T) {
	contract := lifecycleFixture(t, nil)
	_ = contract.Start(lifecycleMember(t))
	store := &lifecycleStore{contract: contract}
	handler := application.NewAbandonContractHandler(store, store)

	err := handler.Handle(lifecycleMember(t), "c-1")

	require.NoError(t, err)
	require.False(t, store.saved == nil || store.saved.State != domain.ContractAbandoned)
}
