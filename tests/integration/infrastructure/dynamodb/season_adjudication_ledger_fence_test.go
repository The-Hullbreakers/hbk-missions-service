//go:build integration

package dynamodb_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func ledgerFenceCredit(seasonID string) domain.MeritCredit {
	return domain.MeritCredit{
		SeasonID: seasonID, Category: "BOUNTY_HUNTING",
		Track: domain.TrackIndividual, RecipientID: "player-1", Amount: 100,
		ContractID: "contract-1", ObjectiveID: "objective-1", UnitID: "unit-1",
		ConfirmationID: "contract-1:objective-1:unit-1:player-1",
	}
}

// TEST-INF-FENCE-001: a direct ledger credit before the lock remains legal.
func TestAdjudicationLedgerFence_FreezingBeforeLockSavesCredit(t *testing.T) {
	t.Parallel()
	// Arrange
	client, table, err := manifestStoreFixture(t)
	require.NoError(t, err)
	const seasonID = "season-ledger-prelock"
	seasons := missiondynamo.NewSeasonStore(client, table)
	require.NoError(t, seasons.Create(domain.NewSeasonWithOfficial(seasonID, domain.SeasonStateFreezing, true)))
	store := missiondynamo.NewMeritLedgerStore(client, table)

	// Act
	err = store.Save(ledgerFenceCredit(seasonID))

	// Assert
	require.NoError(t, err)
	credits, readErr := client.Client().Query(t.Context(), &dynamodb.QueryInput{
		TableName: aws.String(table), KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "MERIT_CREDIT"},
		},
		ConsistentRead: aws.Bool(true),
	})
	require.NoError(t, readErr)
	require.Len(t, credits.Items, 1)
	assert.Equal(t, "100", credits.Items[0]["Amount"].(*types.AttributeValueMemberN).Value)
}

// TEST-INF-FENCE-001: the direct ledger port must not bypass the
// adjudication lock, even when called without a contract approval.
func TestAdjudicationLedgerFence_LockedSeasonRejectsCredit(t *testing.T) {
	t.Parallel()
	// Arrange
	client, table, err := manifestStoreFixture(t)
	require.NoError(t, err)
	const seasonID = "season-ledger-locked"
	seasons := missiondynamo.NewSeasonStore(client, table)
	require.NoError(t, seasons.Create(domain.NewSeasonWithOfficial(seasonID, domain.SeasonStateFreezing, true)))
	_, err = missiondynamo.NewAdjudicationLock(client, table).Acquire(seasonID)
	require.NoError(t, err)
	store := missiondynamo.NewMeritLedgerStore(client, table)

	// Act
	err = store.Save(ledgerFenceCredit(seasonID))

	// Assert
	assert.ErrorIs(t, err, domain.ErrSeasonAdjudicating)
	credits, readErr := client.Client().Query(t.Context(), &dynamodb.QueryInput{
		TableName: aws.String(table), KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "MERIT_CREDIT"},
		},
		ConsistentRead: aws.Bool(true),
	})
	require.NoError(t, readErr)
	assert.Empty(t, credits.Items)
}
