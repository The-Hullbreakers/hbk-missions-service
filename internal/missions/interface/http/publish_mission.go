package http

import (
	"encoding/json"
	nethttp "net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
)

type publishMissionRequest struct {
	TenantID          string `json:"tenant_id"`
	OfferID           string `json:"offer_id"`
	DefinitionID      string `json:"definition_id"`
	DefinitionVersion string `json:"definition_version"`
	Capacity          int    `json:"capacity"`
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

		offer, err := app.Handle(req.TenantID, req.OfferID, req.DefinitionID, req.DefinitionVersion, req.Capacity)
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
