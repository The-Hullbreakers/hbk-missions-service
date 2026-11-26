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

// TEST-INF-LANG-001: a bilingual draft authored through the application
// round-trips its English snapshot as a DynamoDB M attribute — not a JSON
// blob — while Spanish scalars and economics stay intact.
func TestLocalizedDefinition_BilingualDraftRoundTrips(t *testing.T) {
	// Arrange
	client, table, setupErr := localizedStoreFixture(t)
	require.NoError(t, setupErr)
	store := missiondynamo.NewDefinitionStore(client, table)
	request := localizedDefinitionRequest()

	// Act
	created, err := application.NewCreateMissionDefinitionHandler(store).Handle(request)

	// Assert
	require.NoError(t, err)
	stored, err := store.Get("def-rescue")
	require.NoError(t, err)
	assert.Equal(t, "def-rescue", stored.DefinitionCode)
	assert.Equal(t, "Rescate", stored.Title)
	assert.Equal(t, "Salvar al piloto", stored.Description)
	assert.Equal(t, 50, stored.CompletionReward)
	require.Len(t, stored.Objectives, 1)
	assert.Equal(t, 10, stored.Objectives[0].MeritPerUnit)
	assert.Equal(t, 2, stored.Objectives[0].UnitLimit)
	assert.Equal(t, "Rescue", stored.ContentEn.Title)
	assert.Equal(t, "Save the pilot", stored.ContentEn.Description)
	assert.Equal(t, []string{"No aids"}, stored.ContentEn.Rules)
	require.Len(t, stored.ContentEn.Objectives, 1)
	assert.Equal(t, "obj-1", stored.ContentEn.Objectives[0].ObjectiveID)
	assert.Equal(t, "Capture", stored.ContentEn.Objectives[0].Name)
	assert.Equal(t, "Capture the pilot", stored.ContentEn.Objectives[0].Description)
	assert.Equal(t, "def-rescue", created.DefinitionCode)
	item, err := client.Client().GetItem(t.Context(), &dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "DEFINITION"},
			"SK": &types.AttributeValueMemberS{Value: "def-rescue"},
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

// TEST-INF-LANG-001: publishing a bilingual definition version persists the
// English snapshot on both the current and version records so listing
// serves English per item.
func TestLocalizedDefinition_PublishedVersionRoundTrips(t *testing.T) {
	// Arrange
	client, table, setupErr := localizedStoreFixture(t)
	require.NoError(t, setupErr)
	store := missiondynamo.NewDefinitionStore(client, table)
	request := localizedDefinitionRequest()
	definition, err := domain.NewMissionDefinition(request.DefinitionCode, request.Title, request.Description, request.Category, request.Mode, request.CompletionReward, request.Objectives, request.Rules)
	require.NoError(t, err)
	definition.ContentEn = request.ContentEn.Clone()
	require.NoError(t, definition.Publish("v1"))
	require.NoError(t, store.Save(definition))

	// Act
	served, err := application.NewListDefinitionsHandler(store).Handle(domain.LanguageEnglish)

	// Assert
	require.NoError(t, err)
	require.Len(t, served, 2)
	for _, item := range served {
		assert.Equal(t, domain.LanguageEnglish, item.ContentLanguage)
		assert.Equal(t, "Rescue", item.Title)
		assert.Equal(t, []string{"No aids"}, item.Rules)
		require.Len(t, item.Objectives, 1)
		assert.Equal(t, "obj-1", item.Objectives[0].ObjectiveID)
		assert.Equal(t, "Capture", item.Objectives[0].Name)
		assert.Equal(t, 10, item.Objectives[0].MeritPerUnit)
	}
	versioned, err := store.GetVersion("def-rescue", "v1")
	require.NoError(t, err)
	assert.Equal(t, "Rescue", versioned.ContentEn.Title)
}

// TEST-INF-LANG-001: a legacy definition persisted before locale
// attributes existed decodes to zero English content and falls back to
// Spanish with an honest label.
func TestLocalizedDefinition_LegacySpanishOnlyFallsBack(t *testing.T) {
	// Arrange
	client, table, setupErr := localizedStoreFixture(t)
	require.NoError(t, setupErr)
	store := missiondynamo.NewDefinitionStore(client, table)
	request := localizedDefinitionRequest()
	definition, err := domain.NewMissionDefinition(request.DefinitionCode, request.Title, request.Description, request.Category, request.Mode, request.CompletionReward, request.Objectives, request.Rules)
	require.NoError(t, err)
	require.NoError(t, store.Save(definition))

	// Act
	served, err := application.NewListDefinitionsHandler(store).Handle(domain.LanguageEnglish)

	// Assert
	require.NoError(t, err)
	require.Len(t, served, 1)
	assert.Equal(t, domain.LanguageSpanish, served[0].ContentLanguage)
	assert.Equal(t, "Rescate", served[0].Title)
	assert.Equal(t, []string{"No ayudas"}, served[0].Rules)
	require.Len(t, served[0].Objectives, 1)
	assert.Equal(t, 10, served[0].Objectives[0].MeritPerUnit)
	item, err := client.Client().GetItem(t.Context(), &dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "DEFINITION"},
			"SK": &types.AttributeValueMemberS{Value: "def-rescue"},
		},
	})
	require.NoError(t, err)
	_, exists := item.Item["ContentEn"]
	assert.False(t, exists)
}
