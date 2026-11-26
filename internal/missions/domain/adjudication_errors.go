package domain

import "errors"

// Sentinel errors for the two-phase adjudication model: internal attempt
// identity, staged fact-set verification, and credited-roster resolution.
var (
	ErrAdjudicationAttemptInvalid = errors.New("adjudication attempt is invalid")
	ErrStagedFactSetInvalid       = errors.New("staged fact set is invalid")
	ErrStagedFactSetMismatch      = errors.New("staged facts do not match the declared count and hash")
	ErrCreditedRosterInvalid      = errors.New("credited roster is invalid")
)
