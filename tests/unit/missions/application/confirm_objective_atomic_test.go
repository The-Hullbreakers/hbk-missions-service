package application_test

import (
	"errors"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

type atomicWriter struct {
	contract *domain.MissionContract
	credit   domain.MeritCredit
	err      error
	calls    int
}

func (w *atomicWriter) Save(c *domain.MissionContract) error {
	w.calls++
	if w.err != nil {
		return w.err
	}
	w.contract = c
	if pending := c.PendingCredits(); len(pending) == 1 {
		w.credit = pending[0]
	}
	return nil
}

func confirmableContract(t *testing.T) *domain.MissionContract {
	t.Helper()
	contract := lifecycleFixture(t, []domain.Objective{
		domain.NewObjectiveNamed("obj-a", "obj-a", "", "BOUNTY", 10, 5, 2),
	})
	err := contract.Start(lifecycleMember(t))
	require.NoError(t, err)
	return contract
}

func confirmHandler(store application.ContractReader, writer *atomicWriter) *application.ConfirmObjectiveHandler {
	return application.NewConfirmObjectiveHandler(
		store, writer,
		fixedSeasonStore{season: domain.NewSeason("season-1", domain.SeasonStateActive)},
	)
}

// cloneContract rebuilds a contract from its observable state so doubles can
// return a fresh copy per Get, matching real persistence semantics: a failed
// write leaves no partial state behind.
func cloneContract(c *domain.MissionContract) *domain.MissionContract {
	clone := domain.NewMissionContractFromState(
		c.ContractID, c.Offer, c.Roster(), c.State,
		c.RosterFrozen(), c.TeamID, c.SeasonID, c.Confirmations(), c.Departures(), c.PendingCredits(),
	)
	clone.Terms = c.Terms
	return clone
}

// reloadingContractStore returns a fresh clone on every Get.
type reloadingContractStore struct {
	contract *domain.MissionContract
}

func (s reloadingContractStore) Get(string) (*domain.MissionContract, error) {
	return cloneContract(s.contract), nil
}

func TestConfirmObjective_CommitsContractAndCreditAtomically(t *testing.T) {
	contract := confirmableContract(t)
	writer := &atomicWriter{}
	handler := confirmHandler(fixedContractStore{contract: contract}, writer)

	credit, err := handler.Handle(lifecycleMember(t), contract.ContractID, "obj-a", "unit-1", "season-1")

	require.NoError(t, err)
	require.False(t, writer.calls != 1 || writer.contract == nil)
	require.Equal(t, 15, writer.credit.Amount)
	require.NotEmpty(t, credit.ConfirmationID)
}

func TestConfirmObjective_FailedCommitLeavesNoPartialState(t *testing.T) {
	contract := confirmableContract(t)
	writer := &atomicWriter{err: errors.New("ledger unavailable")}
	handler := confirmHandler(fixedContractStore{contract: contract}, writer)

	_, err := handler.Handle(lifecycleMember(t), contract.ContractID, "obj-a", "unit-1", "season-1")

	require.Error(t, err)
	require.False(t, writer.contract != nil || writer.credit.Amount != 0)
}

func TestConfirmObjective_RetryAfterFailureIsIdempotent(t *testing.T) {
	contract := confirmableContract(t)
	writer := &atomicWriter{err: errors.New("transient")}
	handler := confirmHandler(reloadingContractStore{contract: contract}, writer)

	_, _ = handler.Handle(lifecycleMember(t), contract.ContractID, "obj-a", "unit-1", "season-1")
	writer.err = nil
	credit, err := handler.Handle(lifecycleMember(t), contract.ContractID, "obj-a", "unit-1", "season-1")

	require.NoError(t, err)
	require.False(t, writer.calls != 2 || writer.credit.ConfirmationID != credit.ConfirmationID)
}

func TestConfirmObjective_RejectsObjectiveOutsideFrozenTerms(t *testing.T) {
	contract := confirmableContract(t)
	writer := &atomicWriter{}
	handler := confirmHandler(fixedContractStore{contract: contract}, writer)

	_, err := handler.Handle(lifecycleMember(t), contract.ContractID, "obj-z", "unit-1", "season-1")

	require.ErrorIs(t, err, domain.ErrObjectiveNotInContract)
	require.Equal(t, 0, writer.calls)
}
