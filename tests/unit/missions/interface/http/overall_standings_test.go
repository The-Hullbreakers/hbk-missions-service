package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

type overallSeasonsStub struct{}

func (overallSeasonsStub) Get(string) (domain.Season, error) {
	return domain.NewSeason("season-1", domain.SeasonStateActive), nil
}

type overallCreditsStub struct{}

func (overallCreditsStub) ListOverallCredits(string) ([]domain.MeritCredit, map[string]bool, error) {
	return nil, nil, nil
}

func TestOverallStandingsHandler_WithoutSessionFailsClosed(t *testing.T) {
	// Arrange
	handler := missionhttp.OverallStandingsHandler(
		application.NewGetOverallStandingsHandler(overallSeasonsStub{}, overallCreditsStub{}))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/missions/seasons/season-1/standings/overall", nil)
	recorder := httptest.NewRecorder()

	// Act
	handler(recorder, request)

	// Assert
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}
