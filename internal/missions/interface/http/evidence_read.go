package http

import (
	"encoding/json"
	"errors"
	nethttp "net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// mediaURLRequest carries the media reference a reviewer wants to read.
type mediaURLRequest struct {
	MediaURI string `json:"media_uri"`
}

// RequestEvidenceReadHandler returns an HTTP handler for issuing
// reviewer-only evidence read grants.
func RequestEvidenceReadHandler(app *application.RequestEvidenceReadHandler) nethttp.Handler {
	mux := nethttp.NewServeMux()
	mux.HandleFunc("POST /api/v1/missions/contracts/{contract_id}/evidence/media-url", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		contractID := r.PathValue("contract_id")

		var req mediaURLRequest
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

		grant, err := app.Handle(identity, contractID, req.MediaURI)
		if err != nil {
			status := nethttp.StatusUnprocessableEntity
			if errors.Is(err, domain.ErrCallerNotAuthorized) {
				status = nethttp.StatusForbidden
			}
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
			return
		}

		w.WriteHeader(nethttp.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"download_url": grant.URL})
	})
	return mux
}
