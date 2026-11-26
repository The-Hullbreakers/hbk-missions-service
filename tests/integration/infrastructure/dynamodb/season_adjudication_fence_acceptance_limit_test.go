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

// TEST-INF-FENCE-001: a season guard must not reclassify the player's
// existing active-contract limit as an adjudication-lock conflict.
func TestAdjudicationAcceptanceFence_ActiveLimitKeepsPriority(t *testing.T) {
	t.Parallel()
	// Arrange
	client, table, err := manifestStoreFixture(t)
	require.NoError(t, err)
	const seasonID, playerID, contractID = "season-accept-limit", "player-1", "contract-new"
	seasons := missiondynamo.NewSeasonStore(client, table)
	require.NoError(t, seasons.Create(domain.NewSeasonWithOfficial(seasonID, domain.SeasonStateActive, true)))
	offer := domain.NewMissionOffer("offer-new", "BOUNTY_HUNTING", seasonID)
	_, err = client.Client().PutItem(t.Context(), &dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item: map[string]types.AttributeValue{
			"PK":           &types.AttributeValueMemberS{Value: "PLAYER_ACTIVE"},
			"SK":           &types.AttributeValueMemberS{Value: playerID},
			"contract_ids": &types.AttributeValueMemberSS{Value: []string{"prior-contract"}},
		},
	})
	require.NoError(t, err)
	store := missiondynamo.NewContractStore(client, table)
	contract := domain.NewMissionContract(contractID, offer, []string{playerID})

	// Act
	err = store.CommitAcceptance(contract, offer, 1)

	// Assert
	require.ErrorIs(t, err, domain.ErrActiveMissionLimitReached)
	assert.NotErrorIs(t, err, domain.ErrSeasonAdjudicating)
	for _, key := range []map[string]types.AttributeValue{
		{"PK": &types.AttributeValueMemberS{Value: "CONTRACT"}, "SK": &types.AttributeValueMemberS{Value: contractID}},
		{"PK": &types.AttributeValueMemberS{Value: "PLAYER_DEF_ACTIVE#" + playerID}, "SK": &types.AttributeValueMemberS{Value: offer.OfferID}},
	} {
		stored, readErr := client.Client().GetItem(t.Context(), &dynamodb.GetItemInput{
			TableName: aws.String(table), Key: key, ConsistentRead: aws.Bool(true),
		})
		require.NoError(t, readErr)
		assert.Empty(t, stored.Item)
	}
}
