package http

import (
	"encoding/json"
	"errors"
	nethttp "net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// GetMyStandingHandler serves GET
// /api/v1/missions/seasons/{season_id}/leaderboards/{category}/me — the
// caller's own standing; absence returns 404 STANDING_NOT_FOUND.
func GetMyStandingHandler(app *application.GetMyStandingHandler) nethttp.HandlerFunc {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		identity, err := CallerIdentityFromContext(r.Context())
		if err != nil {
			w.WriteHeader(nethttp.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "UNAUTHORIZED"})
			return
		}

		entry, err := app.Handle(identity, r.PathValue("season_id"), r.PathValue("category"), r.URL.Query().Get("track"))
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrStandingNotFound), errors.Is(err, domain.ErrSeasonNotFound):
				w.WriteHeader(nethttp.StatusNotFound)
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
		_ = json.NewEncoder(w).Encode(entry)
	}
}
