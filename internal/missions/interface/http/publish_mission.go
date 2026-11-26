package http

import (
	"encoding/json"
	nethttp "net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
)

// publishMissionRequest carries business fields only; caller identity is
// always derived from the authenticated session, never from the body.
type publishMissionRequest struct {
	OfferID           string `json:"offer_id"`
	DefinitionCode    string `json:"definition_code"`
	DefinitionVersion string `json:"definition_version"`
}

// PublishMissionHandler returns an HTTP handler for publishing mission offers.
func PublishMissionHandler(app *application.PublishMissionHandler) nethttp.Handler {
	mux := nethttp.NewServeMux()
	mux.HandleFunc("POST /api/v1/missions/offers", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		var req publishMissionRequest
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

		offer, err := app.Handle(req.OfferID, req.DefinitionCode, req.DefinitionVersion)
		if err != nil {
			w.WriteHeader(nethttp.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
			return
		}

		w.WriteHeader(nethttp.StatusCreated)
		_ = json.NewEncoder(w).Encode(offer)
	})
	return mux
}
