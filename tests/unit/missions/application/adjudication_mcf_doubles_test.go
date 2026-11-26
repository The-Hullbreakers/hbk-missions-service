package application_test

import (
	"errors"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// TEST-APP-MCF-001 doubles — characterization only. errAdjudicationBlocked
// is a LOCAL sentinel standing in for the future DynamoDB same-write lock
// condition proven at the T-12 boundary.
var errAdjudicationBlocked = errors.New("season adjudication lock engaged")

// mcfContractStore returns a fresh clone on every Get so mutations the
// handler applies in memory never reach the stored original.
type mcfContractStore struct{ contract *domain.MissionContract }

func (s mcfContractStore) Get(string) (*domain.MissionContract, error) {
	return mcfClone(s.contract), nil
}

func mcfClone(c *domain.MissionContract) *domain.MissionContract {
	clone := domain.NewMissionContractFromState(
		c.ContractID, c.Offer, c.Roster(), c.State,
		c.RosterFrozen(), c.TeamID, c.SeasonID, c.Confirmations(), c.Departures(), c.PendingCredits(),
	)
	clone.Terms = c.Terms
	return clone
}

// adjudicationConflictWriter simulates a write rejected by the same-write
// lock condition: the attempt is observed but nothing is persisted.
type adjudicationConflictWriter struct {
	err      error
	attempts int
	saved    *domain.MissionContract
	credits  []domain.MeritCredit
}

func (w *adjudicationConflictWriter) Save(*domain.MissionContract) error {
	w.attempts++
	return w.err
}

func (w *adjudicationConflictWriter) SaveApproved(*domain.MissionContract, []domain.MeritCredit) error {
	w.attempts++
	return w.err
}

type mcfSeasonStore map[string]domain.Season

func (m mcfSeasonStore) Get(id string) (domain.Season, error) {
	s, ok := m[id]
	if !ok {
		return domain.Season{}, domain.ErrSeasonNotFound
	}
	return s, nil
}
