//go:build integration

package dynamodb_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func TestOverallStandings_DynamoDBReadsCreditsAndCompletedContracts(t *testing.T) {
	// Arrange
	client, table, setupErr := localizedStoreFixture(t)
	require.NoError(t, setupErr)
	season := missiondynamo.NewSeasonStore(client, table)
	require.NoError(t, season.Save(domain.NewSeason("season-1", domain.SeasonStateActive)))
	ledger := missiondynamo.NewMeritLedgerStore(client, table)
	for _, credit := range []domain.MeritCredit{
		{SeasonID: "season-1", Track: domain.TrackIndividual, Category: "PVP_AIR", RecipientID: "mira", Amount: 120, ContractID: "c1", ObjectiveID: "o1", ConfirmationID: "confirm-1"},
		{SeasonID: "season-1", Track: domain.TrackIndividual, Category: "RACING", RecipientID: "mira", Amount: 80, ContractID: "c1", ObjectiveID: "o2", ConfirmationID: "confirm-2"},
		{SeasonID: "season-1", Track: domain.TrackIndividual, Category: "PVP_AIR", RecipientID: "kai", Amount: 190, ContractID: "c2", ObjectiveID: "o1", ConfirmationID: "confirm-3"},
		{SeasonID: "season-1", Track: domain.TrackTeam, Category: "PVP_AIR", RecipientID: "team-1", Amount: 600, ContractID: "c3", ObjectiveID: "o1", ConfirmationID: "confirm-4"},
	} {
		require.NoError(t, ledger.Save(credit))
	}
	for id, state := range map[string]domain.ContractState{"c1": domain.ContractCompleted, "c2": domain.ContractInProgress} {
		_, saveErr := client.Client().PutItem(t.Context(), &dynamodb.PutItemInput{TableName: aws.String(table), Item: map[string]types.AttributeValue{"PK": &types.AttributeValueMemberS{Value: "CONTRACT"}, "SK": &types.AttributeValueMemberS{Value: id}, "State": &types.AttributeValueMemberS{Value: string(state)}}})
		require.NoError(t, saveErr)
	}
	identity, identityErr := domain.NewCallerIdentity("kai", nil, nil)
	require.NoError(t, identityErr)
	handler := application.NewGetOverallStandingsHandler(season, missiondynamo.NewOverallCreditsReader(client, table))
	// Act
	got, err := handler.Handle(identity, "season-1", 1, "")
	// Assert
	require.NoError(t, err)
	require.Equal(t, 2, got.Participants)
	require.Len(t, got.Board.Entries, 1)
	require.Equal(t, "mira", got.Board.Entries[0].RecipientID)
	require.Equal(t, 200, got.Board.Entries[0].Total)
	require.Equal(t, 1, got.Board.Entries[0].CompletedMissions)
	require.Equal(t, 2, got.Board.Entries[0].UniqueObjectives)
	require.NotNil(t, got.MyStanding)
	require.Equal(t, "kai", got.MyStanding.RecipientID)
	require.Equal(t, 2, got.MyStanding.Rank)
	require.Equal(t, 0, got.MyStanding.CompletedMissions)
}

func malformedOverallCredit(sk string, item map[string]types.AttributeValue) map[string]types.AttributeValue {
	base := map[string]types.AttributeValue{
		"PK":          &types.AttributeValueMemberS{Value: "MERIT_CREDIT"},
		"SK":          &types.AttributeValueMemberS{Value: sk},
		"SeasonID":    &types.AttributeValueMemberS{Value: "season-1"},
		"Category":    &types.AttributeValueMemberS{Value: "PVP_AIR"},
		"Track":       &types.AttributeValueMemberS{Value: "INDIVIDUAL"},
		"RecipientID": &types.AttributeValueMemberS{Value: "mira"},
		"Amount":      &types.AttributeValueMemberN{Value: "50"},
		"ContractID":  &types.AttributeValueMemberS{Value: "c1"},
		"ObjectiveID": &types.AttributeValueMemberS{Value: "o1"},
	}
	for k, v := range item {
		if v == nil {
			delete(base, k)
			continue
		}
		base[k] = v
	}
	return base
}

func TestOverallStandings_DynamoDBRejectsMalformedCreditFacts(t *testing.T) {
	cases := []struct {
		name string
		item map[string]types.AttributeValue
	}{
		{"missing amount", map[string]types.AttributeValue{"Amount": nil}},
		{"non-numeric amount", map[string]types.AttributeValue{"Amount": &types.AttributeValueMemberS{Value: "abc"}}},
		{"missing recipient", map[string]types.AttributeValue{"RecipientID": nil}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			client, table, setupErr := localizedStoreFixture(t)
			require.NoError(t, setupErr)
			_, putErr := client.Client().PutItem(t.Context(), &dynamodb.PutItemInput{TableName: aws.String(table), Item: malformedOverallCredit("bad-"+tc.name, tc.item)})
			require.NoError(t, putErr)
			reader := missiondynamo.NewOverallCreditsReader(client, table)
			// Act
			_, _, err := reader.ListOverallCredits("season-1")
			// Assert
			require.Error(t, err)
		})
	}
}

func TestOverallStandings_DynamoDBPropagatesQueryFailure(t *testing.T) {
	// Arrange
	client, _, setupErr := localizedStoreFixture(t)
	require.NoError(t, setupErr)
	reader := missiondynamo.NewOverallCreditsReader(client, "overall-missing-table")
	// Act
	_, _, err := reader.ListOverallCredits("season-1")
	// Assert
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read overall credits")
}
