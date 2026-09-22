package application

import (
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// SeasonWriter persists a season.
type SeasonWriter interface {
	Save(season domain.Season) error
}

// ActiveSeasonsReader lists seasons that are currently active.
type ActiveSeasonsReader interface {
	ListActive() ([]domain.Season, error)
}

// CreateSeasonHandler creates a new season.
type CreateSeasonHandler struct {
	activeSeasons ActiveSeasonsReader
	seasons       SeasonWriter
}

// NewCreateSeasonHandler creates a handler with the required ports.
func NewCreateSeasonHandler(activeSeasons ActiveSeasonsReader, seasons SeasonWriter) *CreateSeasonHandler {
	return &CreateSeasonHandler{activeSeasons: activeSeasons, seasons: seasons}
}

// Handle validates and persists a new season.
func (h *CreateSeasonHandler) Handle(tenantID, seasonID string, official bool) (domain.Season, error) {
	active, err := h.activeSeasons.ListActive()
	if err != nil {
		return domain.Season{}, err
	}
	if err := domain.ValidateCanCreateSeason(active); err != nil {
		return domain.Season{}, err
	}

	season, err := domain.NewActiveSeason(seasonID, official)
	if err != nil {
		return domain.Season{}, err
	}
	if err := h.seasons.Save(season); err != nil {
		return domain.Season{}, err
	}

	return season, nil
}
