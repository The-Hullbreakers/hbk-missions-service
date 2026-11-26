package application_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// pendingLister is a hand-written PendingReviewLister double recording calls.
type pendingLister struct {
	contracts []domain.MissionContract
	calls     int
}

func (l *pendingLister) ListPendingReview() ([]domain.MissionContract, error) {
	l.calls++
	return l.contracts, nil
}

func TestListPendingReview_ReviewerGetsQueue(t *testing.T) {
	pending := pendingContract(t, 0)
	lister := &pendingLister{contracts: []domain.MissionContract{*pending}}
	handler := application.NewListPendingReviewHandler(lister)

	got, err := handler.Handle(reviewer(t, "moderator"), domain.LanguageSpanish)

	require.NoError(t, err)
	require.Equal(t, 1, lister.calls)
	require.False(t, len(got) != 1 || got[0].ContractID != pending.ContractID)
}

func TestListPendingReview_MemberDeniedWithoutStoreCall(t *testing.T) {
	lister := &pendingLister{contracts: []domain.MissionContract{*pendingContract(t, 0)}}
	handler := application.NewListPendingReviewHandler(lister)
	member, _ := domain.NewCallerIdentity("player-1", nil, nil)

	_, err := handler.Handle(member, domain.LanguageSpanish)

	require.ErrorIs(t, err, domain.ErrCallerNotAuthorized)
	require.Equal(t, 0, lister.calls)
}

func TestListPendingReview_EmptyQueue(t *testing.T) {
	lister := &pendingLister{contracts: []domain.MissionContract{}}
	handler := application.NewListPendingReviewHandler(lister)

	got, err := handler.Handle(reviewer(t, "admin"), domain.LanguageSpanish)

	require.NoError(t, err)
	require.Len(t, got, 0)
}
