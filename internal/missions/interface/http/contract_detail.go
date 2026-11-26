package http

import (
	"encoding/json"
	"errors"
	nethttp "net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func ContractDetailHandler(app *application.GetContractHandler) nethttp.HandlerFunc {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		identity, err := CallerIdentityFromContext(r.Context())
		if err != nil {
			w.WriteHeader(nethttp.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "UNAUTHORIZED"})
			return
		}

		language, err := requestedMissionLanguage(r)
		if err != nil {
			w.WriteHeader(nethttp.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "INVALID_QUERY_PARAMETER"})
			return
		}
		contract, err := app.Handle(identity, r.PathValue("contract_id"), language)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrCallerNotAuthorized):
				w.WriteHeader(nethttp.StatusForbidden)
			case errors.Is(err, domain.ErrContractNil):
				w.WriteHeader(nethttp.StatusNotFound)
			default:
				w.WriteHeader(nethttp.StatusInternalServerError)
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
			return
		}

		detail := toContractDetail(&contract.MissionContract)
		detail.ContentLanguage = string(contract.ContentLanguage)
		if !isAdjudicator(identity) {
			detail = memberContractDetail(detail, identity.PlayerID())
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Language", string(contract.ContentLanguage))
		w.WriteHeader(nethttp.StatusOK)
		_ = json.NewEncoder(w).Encode(detail)
	}
}

// isAdjudicator reports whether the caller holds a review role.
func isAdjudicator(identity domain.CallerIdentity) bool {
	return identity.HasRole("admin") || identity.HasRole("moderator")
}
