package http

import (
	"encoding/json"
	nethttp "net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
)

// ListDefinitionsHandler returns an HTTP handler for listing mission definitions.
func ListDefinitionsHandler(app *application.ListDefinitionsHandler) nethttp.Handler {
	mux := nethttp.NewServeMux()
	mux.HandleFunc("GET /api/v1/missions/definitions", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		language, err := requestedMissionLanguage(r)
		if err != nil {
			w.WriteHeader(nethttp.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "INVALID_QUERY_PARAMETER"})
			return
		}
		definitions, err := app.Handle(language)
		if err != nil {
			w.WriteHeader(nethttp.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "INTERNAL_ERROR"})
			return
		}
		if len(definitions) > 0 {
			served := definitions[0].ContentLanguage
			homogeneous := true
			for _, definition := range definitions[1:] {
				if definition.ContentLanguage != served {
					homogeneous = false
					break
				}
			}
			if homogeneous {
				w.Header().Set("Content-Language", string(served))
			}
		}
		w.WriteHeader(nethttp.StatusOK)
		_ = json.NewEncoder(w).Encode(definitions)
	})
	return mux
}
