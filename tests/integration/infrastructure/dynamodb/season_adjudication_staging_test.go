//go:build integration

package dynamodb_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// stageTestLockStub returns a fixed attempt — the lock itself is proven by
// TEST-INF-ADJ-001; staging is the unit under test here.
type stageTestLockStub struct{ attempt domain.AdjudicationAttempt }

func (s *stageTestLockStub) Acquire(string) (domain.AdjudicationAttempt, error) {
	return s.attempt, nil
}

// stageTestSnapshotStub serves a fixed snapshot and records the attempt it
// was read under; it carries no assertions.
type stageTestSnapshotStub struct {
	snap application.AdjudicationSnapshot
	got  domain.AdjudicationAttempt
}

func (s *stageTestSnapshotStub) Read(attempt domain.AdjudicationAttempt) (application.AdjudicationSnapshot, error) {
	s.got = attempt
	return s.snap, nil
}

// TEST-INF-FACT-001 — facts stage INDEPENDENTLY of any transaction cap: 200
// ranked entries must all land as staged rows before the manifest commits.
// require.NoError is used only for fixture seeding.
func TestSeasonAdjudicationStaging_Given200RankedEntries_WhenStaged_ThenAllFactsPersistedIndependently(t *testing.T) {
	// Arrange — fresh per-run table, FREEZING official season, fixed
	// attempt, and a snapshot of one INDIVIDUAL bucket with 200 entries.
	ctx := context.Background()
	tableName := fmt.Sprintf("hbk-chg0051-stage-%d", time.Now().UnixNano())
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	require.NoError(t, client.CreateMissionsTable(ctx, tableName))
	t.Logf("created test-owned table (left intact): %s", tableName)

	const seasonID = "season-stage-1"
	store := missiondynamo.NewSeasonStore(client, tableName)
	require.NoError(t, store.Create(domain.NewSeasonWithOfficial(seasonID, domain.SeasonStateFreezing, true)))

	attempt, err := domain.NewAdjudicationAttempt(seasonID, time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	entries := make([]application.AdjudicationEntry, 200)
	for i := range entries {
		entries[i] = application.AdjudicationEntry{
			RecipientID: fmt.Sprintf("p-%03d", i+1), Total: 200 - i,
		}
	}
	snapshot := &stageTestSnapshotStub{snap: application.AdjudicationSnapshot{
		Buckets: []application.AdjudicationBucket{
			{Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, Entries: entries},
		},
	}}
	committer := &lockTestCommitStub{}
	stager := missiondynamo.NewPlacementFactStager(client, tableName)
	finalizer := application.NewSeasonFinalizer(store, &stageTestLockStub{attempt: attempt},
		snapshot, &lockTestRosterStub{}, stager, &lockTestVerifyStub{}, committer)

	// Act
	finalized, err := finalizer.Handle(seasonID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, domain.SeasonStateFinalized, finalized.State)
	assert.Equal(t, domain.SeasonStateFreezing, committer.season.State)
	require.Len(t, committer.facts, 200, "every ranked entry produces one fact")

	result, readErr := client.Client().Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "FACT#" + seasonID},
		},
		ConsistentRead: aws.Bool(true),
	})
	require.NoError(t, readErr)
	assert.Len(t, result.Items, 200, "200 staged rows — no 100-item transaction cap")

	wantSK := attempt.ID + "#BOUNTY_HUNTING#" + string(domain.TrackIndividual) + "#p-001"
	require.Equal(t, "p-001", committer.facts[0].RecipientID)
	var row map[string]types.AttributeValue
	for _, item := range result.Items {
		if s, ok := item["SK"].(*types.AttributeValueMemberS); ok && s.Value == wantSK {
			row = item
		}
	}
	require.NotNil(t, row, "staged row for p-001 must exist under the deterministic SK")
	require.IsType(t, &types.AttributeValueMemberS{}, row["Digest"])
	assert.Equal(t, committer.facts[0].Digest(),
		row["Digest"].(*types.AttributeValueMemberS).Value,
		"the persisted digest must equal the canonical fact digest")
}
