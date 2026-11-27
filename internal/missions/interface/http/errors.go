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
	case errors.Is(err, domain.ErrCallerNotAuthorized):
		return "FORBIDDEN"
	case errors.Is(err, domain.ErrNotTeamMember):
		return "NOT_TEAM_MEMBER"
	case errors.Is(err, domain.ErrContractNotInProgress):
		return "CONTRACT_NOT_IN_PROGRESS"
	case errors.Is(err, domain.ErrContractTerminal):
		return "CONTRACT_TERMINAL"
	case errors.Is(err, domain.ErrContractStateInvalid):
		return "CONTRACT_STATE_INVALID"
	case errors.Is(err, domain.ErrContractNotPendingValidation):
		return "CONTRACT_NOT_PENDING_VALIDATION"
	case errors.Is(err, domain.ErrContractTermsInvalid):
		return "CONTRACT_STATE_INVALID"
	case errors.Is(err, domain.ErrObjectiveNotInContract):
		return "OBJECTIVE_NOT_IN_CONTRACT"
	case errors.Is(err, domain.ErrSeasonNotActive):
		return "SEASON_NOT_ACTIVE"
	case errors.Is(err, domain.ErrSeasonAdjudicating):
		return "SEASON_ADJUDICATING"
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
	case errors.Is(err, domain.ErrOfferInvalid):
		return "INVALID_MISSION_OFFER"
	case errors.Is(err, domain.ErrOfferNotAvailable):
		return "OFFER_NOT_AVAILABLE"
	case errors.Is(err, domain.ErrSeasonInvalid):
		return "INVALID_SEASON"
	case errors.Is(err, domain.ErrActiveSeasonAlreadyExists):
		return "ACTIVE_SEASON_EXISTS"
	case errors.Is(err, domain.ErrSeasonAlreadyExists):
		return "SEASON_ALREADY_EXISTS"
	case errors.Is(err, domain.ErrTeamNotAllowed):
		return "TEAM_NOT_ALLOWED"
	case errors.Is(err, domain.ErrTeammateNotMember):
		return "TEAMMATE_NOT_MEMBER"
	case errors.Is(err, domain.ErrTeammateLimitReached):
		return "TEAMMATE_LIMIT_REACHED"
	case errors.Is(err, domain.ErrMembershipUnavailable):
		return "MEMBERSHIP_UNAVAILABLE"
	case errors.Is(err, domain.ErrNotInvited):
		return "NOT_INVITED"
	case errors.Is(err, domain.ErrInvitationNotPending):
		return "INVITATION_NOT_PENDING"
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
	case errors.Is(err, domain.ErrStandingNotFound):
		return "STANDING_NOT_FOUND"
	case errors.Is(err, domain.ErrSeasonStateInvalid):
		return "SEASON_STATE_INVALID"
	case errors.Is(err, domain.ErrInvalidPolicy):
		return "INVALID_POLICY"
	case errors.Is(err, domain.ErrPhaseNotUnlocked):
		return "PHASE_NOT_UNLOCKED"
	case errors.Is(err, domain.ErrDisputesPending):
		return "DISPUTES_PENDING"
	case errors.Is(err, domain.ErrDisputeStateInvalid):
		return "DISPUTE_STATE_INVALID"
	case errors.Is(err, domain.ErrDisputeNotFound):
		return "DISPUTE_NOT_FOUND"
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
	case errors.Is(err, domain.ErrEvidenceMediaInvalid):
		return "EVIDENCE_MEDIA_INVALID"
	case errors.Is(err, domain.ErrContractEvidenceRequired):
		return "EVIDENCE_REQUIRED"
	case errors.Is(err, domain.ErrUnitNotStaged):
		return "UNIT_NOT_STAGED"
	case errors.Is(err, domain.ErrInvalidOfferStatusFilter),
		errors.Is(err, domain.ErrInvalidOfferModeFilter),
		errors.Is(err, domain.ErrInvalidOfferListFilter),
		errors.Is(err, domain.ErrInvalidContractStatusFilter):
		return "INVALID_QUERY_PARAMETER"
	default:
		return "INVALID_REQUEST"
	}
}
