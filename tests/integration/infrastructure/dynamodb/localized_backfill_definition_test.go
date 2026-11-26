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

func rawDefinitionItem(t *testing.T, client *missiondynamo.DynamoDBClient, table, key string) (map[string]types.AttributeValue, error) {
	item, err := client.Client().GetItem(t.Context(), &dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "DEFINITION"},
			"SK": &types.AttributeValueMemberS{Value: key},
		},
	})
	if err != nil {
		return nil, err
	}
	return item.Item, nil
}

// TEST-INF-LANG-002: apply enriches the historical Spanish-only published
// definition on BOTH the current and the version rows — English snapshot
// only, Spanish baseline/economics untouched.
func TestBackfillDefinitionApplyCurrentAndVersion(t *testing.T) {
	// Arrange
	client, table, setupErr := localizedStoreFixture(t)
	require.NoError(t, setupErr)
	store := missiondynamo.NewDefinitionStore(client, table)
	definition, err := domain.NewMissionDefinition("def-rescue", "Rescate", "Salvar al piloto",
		"BOUNTY_HUNTING", domain.MissionModeSolo, 50,
		[]domain.Objective{
			domain.NewObjectiveNamed("obj-1", "Capturar", "Capturar al piloto", "BOUNTY_HUNTING", 10, 0, 2),
		}, []string{"No ayudas"})
	require.NoError(t, err)
	definition.State = domain.DefinitionStatePublished
	definition.Version = "v1"
	require.NoError(t, store.Save(definition))
	currentBefore, err := rawDefinitionItem(t, client, table, "def-rescue")
	require.NoError(t, err)
	versionBefore, err := rawDefinitionItem(t, client, table, "def-rescue#v1")
	require.NoError(t, err)
	reviewed := backfillReviewed()

	// Act
	report, err := missiondynamo.NewTranslationBackfiller(client, table).Rehearse(t.Context(), reviewed, true)

	// Assert
	require.NoError(t, err)
	assert.Contains(t, report.Updated, missiondynamo.BackfillRecord{Partition: "DEFINITION", ID: "def-rescue#v1"})
	assert.Contains(t, report.Updated, missiondynamo.BackfillRecord{Partition: "DEFINITION", ID: "def-rescue"})
	versioned, err := store.GetVersion("def-rescue", "v1")
	require.NoError(t, err)
	served := application.LocalizeDefinition(versioned, domain.LanguageEnglish)
	assert.Equal(t, domain.LanguageEnglish, served.ContentLanguage)
	assert.Equal(t, "Rescue", served.Title)
	assert.Equal(t, "Save the pilot", served.Description)
	require.Len(t, versioned.ContentEn.Objectives, 1)
	assert.Equal(t, "Capture", versioned.ContentEn.Objectives[0].Name)
	assert.Equal(t, domain.DefinitionStatePublished, versioned.State)
	assert.Equal(t, "v1", versioned.Version)
	assert.Equal(t, "Rescate", versioned.Title)
	assert.Equal(t, 50, versioned.CompletionReward)
	currentAfter, err := rawDefinitionItem(t, client, table, "def-rescue")
	require.NoError(t, err)
	versionAfter, err := rawDefinitionItem(t, client, table, "def-rescue#v1")
	require.NoError(t, err)
	assert.Equal(t, currentBefore, itemWithoutContentEn(currentAfter))
	assert.Equal(t, versionBefore, itemWithoutContentEn(versionAfter))
}
