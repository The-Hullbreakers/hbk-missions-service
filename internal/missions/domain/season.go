package domain

import (
	"strings"
	"time"
)

// SeasonState represents the lifecycle state of a season.
type SeasonState string

const (
	SeasonStateActive    SeasonState = "ACTIVE"
	SeasonStateFreezing  SeasonState = "FREEZING"
	SeasonStateFinalized SeasonState = "FINALIZED"
	SeasonStateArchived  SeasonState = "ARCHIVED"
)

// Season is a competitive window with a state that guards credit acceptance.
type Season struct {
	SeasonID string          `json:"SeasonID"`
	State    SeasonState     `json:"State"`
	Official bool            `json:"Official"`
	StartsAt time.Time       `json:"starts_at,omitempty"`
	Schedule *SeasonSchedule `json:"schedule,omitempty"`
}

// NewSeason creates a season with the given identity and state.
// It defaults to Official=true to keep existing callers backward-compatible.
func NewSeason(id string, state SeasonState) Season {
	return NewSeasonWithOfficial(id, state, true)
}

// NewSeasonWithOfficial creates a season with an explicit official-statistics flag.
func NewSeasonWithOfficial(id string, state SeasonState, official bool) Season {
	return Season{
		SeasonID: id,
		State:    state,
		Official: official,
	}
}

// NewActiveSeason creates a new ACTIVE season after validating the identity.
func NewActiveSeason(seasonID string, official bool) (Season, error) {
	if seasonID == "" {
		return Season{}, ErrSeasonInvalid
	}
	return NewSeasonWithOfficial(seasonID, SeasonStateActive, official), nil
}

// SeasonStatusFilter is a business bucket for listing seasons.
type SeasonStatusFilter string

const (
	SeasonStatusFilterActive   SeasonStatusFilter = "ACTIVE"
	SeasonStatusFilterInactive SeasonStatusFilter = "INACTIVE"
	SeasonStatusFilterAll      SeasonStatusFilter = "ALL"
)

// Match returns true if the season belongs to the filter bucket.
func (f SeasonStatusFilter) Match(season Season) bool {
	switch f {
	case SeasonStatusFilterAll:
		return true
	case SeasonStatusFilterActive:
		return season.State == SeasonStateActive
	case SeasonStatusFilterInactive:
		return season.State != SeasonStateActive
	default:
		return false
	}
}

// ParseSeasonStatusFilter converts a string to a typed filter.
// Empty string defaults to ACTIVE. Unknown values return an error.
func ParseSeasonStatusFilter(raw string) (SeasonStatusFilter, error) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "", string(SeasonStatusFilterActive):
		return SeasonStatusFilterActive, nil
	case string(SeasonStatusFilterInactive):
		return SeasonStatusFilterInactive, nil
	case string(SeasonStatusFilterAll):
		return SeasonStatusFilterAll, nil
	default:
		return "", ErrInvalidSeasonStatusFilter
	}
}

// Freeze transitions an ACTIVE season to FREEZING.
func Freeze(season Season) (Season, error) {
	if season.State != SeasonStateActive {
		return Season{}, ErrInvalidSeasonStateTransition
	}
	season.State = SeasonStateFreezing
	return season, nil
}

// Finalize transitions a FREEZING season to FINALIZED.
func Finalize(season Season) (Season, error) {
	if season.State != SeasonStateFreezing {
		return Season{}, ErrInvalidSeasonStateTransition
	}
	season.State = SeasonStateFinalized
	return season, nil
}

// CanAcceptMerits returns true when the season can still record merit credits
// for the given contract. ACTIVE seasons always accept; FREEZING seasons accept
// only contracts that are already in progress.
func (s Season) CanAcceptMerits(contract *MissionContract) bool {
	if s.State == SeasonStateActive {
		return true
	}
	if s.State == SeasonStateFreezing && contract != nil && contract.State == ContractInProgress {
		return true
	}
	return false
}

// ValidateCanCreateSeason enforces the season-creation invariants in a single
// scope: an existing season_id is never overwritten, and at most one season
// may be ACTIVE at a time.
func ValidateCanCreateSeason(newID string, existing []Season) error {
	for _, season := range existing {
		if season.SeasonID == newID {
			return ErrSeasonAlreadyExists
		}
	}
	for _, season := range existing {
		if season.State == SeasonStateActive {
			return ErrActiveSeasonAlreadyExists
		}
	}
	return nil
}
