package http

import (
	"encoding/json"
	"errors"
	nethttp "net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// ListOffersHandler returns an HTTP handler for the mission board discovery.
func ListOffersHandler(app *application.ListOffersHandler) nethttp.Handler {
	mux := nethttp.NewServeMux()
	mux.HandleFunc("GET /api/v1/missions/offers", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		language, err := requestedMissionLanguage(r)
		if err != nil {
			w.WriteHeader(nethttp.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "INVALID_QUERY_PARAMETER"})
			return
		}
		filter, err := domain.ParseOfferListFilter(offerListQuery(r))
		if err != nil {
			w.WriteHeader(nethttp.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
			return
		}

		page, err := app.Handle(callerID(r), filter, language)
		if err != nil {
			writeListOffersError(w, err)
			return
		}

		if len(page.Offers) > 0 {
			served := page.Offers[0].ContentLanguage
			homogeneous := true
			for _, offer := range page.Offers[1:] {
				if offer.ContentLanguage != served {
					homogeneous = false
					break
				}
			}
			if homogeneous {
				w.Header().Set("Content-Language", string(served))
			}
		}
		w.WriteHeader(nethttp.StatusOK)
		_ = json.NewEncoder(w).Encode(toOffersResponse(page))
	})
	return mux
}

// offerListQuery reads every supported list query parameter.
func offerListQuery(r *nethttp.Request) domain.OfferListQuery {
	q := r.URL.Query()
	return domain.OfferListQuery{
		SeasonID: q.Get("season_id"),
		Status:   q.Get("status"),
		Category: q.Get("category"),
		Mode:     q.Get("mode"),
		Q:        q.Get("q"),
		Limit:    q.Get("limit"),
		Offset:   q.Get("offset"),
		Sort:     q.Get("sort"),
	}
}

// writeListOffersError maps known domain rejections to 422 with their public
// code; unrecognized failures remain 500.
func writeListOffersError(w nethttp.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNoActiveSeason),
		errors.Is(err, domain.ErrActiveSeasonAlreadyExists):
		w.WriteHeader(nethttp.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
	default:
		w.WriteHeader(nethttp.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"code": "INTERNAL_ERROR"})
	}
}

func callerID(r *nethttp.Request) string {
	if v, ok := r.Context().Value(callerIDKey).(string); ok {
		return v
	}
	return ""
}
