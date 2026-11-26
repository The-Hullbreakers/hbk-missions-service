package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

const overallDefaultLimit = 10
const overallMaxLimit = 50

type overallStandingsResponse struct {
	SeasonID     string                     `json:"season_id"`
	Track        domain.Track               `json:"track"`
	Participants int                        `json:"participants"`
	Items        []leaderboardEntryResponse `json:"items"`
	NextCursor   string                     `json:"next_cursor"`
	MyStanding   *leaderboardEntryResponse  `json:"my_standing"`
}

// OverallStandingsHandler serves a member's season-wide individual ranking.
func OverallStandingsHandler(app *application.GetOverallStandingsHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, err := CallerIdentityFromContext(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "UNAUTHORIZED"})
			return
		}
		limit := overallDefaultLimit
		if raw := r.URL.Query().Get("limit"); raw != "" {
			parsed, parseErr := strconv.Atoi(raw)
			if parseErr != nil || parsed < 1 || parsed > overallMaxLimit {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"code": "INVALID_QUERY_PARAMETER"})
				return
			}
			limit = parsed
		}
		result, err := app.Handle(identity, r.PathValue("season_id"), limit, r.URL.Query().Get("cursor"))
		if err != nil {
			status := http.StatusInternalServerError
			switch {
			case errors.Is(err, domain.ErrSeasonNotFound):
				status = http.StatusNotFound
			case errors.Is(err, domain.ErrInvalidCursor):
				status = http.StatusBadRequest
			}
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
			return
		}
		page := toLeaderboardResponse(result.Board)
		var mine *leaderboardEntryResponse
		if result.MyStanding != nil {
			standing := result.MyStanding
			mine = &leaderboardEntryResponse{
				Rank: standing.Rank, RecipientID: standing.RecipientID,
				Track: string(standing.Track), Total: standing.Total,
				CompletedMissions: standing.CompletedMissions,
				UniqueObjectives:  standing.UniqueObjectives,
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(overallStandingsResponse{
			SeasonID: result.SeasonID, Track: domain.TrackIndividual,
			Participants: result.Participants, Items: page.Items,
			NextCursor: page.NextCursor, MyStanding: mine,
		})
	}
}
