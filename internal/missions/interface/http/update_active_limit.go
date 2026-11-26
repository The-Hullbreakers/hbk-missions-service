package http

import (
	"encoding/json"
	"errors"
	nethttp "net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// updateActiveLimitRequest is the PUT body for the active-contract limit.
type updateActiveLimitRequest struct {
	Default   int            `json:"default"`
	Overrides map[string]int `json:"overrides"`
}

// UpdateActiveLimitHandler serves PUT
// /api/v1/missions/policies/active-contract-limit — admin-only tuning of the
// global limit plus optional per-category overrides; non-positive values
// return 422 INVALID_POLICY.
func UpdateActiveLimitHandler(app *application.UpdateActiveLimitHandler) nethttp.HandlerFunc {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		var req updateActiveLimitRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(nethttp.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "INVALID_REQUEST"})
			return
		}

		if err := app.Handle(req.Default, req.Overrides); err != nil {
			switch {
			case errors.Is(err, domain.ErrInvalidPolicy):
				w.WriteHeader(nethttp.StatusUnprocessableEntity)
			default:
				w.WriteHeader(nethttp.StatusInternalServerError)
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
			return
		}

		w.WriteHeader(nethttp.StatusNoContent)
	}
}
