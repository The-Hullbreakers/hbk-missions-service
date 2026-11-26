package http

import (
	"encoding/json"
	nethttp "net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// ListContractsHandler returns an HTTP handler for listing the caller's contracts.
func ListContractsHandler(app *application.ListPlayerContractsHandler) nethttp.Handler {
	mux := nethttp.NewServeMux()
	mux.HandleFunc("GET /api/v1/missions/contracts", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		language, err := requestedMissionLanguage(r)
		if err != nil {
			w.WriteHeader(nethttp.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "INVALID_QUERY_PARAMETER"})
			return
		}
		filter, err := domain.ParseContractListFilter(domain.ContractListQuery{
			SeasonID: r.URL.Query().Get("season_id"),
			Status:   r.URL.Query().Get("status"),
		})
		if err != nil {
			w.WriteHeader(nethttp.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
			return
		}

		contracts, err := app.Handle(callerID(r), filter, language)
		if err != nil {
			w.WriteHeader(nethttp.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "INTERNAL_ERROR"})
			return
		}

		if len(contracts) > 0 {
			served := contracts[0].ContentLanguage
			homogeneous := true
			for _, contract := range contracts[1:] {
				if contract.ContentLanguage != served {
					homogeneous = false
					break
				}
			}
			if homogeneous {
				w.Header().Set("Content-Language", string(served))
			}
		}
		items := make([]contractListResponse, 0, len(contracts))
		for _, contract := range contracts {
			items = append(items, toContractListResponse(contract))
		}
		w.WriteHeader(nethttp.StatusOK)
		_ = json.NewEncoder(w).Encode(items)
	})
	return mux
}
