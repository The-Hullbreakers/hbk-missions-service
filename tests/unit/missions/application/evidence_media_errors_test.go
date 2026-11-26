package application_test

import (
	"errors"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestConfirmObjectiveHandler_WithMedia_PropagatesStoreError(t *testing.T) {
	handler := application.NewConfirmObjectiveHandlerWithEvidence(
		failingContractStore{}, &atomicWriter{}, fixedSeasonStore{}, &fakeEvidenceStore{})

	_, err := handler.HandleWithMedia(testIdentity(t, "player-1", nil),
		"contract-1", "objective-1", "target-1", "evidence/contract-1/x", "season-1")

	require.Error(t, err)
}

func TestConfirmObjectiveHandler_WithMedia_RejectsOutsider(t *testing.T) {
	contract := newInProgressContract()
	handler := application.NewConfirmObjectiveHandlerWithEvidence(
		fixedContractStore{contract: contract}, &atomicWriter{},
		fixedSeasonStore{season: domain.NewSeason("season-1", domain.SeasonStateActive)}, &fakeEvidenceStore{})

	_, err := handler.HandleWithMedia(testIdentity(t, "player-2", nil),
		"contract-1", "objective-1", "target-1", "evidence/contract-1/x", "season-1")

	require.ErrorIs(t, err, domain.ErrCallerNotAuthorized)
}

func TestConfirmObjectiveHandler_WithMedia_PropagatesExistsError(t *testing.T) {
	contract := newInProgressContract()
	broken := errors.New("head failed")
	handler := application.NewConfirmObjectiveHandlerWithEvidence(
		fixedContractStore{contract: contract}, &atomicWriter{},
		fixedSeasonStore{season: domain.NewSeason("season-1", domain.SeasonStateActive)},
		&fakeEvidenceStore{existsErr: broken})

	_, err := handler.HandleWithMedia(testIdentity(t, "player-1", nil),
		"contract-1", "objective-1", "target-1", "evidence/contract-1/x", "season-1")

	require.ErrorIs(t, err, broken)
}

func TestConfirmObjectiveHandler_WithMedia_PropagatesSeasonError(t *testing.T) {
	contract := newInProgressContract()
	store := &fakeEvidenceStore{exists: map[string]bool{"evidence/contract-1/x": true}}
	handler := application.NewConfirmObjectiveHandlerWithEvidence(
		fixedContractStore{contract: contract}, &atomicWriter{}, failingSeasonStore{}, store)

	_, err := handler.HandleWithMedia(testIdentity(t, "player-1", nil),
		"contract-1", "objective-1", "target-1", "evidence/contract-1/x", "season-1")

	require.Error(t, err)
}

func TestConfirmObjectiveHandler_WithMedia_PropagatesDomainError(t *testing.T) {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	store := &fakeEvidenceStore{exists: map[string]bool{"evidence/contract-1/x": true}}
	handler := application.NewConfirmObjectiveHandlerWithEvidence(
		fixedContractStore{contract: contract}, &atomicWriter{},
		fixedSeasonStore{season: domain.NewSeason("season-1", domain.SeasonStateActive)}, store)

	_, err := handler.HandleWithMedia(testIdentity(t, "player-1", nil),
		"contract-1", "objective-1", "target-1", "evidence/contract-1/x", "season-1")

	require.ErrorIs(t, err, domain.ErrContractNotInProgress)
}

func TestConfirmObjectiveHandler_WithMedia_PropagatesWriterError(t *testing.T) {
	contract := newInProgressContract()
	broken := errors.New("commit failed")
	store := &fakeEvidenceStore{exists: map[string]bool{"evidence/contract-1/x": true}}
	handler := application.NewConfirmObjectiveHandlerWithEvidence(
		fixedContractStore{contract: contract}, &atomicWriter{err: broken},
		fixedSeasonStore{season: domain.NewSeason("season-1", domain.SeasonStateActive)}, store)

	_, err := handler.HandleWithMedia(testIdentity(t, "player-1", nil),
		"contract-1", "objective-1", "target-1", "evidence/contract-1/x", "season-1")

	require.ErrorIs(t, err, broken)
}

func TestRequestEvidenceReadHandler_PropagatesStoreError(t *testing.T) {
	handler := application.NewRequestEvidenceReadHandler(failingContractStore{}, &fakeEvidenceStore{})
	reviewer, _ := domain.NewCallerIdentity("ops-1", []string{"admin"}, nil)

	_, err := handler.Handle(reviewer, "contract-1", "evidence/contract-1/x")

	require.Error(t, err)
}

func TestRequestEvidenceReadHandler_PropagatesExistsError(t *testing.T) {
	contract := newInProgressContract()
	broken := errors.New("head failed")
	handler := application.NewRequestEvidenceReadHandler(
		fixedContractStore{contract: contract}, &fakeEvidenceStore{existsErr: broken})
	reviewer, _ := domain.NewCallerIdentity("ops-1", []string{"admin"}, nil)

	_, err := handler.Handle(reviewer, "contract-1", "evidence/contract-1/x")

	require.ErrorIs(t, err, broken)
}

func TestRequestEvidenceReadHandler_PropagatesReadError(t *testing.T) {
	contract := newInProgressContract()
	broken := errors.New("sign failed")
	store := &fakeEvidenceStore{exists: map[string]bool{"evidence/contract-1/x": true}, readErr: broken}
	handler := application.NewRequestEvidenceReadHandler(
		fixedContractStore{contract: contract}, store)
	reviewer, _ := domain.NewCallerIdentity("ops-1", []string{"admin"}, nil)

	_, err := handler.Handle(reviewer, "contract-1", "evidence/contract-1/x")

	require.ErrorIs(t, err, broken)
}
