package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// FreezeSeasonHandler freezes an active season.
type FreezeSeasonHandler struct {
	seasons SeasonReader
	writer  SeasonWriter
}

// NewFreezeSeasonHandler creates a handler with the required ports.
func NewFreezeSeasonHandler(seasons SeasonReader, writer SeasonWriter) *FreezeSeasonHandler {
	return &FreezeSeasonHandler{seasons: seasons, writer: writer}
}

// Handle loads the season, transitions it to FREEZING, and persists it.
func (h *FreezeSeasonHandler) Handle(seasonID string) (domain.Season, error) {
	season, err := h.seasons.Get(seasonID)
	if err != nil {
		return domain.Season{}, err
	}

	season, err = domain.Freeze(season)
	if err != nil {
		return domain.Season{}, err
	}

	if err := h.writer.Save(season); err != nil {
		return domain.Season{}, err
	}

	return season, nil
}
