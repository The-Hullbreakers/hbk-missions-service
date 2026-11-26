package composition_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	missioncomposition "github.com/hbk/hbk-missions-service/internal/missions/composition"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// TEST-E2E-MB-003 (REG-0053-01): a solo player leaving their last-member
// contract sees the same definition reappear on the board and retakes it
// under a personal active limit of 1 — the abandoned attempt must free
// both the reservation and the active slot.
func TestSoloLeaveJourney_LastMemberRetakes(t *testing.T) {
	// Arrange
	client, table := chg0053Table(t, "e2e-solo-retake")

	seasonStore := missiondynamo.NewSeasonStore(client, table)
	require.NoError(t, seasonStore.Save(domain.NewSeason("season-1", domain.SeasonStateActive)))
	offer := domain.NewMissionOffer("offer-x", "BOUNTY_HUNTING", "season-1")
	offer.DefinitionCode = "def-x"
	offer.DefinitionVersion = "legacy"
	offerStore := missiondynamo.NewOfferStore(client, table)
	require.NoError(t, offerStore.Save(offer))
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	member, err := domain.NewCallerIdentity("player-1", nil, nil)
	require.NoError(t, err)
	require.NoError(t, contract.Start(member))
	contractStore := missiondynamo.NewContractStore(client, table)
	require.NoError(t, contractStore.Save(contract))
	policyStore := missiondynamo.NewPolicyStore(client, table)
	require.NoError(t, policyStore.SaveActiveLimit(domain.NewActiveLimitPolicy(1, nil)))
	seedPlayerActive(t, client, table, "player-1", "contract-1")
	seedPlayerReservation(t, client, table, "player-1", "def-x", "contract-1")

	server, err := missioncomposition.NewServer(missioncomposition.Config{
		TableName:   table,
		DynamoDB:    client,
		UserSession: tokenSession{},
	})
	require.NoError(t, err)
	ts := httptest.NewServer(server)
	defer ts.Close()

	// Act
	outcome, actErr := runRetakeJourney(ts.URL, token("player-1"), "contract-1", "offer-x")

	// Assert
	require.NoError(t, actErr)
	require.Equal(t, http.StatusOK, outcome.leaveStatus)
	require.Equal(t, http.StatusOK, outcome.boardStatus)
	require.Equal(t, []string{"offer-x"}, outcome.boardOfferIDs)
	require.Equal(t, http.StatusCreated, outcome.retakeStatus, "retake rejected: %s", outcome.retakeCode)
}
