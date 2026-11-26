//go:build integration

package dynamodb_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// contractBackfillFixture builds an accepted contract whose offer text
// matches the reviewed Spanish source but whose terms may diverge.
func contractBackfillFixture(contractID, objectiveName string) (*domain.MissionContract, error) {
	offer := spanishBackfillOffer("offer-snapshot", "Rescate", "v1")
	objectives := append([]domain.Objective(nil), offer.Objectives...)
	objectives[0].Name = objectiveName
	terms, err := domain.NewContractTerms(objectives)
	if err != nil {
		return nil, err
	}
	contract := domain.NewMissionContract(contractID, offer, []string{"player-1"})
	contract.Terms = terms
	return contract, nil
}

func rawContractItem(ctx context.Context, client *missiondynamo.DynamoDBClient, table, contractID string) (map[string]types.AttributeValue, error) {
	item, err := client.Client().GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "CONTRACT"},
			"SK": &types.AttributeValueMemberS{Value: contractID},
		},
	})
	if err != nil {
		return nil, err
	}
	return item.Item, nil
}

// TEST-INF-LANG-002: a dry run reports the matching contract eligible and
// the terms-drifting contract unresolved — nothing is mutated.
func TestBackfillContractDryRunReportsTermsDrift(t *testing.T) {
	// Arrange
	client, table, setupErr := localizedStoreFixture(t)
	require.NoError(t, setupErr)
	store := missiondynamo.NewContractStore(client, table)
	match, err := contractBackfillFixture("contract-match", "Capturar")
	require.NoError(t, err)
	drift, err := contractBackfillFixture("contract-drift", "Otro objetivo")
	require.NoError(t, err)
	require.NoError(t, store.Save(match))
	require.NoError(t, store.Save(drift))
	matchBefore, err := rawContractItem(t.Context(), client, table, "contract-match")
	require.NoError(t, err)
	driftBefore, err := rawContractItem(t.Context(), client, table, "contract-drift")
	require.NoError(t, err)
	reviewed := backfillReviewed()

	// Act
	report, err := missiondynamo.NewTranslationBackfiller(client, table).Rehearse(t.Context(), reviewed, false)

	// Assert
	require.NoError(t, err)
	assert.Contains(t, report.Eligible, missiondynamo.BackfillRecord{Partition: "CONTRACT", ID: "contract-match"})
	assert.Empty(t, report.Updated)
	assert.Contains(t, report.Unresolved, missiondynamo.BackfillRecord{Partition: "CONTRACT", ID: "contract-drift", Reason: "SPANISH_SNAPSHOT_MISMATCH"})
	matchAfter, err := rawContractItem(t.Context(), client, table, "contract-match")
	require.NoError(t, err)
	driftAfter, err := rawContractItem(t.Context(), client, table, "contract-drift")
	require.NoError(t, err)
	assert.Equal(t, matchBefore, matchAfter)
	assert.Equal(t, driftBefore, driftAfter)
}
