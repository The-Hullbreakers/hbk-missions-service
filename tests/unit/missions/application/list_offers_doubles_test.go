package application_test

import (
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

type offerListReaderStub struct {
	offers    []domain.MissionOffer
	err       error
	gotFilter domain.OfferListFilter
}

func (s *offerListReaderStub) List(filter domain.OfferListFilter) ([]domain.MissionOffer, error) {
	s.gotFilter = filter
	return s.offers, s.err
}

// playerContractsReaderStub records the history query; ListByPlayer keeps
// roster history (departed members still match).
type playerContractsReaderStub struct {
	contracts []domain.MissionContract
	err       error
	gotPlayer string
	gotFilter domain.ContractListFilter
}

func (s *playerContractsReaderStub) ListByPlayer(playerID string, filter domain.ContractListFilter) ([]domain.MissionContract, error) {
	s.gotPlayer = playerID
	s.gotFilter = filter
	var out []domain.MissionContract
	for _, c := range s.contracts {
		if c.IncludesPlayer(playerID) && filter.Match(c) {
			out = append(out, c)
		}
	}
	return out, s.err
}

// boardContractsStub mirrors the store's active-membership semantics:
// ListActiveByPlayer returns non-terminal contracts where the player is an
// active member (departed roster members do not match).
type boardContractsStub struct {
	contracts       []domain.MissionContract
	err             error
	gotActivePlayer string
	activeCalls     int
}

func (s *boardContractsStub) ListActiveByPlayer(playerID string) ([]domain.MissionContract, error) {
	s.gotActivePlayer = playerID
	s.activeCalls++
	var out []domain.MissionContract
	for _, c := range s.contracts {
		if !c.State.IsTerminal() && c.IncludesPlayer(playerID) && !c.IsDeparted(playerID) {
			out = append(out, c)
		}
	}
	return out, s.err
}

type activeSeasonsStub struct {
	seasons []domain.Season
	err     error
}

func (s *activeSeasonsStub) ListActive() ([]domain.Season, error) {
	return s.seasons, s.err
}
