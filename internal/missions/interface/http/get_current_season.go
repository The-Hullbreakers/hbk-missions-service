package http

import (
	"encoding/json"
	"errors"
	nethttp "net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// GetCurrentSeasonHandler serves GET /api/v1/missions/seasons/current — the
// single active season; none returns 404 NO_ACTIVE_SEASON and a broken
// single-active invariant returns 422.
func GetCurrentSeasonHandler(app *application.GetCurrentSeasonHandler) nethttp.HandlerFunc {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		season, err := app.Handle()
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrNoActiveSeason):
				w.WriteHeader(nethttp.StatusNotFound)
			case errors.Is(err, domain.ErrActiveSeasonAlreadyExists):
				w.WriteHeader(nethttp.StatusUnprocessableEntity)
			default:
				w.WriteHeader(nethttp.StatusInternalServerError)
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(nethttp.StatusOK)
		_ = json.NewEncoder(w).Encode(season)
	}
}
