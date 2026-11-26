package http_test

import (
	"errors"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

func TestPublicCode_MapsDomainErrors(t *testing.T) {
	cases := []struct {
		err      error
		expected string
	}{
		{domain.ErrContractNil, "CONTRACT_NOT_FOUND"},
		{domain.ErrContractNotInProgress, "CONTRACT_NOT_IN_PROGRESS"},
		{domain.ErrContractTerminal, "CONTRACT_TERMINAL"},
		{domain.ErrSeasonNotActive, "SEASON_NOT_ACTIVE"},
		{domain.ErrSeasonAdjudicating, "SEASON_ADJUDICATING"},
		{domain.ErrCategoryMismatch, "CATEGORY_MISMATCH"},
		{domain.ErrRosterEmpty, "ROSTER_EMPTY"},
		{domain.ErrDuplicateConfirmation, "DUPLICATE_CONFIRMATION"},
		{domain.ErrObjectiveLimitReached, "OBJECTIVE_LIMIT_REACHED"},
		{domain.ErrTeamIDRequired, "TEAM_ID_REQUIRED"},
		{domain.ErrActiveMissionLimitReached, "ACTIVE_MISSION_LIMIT_REACHED"},
		{domain.ErrOfferInvalid, "INVALID_MISSION_OFFER"},
		{domain.ErrSeasonInvalid, "INVALID_SEASON"},
		{domain.ErrActiveSeasonAlreadyExists, "ACTIVE_SEASON_EXISTS"},
		{errors.New("unknown error"), "INVALID_REQUEST"},
	}

	for _, c := range cases {
		code := missionhttp.PublicCode(c.err)
		if code != c.expected {
			t.Fatalf("expected %s for %v, got %s", c.expected, c.err, code)
		}
	}
}
