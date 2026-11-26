package http

import (
	"encoding/json"
	"errors"
	nethttp "net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// acceptMissionRequest carries option fields only; player identity
// are always derived from the authenticated session, never from the body.
type acceptMissionRequest struct {
	TeamID    string   `json:"team_id"`
	Teammates []string `json:"teammates"`
}

// AcceptMissionHandler returns an HTTP handler for accepting mission offers.
func AcceptMissionHandler(app *application.AcceptMissionHandler) nethttp.Handler {
	mux := nethttp.NewServeMux()
	mux.HandleFunc("POST /api/v1/missions/offers/{offer_id}/accept", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		offerID := r.PathValue("offer_id")

		var req acceptMissionRequest
		if r.Body != nil && r.ContentLength != 0 {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(nethttp.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"code": "BAD_REQUEST"})
				return
			}
		}

		identity, err := CallerIdentityFromContext(r.Context())
		if err != nil {
			w.WriteHeader(nethttp.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "UNAUTHORIZED"})
			return
		}

		contract, err := app.Handle(identity, offerID, req.TeamID, req.Teammates)
		if err != nil {
			status := nethttp.StatusUnprocessableEntity
			if errors.Is(err, domain.ErrMembershipUnavailable) {
				status = nethttp.StatusServiceUnavailable
			}
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
			return
		}

		w.WriteHeader(nethttp.StatusCreated)
		_ = json.NewEncoder(w).Encode(contract)
	})
	return mux
}
