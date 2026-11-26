package http

import (
	"encoding/json"
	nethttp "net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// ListOfferFacetsHandler returns an HTTP handler for the board facet values.
func ListOfferFacetsHandler(app *application.ListOfferFacetsHandler) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		q := r.URL.Query()
		filter, err := domain.ParseOfferListFilter(domain.OfferListQuery{
			SeasonID: q.Get("season_id"),
			Status:   q.Get("status"),
		})
		if err != nil {
			w.WriteHeader(nethttp.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
			return
		}

		facets, err := app.Handle(filter)
		if err != nil {
			writeListOffersError(w, err)
			return
		}

		w.WriteHeader(nethttp.StatusOK)
		_ = json.NewEncoder(w).Encode(facets)
	})
}
