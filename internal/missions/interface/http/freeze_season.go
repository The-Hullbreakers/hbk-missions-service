package http

import (
	"encoding/json"
	"errors"
	nethttp "net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// FreezeSeasonHandler returns an HTTP handler for freezing a season.
func FreezeSeasonHandler(app *application.FreezeSeasonHandler) nethttp.Handler {
	mux := nethttp.NewServeMux()
	mux.HandleFunc("POST /api/v1/missions/seasons/{id}/freeze", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		season, err := app.Handle(r.PathValue("id"))
		if err != nil {
			writeSeasonError(w, err)
			return
		}

		w.WriteHeader(nethttp.StatusOK)
		_ = json.NewEncoder(w).Encode(season)
	})
	return mux
}

// FinalizeSeasonHandler returns an HTTP handler for finalizing a season.
func FinalizeSeasonHandler(app *application.FinalizeSeasonHandler) nethttp.Handler {
	mux := nethttp.NewServeMux()
	mux.HandleFunc("POST /api/v1/missions/seasons/{id}/finalize", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		season, err := app.Handle(r.PathValue("id"))
		if err != nil {
			writeSeasonError(w, err)
			return
		}

		w.WriteHeader(nethttp.StatusOK)
		_ = json.NewEncoder(w).Encode(season)
	})
	return mux
}

func writeSeasonError(w nethttp.ResponseWriter, err error) {
	status := nethttp.StatusUnprocessableEntity
	if errors.Is(err, domain.ErrSeasonNotFound) {
		status = nethttp.StatusNotFound
	}

	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
}
