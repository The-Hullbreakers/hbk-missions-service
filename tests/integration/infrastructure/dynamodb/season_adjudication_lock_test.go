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

// TEST-INF-ADJ-001 — the real SeasonFinalizer driving the proposed lock
// adapter must persist a stable attempt identity and adjudication cutoff on
// the season item while the public state stays FREEZING. Setup uses the
// sanctioned require.NoError exception (fixture seeding only); the adapter
// itself is RED until the boundary lands.
func TestSeasonAdjudicationLock_GivenFreezingSeason_WhenFinalizationBegins_ThenStableLockAndCutoffPersisted(t *testing.T) {
	// Arrange — fresh test-owned table (never created elsewhere, never
	// cleared), a FREEZING official season, and capture-only doubles.
	ctx := context.Background()
	tableName := fmt.Sprintf("hbk-chg0051-adj-%d", time.Now().UnixNano())
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	require.NoError(t, client.CreateMissionsTable(ctx, tableName))
	t.Logf("created test-owned table (left intact): %s", tableName)

	const seasonID = "season-lock-1"
	store := missiondynamo.NewSeasonStore(client, tableName)
	require.NoError(t, store.Create(domain.NewSeasonWithOfficial(seasonID, domain.SeasonStateFreezing, true)))

	snapshot := &lockTestSnapshotStub{}
	roster := &lockTestRosterStub{}
	stager := &lockTestStageStub{}
	verifier := &lockTestVerifyStub{}
	committer := &lockTestCommitStub{}
	lock := missiondynamo.NewAdjudicationLock(client, tableName)
	finalizer := application.NewSeasonFinalizer(store, lock, snapshot, roster, stager, verifier, committer)

	// Act
	finalized, err := finalizer.Handle(seasonID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, domain.SeasonStateFinalized, finalized.State)
	require.Equal(t, "ADJ#"+seasonID, snapshot.got.ID, "snapshot must observe the minted attempt")
	assert.False(t, snapshot.got.Cutoff.IsZero(), "cutoff must be minted once by the lock")
	assert.Equal(t, snapshot.got, stager.attempt, "staging replays the same attempt")
	assert.Equal(t, snapshot.got, verifier.attempt, "verify sees the same attempt")
	assert.Equal(t, snapshot.got, committer.attempt, "commit sees the same attempt")
	assert.False(t, roster.called, "no TEAM entries means no roster resolution")
	assert.Equal(t, domain.SeasonStateFreezing, committer.season.State,
		"the committer must receive the still-FREEZING season")

	result, readErr := client.Client().GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "SEASON"},
			"SK": &types.AttributeValueMemberS{Value: seasonID},
		},
		ConsistentRead: aws.Bool(true),
	})
	require.NoError(t, readErr)
	require.NotNil(t, result.Item)
	require.IsType(t, &types.AttributeValueMemberS{}, result.Item["State"])
	assert.Equal(t, string(domain.SeasonStateFreezing),
		result.Item["State"].(*types.AttributeValueMemberS).Value,
		"public state stays FREEZING until the manifest commits")
	require.IsType(t, &types.AttributeValueMemberS{}, result.Item["AdjudicationLock"])
	assert.Equal(t, snapshot.got.ID,
		result.Item["AdjudicationLock"].(*types.AttributeValueMemberS).Value,
		"the persisted lock id must equal the attempt identity")
	require.IsType(t, &types.AttributeValueMemberS{}, result.Item["AdjudicationCutoff"])
	rawCutoff := result.Item["AdjudicationCutoff"].(*types.AttributeValueMemberS).Value
	persistedCutoff, parseErr := time.Parse(time.RFC3339Nano, rawCutoff)
	require.NoError(t, parseErr, "cutoff must persist as RFC3339Nano UTC")
	assert.Equal(t, snapshot.got.Cutoff.UTC(), persistedCutoff.UTC(),
		"the persisted cutoff must equal the attempt cutoff")
}
