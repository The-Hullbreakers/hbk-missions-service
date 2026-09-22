package http

import (
	"encoding/json"
	nethttp "net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
)

type evidenceRequest struct {
	ObjectiveID string `json:"objective_id"`
	UnitID      string `json:"unit_id"`
	SeasonID    string `json:"season_id"`
	TenantID    string `json:"tenant_id"`
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

		credit, err := app.Handle(req.TenantID, contractID, req.ObjectiveID, req.UnitID, req.SeasonID)
		if err != nil {
			w.WriteHeader(nethttp.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
			return
		}

		w.WriteHeader(nethttp.StatusCreated)
		_ = json.NewEncoder(w).Encode(credit)
	})
	return mux
}
