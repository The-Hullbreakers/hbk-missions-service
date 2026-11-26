package http

import (
	"encoding/json"
	"errors"
	"io"
	nethttp "net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// StartContractHandler returns an HTTP handler for starting an accepted contract.
func StartContractHandler(app *application.StartContractHandler) nethttp.HandlerFunc {
	return transition(app.Handle)
}

// AbandonContractHandler returns an HTTP handler for abandoning a contract.
func AbandonContractHandler(app *application.AbandonContractHandler) nethttp.HandlerFunc {
	return transition(app.Handle)
}

// LeaveContractHandler returns an HTTP handler for a member leaving a contract.
func LeaveContractHandler(app *application.LeaveContractHandler) nethttp.HandlerFunc {
	return transition(app.Handle)
}

// SubmitContractHandler returns an HTTP handler for submitting a contract for review.
func SubmitContractHandler(app *application.SubmitContractHandler) nethttp.HandlerFunc {
	return transition(app.Handle)
}

// ConfirmInvitationHandler returns an HTTP handler for an invitee confirming
// their team roster invitation.
func ConfirmInvitationHandler(app *application.ConfirmInvitationHandler) nethttp.HandlerFunc {
	return transition(app.Handle)
}

// DeclineInvitationHandler returns an HTTP handler for an invitee declining
// their team roster invitation.
func DeclineInvitationHandler(app *application.DeclineInvitationHandler) nethttp.HandlerFunc {
	return transition(app.Handle)
}

// RejectContractHandler returns an HTTP handler for a reviewer rejecting a pending contract.
func RejectContractHandler(app *application.RejectContractHandler) nethttp.HandlerFunc {
	return transition(app.Handle)
}

// approveRequest carries the reviewer's per-unit verdict; absent or empty
// rejected_units approves every staged unit.
type approveRequest struct {
	RejectedUnits []unitRefRequest `json:"rejected_units"`
}

type unitRefRequest struct {
	ObjectiveID string `json:"objective_id"`
	UnitID      string `json:"unit_id"`
}

// ApproveContractHandler returns an HTTP handler for a reviewer approving a
// pending contract; the committed merit credits are returned in the body.
func ApproveContractHandler(app *application.ApproveContractHandler) nethttp.HandlerFunc {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		identity, err := CallerIdentityFromContext(r.Context())
		if err != nil {
			w.WriteHeader(nethttp.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "UNAUTHORIZED"})
			return
		}

		var req approveRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
			w.WriteHeader(nethttp.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "BAD_REQUEST"})
			return
		}
		rejected := make([]domain.UnitRef, 0, len(req.RejectedUnits))
		for _, unit := range req.RejectedUnits {
			if unit.ObjectiveID == "" || unit.UnitID == "" {
				w.WriteHeader(nethttp.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"code": "BAD_REQUEST"})
				return
			}
			rejected = append(rejected, domain.UnitRef{ObjectiveID: unit.ObjectiveID, UnitID: unit.UnitID})
		}

		credits, err := app.HandleWithVerdict(identity, r.PathValue("contract_id"), rejected)
		if err == nil {
			w.WriteHeader(nethttp.StatusOK)
			_ = json.NewEncoder(w).Encode(credits)
			return
		}
		switch {
		case errors.Is(err, domain.ErrCallerNotAuthorized):
			w.WriteHeader(nethttp.StatusForbidden)
		case errors.Is(err, domain.ErrContractNil):
			w.WriteHeader(nethttp.StatusNotFound)
		default:
			w.WriteHeader(nethttp.StatusUnprocessableEntity)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
	}
}

// transition adapts a lifecycle use case to the HTTP boundary.
func transition(act func(domain.CallerIdentity, string) error) nethttp.HandlerFunc {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		identity, err := CallerIdentityFromContext(r.Context())
		if err != nil {
			w.WriteHeader(nethttp.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "UNAUTHORIZED"})
			return
		}

		err = act(identity, r.PathValue("contract_id"))
		if err == nil {
			w.WriteHeader(nethttp.StatusOK)
			return
		}
		switch {
		case errors.Is(err, domain.ErrCallerNotAuthorized), errors.Is(err, domain.ErrNotInvited):
			w.WriteHeader(nethttp.StatusForbidden)
		case errors.Is(err, domain.ErrContractNil):
			w.WriteHeader(nethttp.StatusNotFound)
		default:
			w.WriteHeader(nethttp.StatusUnprocessableEntity)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
	}
}
