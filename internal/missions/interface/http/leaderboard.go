package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// LeaderboardHandler returns an HTTP handler for listing leaderboards.
func LeaderboardHandler(app *application.ListLeaderboardHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/missions/seasons/{season_id}/leaderboards/{category}", func(w http.ResponseWriter, r *http.Request) {
		if _, idErr := CallerIdentityFromContext(r.Context()); idErr != nil {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "UNAUTHORIZED"})
			return
		}

		seasonID := r.PathValue("season_id")
		category := r.PathValue("category")
		track := r.URL.Query().Get("track")

		limit := 20
		if raw := r.URL.Query().Get("limit"); raw != "" {
			if v, err := strconv.Atoi(raw); err == nil && v > 0 {
				if v > 100 {
					v = 100
				}
				limit = v
			}
		}

		minMerits := 0
		if raw := r.URL.Query().Get("min_merits"); raw != "" {
			if v, err := strconv.Atoi(raw); err == nil && v >= 0 {
				minMerits = v
			}
		}

		cursor := r.URL.Query().Get("cursor")

		leaderboard, err := app.Handle(seasonID, category, track, limit, minMerits, cursor)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, domain.ErrInvalidTrack) || errors.Is(err, domain.ErrInvalidCursor) {
				status = http.StatusBadRequest
			}
			if errors.Is(err, domain.ErrSeasonNotFound) {
				status = http.StatusNotFound
			}
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(toLeaderboardResponse(leaderboard))
	})
	return mux
}

type leaderboardResponse struct {
	Items      []leaderboardEntryResponse `json:"items"`
	NextCursor string                     `json:"next_cursor"`
}

type leaderboardEntryResponse struct {
	Rank              int    `json:"rank"`
	RecipientID       string `json:"recipient_id"`
	Track             string `json:"track"`
	Total             int    `json:"total"`
	CompletedMissions int    `json:"completed_missions"`
	UniqueObjectives  int    `json:"unique_objectives"`
}

func toLeaderboardResponse(leaderboard domain.Leaderboard) leaderboardResponse {
	items := make([]leaderboardEntryResponse, len(leaderboard.Entries))
	for i, e := range leaderboard.Entries {
		items[i] = leaderboardEntryResponse{
			Rank:              e.Rank,
			RecipientID:       e.RecipientID,
			Track:             string(e.Track),
			Total:             e.Total,
			CompletedMissions: e.CompletedMissions,
			UniqueObjectives:  e.UniqueObjectives,
		}
	}
	return leaderboardResponse{
		Items:      items,
		NextCursor: leaderboard.NextCursor,
	}
}
