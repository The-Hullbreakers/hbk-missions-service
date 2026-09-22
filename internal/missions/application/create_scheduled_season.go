package application

import (
	"time"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// CreateScheduledSeasonCommand carries the parameters for a scheduled season.
type CreateScheduledSeasonCommand struct {
	TenantID     string
	SeriesPrefix string
	Iteration    int
	StartsAt     time.Time
	ActiveWeeks  int
	OffWeeks     int
	Official     bool
}

// CreateScheduledSeasonHandler creates a new scheduled season.
type CreateScheduledSeasonHandler struct {
	activeSeasons ActiveSeasonsReader
	seasons       SeasonWriter
}

// NewCreateScheduledSeasonHandler creates a handler with the required ports.
func NewCreateScheduledSeasonHandler(activeSeasons ActiveSeasonsReader, seasons SeasonWriter) *CreateScheduledSeasonHandler {
	return &CreateScheduledSeasonHandler{activeSeasons: activeSeasons, seasons: seasons}
}

// Handle validates and persists a new scheduled season.
func (h *CreateScheduledSeasonHandler) Handle(cmd CreateScheduledSeasonCommand) (domain.Season, error) {
	active, err := h.activeSeasons.ListActive()
	if err != nil {
		return domain.Season{}, err
	}
	if err := domain.ValidateCanCreateSeason(active); err != nil {
		return domain.Season{}, err
	}

	season, err := domain.NewScheduledSeason(cmd.SeriesPrefix, cmd.Iteration, cmd.StartsAt, cmd.ActiveWeeks, cmd.OffWeeks, cmd.Official)
	if err != nil {
		return domain.Season{}, err
	}
	if err := h.seasons.Save(season); err != nil {
		return domain.Season{}, err
	}

	return season, nil
}
