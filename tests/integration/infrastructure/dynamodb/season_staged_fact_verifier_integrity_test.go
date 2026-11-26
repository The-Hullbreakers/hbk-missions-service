//go:build integration

package dynamodb_test

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// TEST-INF-FACT-002: a row with its original digest but changed merits
// cannot satisfy verification of the expected staged fact set.
func TestStagedFactVerifier_TamperedRowBlocksCommit(t *testing.T) {
	t.Parallel()
	// Arrange
	client, table, err := manifestStoreFixture(t)
	require.NoError(t, err)
	const seasonID = "season-verified-tampered"
	attempt, err := domain.NewAdjudicationAttempt(seasonID, time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	facts, err := verifierFacts(seasonID)
	require.NoError(t, err)
	require.NoError(t, missiondynamo.NewPlacementFactStager(client, table).Stage(attempt, facts))
	_, err = client.Client().UpdateItem(t.Context(), &dynamodb.UpdateItemInput{
		TableName: aws.String(table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "FACT#" + seasonID},
			"SK": &types.AttributeValueMemberS{Value: attempt.ID + "#" + facts[0].Category + "#" + string(facts[0].Track) + "#" + facts[0].RecipientID},
		},
		UpdateExpression: aws.String("SET FinalMerits = :tampered"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":tampered": &types.AttributeValueMemberN{Value: "999"},
		},
	})
	require.NoError(t, err)
	verifier := missiondynamo.NewStagedFactVerifier(client, table)

	// Act
	err = verifier.Verify(attempt, facts)

	// Assert
	require.Error(t, err)
}
