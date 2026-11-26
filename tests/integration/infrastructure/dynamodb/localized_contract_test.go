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

// TEST-INF-LANG-002: a contract accepted on a bilingual offer round-trips
// its English snapshot as a top-level DynamoDB M attribute — Offer text
// and Terms objectives both render English.
func TestLocalizedContract_BilingualSnapshotRoundTrips(t *testing.T) {
	// Arrange
	client, table, setupErr := localizedStoreFixture(t)
	require.NoError(t, setupErr)
	store := missiondynamo.NewContractStore(client, table)
	contract, contractErr := localizedBilingualContract()
	require.NoError(t, contractErr)
	member, memberErr := domain.NewCallerIdentity("player-1", nil, nil)
	require.NoError(t, memberErr)
	require.NoError(t, store.Save(contract))
	contract.Offer.ContentEn.Title = "Mutated"
	contract.Offer.ContentEn.Rules[0] = "Mutated rule"
	contract.Offer.ContentEn.Objectives[0].Name = "Mutated objective"

	// Act
	served, err := application.NewGetContractHandler(store).Handle(member, "contract-rescue", domain.LanguageEnglish)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, served)
	assert.Equal(t, domain.LanguageEnglish, served.ContentLanguage)
	assert.Equal(t, "Rescue", served.Offer.Title)
	assert.Equal(t, "Save the pilot", served.Offer.Description)
	assert.Equal(t, []string{"No aids"}, served.Offer.Rules)
	termsObjectives := served.Terms.Objectives()
	require.Len(t, termsObjectives, 1)
	assert.Equal(t, "obj-1", termsObjectives[0].ObjectiveID)
	assert.Equal(t, "Capture", termsObjectives[0].Name)
	assert.Equal(t, "Capture the pilot", termsObjectives[0].Description)
	assert.Equal(t, 10, termsObjectives[0].MeritPerUnit)
	assert.Equal(t, 2, termsObjectives[0].UnitLimit)
	assert.Equal(t, 50, served.Offer.CompletionReward)
	assert.Equal(t, "v1", served.Offer.DefinitionVersion)
	assert.Equal(t, []string{"player-1"}, served.Roster())
	item, err := client.Client().GetItem(t.Context(), &dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "CONTRACT"},
			"SK": &types.AttributeValueMemberS{Value: "contract-rescue"},
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
	firstObjective, isObjectiveMap := objectives.Value[0].(*types.AttributeValueMemberM)
	require.True(t, isObjectiveMap)
	_, isIDString := firstObjective.Value["ObjectiveID"].(*types.AttributeValueMemberS)
	assert.True(t, isIDString)
}

// TEST-INF-LANG-002: a contract persisted before locale attributes existed
// decodes to zero English content and falls back to Spanish.
func TestLocalizedContract_LegacySpanishOnlyFallsBack(t *testing.T) {
	// Arrange
	client, table, setupErr := localizedStoreFixture(t)
	require.NoError(t, setupErr)
	store := missiondynamo.NewContractStore(client, table)
	contract, contractErr := localizedLegacyContract()
	require.NoError(t, contractErr)
	member, memberErr := domain.NewCallerIdentity("player-1", nil, nil)
	require.NoError(t, memberErr)
	require.NoError(t, store.Save(contract))

	// Act
	served, err := application.NewGetContractHandler(store).Handle(member, "contract-old", domain.LanguageEnglish)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, served)
	assert.Equal(t, domain.LanguageSpanish, served.ContentLanguage)
	assert.Equal(t, "Rescate antiguo", served.Offer.Title)
	assert.Equal(t, "Salvar al piloto", served.Offer.Description)
	assert.Equal(t, []string{"No ayudas"}, served.Offer.Rules)
	termsObjectives := served.Terms.Objectives()
	require.Len(t, termsObjectives, 1)
	assert.Equal(t, "Capturar", termsObjectives[0].Name)
	assert.Equal(t, 10, termsObjectives[0].MeritPerUnit)
	item, err := client.Client().GetItem(t.Context(), &dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "CONTRACT"},
			"SK": &types.AttributeValueMemberS{Value: "contract-old"},
		},
	})
	require.NoError(t, err)
	_, exists := item.Item["ContentEn"]
	assert.False(t, exists)
}
