package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

type fakeOfferListReader struct {
	offers []domain.MissionOffer
}

func (r *fakeOfferListReader) List(filter domain.OfferListFilter) ([]domain.MissionOffer, error) {
	var out []domain.MissionOffer
	for _, offer := range r.offers {
		if filter.Match(offer) {
			out = append(out, offer)
		}
	}
	return out, nil
}

type fakePlayerContractsReader struct{}

func (r *fakePlayerContractsReader) ListByPlayer(playerID string, filter domain.ContractListFilter) ([]domain.MissionContract, error) {
	return nil, nil
}

func (r *fakePlayerContractsReader) ListActiveByPlayer(playerID string) ([]domain.MissionContract, error) {
	return nil, nil
}

type fakeActiveSeasons struct {
	seasons []domain.Season
}

func (s *fakeActiveSeasons) ListActive() ([]domain.Season, error) {
	return s.seasons, nil
}

type offerListEnvelope struct {
	Offers []map[string]any `json:"offers"`
	Total  int              `json:"total"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
}

func decodeOfferPage(t *testing.T, rec *httptest.ResponseRecorder) offerListEnvelope {
	t.Helper()
	var result offerListEnvelope
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&result))
	return result
}

func newListOffersHandler() http.Handler {
	app := application.NewListOffersHandler(
		&fakeOfferListReader{offers: []domain.MissionOffer{
			domain.NewMissionOffer("offer-1", "BOUNTY", "season-2026-q4"),
			domain.NewMissionOffer("offer-2", "EXPLORATION", "season-2026-q4"),
		}},
		&fakePlayerContractsReader{},
		&fakeActiveSeasons{seasons: []domain.Season{domain.NewSeason("season-2026-q4", domain.SeasonStateActive)}},
	)
	return missionhttp.ListOffersHandler(app)
}
