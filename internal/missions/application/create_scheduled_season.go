package application

import (
	"time"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// CreateScheduledSeasonCommand carries the parameters for a scheduled season.
type CreateScheduledSeasonCommand struct {
	SeriesPrefix string
	Iteration    int
	StartsAt     time.Time
	ActiveWeeks  int
	OffWeeks     int
	ActiveDays   int
	OffDays      int
	Official     bool
}

// CreateScheduledSeasonHandler creates a new scheduled season.
type CreateScheduledSeasonHandler struct {
	reader  SeasonListReader
	creator SeasonCreator
}

// NewCreateScheduledSeasonHandler creates a handler with the required ports.
func NewCreateScheduledSeasonHandler(reader SeasonListReader, creator SeasonCreator) *CreateScheduledSeasonHandler {
	return &CreateScheduledSeasonHandler{reader: reader, creator: creator}
}

// Handle validates and persists a new scheduled season.
func (h *CreateScheduledSeasonHandler) Handle(cmd CreateScheduledSeasonCommand) (domain.Season, error) {
	var season domain.Season
	var err error
	if cmd.ActiveDays > 0 || cmd.OffDays > 0 {
		season, err = domain.NewScheduledSeasonDays(cmd.SeriesPrefix, cmd.Iteration, cmd.StartsAt, cmd.ActiveDays, cmd.OffDays, cmd.Official)
	} else {
		season, err = domain.NewScheduledSeason(cmd.SeriesPrefix, cmd.Iteration, cmd.StartsAt, cmd.ActiveWeeks, cmd.OffWeeks, cmd.Official)
	}
	if err != nil {
		return domain.Season{}, err
	}
	all, err := h.reader.List(domain.SeasonStatusFilterAll)
	if err != nil {
		return domain.Season{}, err
	}
	if err := domain.ValidateCanCreateSeason(season.SeasonID, all); err != nil {
		return domain.Season{}, err
	}
	if err := h.creator.Create(season); err != nil {
		return domain.Season{}, err
	}
	return season, nil
}
