package http

import (
	"encoding/json"
	nethttp "net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
)

type acceptMissionRequest struct {
	TenantID string `json:"tenant_id"`
	PlayerID string `json:"player_id"`
	TeamID   string `json:"team_id"`
}

// AcceptMissionHandler returns an HTTP handler for accepting mission offers.
func AcceptMissionHandler(app *application.AcceptMissionHandler) nethttp.Handler {
	mux := nethttp.NewServeMux()
	mux.HandleFunc("POST /api/v1/missions/offers/{offer_id}/accept", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		offerID := r.PathValue("offer_id")

		var req acceptMissionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(nethttp.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "BAD_REQUEST"})
			return
		}

		contract, err := app.Handle(req.TenantID, req.PlayerID, offerID, req.TeamID)
		if err != nil {
			w.WriteHeader(nethttp.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
			return
		}

		w.WriteHeader(nethttp.StatusCreated)
		_ = json.NewEncoder(w).Encode(contract)
	})
	return mux
}
