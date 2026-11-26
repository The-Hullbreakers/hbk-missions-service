package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// GetCurrentSeasonHandler serves the single-active-season read; zero actives
// surface NO_ACTIVE_SEASON and multiples surface the invariant breach.
type GetCurrentSeasonHandler struct {
	seasons ActiveSeasonsReader
}

// NewGetCurrentSeasonHandler creates a handler bound to the seasons reader.
func NewGetCurrentSeasonHandler(seasons ActiveSeasonsReader) *GetCurrentSeasonHandler {
	return &GetCurrentSeasonHandler{seasons: seasons}
}

// Handle returns the single ACTIVE season.
func (h *GetCurrentSeasonHandler) Handle() (domain.Season, error) {
	active, err := h.seasons.ListActive()
	if err != nil {
		return domain.Season{}, err
	}
	if len(active) == 0 {
		return domain.Season{}, domain.ErrNoActiveSeason
	}
	if len(active) > 1 {
		return domain.Season{}, domain.ErrActiveSeasonAlreadyExists
	}
	return active[0], nil
}
