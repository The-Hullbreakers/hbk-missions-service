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

// TEST-INF-LANG-001: an offer frozen from a bilingual published definition
// round-trips its English snapshot as a DynamoDB M attribute so reads
// serve English.
func TestLocalizedOffer_BilingualSnapshotRoundTrips(t *testing.T) {
	// Arrange
	client, table, setupErr := localizedStoreFixture(t)
	require.NoError(t, setupErr)
	store := missiondynamo.NewOfferStore(client, table)
	request := localizedDefinitionRequest()
	definition, err := domain.NewMissionDefinition(request.DefinitionCode, request.Title, request.Description, request.Category, request.Mode, request.CompletionReward, request.Objectives, request.Rules)
	require.NoError(t, err)
	definition.ContentEn = request.ContentEn.Clone()
	require.NoError(t, definition.Publish("v1"))
	offer, err := domain.NewMissionOfferFromDefinition("offer-rescue", "season-1", definition)
	require.NoError(t, err)
	require.NoError(t, store.Save(offer))

	// Act
	served, err := application.NewGetOfferHandler(store).Handle("offer-rescue", domain.LanguageEnglish)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, domain.LanguageEnglish, served.ContentLanguage)
	assert.Equal(t, "Rescue", served.Title)
	assert.Equal(t, "Save the pilot", served.Description)
	assert.Equal(t, []string{"No aids"}, served.Rules)
	require.Len(t, served.Objectives, 1)
	assert.Equal(t, "obj-1", served.Objectives[0].ObjectiveID)
	assert.Equal(t, "Capture", served.Objectives[0].Name)
	assert.Equal(t, "Capture the pilot", served.Objectives[0].Description)
	assert.Equal(t, 10, served.Objectives[0].MeritPerUnit)
	assert.Equal(t, 2, served.Objectives[0].UnitLimit)
	assert.Equal(t, 50, served.CompletionReward)
	assert.Equal(t, "v1", served.DefinitionVersion)
	item, err := client.Client().GetItem(t.Context(), &dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "OFFER"},
			"SK": &types.AttributeValueMemberS{Value: "offer-rescue"},
		},
	})
	require.NoError(t, err)
	raw, ok := item.Item["ContentEn"].(*types.AttributeValueMemberM)
	require.True(t, ok, "ContentEn must be a DynamoDB M attribute, not a JSON blob")
	_, isString := raw.Value["Title"].(*types.AttributeValueMemberS)
	assert.True(t, isString)
	_, isRulesList := raw.Value["Rules"].(*types.AttributeValueMemberL)
	assert.True(t, isRulesList)
	objectives, isObjList := raw.Value["Objectives"].(*types.AttributeValueMemberL)
	require.True(t, isObjList)
	require.Len(t, objectives.Value, 1)
	_, isObjectiveMap := objectives.Value[0].(*types.AttributeValueMemberM)
	assert.True(t, isObjectiveMap)
}

// TEST-INF-LANG-001: a legacy offer persisted before locale attributes
// existed decodes to zero English content and falls back to Spanish.
func TestLocalizedOffer_LegacySpanishOnlyFallsBack(t *testing.T) {
	// Arrange
	client, table, setupErr := localizedStoreFixture(t)
	require.NoError(t, setupErr)
	store := missiondynamo.NewOfferStore(client, table)
	offer := domain.NewMissionOffer("offer-old", "BOUNTY_HUNTING", "season-1")
	offer.Title = "Rescate"
	offer.Description = "Salvar al piloto"
	offer.Objectives = []domain.Objective{
		domain.NewObjectiveNamed("obj-1", "Capturar", "", "BOUNTY_HUNTING", 10, 0, 2),
	}
	require.NoError(t, store.Save(offer))

	// Act
	served, err := application.NewGetOfferHandler(store).Handle("offer-old", domain.LanguageEnglish)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, domain.LanguageSpanish, served.ContentLanguage)
	assert.Equal(t, "Rescate", served.Title)
	assert.Equal(t, "Salvar al piloto", served.Description)
	require.Len(t, served.Objectives, 1)
	assert.Equal(t, 10, served.Objectives[0].MeritPerUnit)
	assert.Equal(t, 2, served.Objectives[0].UnitLimit)
	item, err := client.Client().GetItem(t.Context(), &dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "OFFER"},
			"SK": &types.AttributeValueMemberS{Value: "offer-old"},
		},
	})
	require.NoError(t, err)
	_, exists := item.Item["ContentEn"]
	assert.False(t, exists)
}
