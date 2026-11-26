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

// TEST-INF-MCF-001: prepending the season check must not classify a
// duplicate credit cancellation as an adjudication-lock conflict.
func TestAdjudicationApprovalFence_ExistingCreditPreservesDuplicateSignal(t *testing.T) {
	t.Parallel()
	// Arrange
	client, table, err := manifestStoreFixture(t)
	require.NoError(t, err)
	const seasonID, contractID = "season-1", "contract-1"
	seasons := missiondynamo.NewSeasonStore(client, table)
	require.NoError(t, seasons.Create(domain.NewSeasonWithOfficial(seasonID, domain.SeasonStateFreezing, true)))
	contracts := missiondynamo.NewContractStore(client, table)
	pendingTeamFixture(t, contracts)
	pending, err := contracts.Get(contractID)
	require.NoError(t, err)
	require.Len(t, pending.PendingCredits(), 1)
	confirmationID := pending.PendingCredits()[0].ConfirmationID
	_, err = client.Client().PutItem(t.Context(), &dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item: map[string]types.AttributeValue{
			"PK":             &types.AttributeValueMemberS{Value: "MERIT_CREDIT"},
			"SK":             &types.AttributeValueMemberS{Value: confirmationID},
			"ConfirmationID": &types.AttributeValueMemberS{Value: confirmationID},
		},
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	})
	require.NoError(t, err)
	reviewer, err := domain.NewCallerIdentity("mod-1", []string{"admin"}, nil)
	require.NoError(t, err)
	handler := application.NewApproveContractHandler(contracts, contracts)

	// Act
	_, err = handler.Handle(reviewer, contractID)

	// Assert
	require.ErrorIs(t, err, domain.ErrDuplicateConfirmation)
	assert.NotErrorIs(t, err, domain.ErrSeasonAdjudicating)
	stored, readErr := contracts.Get(contractID)
	require.NoError(t, readErr)
	assert.Equal(t, domain.ContractPendingValidation, stored.State)
	credits, readErr := client.Client().Query(t.Context(), &dynamodb.QueryInput{
		TableName: aws.String(table), KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "MERIT_CREDIT"},
		},
		ConsistentRead: aws.Bool(true),
	})
	require.NoError(t, readErr)
	assert.Len(t, credits.Items, 1)
}
