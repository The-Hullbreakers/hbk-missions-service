package http

import (
	"encoding/json"
	"errors"
	nethttp "net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// uploadURLRequest carries the client's declared upload intent; the object
// key is always generated server-side.
type uploadURLRequest struct {
	ContentType   string `json:"content_type"`
	ContentLength int64  `json:"content_length"`
}

// RequestEvidenceUploadHandler returns an HTTP handler for issuing
// contract-scoped evidence upload grants.
func RequestEvidenceUploadHandler(app *application.RequestEvidenceUploadHandler) nethttp.Handler {
	mux := nethttp.NewServeMux()
	mux.HandleFunc("POST /api/v1/missions/contracts/{contract_id}/evidence/upload-url", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		contractID := r.PathValue("contract_id")

		var req uploadURLRequest
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

		grant, err := app.Handle(identity, contractID, req.ContentType, req.ContentLength)
		if err != nil {
			status := nethttp.StatusUnprocessableEntity
			if errors.Is(err, domain.ErrCallerNotAuthorized) {
				status = nethttp.StatusForbidden
			}
			if errors.Is(err, domain.ErrMediaTypeNotAllowed) {
				status = nethttp.StatusBadRequest
			}
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
			return
		}

		w.WriteHeader(nethttp.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"upload_url": grant.URL,
			"media_uri":  string(grant.MediaURI),
		})
	})
	return mux
}
