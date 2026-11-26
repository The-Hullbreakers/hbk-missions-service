package application

import (
	"sort"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// TeamRosterReader resolves the frozen member ids of a team's contracts in a
// season: the rosters captured when those contracts started.
type TeamRosterReader interface {
	Members(seasonID, teamID string) ([]string, error)
}

// FinalizedSeasonPersister atomically persists the finalized season and the
// placement-fact publication intents produced by adjudication.
type FinalizedSeasonPersister interface {
	SaveFinalized(season domain.Season, facts []domain.PlacementFact) error
}

// FinalizeSeasonHandler finalizes a freezing season. Official seasons emit one
// placement fact per ranked entry of every standings bucket; practice seasons
// finalize silently.
type FinalizeSeasonHandler struct {
	seasons   SeasonReader
	buckets   StandingsBucketReader
	standings LeaderboardReader
	rosters   TeamRosterReader
	persister FinalizedSeasonPersister
}

// NewFinalizeSeasonHandler creates a handler with the required ports.
func NewFinalizeSeasonHandler(seasons SeasonReader, buckets StandingsBucketReader, standings LeaderboardReader, rosters TeamRosterReader, persister FinalizedSeasonPersister) *FinalizeSeasonHandler {
	return &FinalizeSeasonHandler{
		seasons:   seasons,
		buckets:   buckets,
		standings: standings,
		rosters:   rosters,
		persister: persister,
	}
}

// Handle loads the season, transitions it to FINALIZED, derives the placement
// facts, and persists season plus intents in one write.
func (h *FinalizeSeasonHandler) Handle(seasonID string) (domain.Season, error) {
	season, err := h.seasons.Get(seasonID)
	if err != nil {
		return domain.Season{}, err
	}

	season, err = domain.Finalize(season)
	if err != nil {
		return domain.Season{}, err
	}

	facts, err := h.placementFacts(season)
	if err != nil {
		return domain.Season{}, err
	}

	if err := h.persister.SaveFinalized(season, facts); err != nil {
		return domain.Season{}, err
	}

	return season, nil
}

// placementFacts builds one fact per ranked entry for official seasons; ranks
// follow the published ordering and tie-break criteria, applied here so the
// adjudication never depends on reader order.
func (h *FinalizeSeasonHandler) placementFacts(season domain.Season) ([]domain.PlacementFact, error) {
	if !season.Official {
		return nil, nil
	}
	buckets, err := h.buckets.ListBuckets(season.SeasonID)
	if err != nil {
		return nil, err
	}
	var facts []domain.PlacementFact
	for _, bucket := range buckets {
		standings, err := h.standings.ListStandings(season.SeasonID, bucket.Category, bucket.Track, 0)
		if err != nil {
			return nil, err
		}
		ordered := make([]domain.Standing, len(standings))
		copy(ordered, standings)
		sort.Slice(ordered, func(i, j int) bool {
			return domain.CompareStanding(ordered[i], ordered[j]) < 0
		})
		for i, entry := range ordered {
			fact, err := h.placementFact(season, bucket, entry, i+1)
			if err != nil {
				return nil, err
			}
			facts = append(facts, fact)
		}
	}
	return facts, nil
}

func (h *FinalizeSeasonHandler) placementFact(season domain.Season, bucket domain.StandingBucket, entry domain.Standing, rank int) (domain.PlacementFact, error) {
	params := domain.PlacementFactParams{
		SeasonID:    season.SeasonID,
		Category:    bucket.Category,
		Track:       bucket.Track,
		RecipientID: entry.RecipientID,
		Rank:        rank,
		FinalMerits: entry.Total,
	}
	if bucket.Track == domain.TrackTeam {
		members, err := h.rosters.Members(season.SeasonID, entry.RecipientID)
		if err != nil {
			return domain.PlacementFact{}, err
		}
		params.MemberIDs = members
	}
	return domain.NewPlacementFact(params)
}
