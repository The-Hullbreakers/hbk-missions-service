package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// SeasonListReader lists seasons according to a business status filter.
type SeasonListReader interface {
	List(filter domain.SeasonStatusFilter) ([]domain.Season, error)
}

// ListSeasonsHandler returns a list of seasons filtered by status.
type ListSeasonsHandler struct {
	reader SeasonListReader
}

// NewListSeasonsHandler creates a handler with the required reader port.
func NewListSeasonsHandler(reader SeasonListReader) *ListSeasonsHandler {
	return &ListSeasonsHandler{reader: reader}
}

// Handle returns the seasons that match the given filter.
func (h *ListSeasonsHandler) Handle(filter domain.SeasonStatusFilter) ([]domain.Season, error) {
	return h.reader.List(filter)
}
