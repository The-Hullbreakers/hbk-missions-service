package dynamodb_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsdynamo "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// TestCreditItem_PersistsMediaURI verifies the ledger item carries the media
// reference so reviewers can resolve evidence reads later (T-10).
func TestCreditItem_PersistsMediaURI(t *testing.T) {
	ctx := t.Context()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	err = client.CreateMissionsTable(ctx, "hbk-missions")
	require.NoError(t, err)
	clearMissionsTable(t, ctx, client, "hbk-missions")

	store := missiondynamo.NewMeritLedgerStore(client, "hbk-missions")
	seedActiveSeason(t, client, "hbk-missions", "season-1")
	credit := domain.MeritCredit{
		SeasonID:       "season-1",
		Category:       "BOUNTY_HUNTING",
		Track:          domain.TrackIndividual,
		RecipientID:    "player-1",
		Amount:         100,
		ContractID:     "contract-1",
		ObjectiveID:    "objective-1",
		UnitID:         "target-1",
		ConfirmationID: "contract-1~objective-1~target-1~player-1",
		MediaURI:       "evidence/contract-1/abc-123",
	}

	err = store.Save(credit)
	require.NoError(t, err)

	out, err := client.Client().GetItem(ctx, &awsdynamo.GetItemInput{
		TableName: aws.String("hbk-missions"),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "MERIT_CREDIT"},
			"SK": &types.AttributeValueMemberS{Value: credit.ConfirmationID},
		},
	})
	require.NoError(t, err)
	attr, ok := out.Item["MediaURI"].(*types.AttributeValueMemberS)
	require.False(t, !ok || attr.Value != "evidence/contract-1/abc-123")
}

// TestCreditItem_OmitsMediaURIWhenEmpty verifies existing credits without
// media stay unchanged (backward compatibility of the item shape).
func TestCreditItem_OmitsMediaURIWhenEmpty(t *testing.T) {
	ctx := t.Context()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	err = client.CreateMissionsTable(ctx, "hbk-missions")
	require.NoError(t, err)
	clearMissionsTable(t, ctx, client, "hbk-missions")

	store := missiondynamo.NewMeritLedgerStore(client, "hbk-missions")
	seedActiveSeason(t, client, "hbk-missions", "season-1")
	credit := domain.MeritCredit{
		SeasonID:       "season-1",
		Category:       "BOUNTY_HUNTING",
		Track:          domain.TrackIndividual,
		RecipientID:    "player-1",
		Amount:         100,
		ContractID:     "contract-1",
		ObjectiveID:    "objective-1",
		UnitID:         "target-1",
		ConfirmationID: "contract-1~objective-1~target-1~player-1",
	}

	err = store.Save(credit)
	require.NoError(t, err)

	out, err := client.Client().GetItem(ctx, &awsdynamo.GetItemInput{
		TableName: aws.String("hbk-missions"),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "MERIT_CREDIT"},
			"SK": &types.AttributeValueMemberS{Value: credit.ConfirmationID},
		},
	})
	require.NoError(t, err)
	_, present := out.Item["MediaURI"]
	require.False(t, present)
}
