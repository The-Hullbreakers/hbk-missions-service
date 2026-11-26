package application_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// reviewStore records every write variant so tests can assert which port was
// used — Save (state only) or SaveApproved (transactional adjudication).
type reviewStore struct {
	contract  *domain.MissionContract
	saved     *domain.MissionContract
	credits   []domain.MeritCredit
	saveCalls int
}

func (s *reviewStore) Get(string) (*domain.MissionContract, error) { return s.contract, nil }
func (s *reviewStore) Save(c *domain.MissionContract) error {
	s.saved = c
	s.saveCalls++
	return nil
}
func (s *reviewStore) SaveApproved(c *domain.MissionContract, credits []domain.MeritCredit) error {
	s.saved = c
	s.credits = credits
	return nil
}

// reviewer builds an adjudicator identity with privileged roles.
func reviewer(t *testing.T, roles ...string) domain.CallerIdentity {
	t.Helper()
	id, err := domain.NewCallerIdentity("mod-1", roles, nil)
	require.NoError(t, err)
	return id
}

// inProgressContract starts the lifecycle fixture so Submit is legal.
func inProgressContract(t *testing.T) *domain.MissionContract {
	t.Helper()
	contract := lifecycleFixture(t, []domain.Objective{
		domain.NewObjectiveNamed("obj-a", "obj-a", "", "BOUNTY", 10, 0, 0),
	})
	err := contract.Start(lifecycleMember(t))
	require.NoError(t, err)
	return contract
}

// pendingContract submits an in-progress contract with one confirmed unit.
func pendingContract(t *testing.T, reward int) *domain.MissionContract {
	t.Helper()
	contract := rewardedLifecycleFixture(t, reward)
	member := lifecycleMember(t)
	err := contract.Start(member)
	require.NoError(t, err)
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	_, err = contract.Confirm("obj-a", "u1", season)
	require.NoError(t, err)
	err = contract.Submit(member)
	require.NoError(t, err)
	return contract
}
