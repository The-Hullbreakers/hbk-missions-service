package application_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestSeasonFinalizer_EmptySnapshotCommitsEmptyManifest(t *testing.T) {
	// Arrange
	store := newRecordingSeasonStateStore()
	seedSeason(store, domain.NewSeasonWithOfficial("season-1", domain.SeasonStateFreezing, true))
	log := &finCallLog{}
	committer := &finCommitStub{log: log}
	finalizer := application.NewSeasonFinalizer(store,
		&finLockStub{attempt: finAttempt(), log: log},
		&finSnapshotStub{snap: application.AdjudicationSnapshot{}, log: log},
		&finRosterStub{}, &finStageStub{log: log},
		&finVerifyStub{log: log}, committer)

	// Act
	got, err := finalizer.Handle("season-1")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, domain.SeasonStateFinalized, got.State)
	assert.Equal(t, []string{"acquire", "snapshot", "stage", "verify", "commit"}, log.calls)
	assert.Empty(t, committer.facts, "zero ranked entries commits an empty manifest")
}

func TestSeasonFinalizer_PracticeSeasonStagesNoFacts(t *testing.T) {
	// Arrange — practice seasons run the full pipeline but never emit
	// placement facts, even if the snapshot carries bucket rows.
	store := newRecordingSeasonStateStore()
	seedSeason(store, domain.NewSeasonWithOfficial("practice-1", domain.SeasonStateFreezing, false))
	attempt := finAttempt()
	attempt.ID, attempt.SeasonID = "ADJ#practice-1", "practice-1"
	log := &finCallLog{}
	committer := &finCommitStub{log: log}
	finalizer := application.NewSeasonFinalizer(store,
		&finLockStub{attempt: attempt, log: log},
		&finSnapshotStub{snap: finSnapshot(), log: log},
		&finRosterStub{}, &finStageStub{log: log},
		&finVerifyStub{log: log}, committer)

	// Act
	got, err := finalizer.Handle("practice-1")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, domain.SeasonStateFinalized, got.State)
	assert.Equal(t, []string{"acquire", "snapshot", "stage", "verify", "commit"}, log.calls)
	assert.Empty(t, committer.facts, "practice seasons commit an empty manifest")
}
