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

// TEST-INF-FENCE-001: an acceptance decided from an earlier ACTIVE read
// cannot commit after the season has acquired the adjudication lock.
func TestAdjudicationAcceptanceFence_StaleActiveReadCannotCommitAfterLock(t *testing.T) {
	t.Parallel()
	// Arrange
	client, table, err := manifestStoreFixture(t)
	require.NoError(t, err)
	const seasonID, contractID, playerID = "season-accept-locked", "contract-accept-locked", "player-1"
	seasons := missiondynamo.NewSeasonStore(client, table)
	require.NoError(t, seasons.Create(domain.NewSeasonWithOfficial(seasonID, domain.SeasonStateFreezing, true)))
	_, err = missiondynamo.NewAdjudicationLock(client, table).Acquire(seasonID)
	require.NoError(t, err)
	offer := domain.NewMissionOffer("offer-accept-locked", "BOUNTY_HUNTING", seasonID)
	contract := domain.NewMissionContract(contractID, offer, []string{playerID})
	store := missiondynamo.NewContractStore(client, table)

	// Act
	err = store.CommitAcceptance(contract, offer, 2)

	// Assert
	assert.ErrorIs(t, err, domain.ErrSeasonAdjudicating)
	for _, key := range []map[string]types.AttributeValue{
		{"PK": &types.AttributeValueMemberS{Value: "CONTRACT"}, "SK": &types.AttributeValueMemberS{Value: contractID}},
		{"PK": &types.AttributeValueMemberS{Value: "PLAYER_ACTIVE"}, "SK": &types.AttributeValueMemberS{Value: playerID}},
		{"PK": &types.AttributeValueMemberS{Value: "PLAYER_DEF_ACTIVE#" + playerID}, "SK": &types.AttributeValueMemberS{Value: offer.OfferID}},
	} {
		stored, readErr := client.Client().GetItem(t.Context(), &dynamodb.GetItemInput{
			TableName: aws.String(table), Key: key, ConsistentRead: aws.Bool(true),
		})
		require.NoError(t, readErr)
		assert.Empty(t, stored.Item, "locked acceptance cannot reserve a contract, player or definition")
	}
}
