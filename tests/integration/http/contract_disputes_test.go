package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

// A member raises a dispute via POST /disputes → 201.
func TestDisputeEndpoint_RaiseReturns201(t *testing.T) {
	// Arrange
	store := &disputeStore{contract: disputedContractFixture()}
	app := application.NewRaiseDisputeHandler(store, store)
	auth := missionhttp.NewAuthMiddleware(memberSession{})
	server := httptest.NewServer(auth.Authenticate(missionhttp.RaiseDisputeHandler(app)))
	defer server.Close()

	// Act
	resp := postJSON(t, server, "/api/v1/missions/contracts/contract-1/disputes",
		`{"objective_id":"obj-a","reason":"wrongly rejected"}`)
	defer resp.Body.Close()

	// Assert
	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

// Leave on a contract with an OPEN dispute → 422 DISPUTES_PENDING.
func TestDisputeEndpoint_LeaveBlockedByPendingDispute(t *testing.T) {
	// Arrange — contract already carries an open dispute
	contract := disputedContractFixture()
	member, _ := domain.NewCallerIdentity("player-1", nil, nil)
	_, err := contract.RaiseDispute(member, "obj-a", "contested")
	require.NoError(t, err)
	store := &disputeStore{contract: contract}
	app := application.NewLeaveContractHandler(store, store)
	auth := missionhttp.NewAuthMiddleware(memberSession{})
	server := httptest.NewServer(auth.Authenticate(missionhttp.LeaveContractHandler(app)))
	defer server.Close()

	// Act
	resp := postJSON(t, server, "/api/v1/missions/contracts/contract-1/leave", `{}`)
	defer resp.Body.Close()

	// Assert
	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	var body map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&body)
	require.Equal(t, "DISPUTES_PENDING", body["code"])
}

// A moderator resolves a dispute → 200; re-resolving → 422 DISPUTE_STATE_INVALID.
func TestDisputeEndpoint_ResolveAndDoubleResolve(t *testing.T) {
	// Arrange
	contract := disputedContractFixture()
	member, _ := domain.NewCallerIdentity("player-1", nil, nil)
	dispute, err := contract.RaiseDispute(member, "obj-a", "contested")
	require.NoError(t, err)
	store := &disputeStore{contract: contract}
	app := application.NewResolveDisputeHandler(store, store)
	auth := missionhttp.NewAuthMiddleware(moderatorSession{})
	server := httptest.NewServer(auth.Authenticate(missionhttp.ResolveDisputeHandler(app)))
	defer server.Close()

	// Act — first resolve
	resp := postJSON(t, server, "/api/v1/missions/contracts/contract-1/disputes/"+dispute.DisputeID+"/resolve", `{}`)
	resp.Body.Close()

	// Assert
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Act — second resolve on the stored contract
	resp2 := postJSON(t, server, "/api/v1/missions/contracts/contract-1/disputes/"+dispute.DisputeID+"/resolve", `{}`)
	defer resp2.Body.Close()

	// Assert
	require.Equal(t, http.StatusUnprocessableEntity, resp2.StatusCode)
	var body map[string]string
	_ = json.NewDecoder(resp2.Body).Decode(&body)
	require.Equal(t, "DISPUTE_STATE_INVALID", body["code"])
}
