package http

import (
	"encoding/json"
	nethttp "net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

type createMissionDefinitionRequest struct {
	TenantID     string `json:"tenant_id"`
	DefinitionID string `json:"definition_id"`
	Category     string `json:"category"`
	Capacity     int    `json:"capacity"`
	Mode         string `json:"mode"`
}

// CreateMissionDefinitionHandler returns an HTTP handler for creating mission definitions.
func CreateMissionDefinitionHandler(app *application.CreateMissionDefinitionHandler) nethttp.Handler {
	mux := nethttp.NewServeMux()
	mux.HandleFunc("POST /api/v1/missions/definitions", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		var req createMissionDefinitionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(nethttp.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "BAD_REQUEST"})
			return
		}

		mode := domain.MissionMode(req.Mode)
		if mode == "" {
			mode = domain.MissionModeSolo
		}

		definition, err := app.Handle(req.TenantID, req.DefinitionID, req.Category, req.Capacity, mode)
		if err != nil {
			w.WriteHeader(nethttp.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
			return
		}

		w.WriteHeader(nethttp.StatusCreated)
		_ = json.NewEncoder(w).Encode(definition)
	})
	return mux
}
