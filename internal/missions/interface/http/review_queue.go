package http

import (
	"encoding/json"
	"errors"
	nethttp "net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// reviewQueueEntry is the response projection a reviewer consumes to size and
// select adjudication work — member summary plus review-relevant counts.
type reviewQueueEntry struct {
	ContractID      string   `json:"contract_id"`
	OfferID         string   `json:"offer_id"`
	DefinitionCode  string   `json:"definition_code"`
	Title           string   `json:"title"`
	ContentLanguage string   `json:"content_language"`
	State           string   `json:"state"`
	Roster          []string `json:"roster"`
	SeasonID        string   `json:"season_id"`
	TeamID          string   `json:"team_id"`
	ConfirmedCount  int      `json:"confirmed_count"`
	PendingCredits  int      `json:"pending_credits"`
}

func toReviewQueueEntry(c application.LocalizedContract) reviewQueueEntry {
	return reviewQueueEntry{
		ContractID:      c.ContractID,
		OfferID:         c.Offer.OfferID,
		DefinitionCode:  c.Offer.DefinitionCode,
		Title:           c.Offer.Title,
		ContentLanguage: string(c.ContentLanguage),
		State:           string(c.State),
		Roster:          c.Roster(),
		SeasonID:        c.SeasonID,
		TeamID:          c.TeamID,
		ConfirmedCount:  len(c.Confirmations()),
		PendingCredits:  len(c.PendingCredits()),
	}
}

// ReviewQueueHandler serves GET /api/v1/missions/review-queue — the
// adjudicator-only listing of PENDING_VALIDATION contracts.
func ReviewQueueHandler(app *application.ListPendingReviewHandler) nethttp.HandlerFunc {
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

		contracts, err := app.Handle(identity, language)
		if err != nil {
			if errors.Is(err, domain.ErrCallerNotAuthorized) {
				w.WriteHeader(nethttp.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
				return
			}
			w.WriteHeader(nethttp.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "INTERNAL_ERROR"})
			return
		}

		entries := make([]reviewQueueEntry, 0, len(contracts))
		for _, contract := range contracts {
			entries = append(entries, toReviewQueueEntry(contract))
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(nethttp.StatusOK)
		_ = json.NewEncoder(w).Encode(entries)
	}
}
