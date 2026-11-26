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

// TEST-INF-MCF-001: a FREEZING season still permits reviewer approval
// before adjudication acquires its lock.
func TestAdjudicationApprovalFence_FreezingBeforeLockCommitsCredits(t *testing.T) {
	t.Parallel()
	// Arrange
	client, table, err := manifestStoreFixture(t)
	require.NoError(t, err)
	const seasonID, contractID = "season-1", "contract-1"
	seasons := missiondynamo.NewSeasonStore(client, table)
	require.NoError(t, seasons.Create(domain.NewSeasonWithOfficial(seasonID, domain.SeasonStateFreezing, true)))
	contracts := missiondynamo.NewContractStore(client, table)
	pendingTeamFixture(t, contracts)
	reviewer, err := domain.NewCallerIdentity("mod-1", []string{"admin"}, nil)
	require.NoError(t, err)
	handler := application.NewApproveContractHandler(contracts, contracts)

	// Act
	credits, err := handler.Handle(reviewer, contractID)

	// Assert
	require.NoError(t, err)
	require.Len(t, credits, 1)
	assert.Equal(t, 100, credits[0].Amount)
	stored, readErr := contracts.Get(contractID)
	require.NoError(t, readErr)
	assert.Equal(t, domain.ContractCompleted, stored.State)
	items, readErr := client.Client().Query(t.Context(), &dynamodb.QueryInput{
		TableName: aws.String(table), KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "MERIT_CREDIT"},
		},
		ConsistentRead: aws.Bool(true),
	})
	require.NoError(t, readErr)
	require.Len(t, items.Items, 1)
	assert.Equal(t, "100", items.Items[0]["Amount"].(*types.AttributeValueMemberN).Value)
	assert.Equal(t, "team-9", items.Items[0]["RecipientID"].(*types.AttributeValueMemberS).Value)
}
