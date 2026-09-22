package application_test

import (
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

type fakeSeasonReader struct {
	seasons []domain.Season
}

func (r *fakeSeasonReader) List(filter domain.SeasonStatusFilter) ([]domain.Season, error) {
	var out []domain.Season
	for _, season := range r.seasons {
		if filter.Match(season) {
			out = append(out, season)
		}
	}
	return out, nil
}

func TestListSeasonsHandler_ReturnsActiveSeasons(t *testing.T) {
	reader := &fakeSeasonReader{
		seasons: []domain.Season{
			domain.NewSeason("season-active", domain.SeasonStateActive),
			domain.NewSeason("season-finalized", domain.SeasonStateFinalized),
		},
	}
	handler := application.NewListSeasonsHandler(reader)

	got, err := handler.Handle(domain.SeasonStatusFilterActive)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(got) != 1 || got[0].SeasonID != "season-active" {
		t.Fatalf("expected one active season, got %+v", got)
	}
}

func TestListSeasonsHandler_ReturnsInactiveSeasons(t *testing.T) {
	reader := &fakeSeasonReader{
		seasons: []domain.Season{
			domain.NewSeason("season-active", domain.SeasonStateActive),
			domain.NewSeason("season-archived", domain.SeasonStateArchived),
		},
	}
	handler := application.NewListSeasonsHandler(reader)

	got, err := handler.Handle(domain.SeasonStatusFilterInactive)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(got) != 1 || got[0].SeasonID != "season-archived" {
		t.Fatalf("expected one inactive season, got %+v", got)
	}
}

func TestListSeasonsHandler_ReturnsAllSeasons(t *testing.T) {
	reader := &fakeSeasonReader{
		seasons: []domain.Season{
			domain.NewSeason("season-active", domain.SeasonStateActive),
			domain.NewSeason("season-archived", domain.SeasonStateArchived),
		},
	}
	handler := application.NewListSeasonsHandler(reader)

	got, err := handler.Handle(domain.SeasonStatusFilterAll)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected two seasons, got %d", len(got))
	}
}
