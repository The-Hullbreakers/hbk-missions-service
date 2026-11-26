package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// SeasonFinalizer runs the two-phase adjudication: acquire (or resume) the
// stable attempt, read the fenced snapshot, derive and stage immutable
// placement facts, verify the staged set, then commit the constant-size
// manifest transaction that transitions FREEZING to FINALIZED. Practice
// seasons run the same pipeline with zero facts.
type SeasonFinalizer struct {
	seasons   SeasonReader
	locks     AdjudicationLock
	snapshots AdjudicationSnapshotReader
	rosters   CreditedRosterReader
	stager    PlacementFactStager
	verifier  StagedFactVerifier
	committer FinalizationCommitter
}

// NewSeasonFinalizer creates the use case with its seven outgoing ports.
func NewSeasonFinalizer(seasons SeasonReader, locks AdjudicationLock, snapshots AdjudicationSnapshotReader, rosters CreditedRosterReader, stager PlacementFactStager, verifier StagedFactVerifier, committer FinalizationCommitter) *SeasonFinalizer {
	return &SeasonFinalizer{
		seasons:   seasons,
		locks:     locks,
		snapshots: snapshots,
		rosters:   rosters,
		stager:    stager,
		verifier:  verifier,
		committer: committer,
	}
}

// Handle finalizes a FREEZING season and returns the FINALIZED season only
// after the commit succeeds.
func (f *SeasonFinalizer) Handle(seasonID string) (domain.Season, error) {
	season, err := f.seasons.Get(seasonID)
	if err != nil {
		return domain.Season{}, err
	}
	finalized, err := domain.Finalize(season)
	if err != nil {
		return domain.Season{}, err
	}
	attempt, err := f.locks.Acquire(seasonID)
	if err != nil {
		return domain.Season{}, err
	}
	snapshot, err := f.snapshots.Read(attempt)
	if err != nil {
		return domain.Season{}, err
	}
	facts, err := f.placementFacts(season, attempt, snapshot)
	if err != nil {
		return domain.Season{}, err
	}
	if err := f.stager.Stage(attempt, facts); err != nil {
		return domain.Season{}, err
	}
	if err := f.verifier.Verify(attempt, facts); err != nil {
		return domain.Season{}, err
	}
	if err := f.committer.Commit(season, attempt, facts); err != nil {
		return domain.Season{}, err
	}
	return finalized, nil
}

// placementFacts derives one fact per ranked entry, in snapshot order.
// Practice seasons stage no facts; TEAM entries resolve member_ids only
// from their credited contracts.
func (f *SeasonFinalizer) placementFacts(season domain.Season, attempt domain.AdjudicationAttempt, snapshot AdjudicationSnapshot) ([]domain.PlacementFact, error) {
	if !season.Official {
		return nil, nil
	}
	var facts []domain.PlacementFact
	for _, bucket := range snapshot.Buckets {
		for i, entry := range bucket.Entries {
			fact, err := f.placementFact(attempt, bucket, entry, i+1)
			if err != nil {
				return nil, err
			}
			facts = append(facts, fact)
		}
	}
	return facts, nil
}

func (f *SeasonFinalizer) placementFact(attempt domain.AdjudicationAttempt, bucket AdjudicationBucket, entry AdjudicationEntry, rank int) (domain.PlacementFact, error) {
	params := domain.PlacementFactParams{
		SeasonID:    attempt.SeasonID,
		Category:    bucket.Category,
		Track:       bucket.Track,
		RecipientID: entry.RecipientID,
		Rank:        rank,
		FinalMerits: entry.Total,
	}
	if bucket.Track == domain.TrackTeam {
		members, err := f.rosters.Members(attempt.SeasonID, bucket.Category, entry.RecipientID, entry.CreditedContractIDs)
		if err != nil {
			return domain.PlacementFact{}, err
		}
		params.MemberIDs = members
	}
	return domain.NewPlacementFact(params)
}
