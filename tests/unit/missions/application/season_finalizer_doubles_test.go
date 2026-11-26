package application_test

import (
	"time"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func finAttempt() domain.AdjudicationAttempt {
	return domain.AdjudicationAttempt{
		ID: "ADJ#season-1", SeasonID: "season-1",
		Cutoff: time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC),
	}
}

func finSnapshot() application.AdjudicationSnapshot {
	return application.AdjudicationSnapshot{Buckets: []application.AdjudicationBucket{
		{Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, Entries: []application.AdjudicationEntry{
			{RecipientID: "p-1", Total: 120, CreditedContractIDs: []string{"c-1"}},
		}},
	}}
}

// finCallLog records outgoing-port invocation order across doubles.
type finCallLog struct{ calls []string }

func (l *finCallLog) add(name string) { l.calls = append(l.calls, name) }

type finLockStub struct {
	attempt  domain.AdjudicationAttempt
	err      error
	acquires int
	log      *finCallLog
}

func (s *finLockStub) Acquire(string) (domain.AdjudicationAttempt, error) {
	s.acquires++
	s.log.add("acquire")
	if s.err != nil {
		return domain.AdjudicationAttempt{}, s.err
	}
	return s.attempt, nil
}

type finSnapshotStub struct {
	snap application.AdjudicationSnapshot
	got  domain.AdjudicationAttempt
	err  error
	log  *finCallLog
}

func (s *finSnapshotStub) Read(attempt domain.AdjudicationAttempt) (application.AdjudicationSnapshot, error) {
	s.log.add("snapshot")
	s.got = attempt
	if s.err != nil {
		return application.AdjudicationSnapshot{}, s.err
	}
	return s.snap, nil
}

type finRosterStub struct {
	members        []string
	gotSeasonID    string
	gotCategory    string
	gotRecipientID string
	gotContractIDs []string
	calls          int
}

func (s *finRosterStub) Members(seasonID, category, recipientID string, contractIDs []string) ([]string, error) {
	s.calls++
	s.gotSeasonID = seasonID
	s.gotCategory = category
	s.gotRecipientID = recipientID
	s.gotContractIDs = contractIDs
	return s.members, nil
}

type finStageStub struct {
	log      *finCallLog
	failAt   int
	err      error
	calls    int
	attempts []domain.AdjudicationAttempt
	facts    [][]domain.PlacementFact
}

func (s *finStageStub) Stage(attempt domain.AdjudicationAttempt, facts []domain.PlacementFact) error {
	s.calls++
	s.log.add("stage")
	s.attempts = append(s.attempts, attempt)
	s.facts = append(s.facts, facts)
	if s.calls == s.failAt {
		return s.err
	}
	return nil
}

type finVerifyStub struct {
	log     *finCallLog
	err     error
	attempt domain.AdjudicationAttempt
}

func (s *finVerifyStub) Verify(attempt domain.AdjudicationAttempt, _ []domain.PlacementFact) error {
	s.log.add("verify")
	s.attempt = attempt
	return s.err
}

type finCommitStub struct {
	log     *finCallLog
	season  domain.Season
	attempt domain.AdjudicationAttempt
	facts   []domain.PlacementFact
	err     error
}

func (s *finCommitStub) Commit(season domain.Season, attempt domain.AdjudicationAttempt, facts []domain.PlacementFact) error {
	s.log.add("commit")
	s.season = season
	s.attempt = attempt
	s.facts = facts
	return s.err
}
