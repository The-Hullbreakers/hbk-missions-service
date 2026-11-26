package application_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// localizedPendingLister is a hand-written PendingReviewLister double that
// records calls and serves a fixture queue or an error.
type localizedPendingLister struct {
	contracts []domain.MissionContract
	err       error
	calls     int
}

func (l *localizedPendingLister) ListPendingReview() ([]domain.MissionContract, error) {
	l.calls++
	return l.contracts, l.err
}

func bilingualQueueContract(contractID string) domain.MissionContract {
	offer := domain.NewMissionOfferWithMode("offer-1", "BOUNTY_HUNTING", domain.MissionModeSolo, "season-1")
	offer.Title = "Cacería"
	offer.Description = "Descripción"
	offer.ContentEn = domain.LocalizedContent{Title: "Hunt", Description: "English briefing"}
	return *domain.NewMissionContractFromState(contractID, offer, []string{"player-1"},
		domain.ContractPendingValidation, true, "", "", nil, nil, nil)
}

func legacyQueueContract(contractID string) domain.MissionContract {
	offer := domain.NewMissionOfferWithMode("offer-2", "BOUNTY_HUNTING", domain.MissionModeSolo, "season-1")
	offer.Title = "Misión legada"
	return *domain.NewMissionContractFromState(contractID, offer, []string{"player-1", "player-2"},
		domain.ContractPendingValidation, true, "", "", nil, nil, nil)
}

// TEST-APP-LANG-001 (ListPendingReview partition): a reviewer's mixed queue
// under lang=en serves each entry in its effective language with one
// contract-level label while ids, state, roster and order stay unchanged.
func TestListPendingReviewLocalized_ServesPerItemLanguage(t *testing.T) {
	// Arrange
	lister := &localizedPendingLister{contracts: []domain.MissionContract{
		bilingualQueueContract("contract-1"),
		legacyQueueContract("contract-2"),
	}}
	reviewer, reviewerErr := domain.NewCallerIdentity("mod-1", []string{"moderator"}, nil)
	handler := application.NewListPendingReviewHandler(lister)

	// Act
	entries, err := handler.Handle(reviewer, domain.LanguageEnglish)

	// Assert
	require.NoError(t, reviewerErr)
	require.NoError(t, err)
	assert.Equal(t, 1, lister.calls)
	require.Len(t, entries, 2)
	assert.Equal(t, "contract-1", entries[0].ContractID)
	assert.Equal(t, "offer-1", entries[0].Offer.OfferID)
	assert.Equal(t, domain.ContractPendingValidation, entries[0].State)
	assert.Equal(t, "Hunt", entries[0].Offer.Title)
	assert.Equal(t, domain.LanguageEnglish, entries[0].ContentLanguage)
	assert.Equal(t, "contract-2", entries[1].ContractID)
	assert.Equal(t, "offer-2", entries[1].Offer.OfferID)
	assert.Equal(t, domain.ContractPendingValidation, entries[1].State)
	assert.Equal(t, "Misión legada", entries[1].Offer.Title)
	assert.Equal(t, []string{"player-1", "player-2"}, entries[1].Roster())
	assert.Equal(t, domain.LanguageSpanish, entries[1].ContentLanguage)
}

// TEST-APP-LANG-001 (ListPendingReview partition): a non-reviewer is
// rejected before the store is touched, regardless of language.
func TestListPendingReviewLocalized_NonReviewerDeniedBeforeStore(t *testing.T) {
	// Arrange
	lister := &localizedPendingLister{contracts: []domain.MissionContract{bilingualQueueContract("contract-1")}}
	member, memberErr := domain.NewCallerIdentity("player-1", nil, nil)
	handler := application.NewListPendingReviewHandler(lister)

	// Act
	_, err := handler.Handle(member, domain.LanguageEnglish)

	// Assert
	require.NoError(t, memberErr)
	assert.ErrorIs(t, err, domain.ErrCallerNotAuthorized)
	assert.Equal(t, 0, lister.calls)
}

// TEST-APP-LANG-001 (ListPendingReview partition): an empty queue stays
// empty for a reviewer.
func TestListPendingReviewLocalized_EmptyQueue(t *testing.T) {
	// Arrange
	admin, adminErr := domain.NewCallerIdentity("admin-1", []string{"admin"}, nil)
	handler := application.NewListPendingReviewHandler(&localizedPendingLister{})

	// Act
	entries, err := handler.Handle(admin, domain.LanguageEnglish)

	// Assert
	require.NoError(t, adminErr)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

// TEST-APP-LANG-001 (ListPendingReview partition): a lister failure
// propagates unchanged to the reviewer.
func TestListPendingReviewLocalized_PropagatesListerError(t *testing.T) {
	// Arrange
	boom := errors.New("store down")
	admin, adminErr := domain.NewCallerIdentity("admin-1", []string{"admin"}, nil)
	handler := application.NewListPendingReviewHandler(&localizedPendingLister{err: boom})

	// Act
	_, err := handler.Handle(admin, domain.LanguageEnglish)

	// Assert
	require.NoError(t, adminErr)
	assert.ErrorIs(t, err, boom)
}
