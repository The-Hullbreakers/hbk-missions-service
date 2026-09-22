package domain

import "errors"

// Sentinel errors returned by ConfirmObjective for stable public mapping.
var (
	ErrContractNil                    = errors.New("contract is nil")
	ErrContractNotInProgress          = errors.New("contract not in progress")
	ErrContractTerminal               = errors.New("contract is in a terminal state")
	ErrSeasonNotActive                = errors.New("season not active")
	ErrCategoryMismatch               = errors.New("category mismatch")
	ErrRosterEmpty                    = errors.New("roster is empty")
	ErrDuplicateConfirmation          = errors.New("duplicate confirmation")
	ErrObjectiveLimitReached          = errors.New("objective unit limit reached")
	ErrTeamIDRequired                 = errors.New("team id is required for team missions")
	ErrActiveMissionLimitReached      = errors.New("active mission limit reached")
	ErrOfferCapacityExhausted         = errors.New("offer capacity exhausted")
	ErrOfferInvalid                   = errors.New("offer is invalid")
	ErrSeasonInvalid                  = errors.New("season is invalid")
	ErrActiveSeasonAlreadyExists      = errors.New("an active season already exists")
	ErrInvalidSeasonStatusFilter      = errors.New("invalid season status filter")
	ErrInvalidSeasonStateTransition   = errors.New("invalid season state transition")
	ErrSeasonNotFound                 = errors.New("season not found")
	ErrNoActiveSeason                 = errors.New("no active season")
	ErrSeasonMismatch                 = errors.New("season mismatch")
	ErrInvalidCursor                  = errors.New("invalid cursor")
	ErrInvalidConfirmationID          = errors.New("invalid confirmation id")
	ErrDefinitionInvalid              = errors.New("definition is invalid")
	ErrDefinitionNotFound             = errors.New("definition not found")
	ErrDefinitionNotPublished         = errors.New("definition is not published")
	ErrMissionDefinitionAlreadyActive = errors.New("mission definition already active")
)
