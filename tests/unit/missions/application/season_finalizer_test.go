package application_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestSeasonFinalizer_StagesVerifiesCommitsUnderStableAttempt(t *testing.T) {
	// Arrange
	store := newRecordingSeasonStateStore()
	seedSeason(store, domain.NewSeasonWithOfficial("season-1", domain.SeasonStateFreezing, true))
	log := &finCallLog{}
	snaps := &finSnapshotStub{snap: finSnapshot(), log: log}
	committer := &finCommitStub{log: log}
	finalizer := application.NewSeasonFinalizer(store,
		&finLockStub{attempt: finAttempt(), log: log}, snaps,
		&finRosterStub{}, &finStageStub{log: log},
		&finVerifyStub{log: log}, committer)

	// Act
	got, err := finalizer.Handle("season-1")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, domain.SeasonStateFinalized, got.State)
	assert.Equal(t, []string{"acquire", "snapshot", "stage", "verify", "commit"}, log.calls)
	assert.Equal(t, "ADJ#season-1", snaps.got.ID, "snapshot must observe the minted attempt")
	assert.Equal(t, finAttempt().Cutoff, snaps.got.Cutoff)
	assert.Equal(t, domain.SeasonStateFreezing, committer.season.State,
		"finalizer must not transition state before the commit")
	assert.Equal(t, "ADJ#season-1", committer.attempt.ID)
	require.Len(t, committer.facts, 1)
}

func TestSeasonFinalizer_RetryReusesMintedAttemptAndIdenticalFacts(t *testing.T) {
	// Arrange — a prior run minted and persisted ADJ#season-1, staged one
	// fact, then crashed before commit; the lock double returns that same
	// attempt (Acquire is idempotent, it must never remint).
	store := newRecordingSeasonStateStore()
	seedSeason(store, domain.NewSeasonWithOfficial("season-1", domain.SeasonStateFreezing, true))
	staged, err := domain.NewPlacementFact(domain.PlacementFactParams{
		SeasonID: "season-1", Category: "BOUNTY_HUNTING",
		Track: domain.TrackIndividual, RecipientID: "p-1",
		Rank: 1, FinalMerits: 120,
	})
	require.NoError(t, err)
	log := &finCallLog{}
	snaps := &finSnapshotStub{snap: finSnapshot(), log: log}
	stager := &finStageStub{log: log}
	committer := &finCommitStub{log: log}
	finalizer := application.NewSeasonFinalizer(store,
		&finLockStub{attempt: finAttempt(), log: log}, snaps,
		&finRosterStub{}, stager, &finVerifyStub{log: log}, committer)

	// Act
	got, err := finalizer.Handle("season-1")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, domain.SeasonStateFinalized, got.State)
	assert.Equal(t, "ADJ#season-1", snaps.got.ID, "retry must reuse the minted attempt")
	assert.Equal(t, finAttempt().Cutoff, snaps.got.Cutoff, "retry must reuse the minted cutoff")
	assert.Equal(t, []domain.PlacementFact{staged}, stager.facts[0],
		"retry must re-stage the identical fact set")
	assert.Equal(t, "ADJ#season-1", committer.attempt.ID)
}

func TestSeasonFinalizer_TeamRosterUsesCreditedContractsOnly(t *testing.T) {
	// Arrange
	store := newRecordingSeasonStateStore()
	seedSeason(store, domain.NewSeasonWithOfficial("season-1", domain.SeasonStateFreezing, true))
	snap := application.AdjudicationSnapshot{Buckets: []application.AdjudicationBucket{
		{Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, Entries: []application.AdjudicationEntry{
			{RecipientID: "p-1", Total: 120, CreditedContractIDs: []string{"c-9"}},
		}},
		{Category: "BOUNTY_HUNTING", Track: domain.TrackTeam, Entries: []application.AdjudicationEntry{
			{RecipientID: "team-a", Total: 300, CreditedContractIDs: []string{"c-1"}},
		}},
	}}
	log := &finCallLog{}
	rosters := &finRosterStub{members: []string{"m-1", "m-2"}}
	committer := &finCommitStub{log: log}
	finalizer := application.NewSeasonFinalizer(store,
		&finLockStub{attempt: finAttempt(), log: log},
		&finSnapshotStub{snap: snap, log: log}, rosters,
		&finStageStub{log: log}, &finVerifyStub{log: log}, committer)

	// Act
	_, err := finalizer.Handle("season-1")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, rosters.calls, "roster lookup only applies to TEAM buckets")
	assert.Equal(t, "season-1", rosters.gotSeasonID)
	assert.Equal(t, "BOUNTY_HUNTING", rosters.gotCategory)
	assert.Equal(t, "team-a", rosters.gotRecipientID, "roster scope is the exact bucket recipient")
	assert.Equal(t, []string{"c-1"}, rosters.gotContractIDs,
		"only credited contracts may feed member_ids")
	require.Len(t, committer.facts, 2)
	assert.Empty(t, committer.facts[0].MemberIDs)
	assert.Equal(t, domain.TrackTeam, committer.facts[1].Track)
	assert.Equal(t, []string{"m-1", "m-2"}, committer.facts[1].MemberIDs)
}
