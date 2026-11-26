package composition_test

import (
	"crypto/rand"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

var chg0053NameCounter atomic.Uint64

// uniqueChg0053Table mints a high-entropy table name (timestamp + pid +
// process counter + random suffix) so a test can never collide with — or
// delete — a table it did not create.
func uniqueChg0053Table(name string) string {
	var entropy [6]byte
	_, _ = rand.Read(entropy[:])
	return fmt.Sprintf("chg0053-%s-%d-%d-%d-%x", name,
		time.Now().UnixNano(), os.Getpid(), chg0053NameCounter.Add(1), entropy[:])
}

// chg0053Table creates a dedicated freshly-named table per test and asserts
// it did not pre-exist. Test tables are intentionally NOT dropped: after rows
// were deleted without authorization, any cleanup of local test tables
// requires a separately approved pass — the table stays until then.
// englishize returns a copy of the draft definition with a structurally
// complete English snapshot so Publish accepts it (CHG-0054).
func englishize(def domain.MissionDefinition) domain.MissionDefinition {
	def.ContentEn = domain.LocalizedContent{Title: "EN " + def.Title, Description: "EN " + def.Description}
	for _, rule := range def.Rules {
		def.ContentEn.Rules = append(def.ContentEn.Rules, "EN "+rule)
	}
	for _, objective := range def.Objectives {
		def.ContentEn.Objectives = append(def.ContentEn.Objectives, domain.LocalizedObjective{
			ObjectiveID: objective.ObjectiveID,
			Name:        "EN " + objective.Name,
			Description: objective.Description,
		})
	}
	return def
}

func chg0053Table(t *testing.T, name string) (*missiondynamo.DynamoDBClient, string) {
	t.Helper()
	ctx := t.Context()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	table := uniqueChg0053Table(name)
	_, err = client.Client().DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(table)})
	var notFound *types.ResourceNotFoundException
	require.ErrorAs(t, err, &notFound, "test table %q unexpectedly already exists", table)
	require.NoError(t, client.CreateMissionsTable(ctx, table))
	return client, table
}

// seedPlayerReservation writes a PLAYER_DEF_ACTIVE guard in the shape the
// GREEN CommitAcceptance will persist (PK PLAYER_DEF_ACTIVE#<player>, SK
// <definition_code>, owner in contract_id) so the retake journey proves
// SaveDeparture releases the departing member's reservation.
func seedPlayerReservation(t *testing.T, client *missiondynamo.DynamoDBClient, table, playerID, definitionCode, ownerContractID string) {
	t.Helper()
	_, err := client.Client().PutItem(t.Context(), &dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item: map[string]types.AttributeValue{
			"PK":          &types.AttributeValueMemberS{Value: "PLAYER_DEF_ACTIVE#" + playerID},
			"SK":          &types.AttributeValueMemberS{Value: definitionCode},
			"contract_id": &types.AttributeValueMemberS{Value: ownerContractID},
		},
	})
	require.NoError(t, err)
}

// seedPlayerActive records a contract id in the player's PLAYER_ACTIVE guard
// set so the retake journey can prove departure frees the slot.
func seedPlayerActive(t *testing.T, client *missiondynamo.DynamoDBClient, table, playerID, contractID string) {
	t.Helper()
	_, err := client.Client().UpdateItem(t.Context(), &dynamodb.UpdateItemInput{
		TableName: aws.String(table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "PLAYER_ACTIVE"},
			"SK": &types.AttributeValueMemberS{Value: playerID},
		},
		UpdateExpression: aws.String("ADD contract_ids :cid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":cid": &types.AttributeValueMemberSS{Value: []string{contractID}},
		},
	})
	require.NoError(t, err)
}
