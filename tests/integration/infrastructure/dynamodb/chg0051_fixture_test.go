package dynamodb_test

import (
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
	"github.com/stretchr/testify/require"
)

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

// seedActiveSeason persists an unlocked ACTIVE season so credit- and
// contract-bearing writes pass the adjudication fence (CHG-0051).
func seedActiveSeason(t *testing.T, client *missiondynamo.DynamoDBClient, table, seasonID string) {
	t.Helper()
	store := missiondynamo.NewSeasonStore(client, table)
	require.NoError(t, store.Create(domain.NewSeason(seasonID, domain.SeasonStateActive)))
}
