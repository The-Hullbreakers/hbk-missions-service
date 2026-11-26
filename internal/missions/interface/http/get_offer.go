package http

import (
	"encoding/json"
	"errors"
	nethttp "net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// GetOfferHandler serves GET /api/v1/missions/offers/{offer_id} — the member
// single-offer read; unknown ids return 404 OFFER_NOT_AVAILABLE.
func GetOfferHandler(app *application.GetOfferHandler) nethttp.HandlerFunc {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		language, err := requestedMissionLanguage(r)
		if err != nil {
			w.WriteHeader(nethttp.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "INVALID_QUERY_PARAMETER"})
			return
		}
		offer, err := app.Handle(r.PathValue("offer_id"), language)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrOfferNotAvailable):
				w.WriteHeader(nethttp.StatusNotFound)
			default:
				w.WriteHeader(nethttp.StatusInternalServerError)
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Language", string(offer.ContentLanguage))
		w.WriteHeader(nethttp.StatusOK)
		_ = json.NewEncoder(w).Encode(toOfferResponse(offer))
	}
}
