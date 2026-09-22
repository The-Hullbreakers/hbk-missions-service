package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// FinalizeSeasonHandler finalizes a freezing season.
type FinalizeSeasonHandler struct {
	seasons SeasonReader
	writer  SeasonWriter
}

// NewFinalizeSeasonHandler creates a handler with the required ports.
func NewFinalizeSeasonHandler(seasons SeasonReader, writer SeasonWriter) *FinalizeSeasonHandler {
	return &FinalizeSeasonHandler{seasons: seasons, writer: writer}
}

// Handle loads the season, transitions it to FINALIZED, and persists it.
func (h *FinalizeSeasonHandler) Handle(seasonID string) (domain.Season, error) {
	season, err := h.seasons.Get(seasonID)
	if err != nil {
		return domain.Season{}, err
	}

	season, err = domain.Finalize(season)
	if err != nil {
		return domain.Season{}, err
	}

	if err := h.writer.Save(season); err != nil {
		return domain.Season{}, err
	}

	return season, nil
}
