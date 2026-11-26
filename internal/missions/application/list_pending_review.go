package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// PendingReviewLister reads every contract awaiting adjudication — one CQRS
// query operation for the reviewer queue.
type PendingReviewLister interface {
	ListPendingReview() ([]domain.MissionContract, error)
}

// ListPendingReviewHandler orchestrates the review-queue listing use case:
// only adjudicators may discover contracts pending validation.
type ListPendingReviewHandler struct {
	lister PendingReviewLister
}

// NewListPendingReviewHandler creates a handler with the lister port.
func NewListPendingReviewHandler(lister PendingReviewLister) *ListPendingReviewHandler {
	return &ListPendingReviewHandler{lister: lister}
}

// Handle returns the pending-review contracts for a caller holding an
// adjudicator role; any other caller is rejected before the store is
// touched. Every queue entry is projected through the shared
// LocalizeContract so each item carries its honest content language.
func (h *ListPendingReviewHandler) Handle(identity domain.CallerIdentity, language domain.Language) ([]LocalizedContract, error) {
	if !isReviewer(identity) {
		return nil, domain.ErrCallerNotAuthorized
	}
	contracts, err := h.lister.ListPendingReview()
	if err != nil {
		return nil, err
	}
	entries := make([]LocalizedContract, 0, len(contracts))
	for _, c := range contracts {
		entries = append(entries, LocalizeContract(c, language))
	}
	return entries, nil
}
