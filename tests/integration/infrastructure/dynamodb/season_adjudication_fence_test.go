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

// TEST-INF-FENCE-001: acquiring the adjudication lock must reject a
// concurrent confirmation in the same write, preserving contract/credits.
func TestAdjudicationFence_LockedSeasonCannotStageContractCredit(t *testing.T) {
	t.Parallel()
	// Arrange
	client, table, err := manifestStoreFixture(t)
	require.NoError(t, err)
	const seasonID, contractID = "season-fenced", "contract-fenced"
	seasonStore := missiondynamo.NewSeasonStore(client, table)
	require.NoError(t, seasonStore.Create(domain.NewSeasonWithOfficial(seasonID, domain.SeasonStateFreezing, true)))
	contractStore := missiondynamo.NewContractStore(client, table)
	offer := domain.NewMissionOffer("offer-fenced", "BOUNTY_HUNTING", seasonID)
	contract := domain.NewMissionContract(contractID, offer, []string{"player-1"})
	terms, err := domain.NewContractTerms([]domain.Objective{
		domain.NewObjectiveNamed("objective-1", "Capture", "", "BOUNTY_HUNTING", 100, 0, 0),
	})
	require.NoError(t, err)
	contract.Terms = terms
	member, err := domain.NewCallerIdentity("player-1", nil, nil)
	require.NoError(t, err)
	require.NoError(t, contract.Start(member))
	require.NoError(t, contractStore.Save(contract))
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
	handler := application.NewConfirmObjectiveHandler(contractStore, contractStore, seasonStore)

	// Act
	_, err = handler.Handle(member, contractID, "objective-1", "unit-1", seasonID)

	// Assert
	assert.ErrorIs(t, err, domain.ErrSeasonAdjudicating)
	after, readErr := client.Client().GetItem(t.Context(), &dynamodb.GetItemInput{
		TableName: aws.String(table), Key: key, ConsistentRead: aws.Bool(true),
	})
	require.NoError(t, readErr)
	assert.Equal(t, before.Item, after.Item, "failed confirmation cannot stage a unit")
	credits, readErr := client.Client().Query(t.Context(), &dynamodb.QueryInput{
		TableName: aws.String(table), KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "MERIT_CREDIT"},
		},
		ConsistentRead: aws.Bool(true),
	})
	require.NoError(t, readErr)
	assert.Empty(t, credits.Items, "failed confirmation cannot materialize credits")
	season, readErr := seasonStore.Get(seasonID)
	require.NoError(t, readErr)
	assert.Equal(t, domain.SeasonStateFreezing, season.State)
}
