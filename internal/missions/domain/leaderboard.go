package domain

import (
	"encoding/base64"
	"encoding/json"
	"sort"
)

// LeaderboardEntry is a single row in the leaderboard.
type LeaderboardEntry struct {
	Rank              int
	RecipientID       string
	Track             Track
	Total             int
	CompletedMissions int
	UniqueObjectives  int
}

// Leaderboard is a paginated projection of standings.
type Leaderboard struct {
	Entries    []LeaderboardEntry
	NextCursor string
}

// BuildLeaderboard orders standings by merits and tie-breaks, then paginates.
func BuildLeaderboard(standings []Standing, limit int, cursor string) (Leaderboard, error) {
	if limit <= 0 {
		limit = 20
	}

	var after Standing
	if cursor != "" {
		decoded, err := decodeCursor(cursor)
		if err != nil {
			return Leaderboard{}, ErrInvalidCursor
		}
		after = decoded
	}

	ordered := make([]Standing, len(standings))
	copy(ordered, standings)
	sort.Slice(ordered, func(i, j int) bool {
		return CompareStanding(ordered[i], ordered[j]) < 0
	})

	var start int
	if cursor != "" {
		for i, s := range ordered {
			if CompareStanding(s, after) > 0 {
				start = i
				break
			}
		}
	}

	end := start + limit
	if end > len(ordered) {
		end = len(ordered)
	}
	page := ordered[start:end]

	entries := make([]LeaderboardEntry, len(page))
	for i, s := range page {
		entries[i] = LeaderboardEntry{
			Rank:              start + i + 1,
			RecipientID:       s.RecipientID,
			Track:             s.Track,
			Total:             s.Total,
			CompletedMissions: s.CompletedMissions,
			UniqueObjectives:  s.UniqueObjectives,
		}
	}

	var nextCursor string
	if end < len(ordered) {
		nextCursor = encodeCursor(ordered[end-1])
	}

	return Leaderboard{
		Entries:    entries,
		NextCursor: nextCursor,
	}, nil
}

func CompareStanding(a, b Standing) int {
	if a.Total != b.Total {
		if a.Total > b.Total {
			return -1
		}
		return 1
	}
	if a.CompletedMissions != b.CompletedMissions {
		if a.CompletedMissions > b.CompletedMissions {
			return -1
		}
		return 1
	}
	if a.UniqueObjectives != b.UniqueObjectives {
		if a.UniqueObjectives > b.UniqueObjectives {
			return -1
		}
		return 1
	}
	if a.RecipientID < b.RecipientID {
		return -1
	}
	if a.RecipientID > b.RecipientID {
		return 1
	}
	return 0
}

type cursorData struct {
	Total             int    `json:"t"`
	CompletedMissions int    `json:"c"`
	UniqueObjectives  int    `json:"u"`
	RecipientID       string `json:"r"`
}

func encodeCursor(s Standing) string {
	data, _ := json.Marshal(cursorData{
		Total:             s.Total,
		CompletedMissions: s.CompletedMissions,
		UniqueObjectives:  s.UniqueObjectives,
		RecipientID:       s.RecipientID,
	})
	return base64.URLEncoding.EncodeToString(data)
}

func decodeCursor(cursor string) (Standing, error) {
	data, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return Standing{}, err
	}
	var c cursorData
	if err := json.Unmarshal(data, &c); err != nil {
		return Standing{}, err
	}
	return Standing{
		Total:             c.Total,
		CompletedMissions: c.CompletedMissions,
		UniqueObjectives:  c.UniqueObjectives,
		RecipientID:       c.RecipientID,
	}, nil
}
