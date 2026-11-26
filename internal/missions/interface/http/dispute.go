package http

import (
	"encoding/json"
	"errors"
	nethttp "net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// disputeRequest carries the raise-dispute body.
type disputeRequest struct {
	ObjectiveID string `json:"objective_id"`
	Reason      string `json:"reason"`
}

// RaiseDisputeHandler returns an HTTP handler for POST
// /api/v1/missions/contracts/{contract_id}/disputes.
func RaiseDisputeHandler(app *application.RaiseDisputeHandler) nethttp.Handler {
	mux := nethttp.NewServeMux()
	mux.HandleFunc("POST /api/v1/missions/contracts/{contract_id}/disputes", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		var req disputeRequest
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
		dispute, err := app.Handle(identity, r.PathValue("contract_id"), req.ObjectiveID, req.Reason)
		if err != nil {
			status := nethttp.StatusUnprocessableEntity
			switch {
			case errors.Is(err, domain.ErrCallerNotAuthorized):
				status = nethttp.StatusForbidden
			case errors.Is(err, domain.ErrContractNil):
				status = nethttp.StatusNotFound
			}
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
			return
		}
		w.WriteHeader(nethttp.StatusCreated)
		_ = json.NewEncoder(w).Encode(dispute)
	})
	return mux
}

// ResolveDisputeHandler returns an HTTP handler for POST
// /api/v1/missions/contracts/{contract_id}/disputes/{dispute_id}/resolve.
func ResolveDisputeHandler(app *application.ResolveDisputeHandler) nethttp.Handler {
	mux := nethttp.NewServeMux()
	mux.HandleFunc("POST /api/v1/missions/contracts/{contract_id}/disputes/{dispute_id}/resolve", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		identity, err := CallerIdentityFromContext(r.Context())
		if err != nil {
			w.WriteHeader(nethttp.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "UNAUTHORIZED"})
			return
		}
		err = app.Handle(identity, r.PathValue("contract_id"), r.PathValue("dispute_id"))
		if err != nil {
			status := nethttp.StatusUnprocessableEntity
			switch {
			case errors.Is(err, domain.ErrCallerNotAuthorized):
				status = nethttp.StatusForbidden
			case errors.Is(err, domain.ErrContractNil), errors.Is(err, domain.ErrDisputeNotFound):
				status = nethttp.StatusNotFound
			}
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
			return
		}
		w.WriteHeader(nethttp.StatusOK)
	})
	return mux
}

// ListDisputesHandler returns an HTTP handler for GET
// /api/v1/missions/contracts/{contract_id}/disputes (reviewer only).
func ListDisputesHandler(app *application.ListDisputesHandler) nethttp.HandlerFunc {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		identity, err := CallerIdentityFromContext(r.Context())
		if err != nil {
			w.WriteHeader(nethttp.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "UNAUTHORIZED"})
			return
		}
		disputes, err := app.Handle(identity, r.PathValue("contract_id"))
		if err != nil {
			status := nethttp.StatusUnprocessableEntity
			switch {
			case errors.Is(err, domain.ErrCallerNotAuthorized):
				status = nethttp.StatusForbidden
			case errors.Is(err, domain.ErrContractNil):
				status = nethttp.StatusNotFound
			}
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(nethttp.StatusOK)
		_ = json.NewEncoder(w).Encode(disputes)
	}
}
