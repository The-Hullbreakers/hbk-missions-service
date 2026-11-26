package application_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func seedSeason(store *recordingSeasonStateStore, season domain.Season) {
	store.seasons[season.SeasonID] = season
}

type recordingFinalizedPersister struct {
	seasons []domain.Season
	facts   [][]domain.PlacementFact
}

func (p *recordingFinalizedPersister) SaveFinalized(season domain.Season, facts []domain.PlacementFact) error {
	p.seasons = append(p.seasons, season)
	p.facts = append(p.facts, facts)
	return nil
}

type staticBucketReader struct{ buckets []domain.StandingBucket }

func (r staticBucketReader) ListBuckets(string) ([]domain.StandingBucket, error) {
	return r.buckets, nil
}

type staticStandingsReader struct{ standings []domain.Standing }

func (r staticStandingsReader) ListStandings(string, string, domain.Track, int) ([]domain.Standing, error) {
	return r.standings, nil
}

type staticRosterReader struct{ members map[string][]string }

func (r staticRosterReader) Members(seasonID, teamID string) ([]string, error) {
	return r.members[seasonID+":"+teamID], nil
}

func TestFinalizeSeasonHandler_OfficialSeasonIntendsOneFactPerRankedEntry(t *testing.T) {
	seasons := newRecordingSeasonStateStore()
	seedSeason(seasons, domain.NewSeasonWithOfficial("season-1", domain.SeasonStateFreezing, true))
	buckets := staticBucketReader{buckets: []domain.StandingBucket{
		{Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual},
	}}
	standings := staticStandingsReader{standings: []domain.Standing{
		{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, RecipientID: "player-1", Total: 150},
		{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, RecipientID: "player-2", Total: 100},
	}}
	persister := &recordingFinalizedPersister{}
	handler := application.NewFinalizeSeasonHandler(seasons, buckets, standings, staticRosterReader{}, persister)

	_, err := handler.Handle("season-1")

	require.NoError(t, err)
	require.Len(t, persister.facts, 1)
	require.Len(t, persister.facts[0], 2)
	require.Equal(t, "player-1", persister.facts[0][0].RecipientID)
	require.Equal(t, 1, persister.facts[0][0].Rank)
	require.Equal(t, 150, persister.facts[0][0].FinalMerits)
	require.Equal(t, "player-2", persister.facts[0][1].RecipientID)
	require.Equal(t, 2, persister.facts[0][1].Rank)
}

func TestFinalizeSeasonHandler_UnofficialSeasonIntendsNoFacts(t *testing.T) {
	seasons := newRecordingSeasonStateStore()
	seedSeason(seasons, domain.NewSeasonWithOfficial("practice-1", domain.SeasonStateFreezing, false))
	buckets := staticBucketReader{buckets: []domain.StandingBucket{
		{Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual},
	}}
	standings := staticStandingsReader{standings: []domain.Standing{
		{SeasonID: "practice-1", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, RecipientID: "player-1", Total: 150},
	}}
	persister := &recordingFinalizedPersister{}
	handler := application.NewFinalizeSeasonHandler(seasons, buckets, standings, staticRosterReader{}, persister)

	_, err := handler.Handle("practice-1")

	require.NoError(t, err)
	require.Empty(t, persister.facts[0])
}

func TestFinalizeSeasonHandler_TeamEntryCarriesFrozenMembers(t *testing.T) {
	seasons := newRecordingSeasonStateStore()
	seedSeason(seasons, domain.NewSeasonWithOfficial("season-1", domain.SeasonStateFreezing, true))
	buckets := staticBucketReader{buckets: []domain.StandingBucket{
		{Category: "BOUNTY_HUNTING", Track: domain.TrackTeam},
	}}
	standings := staticStandingsReader{standings: []domain.Standing{
		{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackTeam, RecipientID: "team-alpha", Total: 300},
	}}
	rosters := staticRosterReader{members: map[string][]string{"season-1:team-alpha": {"user-1", "user-2"}}}
	persister := &recordingFinalizedPersister{}
	handler := application.NewFinalizeSeasonHandler(seasons, buckets, standings, rosters, persister)

	_, err := handler.Handle("season-1")

	require.NoError(t, err)
	require.Len(t, persister.facts[0], 1)
	require.Equal(t, "team-alpha", persister.facts[0][0].RecipientID)
	require.Equal(t, []string{"user-1", "user-2"}, persister.facts[0][0].MemberIDs)
}

func TestFinalizeSeasonHandler_IntentsSavedAtomicallyWithSeason(t *testing.T) {
	seasons := newRecordingSeasonStateStore()
	seedSeason(seasons, domain.NewSeasonWithOfficial("season-1", domain.SeasonStateFreezing, true))
	buckets := staticBucketReader{buckets: []domain.StandingBucket{
		{Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual},
	}}
	standings := staticStandingsReader{standings: []domain.Standing{
		{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, RecipientID: "player-1", Total: 150},
	}}
	persister := &recordingFinalizedPersister{}
	handler := application.NewFinalizeSeasonHandler(seasons, buckets, standings, staticRosterReader{}, persister)

	_, err := handler.Handle("season-1")

	require.NoError(t, err)
	require.Len(t, persister.seasons, 1)
	require.Equal(t, domain.SeasonStateFinalized, persister.seasons[0].State)
	require.Len(t, persister.facts, 1)
}
