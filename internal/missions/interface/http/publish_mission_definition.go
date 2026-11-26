package http

import (
	"encoding/json"
	nethttp "net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
)

// publishMissionDefinitionRequest carries business fields only; caller
// identity is always derived from the authenticated session.
type publishMissionDefinitionRequest struct {
	Version string `json:"version"`
}

// PublishMissionDefinitionHandler returns an HTTP handler for publishing a definition version.
func PublishMissionDefinitionHandler(app *application.PublishMissionDefinitionHandler) nethttp.Handler {
	mux := nethttp.NewServeMux()
	mux.HandleFunc("POST /api/v1/missions/definitions/{definition_code}/publish", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		var req publishMissionDefinitionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(nethttp.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "BAD_REQUEST"})
			return
		}

		_, idErr := CallerIdentityFromContext(r.Context())
		if idErr != nil {
			w.WriteHeader(nethttp.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "UNAUTHORIZED"})
			return
		}

		definitionCode := r.PathValue("definition_code")
		definition, err := app.Handle(definitionCode, req.Version)
		if err != nil {
			w.WriteHeader(nethttp.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
			return
		}

		w.WriteHeader(nethttp.StatusOK)
		_ = json.NewEncoder(w).Encode(definition)
	})
	return mux
}
