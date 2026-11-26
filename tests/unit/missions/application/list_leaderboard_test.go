package application_test

import (
	"errors"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

type fakeLeaderboardSeasonReader struct {
	season domain.Season
	err    error
}

func (r *fakeLeaderboardSeasonReader) Get(seasonID string) (domain.Season, error) {
	return r.season, r.err
}

type fakeLeaderboardReader struct {
	standings []domain.Standing
	err       error
}

func (r *fakeLeaderboardReader) ListStandings(seasonID, category string, track domain.Track, minMerits int) ([]domain.Standing, error) {
	return r.standings, r.err
}

func TestListLeaderboardHandler_ReturnsLeaderboard(t *testing.T) {
	seasonReader := &fakeLeaderboardSeasonReader{season: domain.NewSeason("season-1", domain.SeasonStateActive)}
	leaderboardReader := &fakeLeaderboardReader{
		standings: []domain.Standing{
			{RecipientID: "player-1", Total: 150, Track: domain.TrackIndividual},
			{RecipientID: "player-2", Total: 100, Track: domain.TrackIndividual},
		},
	}
	handler := application.NewListLeaderboardHandler(seasonReader, leaderboardReader)

	got, err := handler.Handle("season-1", "BOUNTY_HUNTING", "individual", 10, 0, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(got.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got.Entries))
	}
	if got.Entries[0].RecipientID != "player-1" {
		t.Fatalf("expected player-1 first, got %s", got.Entries[0].RecipientID)
	}
}

func TestListLeaderboardHandler_RejectsInvalidTrack(t *testing.T) {
	seasonReader := &fakeLeaderboardSeasonReader{season: domain.NewSeason("season-1", domain.SeasonStateActive)}
	leaderboardReader := &fakeLeaderboardReader{}
	handler := application.NewListLeaderboardHandler(seasonReader, leaderboardReader)

	_, err := handler.Handle("season-1", "BOUNTY_HUNTING", "squad", 10, 0, "")
	if !errors.Is(err, domain.ErrInvalidTrack) {
		t.Fatalf("expected ErrInvalidTrack, got %v", err)
	}
}

func TestListLeaderboardHandler_ReturnsNotFound(t *testing.T) {
	seasonReader := &fakeLeaderboardSeasonReader{err: domain.ErrSeasonNotFound}
	leaderboardReader := &fakeLeaderboardReader{}
	handler := application.NewListLeaderboardHandler(seasonReader, leaderboardReader)

	_, err := handler.Handle("season-1", "BOUNTY_HUNTING", "individual", 10, 0, "")
	if !errors.Is(err, domain.ErrSeasonNotFound) {
		t.Fatalf("expected ErrSeasonNotFound, got %v", err)
	}
}

func TestListLeaderboardHandler_ReturnsListStandingsError(t *testing.T) {
	seasonReader := &fakeLeaderboardSeasonReader{season: domain.NewSeason("season-1", domain.SeasonStateActive)}
	leaderboardReader := &fakeLeaderboardReader{err: errors.New("standings unavailable")}
	handler := application.NewListLeaderboardHandler(seasonReader, leaderboardReader)

	_, err := handler.Handle("season-1", "BOUNTY_HUNTING", "individual", 10, 0, "")
	if err == nil {
		t.Fatalf("expected list standings error")
	}
}
