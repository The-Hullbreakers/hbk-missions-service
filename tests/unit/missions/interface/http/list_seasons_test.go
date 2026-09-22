package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

type fakeSeasonListReader struct {
	seasons []domain.Season
}

func (r *fakeSeasonListReader) List(filter domain.SeasonStatusFilter) ([]domain.Season, error) {
	var out []domain.Season
	for _, season := range r.seasons {
		if filter.Match(season) {
			out = append(out, season)
		}
	}
	return out, nil
}

func TestListSeasonsHandler_ReturnsActiveByDefault(t *testing.T) {
	app := application.NewListSeasonsHandler(&fakeSeasonListReader{
		seasons: []domain.Season{
			domain.NewSeason("season-active", domain.SeasonStateActive),
			domain.NewSeason("season-archived", domain.SeasonStateArchived),
		},
	})
	handler := missionhttp.ListSeasonsHandler(app)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/seasons", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if len(result) != 1 || result[0]["SeasonID"] != "season-active" {
		t.Fatalf("expected one active season, got %+v", result)
	}
}

func TestListSeasonsHandler_ReturnsAllWhenRequested(t *testing.T) {
	app := application.NewListSeasonsHandler(&fakeSeasonListReader{
		seasons: []domain.Season{
			domain.NewSeason("season-active", domain.SeasonStateActive),
			domain.NewSeason("season-archived", domain.SeasonStateArchived),
		},
	})
	handler := missionhttp.ListSeasonsHandler(app)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/seasons?status=all", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected two seasons, got %d", len(result))
	}
}

func TestListSeasonsHandler_RejectsUnknownStatus(t *testing.T) {
	app := application.NewListSeasonsHandler(&fakeSeasonListReader{})
	handler := missionhttp.ListSeasonsHandler(app)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/seasons?status=closed", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}

	var result map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if result["code"] != "INVALID_QUERY_PARAMETER" {
		t.Fatalf("expected INVALID_QUERY_PARAMETER, got %v", result["code"])
	}
}
