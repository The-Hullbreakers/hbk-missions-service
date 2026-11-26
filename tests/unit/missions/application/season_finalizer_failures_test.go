package application_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestSeasonFinalizer_AcquireFailureStopsPipeline(t *testing.T) {
	// Arrange
	store := newRecordingSeasonStateStore()
	seedSeason(store, domain.NewSeasonWithOfficial("season-1", domain.SeasonStateFreezing, true))
	log := &finCallLog{}
	finalizer := application.NewSeasonFinalizer(store,
		&finLockStub{err: errors.New("adjudication locked"), log: log},
		&finSnapshotStub{snap: finSnapshot(), log: log},
		&finRosterStub{}, &finStageStub{log: log},
		&finVerifyStub{log: log}, &finCommitStub{log: log})

	// Act
	_, err := finalizer.Handle("season-1")

	// Assert
	require.Error(t, err)
	assert.Equal(t, []string{"acquire"}, log.calls,
		"no snapshot/stage/verify/commit may run without the lock")
}

func TestSeasonFinalizer_SnapshotFailureStopsPipeline(t *testing.T) {
	// Arrange
	store := newRecordingSeasonStateStore()
	seedSeason(store, domain.NewSeasonWithOfficial("season-1", domain.SeasonStateFreezing, true))
	log := &finCallLog{}
	finalizer := application.NewSeasonFinalizer(store,
		&finLockStub{attempt: finAttempt(), log: log},
		&finSnapshotStub{err: errors.New("read timeout"), log: log},
		&finRosterStub{}, &finStageStub{log: log},
		&finVerifyStub{log: log}, &finCommitStub{log: log})

	// Act
	_, err := finalizer.Handle("season-1")

	// Assert
	require.Error(t, err)
	assert.Equal(t, []string{"acquire", "snapshot"}, log.calls,
		"a snapshot failure must prevent staging, verify, and commit")
}

func TestSeasonFinalizer_StageFailureBlocksVerifyAndCommit(t *testing.T) {
	// Arrange
	store := newRecordingSeasonStateStore()
	seedSeason(store, domain.NewSeasonWithOfficial("season-1", domain.SeasonStateFreezing, true))
	log := &finCallLog{}
	finalizer := application.NewSeasonFinalizer(store,
		&finLockStub{attempt: finAttempt(), log: log},
		&finSnapshotStub{snap: finSnapshot(), log: log},
		&finRosterStub{},
		&finStageStub{log: log, failAt: 1, err: errors.New("throttled")},
		&finVerifyStub{log: log}, &finCommitStub{log: log})

	// Act
	_, err := finalizer.Handle("season-1")

	// Assert
	require.Error(t, err)
	assert.Equal(t, []string{"acquire", "snapshot", "stage"}, log.calls,
		"a staging failure must prevent verify and commit")
}

func TestSeasonFinalizer_VerifyFailureBlocksCommit(t *testing.T) {
	// Arrange
	store := newRecordingSeasonStateStore()
	seedSeason(store, domain.NewSeasonWithOfficial("season-1", domain.SeasonStateFreezing, true))
	log := &finCallLog{}
	finalizer := application.NewSeasonFinalizer(store,
		&finLockStub{attempt: finAttempt(), log: log},
		&finSnapshotStub{snap: finSnapshot(), log: log},
		&finRosterStub{}, &finStageStub{log: log},
		&finVerifyStub{log: log, err: errors.New("digest mismatch")},
		&finCommitStub{log: log})

	// Act
	_, err := finalizer.Handle("season-1")

	// Assert
	require.Error(t, err)
	assert.Equal(t, []string{"acquire", "snapshot", "stage", "verify"}, log.calls)
	assert.NotContains(t, log.calls, "commit",
		"a failed staged-set verification must prevent the commit")
}
