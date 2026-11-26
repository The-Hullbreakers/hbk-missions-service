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

// SeasonCreator persists a new season; the adapter must refuse to overwrite
// an existing season id.
type SeasonCreator interface {
	Create(season domain.Season) error
}

// CreateSeasonHandler creates a new season.
type CreateSeasonHandler struct {
	reader  SeasonListReader
	creator SeasonCreator
}

// NewCreateSeasonHandler creates a handler with the required ports.
func NewCreateSeasonHandler(reader SeasonListReader, creator SeasonCreator) *CreateSeasonHandler {
	return &CreateSeasonHandler{reader: reader, creator: creator}
}

// Handle validates and persists a new season.
func (h *CreateSeasonHandler) Handle(seasonID string, official bool) (domain.Season, error) {
	season, err := domain.NewActiveSeason(seasonID, official)
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
