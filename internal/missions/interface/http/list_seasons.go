package http

import (
	"encoding/json"
	nethttp "net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// ListSeasonsHandler returns an HTTP handler for listing seasons.
func ListSeasonsHandler(app *application.ListSeasonsHandler) nethttp.Handler {
	mux := nethttp.NewServeMux()
	mux.HandleFunc("GET /api/v1/seasons", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		filter, err := domain.ParseSeasonStatusFilter(r.URL.Query().Get("status"))
		if err != nil {
			w.WriteHeader(nethttp.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
			return
		}

		seasons, err := app.Handle(filter)
		if err != nil {
			w.WriteHeader(nethttp.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "INTERNAL_ERROR"})
			return
		}

		w.WriteHeader(nethttp.StatusOK)
		_ = json.NewEncoder(w).Encode(seasons)
	})
	return mux
}
