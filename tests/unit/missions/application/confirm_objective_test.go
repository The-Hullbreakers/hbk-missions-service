package application_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

type fixedContractStore struct {
	contract *domain.MissionContract
}

func (s fixedContractStore) Get(string) (*domain.MissionContract, error) {
	return s.contract, nil
}

type fixedContractWriter struct{}

func (fixedContractWriter) Save(*domain.MissionContract) error { return nil }

type fixedSeasonStore struct {
	season domain.Season
}

func (s fixedSeasonStore) Get(string) (domain.Season, error) {
	return s.season, nil
}

type failingContractStore struct{}

func (s failingContractStore) Get(string) (*domain.MissionContract, error) {
	return nil, errors.New("contract not found")
}

type failingSeasonStore struct{}

func (s failingSeasonStore) Get(string) (domain.Season, error) {
	return domain.Season{}, errors.New("season not found")
}

func TestConfirmObjectiveHandler_ReturnsContractStoreError(t *testing.T) {
	handler := application.NewConfirmObjectiveHandlerWithEvidence(
		failingContractStore{}, fixedContractWriter{},
		fixedSeasonStore{season: domain.NewSeason("season-1", domain.SeasonStateActive)},
		&fakeEvidenceStore{},
	)
	_, err := handler.HandleWithMedia(testIdentity(t, "player-1", nil), "contract-1", "objective-1", "target-1", "evidence/contract-1/x", "season-1")
	require.Error(t, err)
}

func TestConfirmObjectiveHandler_ReturnsSeasonStoreError(t *testing.T) {
	contract := confirmableContract(t)
	store, uris := storedEvidence(contract.ContractID, "e1")
	handler := application.NewConfirmObjectiveHandlerWithEvidence(
		fixedContractStore{contract: contract}, fixedContractWriter{}, failingSeasonStore{}, store,
	)
	_, err := handler.HandleWithMedia(lifecycleMember(t), "contract-1", "obj-a", "unit-1", uris[0], "season-1")
	require.Error(t, err)
}

func TestConfirmObjectiveHandler_ReturnsContractWriterError(t *testing.T) {
	contract := confirmableContract(t)
	store, uris := storedEvidence(contract.ContractID, "e1")
	writer := &atomicWriter{err: errors.New("contract writer unavailable")}
	handler := application.NewConfirmObjectiveHandlerWithEvidence(
		fixedContractStore{contract: contract}, writer,
		fixedSeasonStore{season: domain.NewSeason("season-1", domain.SeasonStateActive)}, store,
	)
	_, err := handler.HandleWithMedia(lifecycleMember(t), "contract-1", "obj-a", "unit-1", uris[0], "season-1")
	require.Error(t, err)
}
