package http

import (
	"errors"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// PublicCode maps a domain error to a stable public API code.
func PublicCode(err error) string {
	switch {
	case errors.Is(err, domain.ErrContractNil):
		return "CONTRACT_NOT_FOUND"
	case errors.Is(err, domain.ErrContractNotInProgress):
		return "CONTRACT_NOT_IN_PROGRESS"
	case errors.Is(err, domain.ErrContractTerminal):
		return "CONTRACT_TERMINAL"
	case errors.Is(err, domain.ErrSeasonNotActive):
		return "SEASON_NOT_ACTIVE"
	case errors.Is(err, domain.ErrCategoryMismatch):
		return "CATEGORY_MISMATCH"
	case errors.Is(err, domain.ErrRosterEmpty):
		return "ROSTER_EMPTY"
	case errors.Is(err, domain.ErrDuplicateConfirmation):
		return "DUPLICATE_CONFIRMATION"
	case errors.Is(err, domain.ErrObjectiveLimitReached):
		return "OBJECTIVE_LIMIT_REACHED"
	case errors.Is(err, domain.ErrTeamIDRequired):
		return "TEAM_ID_REQUIRED"
	case errors.Is(err, domain.ErrActiveMissionLimitReached):
		return "ACTIVE_MISSION_LIMIT_REACHED"
	case errors.Is(err, domain.ErrOfferCapacityExhausted):
		return "OFFER_CAPACITY_EXHAUSTED"
	case errors.Is(err, domain.ErrOfferInvalid):
		return "INVALID_MISSION_OFFER"
	case errors.Is(err, domain.ErrSeasonInvalid):
		return "INVALID_SEASON"
	case errors.Is(err, domain.ErrActiveSeasonAlreadyExists):
		return "ACTIVE_SEASON_EXISTS"
	case errors.Is(err, domain.ErrInvalidSeasonStatusFilter):
		return "INVALID_QUERY_PARAMETER"
	case errors.Is(err, domain.ErrInvalidTrack):
		return "INVALID_QUERY_PARAMETER"
	case errors.Is(err, domain.ErrInvalidCursor):
		return "INVALID_QUERY_PARAMETER"
	case errors.Is(err, domain.ErrInvalidSeasonStateTransition):
		return "SEASON_STATE_INVALID"
	case errors.Is(err, domain.ErrSeasonNotFound):
		return "SEASON_NOT_FOUND"
	case errors.Is(err, domain.ErrNoActiveSeason):
		return "NO_ACTIVE_SEASON"
	case errors.Is(err, domain.ErrSeasonMismatch):
		return "SEASON_MISMATCH"
	case errors.Is(err, domain.ErrDefinitionInvalid):
		return "INVALID_MISSION_DEFINITION"
	case errors.Is(err, domain.ErrDefinitionNotFound):
		return "DEFINITION_NOT_FOUND"
	case errors.Is(err, domain.ErrDefinitionNotPublished):
		return "DEFINITION_NOT_PUBLISHED"
	case errors.Is(err, domain.ErrMissionDefinitionAlreadyActive):
		return "MISSION_DEFINITION_ALREADY_ACTIVE"
	default:
		return "INVALID_REQUEST"
	}
}
