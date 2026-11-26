//go:build integration

package dynamodb_test

import (
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func snapshotCredit(seasonID, category, track, recipient, contractID string, amount int) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"PK":             &types.AttributeValueMemberS{Value: "MERIT_CREDIT"},
		"SK":             &types.AttributeValueMemberS{Value: contractID + ":" + recipient},
		"SeasonID":       &types.AttributeValueMemberS{Value: seasonID},
		"Category":       &types.AttributeValueMemberS{Value: category},
		"Track":          &types.AttributeValueMemberS{Value: track},
		"RecipientID":    &types.AttributeValueMemberS{Value: recipient},
		"ContractID":     &types.AttributeValueMemberS{Value: contractID},
		"Amount":         &types.AttributeValueMemberN{Value: strconv.Itoa(amount)},
		"ConfirmationID": &types.AttributeValueMemberS{Value: contractID + ":o:" + recipient},
	}
}

// TEST-INF-SNAP-001: the snapshot aggregates season credits into ranked
// bucket entries with their credited contract ids, read base-table
// consistent so nothing after the freeze can be missed.
func TestAdjudicationSnapshot_AggregatesCreditedEntriesPerBucket(t *testing.T) {
	t.Parallel()
	// Arrange
	client, table, err := manifestStoreFixture(t)
	require.NoError(t, err)
	const seasonID = "season-snapshot-1"
	seasons := missiondynamo.NewSeasonStore(client, table)
	require.NoError(t, seasons.Create(domain.NewSeasonWithOfficial(seasonID, domain.SeasonStateFreezing, true)))
	attempt, err := missiondynamo.NewAdjudicationLock(client, table).Acquire(seasonID)
	require.NoError(t, err)
	for _, row := range []map[string]types.AttributeValue{
		snapshotCredit(seasonID, "BOUNTY_HUNTING", "INDIVIDUAL", "p-1", "c-1", 40),
		snapshotCredit(seasonID, "BOUNTY_HUNTING", "INDIVIDUAL", "p-1", "c-2", 60),
		snapshotCredit(seasonID, "BOUNTY_HUNTING", "INDIVIDUAL", "p-2", "c-3", 10),
		snapshotCredit(seasonID, "BOUNTY_HUNTING", "TEAM", "team-a", "c-4", 5),
		snapshotCredit("other-season", "BOUNTY_HUNTING", "INDIVIDUAL", "p-9", "c-9", 999),
	} {
		_, err = client.Client().PutItem(t.Context(), &dynamodb.PutItemInput{TableName: aws.String(table), Item: row})
		require.NoError(t, err)
	}
	reader := missiondynamo.NewAdjudicationSnapshotReader(client, table)

	// Act
	snapshot, err := reader.Read(attempt)

	// Assert
	require.NoError(t, err)
	require.Len(t, snapshot.Buckets, 2)
	byKey := map[string]int{}
	for i, b := range snapshot.Buckets {
		byKey[b.Category+"#"+string(b.Track)] = i
	}
	require.Contains(t, byKey, "BOUNTY_HUNTING#INDIVIDUAL")
	require.Contains(t, byKey, "BOUNTY_HUNTING#TEAM")
	ind := snapshot.Buckets[byKey["BOUNTY_HUNTING#INDIVIDUAL"]]
	require.Len(t, ind.Entries, 2)
	assert.Equal(t, "p-1", ind.Entries[0].RecipientID)
	assert.Equal(t, 100, ind.Entries[0].Total)
	assert.ElementsMatch(t, []string{"c-1", "c-2"}, ind.Entries[0].CreditedContractIDs)
	assert.Equal(t, "p-2", ind.Entries[1].RecipientID)
	assert.Equal(t, 10, ind.Entries[1].Total)
	tm := snapshot.Buckets[byKey["BOUNTY_HUNTING#TEAM"]]
	require.Len(t, tm.Entries, 1)
	assert.Equal(t, "team-a", tm.Entries[0].RecipientID)
	assert.ElementsMatch(t, []string{"c-4"}, tm.Entries[0].CreditedContractIDs)
}
