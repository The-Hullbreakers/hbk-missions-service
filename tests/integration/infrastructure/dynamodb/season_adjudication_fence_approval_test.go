//go:build integration

package dynamodb_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// TEST-INF-MCF-001: reviewer approval after adjudication begins must
// atomically leave the contract pending and the ledger empty.
func TestAdjudicationApprovalFence_LockedSeasonRejectsCreditsAtomically(t *testing.T) {
	t.Parallel()
	// Arrange
	client, table, err := manifestStoreFixture(t)
	require.NoError(t, err)
	const seasonID, contractID = "season-1", "contract-1"
	seasons := missiondynamo.NewSeasonStore(client, table)
	require.NoError(t, seasons.Create(domain.NewSeasonWithOfficial(seasonID, domain.SeasonStateFreezing, true)))
	contracts := missiondynamo.NewContractStore(client, table)
	pendingTeamFixture(t, contracts)
	key := map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: "CONTRACT"},
		"SK": &types.AttributeValueMemberS{Value: contractID},
	}
	before, err := client.Client().GetItem(t.Context(), &dynamodb.GetItemInput{
		TableName: aws.String(table), Key: key, ConsistentRead: aws.Bool(true),
	})
	require.NoError(t, err)
	_, err = missiondynamo.NewAdjudicationLock(client, table).Acquire(seasonID)
	require.NoError(t, err)
	reviewer, err := domain.NewCallerIdentity("mod-1", []string{"admin"}, nil)
	require.NoError(t, err)
	handler := application.NewApproveContractHandler(contracts, contracts)

	// Act
	_, err = handler.Handle(reviewer, contractID)

	// Assert
	assert.ErrorIs(t, err, domain.ErrSeasonAdjudicating)
	after, readErr := client.Client().GetItem(t.Context(), &dynamodb.GetItemInput{
		TableName: aws.String(table), Key: key, ConsistentRead: aws.Bool(true),
	})
	require.NoError(t, readErr)
	assert.Equal(t, before.Item, after.Item, "failed approval cannot complete the contract")
	credits, readErr := client.Client().Query(t.Context(), &dynamodb.QueryInput{
		TableName: aws.String(table), KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "MERIT_CREDIT"},
		},
		ConsistentRead: aws.Bool(true),
	})
	require.NoError(t, readErr)
	assert.Empty(t, credits.Items, "failed approval cannot materialize any merit credit")
}
