package http

import (
	"encoding/json"
	"errors"
	nethttp "net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// evidenceRequest carries submission fields only; caller identity
// are always derived from the authenticated session, never from the body.
type evidenceRequest struct {
	ObjectiveID string `json:"objective_id"`
	UnitID      string `json:"unit_id"`
	SeasonID    string `json:"season_id"`
	MediaURI    string `json:"media_uri"`
}

// ConfirmEvidenceHandler returns an HTTP handler for confirming objective evidence.
func ConfirmEvidenceHandler(app *application.ConfirmObjectiveHandler) nethttp.Handler {
	mux := nethttp.NewServeMux()
	mux.HandleFunc("POST /api/v1/missions/contracts/{contract_id}/evidence", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		contractID := r.PathValue("contract_id")

		var req evidenceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(nethttp.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "BAD_REQUEST"})
			return
		}

		identity, err := CallerIdentityFromContext(r.Context())
		if err != nil {
			w.WriteHeader(nethttp.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "UNAUTHORIZED"})
			return
		}

		credit, err := app.HandleWithMedia(identity, contractID, req.ObjectiveID, req.UnitID, req.MediaURI, req.SeasonID)
		if err != nil {
			status := nethttp.StatusUnprocessableEntity
			if errors.Is(err, domain.ErrCallerNotAuthorized) {
				status = nethttp.StatusForbidden
			}
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
			return
		}

		w.WriteHeader(nethttp.StatusCreated)
		_ = json.NewEncoder(w).Encode(credit)
	})
	return mux
}
