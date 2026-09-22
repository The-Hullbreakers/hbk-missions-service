package http

import (
	"encoding/json"
	nethttp "net/http"
	"time"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
)

type createSeasonRequest struct {
	TenantID           string           `json:"tenant_id"`
	SeasonID           string           `json:"season_id"`
	OfficialStatistics *bool            `json:"official_statistics,omitempty"`
	StartsAt           string           `json:"starts_at,omitempty"`
	Schedule           *scheduleRequest `json:"schedule,omitempty"`
}

type scheduleRequest struct {
	ActiveWeeks int           `json:"active_weeks"`
	OffWeeks    int           `json:"off_weeks"`
	Series      seriesRequest `json:"series"`
}

type seriesRequest struct {
	Prefix    string `json:"prefix"`
	Iteration int    `json:"iteration"`
}

// CreateSeasonHandler returns an HTTP handler for creating a season.
func CreateSeasonHandler(app *application.CreateSeasonHandler, scheduledApp *application.CreateScheduledSeasonHandler) nethttp.Handler {
	mux := nethttp.NewServeMux()
	mux.HandleFunc("POST /api/v1/seasons", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		var req createSeasonRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(nethttp.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "BAD_REQUEST"})
			return
		}

		official := true
		if req.OfficialStatistics != nil {
			official = *req.OfficialStatistics
		}

		var season any
		var err error
		if req.Schedule == nil {
			season, err = app.Handle(req.TenantID, req.SeasonID, official)
		} else {
			startsAt, parseErr := time.Parse(time.RFC3339, req.StartsAt)
			if parseErr != nil {
				w.WriteHeader(nethttp.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"code": "BAD_REQUEST"})
				return
			}
			season, err = scheduledApp.Handle(application.CreateScheduledSeasonCommand{
				TenantID:     req.TenantID,
				SeriesPrefix: req.Schedule.Series.Prefix,
				Iteration:    req.Schedule.Series.Iteration,
				StartsAt:     startsAt,
				ActiveWeeks:  req.Schedule.ActiveWeeks,
				OffWeeks:     req.Schedule.OffWeeks,
				Official:     official,
			})
		}

		if err != nil {
			w.WriteHeader(nethttp.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
			return
		}

		w.WriteHeader(nethttp.StatusCreated)
		_ = json.NewEncoder(w).Encode(season)
	})
	return mux
}
