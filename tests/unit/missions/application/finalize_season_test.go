package application_test

import (
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	"github.com/stretchr/testify/require"
)

func TestFinalizeSeasonHandler_TransitionsFreezingToFinalized(t *testing.T) {
	store := newRecordingSeasonStateStore()
	store.seasons["season-1"] = domain.NewSeason("season-1", domain.SeasonStateFreezing)
	handler := application.NewFinalizeSeasonHandler(store, staticBucketReader{}, staticStandingsReader{}, staticRosterReader{}, store)

	season, err := handler.Handle("season-1")

	if err != nil {
		t.Fatalf("expected finalize to succeed, got %v", err)
	}
	if season.State != domain.SeasonStateFinalized {
		t.Fatalf("expected FINALIZED, got %v", season.State)
	}
	if store.seasons["season-1"].State != domain.SeasonStateFinalized {
		t.Fatalf("expected persisted season to be FINALIZED")
	}
}

func TestFinalizeSeasonHandler_RejectsNonFreezing(t *testing.T) {
	store := newRecordingSeasonStateStore()
	store.seasons["season-1"] = domain.NewSeason("season-1", domain.SeasonStateActive)
	handler := application.NewFinalizeSeasonHandler(store, staticBucketReader{}, staticStandingsReader{}, staticRosterReader{}, store)

	_, err := handler.Handle("season-1")

	if err != domain.ErrInvalidSeasonStateTransition {
		t.Fatalf("expected ErrInvalidSeasonStateTransition, got %v", err)
	}
}

func TestFinalizeSeasonHandler_ReturnsGetError(t *testing.T) {
	handler := application.NewFinalizeSeasonHandler(failingSeasonStateReader{}, staticBucketReader{}, staticStandingsReader{}, staticRosterReader{}, newRecordingSeasonStateStore())
	_, err := handler.Handle("season-1")
	if err == nil {
		t.Fatalf("expected get error")
	}
}

func TestFinalizeSeasonHandler_ReturnsSaveError(t *testing.T) {
	store := newRecordingSeasonStateStore()
	store.seasons["season-1"] = domain.NewSeason("season-1", domain.SeasonStateFreezing)
	handler := application.NewFinalizeSeasonHandler(store, staticBucketReader{}, staticStandingsReader{}, staticRosterReader{}, failingFinalizedPersister{})
	_, err := handler.Handle("season-1")
	if err == nil {
		t.Fatalf("expected save error")
	}
}

func TestFinalizeSeasonHandler_RanksByPublishedOrderWhenReaderIsUnsorted(t *testing.T) {
	seasons := newRecordingSeasonStateStore()
	seedSeason(seasons, domain.NewSeasonWithOfficial("season-1", domain.SeasonStateFreezing, true))
	buckets := staticBucketReader{buckets: []domain.StandingBucket{
		{Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual},
	}}
	standings := staticStandingsReader{standings: []domain.Standing{
		{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, RecipientID: "player-2", Total: 100},
		{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, RecipientID: "player-1", Total: 150},
	}}
	persister := &recordingFinalizedPersister{}
	handler := application.NewFinalizeSeasonHandler(seasons, buckets, standings, staticRosterReader{}, persister)

	_, err := handler.Handle("season-1")

	require.NoError(t, err)
	require.Len(t, persister.facts[0], 2)
	require.Equal(t, "player-1", persister.facts[0][0].RecipientID)
	require.Equal(t, 1, persister.facts[0][0].Rank)
	require.Equal(t, "player-2", persister.facts[0][1].RecipientID)
	require.Equal(t, 2, persister.facts[0][1].Rank)
}
