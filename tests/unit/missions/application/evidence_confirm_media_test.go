package application_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestConfirmObjectiveHandler_WithMedia_StoresReference(t *testing.T) {
	contract := newInProgressContract()
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	writer := &atomicWriter{}
	store := &fakeEvidenceStore{exists: map[string]bool{"evidence/contract-1/abc-123": true}}
	handler := application.NewConfirmObjectiveHandlerWithEvidence(
		fixedContractStore{contract: contract}, writer,
		fixedSeasonStore{season: season}, store)

	credit, err := handler.HandleWithMedia(testIdentity(t, "player-1", nil),
		"contract-1", "objective-1", "target-1", "evidence/contract-1/abc-123", "season-1")

	require.NoError(t, err)
	require.Equal(t, "evidence/contract-1/abc-123", string(credit.MediaURI))
	require.Equal(t, "evidence/contract-1/abc-123", string(writer.credit.MediaURI))
}

func TestConfirmObjectiveHandler_WithMedia_PhantomRejected(t *testing.T) {
	contract := newInProgressContract()
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	writer := &atomicWriter{}
	store := &fakeEvidenceStore{exists: map[string]bool{}}
	handler := application.NewConfirmObjectiveHandlerWithEvidence(
		fixedContractStore{contract: contract}, writer,
		fixedSeasonStore{season: season}, store)

	_, err := handler.HandleWithMedia(testIdentity(t, "player-1", nil),
		"contract-1", "objective-1", "target-1", "evidence/contract-1/missing", "season-1")

	require.ErrorIs(t, err, domain.ErrEvidenceMediaInvalid)
	require.Equal(t, 0, writer.calls)
}

func TestConfirmObjectiveHandler_WithMedia_ForeignRejected(t *testing.T) {
	contract := newInProgressContract()
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	writer := &atomicWriter{}
	store := &fakeEvidenceStore{exists: map[string]bool{"evidence/contract-2/x": true}}
	handler := application.NewConfirmObjectiveHandlerWithEvidence(
		fixedContractStore{contract: contract}, writer,
		fixedSeasonStore{season: season}, store)

	_, err := handler.HandleWithMedia(testIdentity(t, "player-1", nil),
		"contract-1", "objective-1", "target-1", "evidence/contract-2/x", "season-1")

	require.ErrorIs(t, err, domain.ErrEvidenceMediaInvalid)
	require.Equal(t, 0, writer.calls)
}

func TestRequestEvidenceReadHandler_ModeratorGetsGrant(t *testing.T) {
	contract := newInProgressContract()
	store := &fakeEvidenceStore{exists: map[string]bool{"evidence/contract-1/abc-123": true}}
	handler := application.NewRequestEvidenceReadHandler(
		fixedContractStore{contract: contract}, store)
	reviewer, err := domain.NewCallerIdentity("ops-1", []string{"moderator"}, nil)
	require.NoError(t, err)

	grant, err := handler.Handle(reviewer, "contract-1", "evidence/contract-1/abc-123")

	require.NoError(t, err)
	require.False(t, grant.URL == "" || store.readURI != "evidence/contract-1/abc-123")
}

func TestRequestEvidenceReadHandler_MemberDenied(t *testing.T) {
	contract := newInProgressContract()
	store := &fakeEvidenceStore{exists: map[string]bool{"evidence/contract-1/abc-123": true}}
	handler := application.NewRequestEvidenceReadHandler(
		fixedContractStore{contract: contract}, store)

	_, err := handler.Handle(testIdentity(t, "player-1", nil), "contract-1", "evidence/contract-1/abc-123")

	require.ErrorIs(t, err, domain.ErrCallerNotAuthorized)
	require.Empty(t, store.readURI)
}

func TestRequestEvidenceReadHandler_ForeignURIRejected(t *testing.T) {
	contract := newInProgressContract()
	store := &fakeEvidenceStore{exists: map[string]bool{"evidence/contract-2/x": true}}
	handler := application.NewRequestEvidenceReadHandler(
		fixedContractStore{contract: contract}, store)
	reviewer, err := domain.NewCallerIdentity("ops-1", []string{"admin"}, nil)
	require.NoError(t, err)

	_, err = handler.Handle(reviewer, "contract-1", "evidence/contract-2/x")

	require.ErrorIs(t, err, domain.ErrEvidenceMediaInvalid)
}
