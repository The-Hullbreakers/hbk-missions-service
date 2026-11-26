package domain_test

import (
	"testing"
	"time"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	"github.com/stretchr/testify/require"
)

// aFinalizedSeason returns a season fixture in FINALIZED state.
func aFinalizedSeason(id string, official bool) domain.Season {
	return domain.NewSeasonWithOfficial(id, domain.SeasonStateFinalized, official)
}

// REQ-SCC-005 / SCN-SCC-CREATE-EXISTS: an existing season id can never be
// overwritten — neither its state nor its Official flag.
func TestValidateCanCreateSeason_RejectsExistingID(t *testing.T) {
	existing := []domain.Season{aFinalizedSeason("season-x", true)}

	err := domain.ValidateCanCreateSeason("season-x", existing)
	require.ErrorIs(t, err, domain.ErrSeasonAlreadyExists)
}

func TestValidateCanCreateSeason_ExistingIDTakesPrecedenceOverActiveCheck(t *testing.T) {
	existing := []domain.Season{
		aFinalizedSeason("season-x", true),
		domain.NewSeason("season-active", domain.SeasonStateActive),
	}

	err := domain.ValidateCanCreateSeason("season-x", existing)
	require.ErrorIs(t, err, domain.ErrSeasonAlreadyExists)
}

// The single-ACTIVE invariant still holds for a brand-new id.
func TestValidateCanCreateSeason_RejectsNewIDWhileActiveExists(t *testing.T) {
	existing := []domain.Season{domain.NewSeason("season-active", domain.SeasonStateActive)}

	err := domain.ValidateCanCreateSeason("season-z", existing)
	require.ErrorIs(t, err, domain.ErrActiveSeasonAlreadyExists)
}

func TestValidateCanCreateSeason_AllowsNewIDWithoutActive(t *testing.T) {
	existing := []domain.Season{aFinalizedSeason("season-y", false)}

	err := domain.ValidateCanCreateSeason("season-z", existing)
	require.NoError(t, err)
}

// REQ-SCC-010 / SCN-SCC-RETRY-NEXT: a finalized scheduled season whose
// successor is absent from the list reports pending next-season creation,
// so every scheduler tick retries it.
func scheduledFinalizedSeason(t *testing.T) domain.Season {
	t.Helper()
	schedule, err := domain.NewSeasonSchedule("weekly", 3, 2, 1)
	require.NoError(t, err)
	season, err := domain.NewScheduledSeasonFromSchedule(schedule, time.Now().Add(-10*24*time.Hour), true)
	require.NoError(t, err)
	season.State = domain.SeasonStateFinalized
	return season
}

func TestSuccessorPending_TrueWhenSuccessorAbsent(t *testing.T) {
	season := scheduledFinalizedSeason(t)

	if !season.SuccessorPending([]domain.Season{season}) {
		require.Failf(t, "unexpected failure", "expected pending: successor of %q is absent", season.SeasonID)
	}
}

func TestSuccessorPending_FalseWhenSuccessorExists(t *testing.T) {
	season := scheduledFinalizedSeason(t)
	next, err := season.NextScheduledSeason()
	require.NoError(t, err)

	if season.SuccessorPending([]domain.Season{season, next}) {
		require.Failf(t, "unexpected failure", "expected not pending: successor %q exists", next.SeasonID)
	}
}

func TestSuccessorPending_FalseForNonFinalizedOrManual(t *testing.T) {
	active := scheduledFinalizedSeason(t)
	active.State = domain.SeasonStateActive
	manual := aFinalizedSeason("manual-1", true)

	if active.SuccessorPending([]domain.Season{active}) {
		require.Failf(t, "unexpected failure", "expected not pending: season is not finalized")
	}
	if manual.SuccessorPending([]domain.Season{manual}) {
		require.Failf(t, "unexpected failure", "expected not pending: manual season has no schedule")
	}
}
