package http

import (
	"encoding/json"
	nethttp "net/http"
	"time"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
)

// createSeasonRequest carries business fields only; caller identity is always
// derived from the authenticated session, never from the body.
type createSeasonRequest struct {
	SeasonID           string           `json:"season_id"`
	OfficialStatistics *bool            `json:"official_statistics,omitempty"`
	StartsAt           string           `json:"starts_at,omitempty"`
	Schedule           *scheduleRequest `json:"schedule,omitempty"`
}

type scheduleRequest struct {
	ActiveWeeks *int          `json:"active_weeks"`
	OffWeeks    *int          `json:"off_weeks"`
	ActiveDays  *int          `json:"active_days"`
	OffDays     *int          `json:"off_days"`
	Series      seriesRequest `json:"series"`
}

type seriesRequest struct {
	Prefix    string `json:"prefix"`
	Iteration int    `json:"iteration"`
}

func intOrZero(v *int) int {
	if v == nil {
		return 0
	}
	return *v
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

		_, idErr := CallerIdentityFromContext(r.Context())
		if idErr != nil {
			w.WriteHeader(nethttp.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "UNAUTHORIZED"})
			return
		}

		var season any
		var err error
		if req.Schedule == nil {
			season, err = app.Handle(req.SeasonID, official)
		} else {
			startsAt, parseErr := time.Parse(time.RFC3339, req.StartsAt)
			if parseErr != nil {
				w.WriteHeader(nethttp.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"code": "BAD_REQUEST"})
				return
			}
			cmd := application.CreateScheduledSeasonCommand{
				SeriesPrefix: req.Schedule.Series.Prefix,
				Iteration:    req.Schedule.Series.Iteration,
				StartsAt:     startsAt,
				Official:     official,
			}
			daysPresent := req.Schedule.ActiveDays != nil || req.Schedule.OffDays != nil
			weeksPresent := req.Schedule.ActiveWeeks != nil || req.Schedule.OffWeeks != nil
			if daysPresent && weeksPresent {
				w.WriteHeader(nethttp.StatusUnprocessableEntity)
				_ = json.NewEncoder(w).Encode(map[string]string{"code": "INVALID_SEASON"})
				return
			}
			if daysPresent {
				cmd.ActiveDays = intOrZero(req.Schedule.ActiveDays)
				cmd.OffDays = intOrZero(req.Schedule.OffDays)
			} else {
				cmd.ActiveWeeks = intOrZero(req.Schedule.ActiveWeeks)
				cmd.OffWeeks = intOrZero(req.Schedule.OffWeeks)
			}
			season, err = scheduledApp.Handle(cmd)
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
