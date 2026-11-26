//go:build integration

package dynamodb_test

import (
	"errors"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// Adjudication boundary doubles — capture what the real SeasonFinalizer
// passes through each port. They carry no assertions; every observation is
// asserted by the test's Assert section.
type lockTestSnapshotStub struct{ got domain.AdjudicationAttempt }

func (s *lockTestSnapshotStub) Read(attempt domain.AdjudicationAttempt) (application.AdjudicationSnapshot, error) {
	s.got = attempt
	return application.AdjudicationSnapshot{}, nil
}

type lockTestRosterStub struct{ called bool }

func (s *lockTestRosterStub) Members(string, string, string, []string) ([]string, error) {
	s.called = true
	return nil, errors.New("roster must not be called without TEAM entries")
}

type lockTestStageStub struct {
	attempt domain.AdjudicationAttempt
	facts   []domain.PlacementFact
}

func (s *lockTestStageStub) Stage(attempt domain.AdjudicationAttempt, facts []domain.PlacementFact) error {
	s.attempt = attempt
	s.facts = facts
	return nil
}

type lockTestVerifyStub struct{ attempt domain.AdjudicationAttempt }

func (s *lockTestVerifyStub) Verify(attempt domain.AdjudicationAttempt, _ []domain.PlacementFact) error {
	s.attempt = attempt
	return nil
}

type lockTestCommitStub struct {
	season  domain.Season
	attempt domain.AdjudicationAttempt
	facts   []domain.PlacementFact
}

func (s *lockTestCommitStub) Commit(season domain.Season, attempt domain.AdjudicationAttempt, facts []domain.PlacementFact) error {
	s.season = season
	s.attempt = attempt
	s.facts = facts
	return nil
}
