package http

import (
	"encoding/json"
	"errors"
	nethttp "net/http"
	"strconv"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

const (
	defaultStatisticsTopN = 3
	maxStatisticsTopN     = 50
)

// errInvalidTopN marks a malformed or out-of-range top_n query parameter.
var errInvalidTopN = errors.New("invalid top_n")

// SeasonStatisticsHandler serves GET
// /api/v1/missions/seasons/{season_id}/statistics — the per-season
// dashboard summary grouped by category and track, for any season state.
func SeasonStatisticsHandler(app *application.GetSeasonStatisticsHandler) nethttp.HandlerFunc {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		topN, err := parseTopN(r.URL.Query().Get("top_n"))
		if err != nil {
			w.WriteHeader(nethttp.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "INVALID_QUERY_PARAMETER"})
			return
		}

		stats, err := app.Handle(
			r.PathValue("season_id"),
			r.URL.Query().Get("category"),
			r.URL.Query().Get("track"),
			topN)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrSeasonNotFound),
				errors.Is(err, domain.ErrNoActiveSeason):
				w.WriteHeader(nethttp.StatusNotFound)
			case errors.Is(err, domain.ErrActiveSeasonAlreadyExists):
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
		_ = json.NewEncoder(w).Encode(stats)
	}
}

// parseTopN validates the optional top_n parameter: an integer between 1
// and maxStatisticsTopN, defaulting to defaultStatisticsTopN.
func parseTopN(raw string) (int, error) {
	if raw == "" {
		return defaultStatisticsTopN, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > maxStatisticsTopN {
		return 0, errInvalidTopN
	}
	return n, nil
}
