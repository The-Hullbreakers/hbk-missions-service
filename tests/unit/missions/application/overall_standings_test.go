package application_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

type overallCreditsStub struct {
	credits   []domain.MeritCredit
	completed map[string]bool
	calls     int
	err       error
}

func (s *overallCreditsStub) ListOverallCredits(_ string) ([]domain.MeritCredit, map[string]bool, error) {
	s.calls++
	return s.credits, s.completed, s.err
}

func TestOverallStandings_SumsCategoriesAndShowsCallerOutsideTop(t *testing.T) {
	// Arrange
	reader := &overallCreditsStub{credits: []domain.MeritCredit{
		{SeasonID: "season-1", Category: "PVP_AIR", Track: domain.TrackIndividual, RecipientID: "mira", Amount: 120, ContractID: "contract-1", ObjectiveID: "objective-1"},
		{SeasonID: "season-1", Category: "RACING", Track: domain.TrackIndividual, RecipientID: "mira", Amount: 80, ContractID: "contract-1", ObjectiveID: "objective-1"},
		{SeasonID: "season-1", Category: "PVP_AIR", Track: domain.TrackIndividual, RecipientID: "kai", Amount: 190, ContractID: "contract-2", ObjectiveID: "objective-1"},
		{SeasonID: "season-1", Category: "PVP_AIR", Track: domain.TrackTeam, RecipientID: "team-mira", Amount: 400, ContractID: "contract-3", ObjectiveID: "objective-1"},
	}, completed: map[string]bool{"contract-1": true, "contract-2": true, "contract-3": true}}
	identity, identityErr := domain.NewCallerIdentity("kai", nil, nil)
	handler := application.NewGetOverallStandingsHandler(
		&fakeLeaderboardSeasonReader{season: domain.NewSeason("season-1", domain.SeasonStateActive)}, reader)

	// Act
	got, err := handler.Handle(identity, "season-1", 1, "")

	// Assert
	require.NoError(t, identityErr)
	require.NoError(t, err)
	require.Equal(t, "season-1", got.SeasonID)
	require.Equal(t, 2, got.Participants)
	require.Len(t, got.Board.Entries, 1)
	require.Equal(t, "mira", got.Board.Entries[0].RecipientID)
	require.Equal(t, 200, got.Board.Entries[0].Total)
	require.Equal(t, 1, got.Board.Entries[0].CompletedMissions)
	require.Equal(t, 2, got.Board.Entries[0].UniqueObjectives)
	require.NotNil(t, got.MyStanding)
	require.Equal(t, "kai", got.MyStanding.RecipientID)
	require.Equal(t, 2, got.MyStanding.Rank)
	require.NotEmpty(t, got.Board.NextCursor)
	require.Equal(t, 1, reader.calls)
}

func TestOverallStandings_NewSeasonHasNoInventedRank(t *testing.T) {
	// Arrange
	reader := &overallCreditsStub{}
	identity, identityErr := domain.NewCallerIdentity("new-player", nil, nil)
	handler := application.NewGetOverallStandingsHandler(
		&fakeLeaderboardSeasonReader{season: domain.NewSeason("season-1", domain.SeasonStateActive)}, reader)

	// Act
	got, err := handler.Handle(identity, "season-1", 10, "")

	// Assert
	require.NoError(t, identityErr)
	require.NoError(t, err)
	require.Equal(t, 0, got.Participants)
	require.Empty(t, got.Board.Entries)
	require.Nil(t, got.MyStanding)
	require.Equal(t, 1, reader.calls)
}

func TestOverallStandings_PropagatesCreditsReadFailure(t *testing.T) {
	// Arrange
	want := errors.New("credits store unavailable")
	reader := &overallCreditsStub{err: want}
	identity, identityErr := domain.NewCallerIdentity("kai", nil, nil)
	handler := application.NewGetOverallStandingsHandler(
		&fakeLeaderboardSeasonReader{season: domain.NewSeason("season-1", domain.SeasonStateActive)}, reader)

	// Act
	_, err := handler.Handle(identity, "season-1", 10, "")

	// Assert
	require.NoError(t, identityErr)
	require.ErrorIs(t, err, want)
}

func TestOverallStandings_RejectsMalformedCursor(t *testing.T) {
	// Arrange
	reader := &overallCreditsStub{}
	identity, identityErr := domain.NewCallerIdentity("kai", nil, nil)
	handler := application.NewGetOverallStandingsHandler(
		&fakeLeaderboardSeasonReader{season: domain.NewSeason("season-1", domain.SeasonStateActive)}, reader)

	// Act
	_, err := handler.Handle(identity, "season-1", 10, "not-a-valid-cursor")

	// Assert
	require.NoError(t, identityErr)
	require.ErrorIs(t, err, domain.ErrInvalidCursor)
}
