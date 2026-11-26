package http

import (
	"encoding/json"
	"errors"
	nethttp "net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// StandingsHistoryHandler serves GET
// /api/v1/missions/seasons/{season_id}/standings/history — the immutable final
// leaderboards of a FINALIZED season; other states return 422
// SEASON_STATE_INVALID.
func StandingsHistoryHandler(app *application.ListStandingsHistoryHandler) nethttp.HandlerFunc {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		history, err := app.Handle(
			r.PathValue("season_id"),
			r.URL.Query().Get("category"),
			r.URL.Query().Get("track"))
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrSeasonNotFound):
				w.WriteHeader(nethttp.StatusNotFound)
			case errors.Is(err, domain.ErrSeasonStateInvalid):
				w.WriteHeader(nethttp.StatusUnprocessableEntity)
			case errors.Is(err, domain.ErrInvalidTrack):
				w.WriteHeader(nethttp.StatusBadRequest)
			default:
				w.WriteHeader(nethttp.StatusInternalServerError)
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(nethttp.StatusOK)
		_ = json.NewEncoder(w).Encode(history)
	}
}
